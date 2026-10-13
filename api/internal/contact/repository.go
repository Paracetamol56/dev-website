package contact

import (
	"context"
	"dev/internal/db"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func CreateContact(c context.Context, contact Contact) (*primitive.ObjectID, error) {
	db := db.GetDB()
	collection := db.Collection("contacts")
	result, err := collection.InsertOne(c, contact)
	if err != nil {
		return nil, err
	}
	insertedID := result.InsertedID.(primitive.ObjectID)
	return &insertedID, nil
}
