package models

import (
	"dev/internal/db"
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ErrWordNotAdded = errors.New("word not added")

// Also matches old codes, created before codes were uppercase only
var caseInsensitive = &options.Collation{Locale: "en", Strength: 2}

// Word represents a word in a word cloud.
type Word struct {
	Text      string             `json:"text" bson:"text"`
	UUID      uuid.UUID          `json:"uuid" bson:"uuid"`
	UserAgent string             `json:"userAgent" bson:"userAgent"`
	CreatedAt primitive.DateTime `json:"createdAt" bson:"createdAt"`
}

// WordCloud represents a word cloud object.
type WordCloud struct {
	Id          primitive.ObjectID  `json:"id" bson:"_id,omitempty"`
	UserId      primitive.ObjectID  `json:"user" bson:"user"`
	Name        string              `json:"name" bson:"name"`
	Description string              `json:"description" bson:"description"`
	Code        string              `json:"code" bson:"code"`
	Words       []Word              `json:"words" bson:"words"`
	CreatedAt   primitive.DateTime  `json:"createdAt" bson:"createdAt"`
	UpdatedAt   primitive.DateTime  `json:"updatedAt" bson:"updatedAt"`
	ClosedAt    *primitive.DateTime `json:"closedAt,omitempty" bson:"closedAt,omitempty"`
}

// GetWordCloudByCode retrieves a word cloud by its code.
// It takes a gin.Context and a code string as parameters.
// It returns a pointer to a WordCloud object and an error.
func GetWordCloudByCode(c *gin.Context, code string) (*WordCloud, error) {
	db := db.GetDB()
	collection := db.Collection("word_cloud_sessions")
	var wordCloud WordCloud
	filter := bson.M{"code": code, "closedAt": bson.M{"$exists": false}}
	if err := collection.FindOne(c, filter, options.FindOne().SetCollation(caseInsensitive)).Decode(&wordCloud); err != nil {
		return nil, err
	}
	return &wordCloud, nil
}

func IsCodeInUse(c *gin.Context, code string) (bool, error) {
	collection := db.GetDB().Collection("word_cloud_sessions")
	filter := bson.M{"code": code, "closedAt": bson.M{"$exists": false}}
	count, err := collection.CountDocuments(c, filter, options.Count().SetCollation(caseInsensitive).SetLimit(1))
	return count > 0, err
}

// GetWordCloudById retrieves a word cloud by its ID.
// It takes a gin.Context and an ID of type primitive.ObjectID as parameters.
// It returns a pointer to a WordCloud object and an error.
func GetWordCloudById(c *gin.Context, id primitive.ObjectID) (*WordCloud, error) {
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
func GetWordCloudByUser(c *gin.Context, userId primitive.ObjectID, status string) ([]*WordCloud, error) {
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
func CreateWordCloud(c *gin.Context, wordCloud *WordCloud) (*mongo.InsertOneResult, error) {
	db := db.GetDB()
	collection := db.Collection("word_cloud_sessions")
	result, err := collection.InsertOne(c, wordCloud)
	return result, err
}

// AddWordToWordCloud returns ErrWordNotAdded when the session is closed or the participant
// already sent this word; both are checked within the update itself.
func AddWordToWordCloud(c *gin.Context, sessionId primitive.ObjectID, word *Word) error {
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

func UpdateWordCloud(c *gin.Context, id primitive.ObjectID, set bson.M, unset bson.M) error {
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

func DeleteWordCloud(c *gin.Context, id primitive.ObjectID) error {
	_, err := db.GetDB().Collection("word_cloud_sessions").DeleteOne(c, bson.M{"_id": id})
	return err
}
