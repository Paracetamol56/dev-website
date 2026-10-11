package ormi

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestIsPastDueDate(t *testing.T) {
	at := func(offset time.Duration) primitive.DateTime {
		return primitive.NewDateTimeFromTime(time.Now().Add(offset))
	}
	assert.False(t, isPastDueDate(at(time.Hour)))
	assert.False(t, isPastDueDate(at(-23*time.Hour)), "clients send midnight in their own timezone")
	assert.True(t, isPastDueDate(at(-25*time.Hour)))
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
