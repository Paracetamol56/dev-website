package icon

import (
	"context"
	"dev/internal/testutil"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIconScore(t *testing.T) {
	icon := Icon{Name: "arrow-down", Title: "Arrow Down", Tags: []string{"South", "direction"}}
	assert.Equal(t, 1000, iconScore(icon, []string{"arrow", "down"}))
	assert.Equal(t, 50, iconScore(icon, []string{"arrow"}))
	assert.Equal(t, 20, iconScore(icon, []string{"down"}))
	assert.Equal(t, 10, iconScore(icon, []string{"south"}))
	assert.Equal(t, 5, iconScore(icon, []string{"direct"}))
	assert.Equal(t, 70, iconScore(icon, []string{"down", "arrow"}))
	assert.Equal(t, 0, iconScore(icon, []string{"zap"}))
	assert.Equal(t, 0, iconScore(icon, nil))
	assert.Equal(t, 110, iconScore(Icon{Name: "zap", Title: "Zap", Tags: []string{"zap"}}, []string{"zap"}))
	assert.Equal(t, 1000, iconScore(Icon{Name: "x", Title: "Arrow Down"}, []string{"arrow", "down"}))
}

const (
	lucideTags  = `{"a-arrow-down": ["letter", "text"], "zap": null}`
	simpleIcons = `[
		{"title": "GitHub", "slug": "github", "hex": "181717", "aliases": {"aka": ["gh"]}},
		{"title": "Svelte", "slug": "svelte", "hex": "FF3E00"},
		{"title": "No slug", "hex": "000000"}
	]`
)

func fakeUnpkg(internet *testutil.Internet, lucide string, simple string) {
	internet.Handle("unpkg.com", func(w http.ResponseWriter, r *http.Request) {
		body := map[string]string{
			"/lucide-static@latest/tags.json":             lucide,
			"/simple-icons@latest/data/simple-icons.json": simple,
		}[r.URL.Path]
		if body == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		io.WriteString(w, body)
	})
}

func storedIcons(t *testing.T) []Icon {
	t.Helper()
	icons, err := GetAllIcons(context.Background(), &IconFilter{})
	require.NoError(t, err)
	return icons
}

func TestLucideTitle(t *testing.T) {
	assert.Equal(t, "A Arrow Down", lucideTitle("a-arrow-down"))
	assert.Equal(t, "Zap", lucideTitle("zap"))
	assert.Equal(t, "Trailing ", lucideTitle("trailing-"))
	assert.Equal(t, "", lucideTitle(""))
}

func TestRefreshIcons(t *testing.T) {
	testutil.Clean(t)
	internet := testutil.FakeInternet(t)
	fakeUnpkg(internet, lucideTags, simpleIcons)

	RefreshIcons(context.Background())

	assert.Equal(t, []Icon{
		{Id: "lucide:a-arrow-down", Name: "a-arrow-down", Title: "A Arrow Down", Source: "lucide", Tags: []string{"letter", "text"}},
		{Id: "lucide:zap", Name: "zap", Title: "Zap", Source: "lucide", Tags: []string{}},
		{Id: "simpleicons:github", Name: "github", Title: "GitHub", Source: "simpleicons", Tags: []string{"gh"}, Hex: "181717"},
		{Id: "simpleicons:svelte", Name: "svelte", Title: "Svelte", Source: "simpleicons", Tags: []string{}, Hex: "FF3E00"},
	}, storedIcons(t))
}

func TestRefreshIconsKeepsASourceThatFails(t *testing.T) {
	testutil.Clean(t)
	internet := testutil.FakeInternet(t)
	fakeUnpkg(internet, lucideTags, simpleIcons)
	RefreshIcons(context.Background())

	fakeUnpkg(internet, `{"zap": []}`, ``)
	RefreshIcons(context.Background())
	ids := []string{}
	for _, icon := range storedIcons(t) {
		ids = append(ids, icon.Id)
	}
	assert.Equal(t, []string{"lucide:zap", "simpleicons:github", "simpleicons:svelte"}, ids, "lucide is replaced, simpleicons kept")

	fakeUnpkg(internet, `{}`, `not json`)
	RefreshIcons(context.Background())
	assert.Len(t, storedIcons(t), 3, "an empty or unreadable source is not stored")

	testutil.FakeInternet(t)
	RefreshIcons(context.Background())
	assert.Len(t, storedIcons(t), 3, "network failure")

	fakeUnpkg(testutil.FakeInternet(t), lucideTags, simpleIcons)
	testutil.BreakDB(t)
	RefreshIcons(context.Background())
}

func TestSeedIcons(t *testing.T) {
	testutil.Clean(t)
	internet := testutil.FakeInternet(t)
	fakeUnpkg(internet, lucideTags, simpleIcons)

	SeedIcons(context.Background())
	assert.Len(t, storedIcons(t), 4)
	assert.Len(t, internet.Requests("unpkg.com"), 2)

	SeedIcons(context.Background())
	assert.Len(t, internet.Requests("unpkg.com"), 2, "a filled collection is not seeded again")

	testutil.BreakDB(t)
	SeedIcons(context.Background())
	assert.Len(t, internet.Requests("unpkg.com"), 2)
}

func TestFetchJSONRejectsNilContext(t *testing.T) {
	assert.Error(t, fetchJSON(nil, lucideTagsURL, nil))
}

func TestScheduleIconRefreshSeedsAtStartup(t *testing.T) {
	testutil.Clean(t)
	fakeUnpkg(testutil.FakeInternet(t), lucideTags, simpleIcons)

	ScheduleIconRefresh()

	assert.Eventually(t, func() bool {
		stored, _ := CountIcons(context.Background())
		return stored == 4
	}, 5*time.Second, 20*time.Millisecond)
}
