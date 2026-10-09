package models

import (
	"context"
	"dev/internal/db"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	TodoStateTodo       = "TODO"
	TodoStateInProgress = "IN_PROGRESS"
	TodoStateDone       = "DONE"
	TodoStateCancelled  = "CANCELLED"
)

type OrmiEvent struct {
	PreviousState string             `json:"previousState" bson:"previousState" example:"TODO"`
	NewState      string             `json:"newState" bson:"newState" example:"DONE"`
	UpdatedAt     primitive.DateTime `json:"updatedAt" bson:"updatedAt" swaggertype:"string" format:"date-time"`
}

type Todo struct {
	Id          primitive.ObjectID `json:"id" bson:"_id,omitempty" swaggertype:"string"`
	User        primitive.ObjectID `json:"user,omitempty" bson:"user" swaggertype:"string"`
	Title       string             `json:"title" bson:"title"`
	Description string             `json:"description,omitempty" bson:"description"`
	DueDate     primitive.DateTime `json:"dueDate,omitempty" bson:"dueDate,omitempty" swaggertype:"string" format:"date-time"`
	Labels      []string           `json:"labels,omitempty" bson:"labels"`
	History     []OrmiEvent        `json:"history,omitempty" bson:"history"`
	State       string             `json:"state" bson:"state" enums:"TODO,IN_PROGRESS,DONE,CANCELLED"`
	Position    int                `json:"position" bson:"position"`
	GitURL      string             `json:"gitURL,omitempty" bson:"gitURL,omitempty"`
	GitIssue    uint32             `json:"gitIssue,omitempty" bson:"gitIssue,omitempty"`
	CreatedAt   primitive.DateTime `json:"createdAt" bson:"createdAt,omitempty" swaggertype:"string" format:"date-time"`
}

type TodoDayCount struct {
	Date  string `json:"date" bson:"_id" example:"2026-10-09"`
	Count int    `json:"count" bson:"count"`
}

func GetTodoByUserIdAndId(ctx context.Context, userId primitive.ObjectID, id primitive.ObjectID) (*Todo, error) {
	collection := db.GetDB().Collection("ormi")
	var todo Todo
	if err := collection.FindOne(ctx, bson.M{"user": userId, "_id": id}).Decode(&todo); err != nil {
		return nil, err
	}
	return &todo, nil
}

func GetTodosByUserId(ctx context.Context, userId primitive.ObjectID, state string) ([]Todo, error) {
	collection := db.GetDB().Collection("ormi")

	query := bson.M{"user": userId}
	if state != "" {
		query["state"] = state
	}

	sort := options.Find().SetSort(bson.D{{Key: "position", Value: 1}, {Key: "createdAt", Value: 1}})
	cursor, err := collection.Find(ctx, query, sort)
	if err != nil {
		return nil, err
	}

	todos := []Todo{}
	if err = cursor.All(ctx, &todos); err != nil {
		return nil, err
	}
	return todos, nil
}

func CreateTodo(ctx context.Context, todo *Todo) (*mongo.InsertOneResult, error) {
	return db.GetDB().Collection("ormi").InsertOne(ctx, todo)
}

func UpdateTodo(ctx context.Context, todo *Todo) (*mongo.UpdateResult, error) {
	collection := db.GetDB().Collection("ormi")
	return collection.ReplaceOne(ctx, bson.M{"_id": todo.Id, "user": todo.User}, todo)
}

func DeleteTodo(ctx context.Context, userId primitive.ObjectID, id primitive.ObjectID) (*mongo.DeleteResult, error) {
	return db.GetDB().Collection("ormi").DeleteOne(ctx, bson.M{"user": userId, "_id": id})
}

func CountTodosByUserId(ctx context.Context, userId primitive.ObjectID) (int64, error) {
	return db.GetDB().Collection("ormi").CountDocuments(ctx, bson.M{"user": userId})
}

func ReorderTodos(ctx context.Context, userId primitive.ObjectID, ids []primitive.ObjectID) error {
	writes := make([]mongo.WriteModel, len(ids))
	for position, id := range ids {
		writes[position] = mongo.NewUpdateOneModel().
			SetFilter(bson.M{"_id": id, "user": userId}).
			SetUpdate(bson.M{"$set": bson.M{"position": position}})
	}
	_, err := db.GetDB().Collection("ormi").BulkWrite(ctx, writes)
	return err
}

func GetCompletedTodosPerDay(ctx context.Context, userId primitive.ObjectID, since time.Time, timezone string) ([]TodoDayCount, error) {
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.M{"user": userId}}},
		bson.D{{Key: "$unwind", Value: "$history"}},
		bson.D{{Key: "$match", Value: bson.M{
			"history.newState":  TodoStateDone,
			"history.updatedAt": bson.M{"$gte": primitive.NewDateTimeFromTime(since)},
		}}},
		bson.D{{Key: "$group", Value: bson.M{
			"_id": bson.M{"$dateToString": bson.M{
				"format":   "%Y-%m-%d",
				"date":     "$history.updatedAt",
				"timezone": timezone,
			}},
			"count": bson.M{"$sum": 1},
		}}},
		bson.D{{Key: "$sort", Value: bson.M{"_id": 1}}},
	}

	cursor, err := db.GetDB().Collection("ormi").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}

	days := []TodoDayCount{}
	if err = cursor.All(ctx, &days); err != nil {
		return nil, err
	}
	return days, nil
}
