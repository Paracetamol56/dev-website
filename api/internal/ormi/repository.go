package ormi

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
