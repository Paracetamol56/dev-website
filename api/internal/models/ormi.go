package models

import (
	"context"
	"dev/internal/db"
	"regexp"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	TodoStateTodo       = "TODO"
	TodoStateInProgress = "IN_PROGRESS"
	TodoStateStandby    = "STANDBY"
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
	State       string             `json:"state" bson:"state" enums:"TODO,IN_PROGRESS,STANDBY,DONE,CANCELLED"`
	Position    int                `json:"position" bson:"position"`
	GitURL      string             `json:"gitURL,omitempty" bson:"gitURL,omitempty"`
	GitIssue    uint32             `json:"gitIssue,omitempty" bson:"gitIssue,omitempty"`
	CreatedAt   primitive.DateTime `json:"createdAt" bson:"createdAt,omitempty" swaggertype:"string" format:"date-time"`
}

type TodoDayCount struct {
	Date  string `json:"date" bson:"_id" example:"2026-10-09"`
	Count int    `json:"count" bson:"count"`
}

type TodoStats struct {
	Days          []TodoDayCount `json:"days"`
	FirstYear     int            `json:"firstYear" example:"2024"`
	CurrentStreak int            `json:"currentStreak"`
	LongestStreak int            `json:"longestStreak"`
	Todo          int            `json:"todo"`
	InProgress    int            `json:"inProgress"`
	Standby       int            `json:"standby"`
	Overdue       int            `json:"overdue"`
}

type TodoFilter struct {
	States []string   `form:"state" binding:"omitempty,dive,oneof=TODO IN_PROGRESS STANDBY DONE CANCELLED"`
	Labels []string   `form:"label" binding:"omitempty,max=20,dive,max=50"`
	From   *time.Time `form:"from" time_format:"2006-01-02T15:04:05Z07:00"`
	To     *time.Time `form:"to" time_format:"2006-01-02T15:04:05Z07:00"`
	Sort   string     `form:"sort,default=position" binding:"oneof=position due created updated"`
	Order  string     `form:"order" binding:"omitempty,oneof=asc desc"`
	Limit  int64      `form:"limit,default=50" binding:"min=1,max=200"`
	Offset int64      `form:"offset,default=0" binding:"min=0"`
}

type TodoPage struct {
	Total int    `json:"total" example:"42"`
	Items []Todo `json:"items"`
}

var openTodoStates = []string{TodoStateTodo, TodoStateInProgress, TodoStateStandby}

func GetTodoByUserIdAndId(ctx context.Context, userId primitive.ObjectID, id primitive.ObjectID) (*Todo, error) {
	collection := db.GetDB().Collection("ormi")
	var todo Todo
	if err := collection.FindOne(ctx, bson.M{"user": userId, "_id": id}).Decode(&todo); err != nil {
		return nil, err
	}
	return &todo, nil
}

func GetTodosByUserId(ctx context.Context, userId primitive.ObjectID, filter *TodoFilter) (*TodoPage, error) {
	collection := db.GetDB().Collection("ormi")

	query := bson.M{"user": userId}
	if len(filter.States) > 0 {
		query["state"] = bson.M{"$in": filter.States}
	}

	if len(filter.Labels) > 0 {
		query["labels"] = bson.M{"$all": filter.Labels}
	}

	lastUpdate := bson.M{}
	if filter.From != nil {
		lastUpdate["$gte"] = primitive.NewDateTimeFromTime(*filter.From)
	}
	if filter.To != nil {
		lastUpdate["$lt"] = primitive.NewDateTimeFromTime(*filter.To)
	}

	sortKeys := map[string][]string{
		"position": {"position", "createdAt"},
		"due":      {"dueDate", "_id"},
		"created":  {"createdAt", "_id"},
		"updated":  {"lastUpdate", "_id"},
	}[filter.Sort]
	direction := 1
	if filter.Order == "desc" || (filter.Order == "" && (filter.Sort == "created" || filter.Sort == "updated")) {
		direction = -1
	}
	sort := bson.D{}
	if filter.Sort == "due" {
		sort = append(sort, bson.E{Key: "noDueDate", Value: 1})
	}
	for _, key := range sortKeys {
		sort = append(sort, bson.E{Key: key, Value: direction})
	}

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: query}},
		bson.D{{Key: "$addFields", Value: bson.M{
			"lastUpdate": bson.M{"$last": "$history.updatedAt"},
			"noDueDate":  bson.M{"$eq": bson.A{bson.M{"$type": "$dueDate"}, "missing"}},
		}}},
	}
	if len(lastUpdate) > 0 {
		pipeline = append(pipeline, bson.D{{Key: "$match", Value: bson.M{"lastUpdate": lastUpdate}}})
	}
	pipeline = append(pipeline, bson.D{{Key: "$facet", Value: bson.M{
		"total": bson.A{bson.M{"$count": "count"}},
		"items": bson.A{
			bson.M{"$sort": sort},
			bson.M{"$skip": filter.Offset},
			bson.M{"$limit": filter.Limit},
		},
	}}})

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	var results []struct {
		Total []struct {
			Count int `bson:"count"`
		} `bson:"total"`
		Items []Todo `bson:"items"`
	}
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	page := &TodoPage{Items: []Todo{}}
	if len(results) > 0 {
		if len(results[0].Total) > 0 {
			page.Total = results[0].Total[0].Count
		}
		if results[0].Items != nil {
			page.Items = results[0].Items
		}
	}
	return page, nil
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

func GetCompletedTodosPerDay(ctx context.Context, userId primitive.ObjectID, from time.Time, to time.Time, timezone string) ([]TodoDayCount, error) {
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.M{"user": userId}}},
		bson.D{{Key: "$unwind", Value: "$history"}},
		bson.D{{Key: "$match", Value: bson.M{
			"history.newState": TodoStateDone,
			"history.updatedAt": bson.M{
				"$gte": primitive.NewDateTimeFromTime(from),
				"$lt":  primitive.NewDateTimeFromTime(to),
			},
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

const dayFormat = "2006-01-02"

func todoStreaks(days []TodoDayCount, today time.Time) (current int, longest int) {
	active := make(map[string]bool, len(days))
	for _, day := range days {
		active[day.Date] = true
	}

	run := 0
	for day := today.AddDate(-1, 0, -7); !day.After(today); day = day.AddDate(0, 0, 1) {
		if active[day.Format(dayFormat)] {
			run++
			longest = max(longest, run)
		} else {
			run = 0
		}
	}

	// A day without completion only breaks the streak once it is over
	day := today
	if !active[day.Format(dayFormat)] {
		day = day.AddDate(0, 0, -1)
	}
	for active[day.Format(dayFormat)] {
		current++
		day = day.AddDate(0, 0, -1)
	}
	return current, longest
}

func GetTodoStats(ctx context.Context, userId primitive.ObjectID, location *time.Location) (*TodoStats, error) {
	collection := db.GetDB().Collection("ormi")
	now := time.Now().In(location)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)

	days, err := GetCompletedTodosPerDay(ctx, userId, today.AddDate(-1, 0, -7), today.AddDate(0, 0, 1), location.String())
	if err != nil {
		return nil, err
	}
	stats := &TodoStats{Days: days, FirstYear: today.Year()}

	var first Todo
	oldest := options.FindOne().SetSort(bson.M{"createdAt": 1})
	err = collection.FindOne(ctx, bson.M{"user": userId}, oldest).Decode(&first)
	if err == nil {
		stats.FirstYear = first.CreatedAt.Time().In(location).Year()
	} else if err != mongo.ErrNoDocuments {
		return nil, err
	}

	stats.CurrentStreak, stats.LongestStreak = todoStreaks(days, today)

	cursor, err := collection.Aggregate(ctx, mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.M{"user": userId, "state": bson.M{"$in": openTodoStates}}}},
		bson.D{{Key: "$group", Value: bson.M{"_id": "$state", "count": bson.M{"$sum": 1}}}},
	})
	if err != nil {
		return nil, err
	}
	var states []struct {
		State string `bson:"_id"`
		Count int    `bson:"count"`
	}
	if err = cursor.All(ctx, &states); err != nil {
		return nil, err
	}
	for _, state := range states {
		switch state.State {
		case TodoStateTodo:
			stats.Todo = state.Count
		case TodoStateInProgress:
			stats.InProgress = state.Count
		case TodoStateStandby:
			stats.Standby = state.Count
		}
	}

	overdue, err := collection.CountDocuments(ctx, bson.M{
		"user":    userId,
		"state":   bson.M{"$in": openTodoStates},
		"dueDate": bson.M{"$lt": primitive.NewDateTimeFromTime(today)},
	})
	if err != nil {
		return nil, err
	}
	stats.Overdue = int(overdue)

	return stats, nil
}

func GetTodoLabels(ctx context.Context, userId primitive.ObjectID, search string, limit int) ([]string, error) {
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.M{"user": userId}}},
		bson.D{{Key: "$unwind", Value: "$labels"}},
	}
	if search != "" {
		pattern := primitive.Regex{Pattern: regexp.QuoteMeta(search), Options: "i"}
		pipeline = append(pipeline, bson.D{{Key: "$match", Value: bson.M{"labels": pattern}}})
	}
	pipeline = append(pipeline,
		bson.D{{Key: "$group", Value: bson.M{"_id": "$labels", "count": bson.M{"$sum": 1}}}},
		bson.D{{Key: "$sort", Value: bson.D{{Key: "count", Value: -1}, {Key: "_id", Value: 1}}}},
		bson.D{{Key: "$limit", Value: limit}},
	)

	cursor, err := db.GetDB().Collection("ormi").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	var results []struct {
		Label string `bson:"_id"`
	}
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	labels := make([]string, len(results))
	for i, result := range results {
		labels[i] = result.Label
	}
	return labels, nil
}
