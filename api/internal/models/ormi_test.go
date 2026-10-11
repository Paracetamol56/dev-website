package models

import (
	"testing"
	"time"

	"dev/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func date(t time.Time) primitive.DateTime {
	return primitive.NewDateTimeFromTime(t)
}

func createTodo(t *testing.T, todo Todo) primitive.ObjectID {
	t.Helper()
	result, err := CreateTodo(ctx, &todo)
	require.NoError(t, err)
	return result.InsertedID.(primitive.ObjectID)
}

func history(states ...any) []OrmiEvent {
	events := []OrmiEvent{}
	previous := ""
	for i := 0; i < len(states); i += 2 {
		state := states[i].(string)
		events = append(events, OrmiEvent{PreviousState: previous, NewState: state, UpdatedAt: date(states[i+1].(time.Time))})
		previous = state
	}
	return events
}

func titles(page *TodoPage) []string {
	names := []string{}
	for _, todo := range page.Items {
		names = append(names, todo.Title)
	}
	return names
}

func TestTodoCrud(t *testing.T) {
	testutil.Clean(t)
	userId, otherUser := primitive.NewObjectID(), primitive.NewObjectID()
	now := time.Now().Truncate(time.Millisecond)

	id := createTodo(t, Todo{User: userId, Title: "Write tests", Labels: []string{"api"}, State: TodoStateTodo, CreatedAt: date(now),
		History: history(TodoStateTodo, now)})

	todo, err := GetTodoByUserIdAndId(ctx, userId, id)
	require.NoError(t, err)
	assert.Equal(t, "Write tests", todo.Title)
	assert.Equal(t, []string{"api"}, todo.Labels)
	assert.Equal(t, date(now), todo.CreatedAt)
	_, err = GetTodoByUserIdAndId(ctx, otherUser, id)
	assert.ErrorIs(t, err, mongo.ErrNoDocuments, "todos are private")

	todo.Title = "Write more tests"
	todo.DueDate = date(now.Add(time.Hour))
	result, err := UpdateTodo(ctx, todo)
	require.NoError(t, err)
	assert.EqualValues(t, 1, result.MatchedCount)
	updated, err := GetTodoByUserIdAndId(ctx, userId, id)
	require.NoError(t, err)
	assert.Equal(t, todo, updated)

	todo.User = otherUser
	result, err = UpdateTodo(ctx, todo)
	require.NoError(t, err)
	assert.Zero(t, result.MatchedCount, "a todo cannot be moved to someone else")

	count, err := CountTodosByUserId(ctx, userId)
	require.NoError(t, err)
	assert.EqualValues(t, 1, count)
	count, err = CountTodosByUserId(ctx, otherUser)
	require.NoError(t, err)
	assert.Zero(t, count)

	deleted, err := DeleteTodo(ctx, otherUser, id)
	require.NoError(t, err)
	assert.Zero(t, deleted.DeletedCount)
	deleted, err = DeleteTodo(ctx, userId, id)
	require.NoError(t, err)
	assert.EqualValues(t, 1, deleted.DeletedCount)
	_, err = GetTodoByUserIdAndId(ctx, userId, id)
	assert.ErrorIs(t, err, mongo.ErrNoDocuments)
}

func TestGetTodosByUserId(t *testing.T) {
	testutil.Clean(t)
	userId := primitive.NewObjectID()
	day := func(n int) time.Time { return time.Date(2026, 3, n, 12, 0, 0, 0, time.UTC) }

	createTodo(t, Todo{User: userId, Title: "A", State: TodoStateTodo, Position: 2, CreatedAt: date(day(1)), DueDate: date(day(20)),
		Labels: []string{"home", "urgent"}, History: history(TodoStateTodo, day(1))})
	createTodo(t, Todo{User: userId, Title: "B", State: TodoStateDone, Position: 0, CreatedAt: date(day(2)),
		Labels: []string{"home"}, History: history(TodoStateTodo, day(2), TodoStateInProgress, day(5), TodoStateDone, day(9))})
	createTodo(t, Todo{User: userId, Title: "C", State: TodoStateInProgress, Position: 1, CreatedAt: date(day(3)), DueDate: date(day(10)),
		History: history(TodoStateTodo, day(3), TodoStateInProgress, day(4))})
	createTodo(t, Todo{User: userId, Title: "D", State: TodoStateCancelled, Position: 3, CreatedAt: date(day(4)),
		Labels: []string{"work"}, History: history(TodoStateTodo, day(4), TodoStateCancelled, day(6))})
	createTodo(t, Todo{User: primitive.NewObjectID(), Title: "Someone else's", State: TodoStateTodo, History: history(TodoStateTodo, day(1))})

	list := func(filter TodoFilter) *TodoPage {
		if filter.Sort == "" {
			filter.Sort = "position"
		}
		if filter.Limit == 0 {
			filter.Limit = 50
		}
		page, err := GetTodosByUserId(ctx, userId, &filter)
		require.NoError(t, err)
		return page
	}

	all := list(TodoFilter{})
	assert.Equal(t, 4, all.Total)
	assert.Equal(t, []string{"B", "C", "A", "D"}, titles(all), "by position")
	assert.Equal(t, []string{"D", "A", "C", "B"}, titles(list(TodoFilter{Order: "desc"})))

	assert.Equal(t, []string{"C", "A", "B", "D"}, titles(list(TodoFilter{Sort: "due"})), "todos without a due date come last")
	assert.Equal(t, []string{"A", "C", "D", "B"}, titles(list(TodoFilter{Sort: "due", Order: "desc"})), "even in descending order")
	assert.Equal(t, []string{"D", "C", "B", "A"}, titles(list(TodoFilter{Sort: "created"})), "newest first by default")
	assert.Equal(t, []string{"A", "B", "C", "D"}, titles(list(TodoFilter{Sort: "created", Order: "asc"})))
	assert.Equal(t, []string{"B", "D", "C", "A"}, titles(list(TodoFilter{Sort: "updated"})), "by last state change")
	assert.Equal(t, []string{"A", "C", "D", "B"}, titles(list(TodoFilter{Sort: "updated", Order: "asc"})))

	assert.Equal(t, []string{"C", "A"}, titles(list(TodoFilter{States: []string{TodoStateTodo, TodoStateInProgress}})))
	assert.Equal(t, []string{"B", "A"}, titles(list(TodoFilter{Labels: []string{"home"}})))
	assert.Equal(t, []string{"A"}, titles(list(TodoFilter{Labels: []string{"home", "urgent"}})), "every label is required")
	assert.Empty(t, titles(list(TodoFilter{Labels: []string{"home", "work"}})))

	from, to := day(4), day(9)
	assert.Equal(t, []string{"B", "C", "D"}, titles(list(TodoFilter{From: &from})), "from is inclusive")
	assert.Equal(t, []string{"C", "A", "D"}, titles(list(TodoFilter{To: &to})), "to is exclusive")
	assert.Equal(t, []string{"C", "D"}, titles(list(TodoFilter{From: &from, To: &to})))
	assert.Equal(t, []string{"D"}, titles(list(TodoFilter{From: &from, To: &to, States: []string{TodoStateCancelled}})))

	page := list(TodoFilter{Limit: 2, Offset: 1})
	assert.Equal(t, 4, page.Total, "the total ignores the page")
	assert.Equal(t, []string{"C", "A"}, titles(page))
	beyond := list(TodoFilter{Offset: 10})
	assert.Equal(t, 4, beyond.Total)
	assert.Equal(t, []Todo{}, beyond.Items)

	none, err := GetTodosByUserId(ctx, primitive.NewObjectID(), &TodoFilter{Sort: "position", Limit: 50})
	require.NoError(t, err)
	assert.Equal(t, &TodoPage{Total: 0, Items: []Todo{}}, none)
}

func TestReorderTodos(t *testing.T) {
	testutil.Clean(t)
	userId, otherUser := primitive.NewObjectID(), primitive.NewObjectID()
	a := createTodo(t, Todo{User: userId, Title: "A", Position: 0})
	b := createTodo(t, Todo{User: userId, Title: "B", Position: 1})
	c := createTodo(t, Todo{User: userId, Title: "C", Position: 2})
	foreign := createTodo(t, Todo{User: otherUser, Title: "Foreign", Position: 7})

	require.NoError(t, ReorderTodos(ctx, userId, []primitive.ObjectID{c, foreign, a, b}))

	page, err := GetTodosByUserId(ctx, userId, &TodoFilter{Sort: "position", Limit: 50})
	require.NoError(t, err)
	assert.Equal(t, []string{"C", "A", "B"}, titles(page))
	untouched, err := GetTodoByUserIdAndId(ctx, otherUser, foreign)
	require.NoError(t, err)
	assert.Equal(t, 7, untouched.Position, "todos of other users are not moved")
}

func TestGetCompletedTodosPerDay(t *testing.T) {
	testutil.Clean(t)
	userId := primitive.NewObjectID()
	at := func(value string) time.Time {
		parsed, err := time.Parse(time.RFC3339, value)
		require.NoError(t, err)
		return parsed
	}

	createTodo(t, Todo{User: userId, History: history(TodoStateTodo, at("2026-03-01T10:00:00Z"), TodoStateDone, at("2026-03-02T10:00:00Z"))})
	createTodo(t, Todo{User: userId, History: history(TodoStateTodo, at("2026-03-01T10:00:00Z"), TodoStateDone, at("2026-03-02T23:30:00Z"),
		TodoStateTodo, at("2026-03-03T08:00:00Z"), TodoStateDone, at("2026-03-05T08:00:00Z"))})
	createTodo(t, Todo{User: userId, History: history(TodoStateTodo, at("2026-03-01T10:00:00Z"), TodoStateCancelled, at("2026-03-02T11:00:00Z"))})
	createTodo(t, Todo{User: primitive.NewObjectID(), History: history(TodoStateDone, at("2026-03-02T10:00:00Z"))})

	days, err := GetCompletedTodosPerDay(ctx, userId, at("2026-03-01T00:00:00Z"), at("2026-04-01T00:00:00Z"), "UTC")
	require.NoError(t, err)
	assert.Equal(t, []TodoDayCount{{Date: "2026-03-02", Count: 2}, {Date: "2026-03-05", Count: 1}}, days, "a todo completed twice counts twice")

	days, err = GetCompletedTodosPerDay(ctx, userId, at("2026-03-01T00:00:00Z"), at("2026-04-01T00:00:00Z"), "Europe/Paris")
	require.NoError(t, err)
	assert.Equal(t, []TodoDayCount{{Date: "2026-03-02", Count: 1}, {Date: "2026-03-03", Count: 1}, {Date: "2026-03-05", Count: 1}}, days,
		"days are split in the given timezone")

	days, err = GetCompletedTodosPerDay(ctx, userId, at("2026-03-02T10:00:00Z"), at("2026-03-02T23:30:00Z"), "UTC")
	require.NoError(t, err)
	assert.Equal(t, []TodoDayCount{{Date: "2026-03-02", Count: 1}}, days, "from is inclusive, to exclusive")

	days, err = GetCompletedTodosPerDay(ctx, userId, at("2025-01-01T00:00:00Z"), at("2026-01-01T00:00:00Z"), "UTC")
	require.NoError(t, err)
	assert.Equal(t, []TodoDayCount{}, days)

	_, err = GetCompletedTodosPerDay(ctx, userId, at("2026-03-01T00:00:00Z"), at("2026-04-01T00:00:00Z"), "Mars/Olympus")
	assert.Error(t, err, "unknown timezone")
}

func TestTodoStreaks(t *testing.T) {
	today := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	days := func(dates ...string) []TodoDayCount {
		counts := []TodoDayCount{}
		for _, date := range dates {
			counts = append(counts, TodoDayCount{Date: date, Count: 1})
		}
		return counts
	}
	cases := []struct {
		name             string
		days             []TodoDayCount
		current, longest int
	}{
		{"nothing", days(), 0, 0},
		{"today only", days("2026-03-10"), 1, 1},
		{"today pending does not break the streak", days("2026-03-08", "2026-03-09"), 2, 2},
		{"streak through today", days("2026-03-08", "2026-03-09", "2026-03-10"), 3, 3},
		{"broken two days ago", days("2026-03-06", "2026-03-07", "2026-03-08"), 0, 3},
		{"longest in the past", days("2026-01-01", "2026-01-02", "2026-01-03", "2026-01-04", "2026-03-10"), 1, 4},
		{"across a month", days("2026-02-27", "2026-02-28", "2026-03-01"), 0, 3},
		{"older than the window", days("2024-01-01", "2024-01-02"), 0, 0},
	}
	for _, c := range cases {
		current, longest := todoStreaks(c.days, today)
		assert.Equal(t, c.current, current, c.name)
		assert.Equal(t, c.longest, longest, c.name)
	}
}

func TestGetTodoStats(t *testing.T) {
	testutil.Clean(t)
	userId := primitive.NewObjectID()
	now := time.Now().UTC()
	noon := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, time.UTC)
	daysAgo := func(n int) time.Time { return noon.AddDate(0, 0, -n) }

	empty, err := GetTodoStats(ctx, userId, time.UTC)
	require.NoError(t, err)
	assert.Equal(t, &TodoStats{Days: []TodoDayCount{}, FirstYear: now.Year()}, empty)

	createTodo(t, Todo{User: userId, State: TodoStateDone, CreatedAt: date(time.Date(2021, 6, 1, 0, 0, 0, 0, time.UTC)),
		History: history(TodoStateDone, daysAgo(1))})
	createTodo(t, Todo{User: userId, State: TodoStateDone, CreatedAt: date(daysAgo(40)), History: history(TodoStateDone, daysAgo(2))})
	createTodo(t, Todo{User: userId, State: TodoStateDone, CreatedAt: date(daysAgo(40)), History: history(TodoStateDone, daysAgo(2))})
	for n := 10; n < 15; n++ {
		createTodo(t, Todo{User: userId, State: TodoStateDone, CreatedAt: date(daysAgo(40)), History: history(TodoStateDone, daysAgo(n))})
	}
	createTodo(t, Todo{User: userId, State: TodoStateDone, CreatedAt: date(daysAgo(500)), History: history(TodoStateDone, daysAgo(400))})
	createTodo(t, Todo{User: userId, State: TodoStateTodo, CreatedAt: date(daysAgo(3)), DueDate: date(daysAgo(1))})
	createTodo(t, Todo{User: userId, State: TodoStateTodo, CreatedAt: date(daysAgo(3)), DueDate: date(noon.AddDate(0, 0, 3))})
	createTodo(t, Todo{User: userId, State: TodoStateInProgress, CreatedAt: date(daysAgo(3)), DueDate: date(daysAgo(5))})
	createTodo(t, Todo{User: userId, State: TodoStateStandby, CreatedAt: date(daysAgo(3))})
	createTodo(t, Todo{User: userId, State: TodoStateDone, CreatedAt: date(daysAgo(3)), DueDate: date(daysAgo(5))})
	createTodo(t, Todo{User: userId, State: TodoStateCancelled, CreatedAt: date(daysAgo(3)), DueDate: date(daysAgo(5))})
	createTodo(t, Todo{User: primitive.NewObjectID(), State: TodoStateTodo, DueDate: date(daysAgo(5)), History: history(TodoStateDone, noon)})

	stats, err := GetTodoStats(ctx, userId, time.UTC)
	require.NoError(t, err)
	assert.Equal(t, 2021, stats.FirstYear, "year of the oldest todo")
	assert.Equal(t, 2, stats.CurrentStreak, "yesterday and the day before, today is not over")
	assert.Equal(t, 5, stats.LongestStreak)
	assert.Equal(t, 2, stats.Todo)
	assert.Equal(t, 1, stats.InProgress)
	assert.Equal(t, 1, stats.Standby)
	assert.Equal(t, 2, stats.Overdue, "open todos due before today")
	require.Len(t, stats.Days, 7, "completions older than a year are left out")
	assert.Equal(t, TodoDayCount{Date: daysAgo(2).Format(dayFormat), Count: 2}, stats.Days[5])
	assert.Equal(t, TodoDayCount{Date: daysAgo(1).Format(dayFormat), Count: 1}, stats.Days[6])

	tokyo, err := time.LoadLocation("Asia/Tokyo")
	require.NoError(t, err)
	inTokyo, err := GetTodoStats(ctx, userId, tokyo)
	require.NoError(t, err)
	assert.Equal(t, 2021, inTokyo.FirstYear)
	assert.Len(t, inTokyo.Days, 7)
}

func TestGetTodoLabels(t *testing.T) {
	testutil.Clean(t)
	userId := primitive.NewObjectID()
	createTodo(t, Todo{User: userId, Labels: []string{"home", "Garden"}})
	createTodo(t, Todo{User: userId, Labels: []string{"home", "work"}})
	createTodo(t, Todo{User: userId, Labels: []string{"home", "work", "c++"}})
	createTodo(t, Todo{User: userId})
	createTodo(t, Todo{User: primitive.NewObjectID(), Labels: []string{"secret"}})

	labels, err := GetTodoLabels(ctx, userId, "", 10)
	require.NoError(t, err)
	assert.Equal(t, []string{"home", "work", "Garden", "c++"}, labels, "most used first, then alphabetically")

	labels, err = GetTodoLabels(ctx, userId, "", 2)
	require.NoError(t, err)
	assert.Equal(t, []string{"home", "work"}, labels)
	labels, err = GetTodoLabels(ctx, userId, "gArD", 10)
	require.NoError(t, err)
	assert.Equal(t, []string{"Garden"}, labels, "the search ignores case")
	labels, err = GetTodoLabels(ctx, userId, "o", 10)
	require.NoError(t, err)
	assert.Equal(t, []string{"home", "work"}, labels)
	labels, err = GetTodoLabels(ctx, userId, "c++", 10)
	require.NoError(t, err)
	assert.Equal(t, []string{"c++"}, labels, "the search is not a pattern")
	labels, err = GetTodoLabels(ctx, userId, "secret", 10)
	require.NoError(t, err)
	assert.Equal(t, []string{}, labels)
}

func TestTodoDatabaseFailures(t *testing.T) {
	testutil.BreakDB(t)
	userId, id := primitive.NewObjectID(), primitive.NewObjectID()

	_, err := GetTodoByUserIdAndId(ctx, userId, id)
	assert.Error(t, err)
	_, err = GetTodosByUserId(ctx, userId, &TodoFilter{Sort: "position", Limit: 50})
	assert.Error(t, err)
	_, err = CreateTodo(ctx, &Todo{})
	assert.Error(t, err)
	_, err = UpdateTodo(ctx, &Todo{})
	assert.Error(t, err)
	_, err = DeleteTodo(ctx, userId, id)
	assert.Error(t, err)
	_, err = CountTodosByUserId(ctx, userId)
	assert.Error(t, err)
	assert.Error(t, ReorderTodos(ctx, userId, []primitive.ObjectID{id}))
	_, err = GetCompletedTodosPerDay(ctx, userId, time.Now(), time.Now(), "UTC")
	assert.Error(t, err)
	_, err = GetTodoStats(ctx, userId, time.UTC)
	assert.Error(t, err)
	_, err = GetTodoLabels(ctx, userId, "", 10)
	assert.Error(t, err)
}
