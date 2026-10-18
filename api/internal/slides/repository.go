package slides

import (
	"context"
	"dev/internal/db"
	"regexp"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func collection() *mongo.Collection {
	return db.GetDB().Collection("slides")
}

func EnsureIndexes(ctx context.Context) error {
	_, err := collection().Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "user", Value: 1}, {Key: "updatedAt", Value: -1}},
	})
	return err
}

func CreateDeck(ctx context.Context, deck *Deck) (*mongo.InsertOneResult, error) {
	return collection().InsertOne(ctx, deck)
}

func GetDeckByUserIdAndId(ctx context.Context, userId primitive.ObjectID, id primitive.ObjectID) (*Deck, error) {
	var deck Deck
	if err := collection().FindOne(ctx, bson.M{"_id": id, "user": userId}).Decode(&deck); err != nil {
		return nil, err
	}
	return &deck, nil
}

// GetDecksByUserId lists decks without their content, last updated first
func GetDecksByUserId(ctx context.Context, userId primitive.ObjectID, search string, page int, limit int) ([]DeckSummary, int64, error) {
	filter := bson.M{"user": userId}
	if search != "" {
		filter["title"] = primitive.Regex{Pattern: regexp.QuoteMeta(search), Options: "i"}
	}

	total, err := collection().CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	cursor, err := collection().Find(ctx, filter, options.Find().
		SetProjection(bson.M{"content": 0}).
		SetSort(bson.D{{Key: "updatedAt", Value: -1}, {Key: "_id", Value: -1}}).
		SetSkip(int64((page-1)*limit)).
		SetLimit(int64(limit)))
	if err != nil {
		return nil, 0, err
	}
	decks := []DeckSummary{}
	if err := cursor.All(ctx, &decks); err != nil {
		return nil, 0, err
	}
	return decks, total, nil
}

func GetDeckById(ctx context.Context, id primitive.ObjectID) (*Deck, error) {
	var deck Deck
	if err := collection().FindOne(ctx, bson.M{"_id": id}).Decode(&deck); err != nil {
		return nil, err
	}
	return &deck, nil
}

func SetDeckFields(ctx context.Context, id primitive.ObjectID, set bson.M) error {
	_, err := collection().UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": set})
	return err
}

func DeleteDeck(ctx context.Context, userId primitive.ObjectID, id primitive.ObjectID) (*mongo.DeleteResult, error) {
	return collection().DeleteOne(ctx, bson.M{"_id": id, "user": userId})
}

func GetDecksWithoutOutline(ctx context.Context) ([]Deck, error) {
	cursor, err := collection().Find(ctx, bson.M{"outline": bson.M{"$exists": false}})
	if err != nil {
		return nil, err
	}
	decks := []Deck{}
	err = cursor.All(ctx, &decks)
	return decks, err
}

func SetDeckOutline(ctx context.Context, id primitive.ObjectID, outline []OutlineEntry) error {
	_, err := collection().UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"outline": outline}, "$unset": bson.M{"slides": ""}})
	return err
}

func SetMissingSettings(ctx context.Context, settings Settings) (int64, error) {
	result, err := collection().UpdateMany(ctx, bson.M{"settings": bson.M{"$exists": false}}, bson.M{"$set": bson.M{"settings": settings}})
	if err != nil {
		return 0, err
	}
	return result.ModifiedCount, nil
}
