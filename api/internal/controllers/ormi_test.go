package controllers

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"testing"
	"time"

	"dev/internal/models"
	"dev/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func postTodo(t *testing.T, userId primitive.ObjectID, body map[string]any) models.Todo {
	t.Helper()
	response := do(t, http.MethodPost, "/ormi", body, as(userId))
	require.Equal(t, http.StatusCreated, response.Code, response.Body.String())
	return decode[models.Todo](t, response)
}

func patchTodo(t *testing.T, userId primitive.ObjectID, id primitive.ObjectID, body any) models.Todo {
	t.Helper()
	response := do(t, http.MethodPatch, "/ormi/"+id.Hex(), body, as(userId))
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	return decode[models.Todo](t, response)
}

func listTodos(t *testing.T, userId primitive.ObjectID, query string) models.TodoPage {
	t.Helper()
	response := do(t, http.MethodGet, "/ormi?"+query, nil, as(userId))
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	return decode[models.TodoPage](t, response)
}

func todoTitles(page models.TodoPage) []string {
	titles := []string{}
	for _, todo := range page.Items {
		titles = append(titles, todo.Title)
	}
	return titles
}

func TestPostTodo(t *testing.T) {
	testutil.Clean(t)
	userId := primitive.NewObjectID()
	due := time.Now().Add(48 * time.Hour).UTC().Truncate(time.Second)

	todo := postTodo(t, userId, map[string]any{
		"title": "Write tests", "description": "All of them", "dueDate": due, "labels": []string{"api", "go"},
		"gitURL": "https://github.com/Paracetamol56/dev-showcase", "gitIssue": 12,
	})
	assert.False(t, todo.Id.IsZero())
	assert.Equal(t, userId, todo.User)
	assert.Equal(t, "Write tests", todo.Title)
	assert.Equal(t, "All of them", todo.Description)
	assert.Equal(t, due, todo.DueDate.Time().UTC())
	assert.Equal(t, []string{"api", "go"}, todo.Labels)
	assert.Equal(t, "https://github.com/Paracetamol56/dev-showcase", todo.GitURL)
	assert.EqualValues(t, 12, todo.GitIssue)
	assert.Equal(t, models.TodoStateTodo, todo.State)
	assert.Equal(t, 0, todo.Position)
	assert.WithinDuration(t, time.Now(), todo.CreatedAt.Time(), 5*time.Second)
	require.Len(t, todo.History, 1)
	assert.Equal(t, "", todo.History[0].PreviousState)
	assert.Equal(t, models.TodoStateTodo, todo.History[0].NewState)

	second := postTodo(t, userId, map[string]any{"title": "Second"})
	assert.Equal(t, 1, second.Position, "new todos go to the end of the list")
	assert.Zero(t, second.DueDate)
	assert.Equal(t, 0, postTodo(t, primitive.NewObjectID(), map[string]any{"title": "Someone else's"}).Position)

	midnight := time.Now().UTC().Truncate(24 * time.Hour)
	assert.Equal(t, midnight, postTodo(t, userId, map[string]any{"title": "Due today", "dueDate": midnight}).DueDate.Time().UTC(),
		"today is not in the past, whatever the timezone of the client")
}

func TestPostTodoRejections(t *testing.T) {
	testutil.Clean(t)
	userId := primitive.NewObjectID()

	for name, body := range map[string]any{
		"not json":         "not json",
		"no title":         map[string]any{"description": "No title"},
		"short title":      map[string]any{"title": "A"},
		"long title":       map[string]any{"title": repeat("a", 101)},
		"long description": map[string]any{"title": "Valid", "description": repeat("a", 1001)},
		"git url":          map[string]any{"title": "Valid", "gitURL": "not a url"},
		"negative issue":   map[string]any{"title": "Valid", "gitIssue": -1},
		"due date":         map[string]any{"title": "Valid", "dueDate": "tomorrow"},
		"labels":           map[string]any{"title": "Valid", "labels": "one"},
	} {
		assert.Equal(t, http.StatusBadRequest, do(t, http.MethodPost, "/ormi", body, as(userId)).Code, name)
	}

	past := do(t, http.MethodPost, "/ormi", map[string]any{"title": "Too late", "dueDate": time.Now().Add(-25 * time.Hour)}, as(userId))
	assert.Equal(t, http.StatusBadRequest, past.Code)
	assert.Equal(t, "Due date should not be in the past", errorOf(t, past))
	assert.Zero(t, count(t, "ormi"))
	assert.Equal(t, http.StatusUnauthorized, do(t, http.MethodPost, "/ormi", map[string]any{"title": "Anonymous"}).Code)

	testutil.BreakDB(t)
	assert.Equal(t, http.StatusInternalServerError, do(t, http.MethodPost, "/ormi", map[string]any{"title": "Valid"}, as(userId)).Code)
}

func TestIsPastDueDate(t *testing.T) {
	at := func(offset time.Duration) primitive.DateTime {
		return primitive.NewDateTimeFromTime(time.Now().Add(offset))
	}
	assert.False(t, isPastDueDate(at(time.Hour)))
	assert.False(t, isPastDueDate(at(-23*time.Hour)), "clients send midnight in their own timezone")
	assert.True(t, isPastDueDate(at(-25*time.Hour)))
}

func TestOptionalDate(t *testing.T) {
	var body struct {
		DueDate OptionalDate `json:"dueDate"`
	}
	require.NoError(t, json.Unmarshal([]byte(`{}`), &body))
	assert.False(t, body.DueDate.Set, "absent")

	require.NoError(t, json.Unmarshal([]byte(`{"dueDate": null}`), &body))
	assert.True(t, body.DueDate.Set, "null clears the date")
	assert.Zero(t, body.DueDate.Value)

	require.NoError(t, json.Unmarshal([]byte(`{"dueDate": "2026-03-01T00:00:00Z"}`), &body))
	assert.True(t, body.DueDate.Set)
	assert.Equal(t, time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC), body.DueDate.Value.Time().UTC())

	assert.Error(t, json.Unmarshal([]byte(`{"dueDate": "soon"}`), &body))
}

func TestGetTodo(t *testing.T) {
	testutil.Clean(t)
	userId := primitive.NewObjectID()
	todo := postTodo(t, userId, map[string]any{"title": "Mine"})

	response := do(t, http.MethodGet, "/ormi/"+todo.Id.Hex(), nil, as(userId))
	require.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, todo, decode[models.Todo](t, response))

	private := do(t, http.MethodGet, "/ormi/"+todo.Id.Hex(), nil, as(primitive.NewObjectID()))
	assert.Equal(t, http.StatusNotFound, private.Code, "todos of other users do not exist")
	assert.Equal(t, "Todo not found", errorOf(t, private))
	assert.Equal(t, http.StatusNotFound, do(t, http.MethodGet, "/ormi/"+primitive.NewObjectID().Hex(), nil, as(userId)).Code)
	assert.Equal(t, http.StatusBadRequest, do(t, http.MethodGet, "/ormi/first", nil, as(userId)).Code)

	testutil.BreakDB(t)
	assert.Equal(t, http.StatusInternalServerError, do(t, http.MethodGet, "/ormi/"+todo.Id.Hex(), nil, as(userId)).Code)
}

func TestGetTodos(t *testing.T) {
	testutil.Clean(t)
	userId := primitive.NewObjectID()
	a := postTodo(t, userId, map[string]any{"title": "Task A", "labels": []string{"home"}, "dueDate": time.Now().Add(72 * time.Hour)})
	b := postTodo(t, userId, map[string]any{"title": "Task B", "labels": []string{"home", "urgent"}, "dueDate": time.Now().Add(24 * time.Hour)})
	postTodo(t, userId, map[string]any{"title": "Task C"})
	postTodo(t, primitive.NewObjectID(), map[string]any{"title": "Someone else's"})
	patchTodo(t, userId, b.Id, map[string]any{"state": "DONE"})

	all := listTodos(t, userId, "")
	assert.Equal(t, 3, all.Total)
	assert.Equal(t, []string{"Task A", "Task B", "Task C"}, todoTitles(all), "custom order by default")
	assert.Equal(t, a.Id, all.Items[0].Id)

	assert.Equal(t, []string{"Task A", "Task C"}, todoTitles(listTodos(t, userId, "state=TODO&state=IN_PROGRESS")))
	assert.Equal(t, []string{"Task B"}, todoTitles(listTodos(t, userId, "state=DONE")))
	assert.Equal(t, []string{"Task A", "Task B"}, todoTitles(listTodos(t, userId, "label=home")))
	assert.Equal(t, []string{"Task B"}, todoTitles(listTodos(t, userId, "label=home&label=urgent")))
	assert.Equal(t, []string{"Task B", "Task A", "Task C"}, todoTitles(listTodos(t, userId, "sort=due")))
	assert.Equal(t, []string{"Task C", "Task B", "Task A"}, todoTitles(listTodos(t, userId, "sort=created")))
	assert.Equal(t, []string{"Task A", "Task B", "Task C"}, todoTitles(listTodos(t, userId, "sort=created&order=asc")))
	assert.Equal(t, "Task B", todoTitles(listTodos(t, userId, "sort=updated"))[0], "the last state change comes first")

	page := listTodos(t, userId, "limit=1&offset=1")
	assert.Equal(t, 3, page.Total)
	assert.Equal(t, []string{"Task B"}, todoTitles(page))

	window := url.Values{}
	window.Set("from", time.Now().Add(-time.Minute).Format(time.RFC3339))
	window.Set("to", time.Now().Add(time.Minute).Format(time.RFC3339))
	assert.Equal(t, 3, listTodos(t, userId, window.Encode()).Total)
	window.Set("from", time.Now().Add(time.Minute).Format(time.RFC3339))
	window.Set("to", time.Now().Add(time.Hour).Format(time.RFC3339))
	assert.Zero(t, listTodos(t, userId, window.Encode()).Total)

	empty := do(t, http.MethodGet, "/ormi", nil, as(primitive.NewObjectID()))
	assert.JSONEq(t, `{"total": 0, "items": []}`, empty.Body.String())

	labels := url.Values{}
	for i := range 21 {
		labels.Add("label", strconv.Itoa(i))
	}
	for _, query := range []string{
		"state=FINISHED", "sort=priority", "order=up", "limit=0", "limit=201", "offset=-1", "from=yesterday", "to=2026-03-01",
		"label=" + repeat("a", 51), labels.Encode(),
	} {
		assert.Equal(t, http.StatusBadRequest, do(t, http.MethodGet, "/ormi?"+query, nil, as(userId)).Code, query)
	}

	testutil.BreakDB(t)
	assert.Equal(t, http.StatusInternalServerError, do(t, http.MethodGet, "/ormi", nil, as(userId)).Code)
}

func TestPatchTodo(t *testing.T) {
	testutil.Clean(t)
	userId := primitive.NewObjectID()
	due := time.Now().Add(48 * time.Hour).UTC().Truncate(time.Second)
	todo := postTodo(t, userId, map[string]any{"title": "Original", "description": "Text", "labels": []string{"a"}, "dueDate": due})

	patched := patchTodo(t, userId, todo.Id, map[string]any{"title": "Renamed"})
	assert.Equal(t, "Renamed", patched.Title)
	assert.Equal(t, "Text", patched.Description, "fields left out are kept")
	assert.Equal(t, []string{"a"}, patched.Labels)
	assert.Equal(t, due, patched.DueDate.Time().UTC())
	assert.Len(t, patched.History, 1)

	patched = patchTodo(t, userId, todo.Id, map[string]any{
		"description": "", "labels": []string{}, "gitURL": "https://example.test/repo", "gitIssue": 3,
	})
	assert.Empty(t, patched.Description)
	assert.Empty(t, patched.Labels)
	assert.Equal(t, "https://example.test/repo", patched.GitURL)
	assert.EqualValues(t, 3, patched.GitIssue)

	patched = patchTodo(t, userId, todo.Id, `{"dueDate": null}`)
	assert.Zero(t, patched.DueDate, "null removes the due date")
	stored := decode[models.Todo](t, do(t, http.MethodGet, "/ormi/"+todo.Id.Hex(), nil, as(userId)))
	assert.Zero(t, stored.DueDate)
	assert.Equal(t, "Renamed", stored.Title)

	patched = patchTodo(t, userId, todo.Id, map[string]any{"state": "IN_PROGRESS"})
	assert.Equal(t, models.TodoStateInProgress, patched.State)
	patched = patchTodo(t, userId, todo.Id, map[string]any{"state": "IN_PROGRESS"})
	require.Len(t, patched.History, 2, "setting the same state is not an event")
	patched = patchTodo(t, userId, todo.Id, map[string]any{"state": "DONE"})
	require.Len(t, patched.History, 3)
	assert.Equal(t, models.OrmiEvent{PreviousState: "IN_PROGRESS", NewState: "DONE", UpdatedAt: patched.History[2].UpdatedAt}, patched.History[2])
	assert.WithinDuration(t, time.Now(), patched.History[2].UpdatedAt.Time(), 5*time.Second)
	stored = decode[models.Todo](t, do(t, http.MethodGet, "/ormi/"+todo.Id.Hex(), nil, as(userId)))
	assert.Equal(t, patched, stored)
}

func TestPatchTodoDueDate(t *testing.T) {
	testutil.Clean(t)
	userId := primitive.NewObjectID()
	overdue := primitive.NewDateTimeFromTime(time.Now().Add(-72 * time.Hour).Truncate(time.Second))
	result, err := models.CreateTodo(ctx, &models.Todo{User: userId, Title: "Overdue", State: models.TodoStateTodo, DueDate: overdue})
	require.NoError(t, err)
	id := result.InsertedID.(primitive.ObjectID)

	patched := patchTodo(t, userId, id, map[string]any{"title": "Still overdue", "dueDate": overdue})
	assert.Equal(t, overdue, patched.DueDate, "an unchanged past date does not block other edits")
	assert.Equal(t, "Still overdue", patched.Title)

	past := do(t, http.MethodPatch, "/ormi/"+id.Hex(), map[string]any{"title": "Moved", "dueDate": time.Now().Add(-48 * time.Hour)}, as(userId))
	assert.Equal(t, http.StatusBadRequest, past.Code)
	assert.Equal(t, "Due date should not be in the past", errorOf(t, past))
	assert.Equal(t, "Still overdue", decode[models.Todo](t, do(t, http.MethodGet, "/ormi/"+id.Hex(), nil, as(userId))).Title, "nothing is saved")

	future := time.Now().Add(48 * time.Hour).UTC().Truncate(time.Second)
	assert.Equal(t, future, patchTodo(t, userId, id, map[string]any{"dueDate": future}).DueDate.Time().UTC())
}

func TestPatchTodoRejections(t *testing.T) {
	testutil.Clean(t)
	userId := primitive.NewObjectID()
	todo := postTodo(t, userId, map[string]any{"title": "Original"})
	path := "/ormi/" + todo.Id.Hex()

	for name, body := range map[string]any{
		"not json":      "not json",
		"short title":   map[string]any{"title": "A"},
		"long text":     map[string]any{"description": repeat("a", 1001)},
		"unknown state": map[string]any{"state": "FINISHED"},
		"git url":       map[string]any{"gitURL": "not a url"},
		"due date":      map[string]any{"dueDate": "soon"},
	} {
		assert.Equal(t, http.StatusBadRequest, do(t, http.MethodPatch, path, body, as(userId)).Code, name)
	}
	assert.Equal(t, http.StatusBadRequest, do(t, http.MethodPatch, "/ormi/first", map[string]any{}, as(userId)).Code)
	assert.Equal(t, http.StatusNotFound, do(t, http.MethodPatch, path, map[string]any{"title": "Stolen"}, as(primitive.NewObjectID())).Code)
	assert.Equal(t, http.StatusNotFound, do(t, http.MethodPatch, "/ormi/"+primitive.NewObjectID().Hex(), map[string]any{}, as(userId)).Code)
	assert.Equal(t, "Original", decode[models.Todo](t, do(t, http.MethodGet, path, nil, as(userId))).Title)

	testutil.BreakDB(t)
	assert.Equal(t, http.StatusInternalServerError, do(t, http.MethodPatch, path, map[string]any{"title": "Renamed"}, as(userId)).Code)
}

func TestDeleteTodo(t *testing.T) {
	testutil.Clean(t)
	userId := primitive.NewObjectID()
	todo := postTodo(t, userId, map[string]any{"title": "Temporary"})
	path := "/ormi/" + todo.Id.Hex()

	assert.Equal(t, http.StatusNotFound, do(t, http.MethodDelete, path, nil, as(primitive.NewObjectID())).Code)
	assert.Equal(t, http.StatusBadRequest, do(t, http.MethodDelete, "/ormi/first", nil, as(userId)).Code)
	assert.Equal(t, 1, count(t, "ormi"))

	deleted := do(t, http.MethodDelete, path, nil, as(userId))
	assert.Equal(t, http.StatusNoContent, deleted.Code)
	assert.Empty(t, deleted.Body.String())
	assert.Zero(t, count(t, "ormi"))
	assert.Equal(t, http.StatusNotFound, do(t, http.MethodDelete, path, nil, as(userId)).Code)

	testutil.BreakDB(t)
	assert.Equal(t, http.StatusInternalServerError, do(t, http.MethodDelete, path, nil, as(userId)).Code)
}

func TestPutTodoOrder(t *testing.T) {
	testutil.Clean(t)
	userId := primitive.NewObjectID()
	a := postTodo(t, userId, map[string]any{"title": "Task A"})
	b := postTodo(t, userId, map[string]any{"title": "Task B"})
	c := postTodo(t, userId, map[string]any{"title": "Task C"})

	response := do(t, http.MethodPut, "/ormi/order", map[string]any{"ids": []string{c.Id.Hex(), a.Id.Hex(), b.Id.Hex()}}, as(userId))
	assert.Equal(t, http.StatusNoContent, response.Code, response.Body.String())
	assert.Equal(t, []string{"Task C", "Task A", "Task B"}, todoTitles(listTodos(t, userId, "")))

	assert.Equal(t, http.StatusNoContent, do(t, http.MethodPut, "/ormi/order", map[string]any{"ids": []string{b.Id.Hex()}}, as(primitive.NewObjectID())).Code)
	assert.Equal(t, []string{"Task C", "Task A", "Task B"}, todoTitles(listTodos(t, userId, "")), "someone else cannot reorder them")

	tooMany := make([]string, 1001)
	for i := range tooMany {
		tooMany[i] = a.Id.Hex()
	}
	for name, body := range map[string]any{
		"not json":   "not json",
		"no ids":     map[string]any{},
		"empty":      map[string]any{"ids": []string{}},
		"too many":   map[string]any{"ids": tooMany},
		"short id":   map[string]any{"ids": []string{"abc"}},
		"not hex":    map[string]any{"ids": []string{repeat("z", 24)}},
		"0x prefix":  map[string]any{"ids": []string{"0x" + repeat("a", 22)}},
		"not a list": map[string]any{"ids": a.Id.Hex()},
	} {
		assert.Equal(t, http.StatusBadRequest, do(t, http.MethodPut, "/ormi/order", body, as(userId)).Code, name)
	}

	testutil.BreakDB(t)
	assert.Equal(t, http.StatusInternalServerError, do(t, http.MethodPut, "/ormi/order", map[string]any{"ids": []string{a.Id.Hex()}}, as(userId)).Code)
}

func TestGetTodoStats(t *testing.T) {
	testutil.Clean(t)
	userId := primitive.NewObjectID()
	first := postTodo(t, userId, map[string]any{"title": "First"})
	postTodo(t, userId, map[string]any{"title": "Second"})
	third := postTodo(t, userId, map[string]any{"title": "Third"})
	patchTodo(t, userId, first.Id, map[string]any{"state": "DONE"})
	patchTodo(t, userId, third.Id, map[string]any{"state": "STANDBY"})

	response := do(t, http.MethodGet, "/ormi/stats", nil, as(userId))
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	stats := decode[models.TodoStats](t, response)
	today := time.Now().UTC().Format("2006-01-02")
	assert.Equal(t, []models.TodoDayCount{{Date: today, Count: 1}}, stats.Days)
	assert.Equal(t, time.Now().UTC().Year(), stats.FirstYear)
	assert.Equal(t, 1, stats.CurrentStreak)
	assert.Equal(t, 1, stats.LongestStreak)
	assert.Equal(t, 1, stats.Todo)
	assert.Equal(t, 0, stats.InProgress)
	assert.Equal(t, 1, stats.Standby)
	assert.Equal(t, 0, stats.Overdue)

	inAuckland := do(t, http.MethodGet, "/ormi/stats?tz=Pacific/Auckland", nil, as(userId))
	require.Equal(t, http.StatusOK, inAuckland.Code)
	auckland, err := time.LoadLocation("Pacific/Auckland")
	require.NoError(t, err)
	assert.Equal(t, time.Now().In(auckland).Format("2006-01-02"), decode[models.TodoStats](t, inAuckland).Days[0].Date)

	nothing := do(t, http.MethodGet, "/ormi/stats", nil, as(primitive.NewObjectID()))
	assert.JSONEq(t, `{"days": [], "firstYear": `+strconv.Itoa(time.Now().UTC().Year())+`, "currentStreak": 0, "longestStreak": 0,
		"todo": 0, "inProgress": 0, "standby": 0, "overdue": 0}`, nothing.Body.String())

	invalid := do(t, http.MethodGet, "/ormi/stats?tz=Mars/Olympus", nil, as(userId))
	assert.Equal(t, http.StatusBadRequest, invalid.Code)
	assert.Equal(t, "Invalid timezone", errorOf(t, invalid))

	testutil.BreakDB(t)
	assert.Equal(t, http.StatusInternalServerError, do(t, http.MethodGet, "/ormi/stats", nil, as(userId)).Code)
}

func TestGetTodoYearStats(t *testing.T) {
	testutil.Clean(t)
	userId := primitive.NewObjectID()
	at := func(value string) primitive.DateTime {
		parsed, err := time.Parse(time.RFC3339, value)
		require.NoError(t, err)
		return primitive.NewDateTimeFromTime(parsed)
	}
	for _, completion := range []string{"2024-12-31T23:30:00Z", "2025-01-01T08:00:00Z", "2025-06-15T08:00:00Z", "2025-06-15T09:00:00Z", "2026-01-01T00:00:00Z"} {
		_, err := models.CreateTodo(ctx, &models.Todo{User: userId, State: models.TodoStateDone,
			History: []models.OrmiEvent{{NewState: models.TodoStateDone, UpdatedAt: at(completion)}}})
		require.NoError(t, err)
	}

	response := do(t, http.MethodGet, "/ormi/stats/2025", nil, as(userId))
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	assert.JSONEq(t, `[{"date": "2025-01-01", "count": 1}, {"date": "2025-06-15", "count": 2}]`, response.Body.String())

	inParis := do(t, http.MethodGet, "/ormi/stats/2025?tz=Europe/Paris", nil, as(userId))
	assert.JSONEq(t, `[{"date": "2025-01-01", "count": 2}, {"date": "2025-06-15", "count": 2}]`, inParis.Body.String(),
		"the year starts at midnight in the given timezone")
	assert.JSONEq(t, `[]`, do(t, http.MethodGet, "/ormi/stats/2023", nil, as(userId)).Body.String())

	for _, year := range []string{"last", "1969", "10000", "-5"} {
		invalid := do(t, http.MethodGet, "/ormi/stats/"+year, nil, as(userId))
		assert.Equal(t, http.StatusBadRequest, invalid.Code, year)
		assert.Equal(t, "Invalid year", errorOf(t, invalid))
	}
	assert.Equal(t, http.StatusBadRequest, do(t, http.MethodGet, "/ormi/stats/2025?tz=Mars/Olympus", nil, as(userId)).Code)

	testutil.BreakDB(t)
	assert.Equal(t, http.StatusInternalServerError, do(t, http.MethodGet, "/ormi/stats/2025", nil, as(userId)).Code)
}

func TestGetTodoLabels(t *testing.T) {
	testutil.Clean(t)
	userId := primitive.NewObjectID()
	postTodo(t, userId, map[string]any{"title": "One", "labels": []string{"home", "garden"}})
	postTodo(t, userId, map[string]any{"title": "Two", "labels": []string{"home"}})
	postTodo(t, primitive.NewObjectID(), map[string]any{"title": "Other", "labels": []string{"secret"}})
	for i := range 12 {
		postTodo(t, userId, map[string]any{"title": "Many", "labels": []string{"label-" + strconv.Itoa(i)}})
	}

	all := decode[[]string](t, do(t, http.MethodGet, "/ormi/labels", nil, as(userId)))
	assert.Len(t, all, 10, "at most ten suggestions")
	assert.Equal(t, []string{"home", "garden"}, all[:2])
	assert.Equal(t, []string{"garden"}, decode[[]string](t, do(t, http.MethodGet, "/ormi/labels?q=GARD", nil, as(userId))))
	assert.JSONEq(t, `[]`, do(t, http.MethodGet, "/ormi/labels?q=secret", nil, as(userId)).Body.String())
	assert.Equal(t, http.StatusBadRequest, do(t, http.MethodGet, "/ormi/labels?q="+repeat("a", 51), nil, as(userId)).Code)

	testutil.BreakDB(t)
	assert.Equal(t, http.StatusInternalServerError, do(t, http.MethodGet, "/ormi/labels", nil, as(userId)).Code)
}
