package models

import (
	"github.com/Paracetamol56/dev-website/api/db"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Word represents a word in a word cloud.
type Word struct {
	Text      string             `json:"text" bson:"text"`
	IP        string             `json:"ip" bson:"ip"`
	UserAgent string             `json:"userAgent" bson:"userAgent"`
	CreatedAt primitive.DateTime `json:"createdAt" bson:"createdAt"`
}

// WordCloud represents a word cloud object.
type WordCloud struct {
	Id          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserId      primitive.ObjectID `json:"user" bson:"user"`
	Name        string             `json:"name" bson:"name"`
	Description string             `json:"description" bson:"description"`
	Code        string             `json:"code" bson:"code"`
	Open        bool               `json:"open" bson:"open"`
	Words       []Word             `json:"words" bson:"words"`
	CreatedAt   primitive.DateTime `json:"createdAt" bson:"createdAt"`
	UpdatedAt   primitive.DateTime `json:"updatedAt" bson:"updatedAt"`
	ClosedAt    primitive.DateTime `json:"closedAt,omitempty" bson:"closedAt,omitempty"`
}

// GetWordCloudByCode retrieves a word cloud by its code.
// It takes a gin.Context and a code string as parameters.
// It returns a pointer to a WordCloud object and an error.
func GetWordCloudByCode(c *gin.Context, code string) (*WordCloud, error) {
	db := db.GetDB()
	collection := db.Collection("word_cloud_sessions")
	var wordCloud WordCloud
	if err := collection.FindOne(c, bson.M{"code": code}).Decode(&wordCloud); err != nil {
		return nil, err
	}
	return &wordCloud, nil
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
func GetWordCloudByUser(c *gin.Context, userId primitive.ObjectID) ([]*WordCloud, error) {
	db := db.GetDB()
	collection := db.Collection("word_cloud_sessions")
	var wordClouds []*WordCloud
	cursor, err := collection.Find(c, bson.M{"user": userId})
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

// AddWordToWordCloud adds a word to a word cloud.
// It takes a gin.Context, a pointer to a WordCloud object, and a pointer to a Word object as parameters.
// It returns a pointer to a mongo.UpdateResult object and an error.
func AddWordToWordCloud(c *gin.Context, wordCloud *WordCloud, word *Word) (*mongo.UpdateResult, error) {
	db := db.GetDB()
	collection := db.Collection("word_cloud_sessions")
	filter := bson.M{"_id": wordCloud.Id}
	update := bson.M{"$push": bson.M{"words": word}}
	result, err := collection.UpdateOne(c, filter, update)
	return result, err
}
