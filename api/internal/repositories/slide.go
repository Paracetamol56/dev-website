package repositories

import (
	"context"
	"dev/internal/db"
	"dev/internal/models"
	"dev/internal/utils"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type SlideRepository struct {
	collection *mongo.Collection
}

func NewSlideDeckRepository() *SlideRepository {
	db := db.GetDB()
	collection := db.Collection("slides")

	return &SlideRepository{collection: collection}
}

func (r *SlideRepository) Create(c context.Context, slide *models.SlideDeck) (*mongo.InsertOneResult, error) {
	slide.Id = primitive.NewObjectID()
	slide.Revision = 0
	slide.CreatedAt = time.Now()
	slide.UpdatedAt = time.Now()
	slide.ArchivedAt = nil

	if slide.Status == "" {
		slide.Status = models.StatusDraft
	}

	return r.collection.InsertOne(
		c,
		slide,
	)
}

func (r *SlideRepository) GetById(c context.Context, id primitive.ObjectID) (*models.SlideDeck, error) {
	var slide models.SlideDeck

	err := r.collection.FindOne(
		c,
		bson.M{
			"_id":         id,
			"archived_at": bson.M{"$exists": false},
			"status":      models.StatusPublished,
		}).Decode(&slide)
	if err != nil {
		return nil, err
	}
	return &slide, nil
}

func (r *SlideRepository) GetByIdForUser(c context.Context, id primitive.ObjectID, userId primitive.ObjectID) (*models.SlideDeck, error) {
	var slide models.SlideDeck

	err := r.collection.FindOne(
		c,
		bson.M{"_id": id, "user_id": userId},
	).Decode(&slide)
	if err != nil {
		return nil, err
	}
	return &slide, nil
}

func (r *SlideRepository) GetByUser(c context.Context, userId primitive.ObjectID, archived bool, pagination *utils.Pagination) (*utils.PaginatedFacetResult[models.SlideDeck], error) {
	filter := bson.M{
		"user_id": userId,
	}
	if archived {
		filter["archived_at"] = bson.M{"$exists": true}
	} else {
		filter["archived_at"] = bson.M{"$exists": false}
	}

	cursor, err := r.collection.Aggregate(
		c,
		bson.A{
			bson.M{"$match": filter},
			pagination.GetFacetStage(),
		},
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(c)
	var result utils.PaginatedFacetResult[models.SlideDeck]
	if !cursor.Next(c) {
		return nil, mongo.ErrNoDocuments
	}
	if err := cursor.Decode(&result); err != nil {
		return nil, err
	}
	fmt.Println(result)
	return &result, nil
}

func (r *SlideRepository) Update(c context.Context, userId primitive.ObjectID, slide *models.SlideDeck) (*mongo.UpdateResult, error) {
	slide.UpdatedAt = time.Now()
	filter := bson.M{
		"_id":     slide.Id,
		"user_id": userId,
	}

	result, err := r.collection.UpdateOne(
		c,
		filter,
		bson.M{
			"$set": slide,
		},
	)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *SlideRepository) UpdateContent(c context.Context, userId primitive.ObjectID, id primitive.ObjectID, content string) (*mongo.UpdateResult, error) {
	filter := bson.M{
		"_id":     id,
		"user_id": userId,
	}

	update := bson.M{
		"$set": bson.M{
			"content":   content,
			"updatedAt": time.Now(),
		},
		"$inc": bson.M{
			"revision": 1,
		},
	}

	result, err := r.collection.UpdateOne(
		c,
		filter,
		update,
	)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *SlideRepository) AddOperation(c context.Context, userId primitive.ObjectID, id primitive.ObjectID, operation *models.SlideOperation) (*mongo.UpdateResult, error) {
	filter := bson.M{
		"_id":     id,
		"user_id": userId,
	}

	update := bson.M{
		"$push": bson.M{
			"operations": bson.M{
				"$each":     []models.SlideOperation{*operation},
				"$position": 0,
				"$slice":    -100,
			},
			"$inc": bson.M{
				"revision": 1,
			},
			"$set": bson.M{
				"updatedAt": time.Now(),
			},
		},
	}

	result, err := r.collection.UpdateOne(
		c,
		filter,
		update,
	)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *SlideRepository) Archive(c context.Context, userId primitive.ObjectID, id primitive.ObjectID) (*mongo.UpdateResult, error) {
	filter := bson.M{
		"_id":     id,
		"user_id": userId,
	}

	update := bson.M{
		"$set": bson.M{
			"updated_at":  time.Now(),
			"archived_at": time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(
		c,
		filter,
		update,
	)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *SlideRepository) Unarchive(c context.Context, userId primitive.ObjectID, id primitive.ObjectID) (*mongo.UpdateResult, error) {
	filter := bson.M{
		"_id":     id,
		"user_id": userId,
	}

	update := bson.M{
		"$set": bson.M{
			"updated_at":  time.Now(),
			"archived_at": nil,
		},
	}

	result, err := r.collection.UpdateOne(
		c,
		filter,
		update,
	)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *SlideRepository) Delete(c context.Context, userId primitive.ObjectID, id primitive.ObjectID) (*mongo.DeleteResult, error) {
	filter := bson.M{
		"_id":     id,
		"user_id": userId,
	}

	result, err := r.collection.DeleteOne(
		c,
		filter,
	)
	if err != nil {
		return nil, err
	}
	return result, nil
}
