package models

import (
	"github.com/Paracetamol56/dev-website/api/db"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type OrmiEvent struct {
	EventType     string `json:"eventType" bson:"eventType"`
	PreviousState string `json:"previousState" bson:"previousState"`
	NewState      string `json:"newState" bson:"newState"`
	TimeStamp     string `json:"timeStamp" bson:"timeStamp"`
}

type Todo struct {
	Id          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	User        primitive.ObjectID `json:"user,omitempty" bson:"user"`
	Title       string             `json:"title" bson:"title"`
	Description string             `json:"description" bson:"description"`
	Labels      []string           `json:"labels" bson:"labels"`
	History     []OrmiEvent        `json:"history" bson:"history"`
	GitURL      string             `json:"gitURL,omitempty" bson:"gitURL,omitempty"`
	GitIssue    uint32             `json:"gitIssue,omitempty" bson:"gitIssue,omitempty"`
	CreatedAt   primitive.DateTime `json:"createdAt" bson:"createdAt,omitempty"`
}

func CreateTodo(c *gin.Context, todo *Todo) (*mongo.InsertOneResult, error) {
	db := db.GetDB()
	collection := db.Collection("ormi")
	result, err := collection.InsertOne(c, todo)
	return result, err
}

func GetTodoById(c *gin.Context, id primitive.ObjectID) (*Todo, error) {
	db := db.GetDB()
	collection := db.Collection("ormi")
	var todo Todo
	if err := collection.FindOne(c, bson.M{"_id": id}).Decode(&todo); err != nil {
		return nil, err
	}
	return &todo, nil
}

func GetTodosByUserId(c *gin.Context, userId primitive.ObjectID) ([]Todo, error) {
	db := db.GetDB()
	collection := db.Collection("ormi")
	cursor, err := collection.Find(c, bson.M{"user": userId})
	if err != nil {
		return nil, err
	}
	var todos []Todo
	if err = cursor.All(c, &todos); err != nil {
		return nil, err
	}
	return todos, nil
}
