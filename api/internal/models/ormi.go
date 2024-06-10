package models

import (
	"github.com/Paracetamol56/dev-website/api/db"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// State could be : "TODO", "IN_PROGRESS", "DONE", "CANCELLED"

type OrmiEvent struct {
	PreviousState string             `json:"previousState" bson:"previousState"`
	NewState      string             `json:"newState" bson:"newState"`
	UpdatedAt     primitive.DateTime `json:"updatedAt" bson:"updatedAt"`
}

type Todo struct {
	Id          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	User        primitive.ObjectID `json:"user,omitempty" bson:"user"`
	Title       string             `json:"title" bson:"title"`
	Description string             `json:"description,omitempty" bson:"description"`
	DueDate     primitive.DateTime `json:"dueDate,omitempty" bson:"dueDate,omitempty"`
	Labels      []string           `json:"labels,omitempty" bson:"labels"`
	History     []OrmiEvent        `json:"history,omitempty" bson:"history"`
	State       string             `json:"state,omitempty" bson:"state,omitempty"`
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

func GetTodoByUserIdAndId(c *gin.Context, userId primitive.ObjectID, id primitive.ObjectID) (*Todo, error) {
	db := db.GetDB()
	collection := db.Collection("ormi")
	var todo Todo
	if err := collection.FindOne(c, bson.M{"user": userId, "_id": id}).Decode(&todo); err != nil {
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

func GetTodosByUserIdAndState(c *gin.Context, userId primitive.ObjectID, state string) ([]Todo, error) {
	db := db.GetDB()
	collection := db.Collection("ormi")
	// Find todos with the last state in history equals to the given state
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.D{{Key: "user", Value: userId}}}},
		bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$history"}}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "history.updatedAt", Value: 1}}}},
		bson.D{
			{Key: "$group",
				Value: bson.D{
					{Key: "_id", Value: "$_id"},
					{Key: "title", Value: bson.D{{Key: "$first", Value: "$title"}}},
					{Key: "dueDate", Value: bson.D{{Key: "$first", Value: "$dueDate"}}},
					{Key: "state", Value: bson.D{{Key: "$first", Value: "$history.newState"}}},
					{Key: "createdAt", Value: bson.D{{Key: "$first", Value: "$createdAt"}}},
				},
			},
		},
		bson.D{{Key: "$match", Value: bson.D{{Key: "state", Value: state}}}},
	}
	cursor, err := collection.Aggregate(c, pipeline)
	if err != nil {
		return nil, err
	}
	var todos []Todo
	if err = cursor.All(c, &todos); err != nil {
		return nil, err
	}
	return todos, nil
}
