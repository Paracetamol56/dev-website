package chip

import (
	"context"
	"dev/internal/db"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetAllMicroprocessors(c context.Context, filter *MicroprocessorFilter) ([]Microprocessor, error) {
	db := db.GetDB()
	collection := db.Collection("microprocessors")

	query := bson.M{}
	if filter.Type != "" {
		query["type"] = filter.Type
	}
	if filter.Vendor != "" {
		query["vendor"] = filter.Vendor
	}

	cursor, err := collection.Find(c, query, nil)
	if err != nil {
		return nil, err
	}

	value := []Microprocessor{}
	if err = cursor.All(c, &value); err != nil {
		return nil, err
	}

	return value, nil
}

func GetMicroprocessorById(c context.Context, id string) (*Microprocessor, error) {
	collection := db.GetDB().Collection("microprocessors")

	ids := bson.A{id}
	if objectId, err := primitive.ObjectIDFromHex(id); err == nil {
		ids = append(ids, objectId)
	}

	var value Microprocessor
	if err := collection.FindOne(c, bson.M{"_id": bson.M{"$in": ids}}).Decode(&value); err != nil {
		return nil, err
	}

	return &value, nil
}

func CountMicroprocessors(ctx context.Context, source string) (int64, error) {
	query := bson.M{}
	if source != "" {
		query["source"] = source
	}
	return db.GetDB().Collection("microprocessors").CountDocuments(ctx, query)
}

func ReplaceMicroprocessors(ctx context.Context, source string, chips []Microprocessor) error {
	collection := db.GetDB().Collection("microprocessors")

	ids := make([]string, 0, len(chips))
	writes := make([]mongo.WriteModel, 0, len(chips))
	for _, chip := range chips {
		ids = append(ids, chip.Id)
		writes = append(writes, mongo.NewReplaceOneModel().
			SetFilter(bson.M{"_id": chip.Id}).
			SetReplacement(chip).
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
