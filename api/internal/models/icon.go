package models

import (
	"context"
	"dev/internal/db"
	"regexp"
	"sort"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	IconSourceLucide      = "lucide"
	IconSourceSimpleIcons = "simpleicons"
)

type Icon struct {
	Id     string   `json:"id" bson:"_id" example:"lucide:a-arrow-down"`
	Name   string   `json:"name" bson:"name" example:"a-arrow-down"`
	Title  string   `json:"title" bson:"title" example:"A Arrow Down"`
	Source string   `json:"source" bson:"source" example:"lucide"`
	Tags   []string `json:"tags" bson:"tags"`
	Hex    string   `json:"hex,omitempty" bson:"hex,omitempty" example:"ECD53F"`
}

type IconFilter struct {
	Source string `form:"source,omitempty" binding:"omitempty,oneof=lucide simpleicons"`
}

func GetAllIcons(ctx context.Context, filter *IconFilter) ([]Icon, error) {
	collection := db.GetDB().Collection("icons")

	query := bson.M{}
	if filter.Source != "" {
		query["source"] = filter.Source
	}

	cursor, err := collection.Find(ctx, query, options.Find().SetSort(bson.D{{Key: "source", Value: 1}, {Key: "name", Value: 1}}))
	if err != nil {
		return nil, err
	}

	value := []Icon{}
	if err = cursor.All(ctx, &value); err != nil {
		return nil, err
	}

	return value, nil
}

func CountIcons(ctx context.Context) (int64, error) {
	return db.GetDB().Collection("icons").CountDocuments(ctx, bson.M{})
}

// ReplaceIcons makes the stored icons of a source match the given set: icons are upserted
// first and stale ones removed after, so the collection is never empty during a refresh.
func ReplaceIcons(ctx context.Context, source string, icons []Icon) error {
	collection := db.GetDB().Collection("icons")

	ids := make([]string, 0, len(icons))
	writes := make([]mongo.WriteModel, 0, len(icons))
	for _, icon := range icons {
		ids = append(ids, icon.Id)
		writes = append(writes, mongo.NewReplaceOneModel().
			SetFilter(bson.M{"_id": icon.Id}).
			SetReplacement(icon).
			SetUpsert(true))
	}

	if len(writes) > 0 {
		if _, err := collection.BulkWrite(ctx, writes, options.BulkWrite().SetOrdered(false)); err != nil {
			return err
		}
	}

	_, err := collection.DeleteMany(ctx, bson.M{"source": source, "_id": bson.M{"$nin": ids}})
	return err
}

type IconSearch struct {
	Query   string   `form:"q" binding:"max=100"`
	Sources []string `form:"source" binding:"dive,oneof=lucide simpleicons"`
	Tag     string   `form:"tag,omitempty" binding:"max=100"`
	Limit   int      `form:"limit,default=50" binding:"min=1,max=500"`
	Offset  int      `form:"offset,default=0" binding:"min=0"`
}

type IconSearchResult struct {
	Total int    `json:"total" example:"42"`
	Items []Icon `json:"items"`
}

// SearchIcons returns icons matching every word of the query (in the name, title or tags),
// ranked so that exact and prefix name matches come before title and tag matches.
func SearchIcons(ctx context.Context, search *IconSearch) (*IconSearchResult, error) {
	collection := db.GetDB().Collection("icons")

	terms := strings.Fields(strings.ToLower(search.Query))
	conditions := bson.A{}
	if len(search.Sources) > 0 {
		conditions = append(conditions, bson.M{"source": bson.M{"$in": search.Sources}})
	}
	if search.Tag != "" {
		conditions = append(conditions, bson.M{"tags": bson.M{"$regex": "^" + regexp.QuoteMeta(search.Tag) + "$", "$options": "i"}})
	}
	for _, term := range terms {
		pattern := bson.M{"$regex": regexp.QuoteMeta(term), "$options": "i"}
		conditions = append(conditions, bson.M{"$or": bson.A{
			bson.M{"name": pattern},
			bson.M{"title": pattern},
			bson.M{"tags": pattern},
		}})
	}
	query := bson.M{}
	if len(conditions) > 0 {
		query["$and"] = conditions
	}

	cursor, err := collection.Find(ctx, query)
	if err != nil {
		return nil, err
	}
	icons := []Icon{}
	if err = cursor.All(ctx, &icons); err != nil {
		return nil, err
	}

	scores := make(map[string]int, len(icons))
	for _, icon := range icons {
		scores[icon.Id] = iconScore(icon, terms)
	}
	sort.SliceStable(icons, func(i, j int) bool {
		a, b := icons[i], icons[j]
		if scores[a.Id] != scores[b.Id] {
			return scores[a.Id] > scores[b.Id]
		}
		if a.Name != b.Name {
			return a.Name < b.Name
		}
		return a.Source < b.Source
	})

	start := min(search.Offset, len(icons))
	end := min(start+search.Limit, len(icons))
	return &IconSearchResult{Total: len(icons), Items: icons[start:end]}, nil
}

func iconScore(icon Icon, terms []string) int {
	name, title := strings.ToLower(icon.Name), strings.ToLower(icon.Title)
	if len(terms) > 1 && (name == strings.Join(terms, "-") || title == strings.Join(terms, " ")) {
		return 1000
	}
	score := 0
	for _, term := range terms {
		switch {
		case name == term || title == term:
			score += 100
		case strings.HasPrefix(name, term) || strings.HasPrefix(title, term):
			score += 50
		case strings.Contains(name, term) || strings.Contains(title, term):
			score += 20
		}
		for _, tag := range icon.Tags {
			tag = strings.ToLower(tag)
			if tag == term {
				score += 10
				break
			}
			if strings.Contains(tag, term) {
				score += 5
				break
			}
		}
	}
	return score
}
