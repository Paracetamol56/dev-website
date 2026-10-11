package wordcloud

import (
	"context"
	"dev/internal/db"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Also matches old codes, created before codes were uppercase only
var caseInsensitive = &options.Collation{Locale: "en", Strength: 2}

// GetWordCloudByCode retrieves a word cloud by its code.
// It takes a gin.Context and a code string as parameters.
// It returns a pointer to a WordCloud object and an error.
func GetWordCloudByCode(c context.Context, code string) (*WordCloud, error) {
	db := db.GetDB()
	collection := db.Collection("word_cloud_sessions")
	var wordCloud WordCloud
	filter := bson.M{"code": code, "closedAt": bson.M{"$exists": false}}
	if err := collection.FindOne(c, filter, options.FindOne().SetCollation(caseInsensitive)).Decode(&wordCloud); err != nil {
		return nil, err
	}
	return &wordCloud, nil
}

func IsCodeInUse(c context.Context, code string) (bool, error) {
	collection := db.GetDB().Collection("word_cloud_sessions")
	filter := bson.M{"code": code, "closedAt": bson.M{"$exists": false}}
	count, err := collection.CountDocuments(c, filter, options.Count().SetCollation(caseInsensitive).SetLimit(1))
	return count > 0, err
}

// GetWordCloudById retrieves a word cloud by its ID.
// It takes a gin.Context and an ID of type primitive.ObjectID as parameters.
// It returns a pointer to a WordCloud object and an error.
func GetWordCloudById(c context.Context, id primitive.ObjectID) (*WordCloud, error) {
	db := db.GetDB()
	collection := db.Collection("word_cloud_sessions")
	var wordCloud WordCloud
	if err := collection.FindOne(c, bson.M{"_id": id}).Decode(&wordCloud); err != nil {
		return nil, err
	}
	return &wordCloud, nil
}

// GetWordCloudByUser retrieves all word clouds associated with a user.
// It takes a gin.Context and a userID of type primitive.ObjectID as parameters.
// It returns a slice of pointers to WordCloud objects and an error.
func GetWordCloudByUser(c context.Context, userId primitive.ObjectID, status string) ([]*WordCloud, error) {
	db := db.GetDB()
	collection := db.Collection("word_cloud_sessions")
	var filter bson.M
	switch status {
	case "open":
		filter = bson.M{"user": userId, "closedAt": bson.M{"$exists": false}}
	case "closed":
		filter = bson.M{"user": userId, "closedAt": bson.M{"$exists": true}}
	default:
		return nil, fmt.Errorf("invalid status")
	}
	var wordClouds []*WordCloud
	cursor, err := collection.Find(c, filter)
	if err != nil {
		return nil, err
	}
	if err = cursor.All(c, &wordClouds); err != nil {
		return nil, err
	}
	return wordClouds, nil
}

// CreateWordCloud creates a new word cloud.
// It takes a gin.Context and a pointer to a WordCloud object as parameters.
// It returns a pointer to a mongo.InsertOneResult object and an error.
func CreateWordCloud(c context.Context, wordCloud *WordCloud) (*mongo.InsertOneResult, error) {
	db := db.GetDB()
	collection := db.Collection("word_cloud_sessions")
	result, err := collection.InsertOne(c, wordCloud)
	return result, err
}

// AddWordToWordCloud returns ErrWordNotAdded when the session is closed or the participant
// already sent this word; both are checked within the update itself.
func AddWordToWordCloud(c context.Context, sessionId primitive.ObjectID, word *Word) error {
	collection := db.GetDB().Collection("word_cloud_sessions")
	filter := bson.M{
		"_id":      sessionId,
		"closedAt": bson.M{"$exists": false},
		"words":    bson.M{"$not": bson.M{"$elemMatch": bson.M{"text": word.Text, "uuid": word.UUID}}},
	}
	result, err := collection.UpdateOne(c, filter, bson.M{"$push": bson.M{"words": word}})
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrWordNotAdded
	}
	return nil
}

func UpdateWordCloud(c context.Context, id primitive.ObjectID, set bson.M, unset bson.M) error {
	update := bson.M{"$set": bson.M{"updatedAt": primitive.NewDateTimeFromTime(time.Now())}}
	for key, value := range set {
		update["$set"].(bson.M)[key] = value
	}
	if len(unset) > 0 {
		update["$unset"] = unset
	}
	_, err := db.GetDB().Collection("word_cloud_sessions").UpdateOne(c, bson.M{"_id": id}, update)
	return err
}

func DeleteWordCloud(c context.Context, id primitive.ObjectID) error {
	_, err := db.GetDB().Collection("word_cloud_sessions").DeleteOne(c, bson.M{"_id": id})
	return err
}

// EnsureWordIds gives an id to words stored before words had one. The update only applies
// if no word was added since the session was read.
func EnsureWordIds(c context.Context, wordCloud *WordCloud) error {
	missing := false
	for i := range wordCloud.Words {
		if wordCloud.Words[i].Id.IsZero() {
			wordCloud.Words[i].Id = primitive.NewObjectID()
			missing = true
		}
	}
	if !missing {
		return nil
	}
	filter := bson.M{"_id": wordCloud.Id, "words": bson.M{"$size": len(wordCloud.Words)}}
	_, err := db.GetDB().Collection("word_cloud_sessions").UpdateOne(c, filter, bson.M{"$set": bson.M{"words": wordCloud.Words}})
	return err
}

func SetWordHidden(c context.Context, sessionId primitive.ObjectID, wordId primitive.ObjectID, hidden bool) (bool, error) {
	result, err := db.GetDB().Collection("word_cloud_sessions").UpdateOne(c,
		bson.M{"_id": sessionId, "words._id": wordId},
		bson.M{"$set": bson.M{"words.$.hidden": hidden}})
	if err != nil {
		return false, err
	}
	return result.MatchedCount > 0, nil
}

// MergeWords rewrites every word whose text is in from to into, keeping the submitted text.
func MergeWords(c context.Context, sessionId primitive.ObjectID, from []string, into string) error {
	merged := bson.M{"$map": bson.M{
		"input": "$words",
		"in": bson.M{"$cond": bson.A{
			bson.M{"$in": bson.A{"$$this.text", from}},
			bson.M{"$mergeObjects": bson.A{"$$this", bson.M{
				"text":         into,
				"originalText": bson.M{"$ifNull": bson.A{"$$this.originalText", "$$this.text"}},
			}}},
			"$$this",
		}},
	}}
	_, err := db.GetDB().Collection("word_cloud_sessions").UpdateOne(c,
		bson.M{"_id": sessionId},
		mongo.Pipeline{{{Key: "$set", Value: bson.M{"words": merged}}}})
	return err
}
