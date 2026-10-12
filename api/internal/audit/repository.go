package audit

import (
	"context"
	"dev/internal/db"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const collectionName = "audit"

func EnsureIndexes(ctx context.Context) error {
	_, err := db.GetDB().Collection(collectionName).Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "createdAt", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(int32(Retention.Seconds())),
		},
		{Keys: bson.D{{Key: "userId", Value: 1}, {Key: "createdAt", Value: -1}}},
		{Keys: bson.D{{Key: "action", Value: 1}, {Key: "createdAt", Value: -1}}},
	})
	return err
}

func CreateEntry(ctx context.Context, entry *Entry) error {
	result, err := db.GetDB().Collection(collectionName).InsertOne(ctx, entry)
	if err != nil {
		return err
	}
	entry.Id = result.InsertedID.(primitive.ObjectID)
	return nil
}

// GetEntriesByUserId returns the entries of a user, newest first
func GetEntriesByUserId(ctx context.Context, userId primitive.ObjectID) ([]Entry, error) {
	cursor, err := db.GetDB().Collection(collectionName).Find(
		ctx,
		bson.M{"userId": userId},
		options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}),
	)
	if err != nil {
		return nil, err
	}
	entries := []Entry{}
	err = cursor.All(ctx, &entries)
	return entries, err
}
