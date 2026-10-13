//go:build e2e

package e2e

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type todoPage struct {
	Total int              `json:"total"`
	Items []map[string]any `json:"items"`
}

func (browser *browser) todos(query string) (int, []string) {
	browser.t.Helper()
	var page todoPage
	browser.get("/ormi" + query).expect(http.StatusOK).into(&page)
	titles := []string{}
	for _, todo := range page.Items {
		titles = append(titles, todo["title"].(string))
	}
	return page.Total, titles
}

func (browser *browser) addTodo(body map[string]any) string {
	browser.t.Helper()
	return browser.post("/ormi", body).expect(http.StatusCreated).object()["id"].(string)
}

func TestOrmi(t *testing.T) {
	app := start(t)
	ada := app.visitor()
	app.visitor().get("/ormi").expect(http.StatusUnauthorized)
	ada.login(app.address("ada"))

	total, titles := ada.todos("")
	assert.Zero(t, total)
	assert.Empty(t, titles)

	nextWeek := time.Now().Add(7 * 24 * time.Hour).UTC().Format(time.RFC3339)
	tomorrow := time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339)
	tests := ada.addTodo(map[string]any{"title": "Write tests", "labels": []string{"api", "quality"}, "dueDate": nextWeek})
	refactor := ada.addTodo(map[string]any{"title": "Refactor the API", "labels": []string{"api"}, "dueDate": tomorrow})
	release := ada.addTodo(map[string]any{"title": "Release", "description": "Tag and deploy"})
	ada.post("/ormi", map[string]any{"title": "x"}).expect(http.StatusBadRequest)
	ada.post("/ormi", map[string]any{"title": "Too late", "dueDate": "2020-01-01T00:00:00Z"}).expect(http.StatusBadRequest)

	total, titles = ada.todos("")
	assert.Equal(t, 3, total)
	assert.Equal(t, []string{"Write tests", "Refactor the API", "Release"}, titles)

	ada.call(http.MethodPut, "/ormi/order", map[string]any{"ids": []string{release, tests, refactor}}).expect(http.StatusNoContent)
	_, titles = ada.todos("")
	assert.Equal(t, []string{"Release", "Write tests", "Refactor the API"}, titles)
	_, titles = ada.todos("?sort=due")
	assert.Equal(t, []string{"Refactor the API", "Write tests", "Release"}, titles)

	started := ada.patch("/ormi/"+tests, map[string]any{"state": "IN_PROGRESS", "description": "Unit and end to end"}).expect(http.StatusOK).object()
	assert.Equal(t, "IN_PROGRESS", started["state"])
	done := ada.patch("/ormi/"+tests, map[string]any{"state": "DONE"}).expect(http.StatusOK).object()
	assert.Len(t, done["history"], 3)
	ada.patch("/ormi/"+release, map[string]any{"state": "STANDBY"}).expect(http.StatusOK)
	cleared := ada.patch("/ormi/"+refactor, `{"dueDate": null}`).expect(http.StatusOK).object()
	assert.NotContains(t, cleared, "dueDate")
	ada.patch("/ormi/"+refactor, map[string]any{"state": "ARCHIVED"}).expect(http.StatusBadRequest)

	total, titles = ada.todos("?state=TODO&state=IN_PROGRESS&state=STANDBY")
	assert.Equal(t, 2, total, "the default view of the app: open todos")
	assert.Equal(t, []string{"Release", "Refactor the API"}, titles)
	_, titles = ada.todos("?state=DONE")
	assert.Equal(t, []string{"Write tests"}, titles)
	_, titles = ada.todos("?label=api&label=quality")
	assert.Equal(t, []string{"Write tests"}, titles)
	total, titles = ada.todos("?limit=1&offset=1")
	assert.Equal(t, 3, total)
	assert.Equal(t, []string{"Write tests"}, titles)
	_, titles = ada.todos("?sort=updated")
	assert.Equal(t, "Release", titles[0])
	ada.get("/ormi?limit=500").expect(http.StatusBadRequest)

	var stats struct {
		Days []struct {
			Date  string `json:"date"`
			Count int    `json:"count"`
		} `json:"days"`
		CurrentStreak int `json:"currentStreak"`
		LongestStreak int `json:"longestStreak"`
		Todo          int `json:"todo"`
		InProgress    int `json:"inProgress"`
		Standby       int `json:"standby"`
		Overdue       int `json:"overdue"`
	}
	ada.get("/ormi/stats?tz=UTC").expect(http.StatusOK).into(&stats)
	today := time.Now().UTC().Format("2006-01-02")
	require.Len(t, stats.Days, 1)
	assert.Equal(t, today, stats.Days[0].Date)
	assert.Equal(t, 1, stats.Days[0].Count)
	assert.Equal(t, 1, stats.CurrentStreak)
	assert.Equal(t, 1, stats.Todo)
	assert.Equal(t, 0, stats.InProgress)
	assert.Equal(t, 1, stats.Standby)
	year := ada.get("/ormi/stats/" + today[:4]).expect(http.StatusOK).list()
	require.Len(t, year, 1)
	assert.Equal(t, today, year[0]["date"])
	ada.get("/ormi/stats?tz=Nowhere").expect(http.StatusBadRequest)

	var labels []string
	ada.get("/ormi/labels").expect(http.StatusOK).into(&labels)
	assert.Equal(t, []string{"api", "quality"}, labels)
	ada.get("/ormi/labels?q=QUAL").expect(http.StatusOK).into(&labels)
	assert.Equal(t, []string{"quality"}, labels)

	grace := app.visitor()
	grace.login(app.address("grace"))
	total, _ = grace.todos("")
	assert.Zero(t, total, "todos are private")
	grace.get("/ormi/" + tests).expect(http.StatusNotFound)
	grace.patch("/ormi/"+tests, map[string]any{"title": "Stolen"}).expect(http.StatusNotFound)
	grace.delete("/ormi/" + tests).expect(http.StatusNotFound)
	grace.call(http.MethodPut, "/ormi/order", map[string]any{"ids": []string{refactor, release}}).expect(http.StatusNoContent)
	grace.get("/ormi/labels").expect(http.StatusOK).into(&labels)
	assert.Empty(t, labels)
	_, titles = ada.todos("")
	assert.Equal(t, []string{"Release", "Write tests", "Refactor the API"}, titles, "and cannot be touched by someone else")
	assert.Equal(t, "Write tests", ada.get("/ormi/" + tests).expect(http.StatusOK).object()["title"])

	ada.delete("/ormi/" + release).expect(http.StatusNoContent)
	ada.delete("/ormi/" + release).expect(http.StatusNotFound)
	ada.get("/ormi/" + release).expect(http.StatusNotFound)
	total, _ = ada.todos("")
	assert.Equal(t, 2, total)
}
