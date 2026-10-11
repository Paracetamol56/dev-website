package ormi

import (
	"context"
	"dev/internal/utils"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

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

// Clients send midnight in their own timezone, hence the day of slack
func isPastDueDate(dueDate primitive.DateTime) bool {
	return dueDate.Time().Before(time.Now().Add(-24 * time.Hour))
}

const suggestedLabels = 10

func ListTodos(ctx context.Context, userId primitive.ObjectID, filter *TodoFilter) (*TodoPage, error) {
	return GetTodosByUserId(ctx, userId, filter)
}

func GetTodo(ctx context.Context, userId primitive.ObjectID, id primitive.ObjectID) (*Todo, error) {
	todo, err := GetTodoByUserIdAndId(ctx, userId, id)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, utils.NotFound("Todo not found")
	}
	return todo, err
}

// AddTodo creates a todo in the TODO state, at the end of the list
func AddTodo(ctx context.Context, userId primitive.ObjectID, body TodoBody) (*Todo, error) {
	if body.DueDate != 0 && isPastDueDate(body.DueDate) {
		return nil, utils.Invalid("Due date should not be in the past")
	}

	position, err := CountTodosByUserId(ctx, userId)
	if err != nil {
		return nil, err
	}

	now := primitive.NewDateTimeFromTime(time.Now())
	result, err := CreateTodo(ctx, &Todo{
		User:        userId,
		Title:       body.Title,
		Description: body.Description,
		DueDate:     body.DueDate,
		Labels:      body.Labels,
		State:       TodoStateTodo,
		Position:    int(position),
		History:     []OrmiEvent{{PreviousState: "", NewState: TodoStateTodo, UpdatedAt: now}},
		GitURL:      body.GitURL,
		GitIssue:    body.GitIssue,
		CreatedAt:   now,
	})
	if err != nil {
		return nil, err
	}

	return GetTodoByUserIdAndId(ctx, userId, result.InsertedID.(primitive.ObjectID))
}

// ChangeTodo applies the given fields; a new state is appended to the history
func ChangeTodo(ctx context.Context, userId primitive.ObjectID, id primitive.ObjectID, body PatchTodoBody) (*Todo, error) {
	todo, err := GetTodo(ctx, userId, id)
	if err != nil {
		return nil, err
	}

	if body.Title != nil {
		todo.Title = *body.Title
	}
	if body.Description != nil {
		todo.Description = *body.Description
	}
	if body.DueDate.Set && body.DueDate.Value != todo.DueDate {
		if body.DueDate.Value != 0 && isPastDueDate(body.DueDate.Value) {
			return nil, utils.Invalid("Due date should not be in the past")
		}
		todo.DueDate = body.DueDate.Value
	}
	if body.Labels != nil {
		todo.Labels = *body.Labels
	}
	if body.GitURL != nil {
		todo.GitURL = *body.GitURL
	}
	if body.GitIssue != nil {
		todo.GitIssue = *body.GitIssue
	}
	if body.State != nil && todo.State != *body.State {
		todo.History = append(todo.History, OrmiEvent{
			PreviousState: todo.State,
			NewState:      *body.State,
			UpdatedAt:     primitive.NewDateTimeFromTime(time.Now()),
		})
		todo.State = *body.State
	}

	if _, err := UpdateTodo(ctx, todo); err != nil {
		return nil, err
	}
	return todo, nil
}

func RemoveTodo(ctx context.Context, userId primitive.ObjectID, id primitive.ObjectID) error {
	result, err := DeleteTodo(ctx, userId, id)
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return utils.NotFound("Todo not found")
	}
	return nil
}

func OrderTodos(ctx context.Context, userId primitive.ObjectID, ids []primitive.ObjectID) error {
	return ReorderTodos(ctx, userId, ids)
}

func loadLocation(timezone string) (*time.Location, error) {
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, utils.Invalid("Invalid timezone")
	}
	return location, nil
}

func Stats(ctx context.Context, userId primitive.ObjectID, timezone string) (*TodoStats, error) {
	location, err := loadLocation(timezone)
	if err != nil {
		return nil, err
	}
	return GetTodoStats(ctx, userId, location)
}

// YearStats counts the todos completed per day of a calendar year
func YearStats(ctx context.Context, userId primitive.ObjectID, year int, timezone string) ([]TodoDayCount, error) {
	if year < 1970 || year > 9999 {
		return nil, utils.Invalid("Invalid year")
	}
	location, err := loadLocation(timezone)
	if err != nil {
		return nil, err
	}

	from := time.Date(year, time.January, 1, 0, 0, 0, 0, location)
	return GetCompletedTodosPerDay(ctx, userId, from, from.AddDate(1, 0, 0), timezone)
}

func SuggestLabels(ctx context.Context, userId primitive.ObjectID, search string) ([]string, error) {
	return GetTodoLabels(ctx, userId, search, suggestedLabels)
}
