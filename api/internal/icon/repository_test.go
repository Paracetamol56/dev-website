package icon

import (
	"dev/internal/testutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var iconFixtures = []Icon{
	{Id: "lucide:arrow-down", Name: "arrow-down", Title: "Arrow Down", Source: "lucide", Tags: []string{"direction", "south"}},
	{Id: "lucide:arrow-down-up", Name: "arrow-down-up", Title: "Arrow Down Up", Source: "lucide", Tags: []string{"sort"}},
	{Id: "lucide:download", Name: "download", Title: "Download", Source: "lucide", Tags: []string{"arrow", "save"}},
	{Id: "lucide:github", Name: "github", Title: "Github", Source: "lucide", Tags: []string{"logo", "version control"}},
	{Id: "simpleicons:github", Name: "github", Title: "GitHub", Source: "simpleicons", Tags: []string{}, Hex: "181717"},
	{Id: "simpleicons:arrow", Name: "arrow", Title: "Arrow", Source: "simpleicons", Tags: []string{"Electronics"}, Hex: "000000"},
	{Id: "simpleicons:cplusplus", Name: "cplusplus", Title: "C++", Source: "simpleicons", Tags: []string{"cpp"}, Hex: "00599C"},
}

func seedIcons(t *testing.T) {
	t.Helper()
	testutil.Clean(t)
	require.NoError(t, ReplaceIcons(ctx, "seed", iconFixtures))
}

func iconIds(icons []Icon) []string {
	ids := []string{}
	for _, icon := range icons {
		ids = append(ids, icon.Id)
	}
	return ids
}

func TestGetAllIcons(t *testing.T) {
	seedIcons(t)

	icons, err := GetAllIcons(ctx, &IconFilter{})
	require.NoError(t, err)
	assert.Equal(t, []string{
		"lucide:arrow-down", "lucide:arrow-down-up", "lucide:download", "lucide:github",
		"simpleicons:arrow", "simpleicons:cplusplus", "simpleicons:github",
	}, iconIds(icons), "sorted by source then name")
	assert.Equal(t, iconFixtures[0], icons[0])

	icons, err = GetAllIcons(ctx, &IconFilter{Source: "simpleicons"})
	require.NoError(t, err)
	assert.Len(t, icons, 3)

	count, err := CountIcons(ctx)
	require.NoError(t, err)
	assert.EqualValues(t, 7, count)

	testutil.Clean(t)
	icons, err = GetAllIcons(ctx, &IconFilter{})
	require.NoError(t, err)
	assert.Equal(t, []Icon{}, icons, "an empty list rather than null")
}

func TestReplaceIcons(t *testing.T) {
	seedIcons(t)

	require.NoError(t, ReplaceIcons(ctx, "lucide", []Icon{
		{Id: "lucide:github", Name: "github", Title: "Renamed", Source: "lucide", Tags: []string{}},
		{Id: "lucide:zap", Name: "zap", Title: "Zap", Source: "lucide", Tags: []string{}},
	}))
	icons, err := GetAllIcons(ctx, &IconFilter{})
	require.NoError(t, err)
	assert.Equal(t, []string{"lucide:github", "lucide:zap", "simpleicons:arrow", "simpleicons:cplusplus", "simpleicons:github"}, iconIds(icons),
		"stale icons of the source are removed, other sources untouched")
	assert.Equal(t, "Renamed", icons[0].Title)

	require.NoError(t, ReplaceIcons(ctx, "lucide", nil))
	icons, err = GetAllIcons(ctx, &IconFilter{Source: "lucide"})
	require.NoError(t, err)
	assert.Empty(t, icons)
}

func TestSearchIcons(t *testing.T) {
	seedIcons(t)
	search := func(search IconSearch) *IconSearchResult {
		if search.Limit == 0 {
			search.Limit = 50
		}
		result, err := SearchIcons(ctx, &search)
		require.NoError(t, err)
		return result
	}

	everything := search(IconSearch{})
	assert.Equal(t, 7, everything.Total)
	assert.Equal(t, []string{
		"simpleicons:arrow", "lucide:arrow-down", "lucide:arrow-down-up", "simpleicons:cplusplus",
		"lucide:download", "lucide:github", "simpleicons:github",
	}, iconIds(everything.Items), "without a query, by name then source")

	arrows := search(IconSearch{Query: "ARROW"})
	assert.Equal(t, []string{"simpleicons:arrow", "lucide:arrow-down", "lucide:arrow-down-up", "lucide:download"}, iconIds(arrows.Items),
		"exact name, then prefixes, then tags")

	assert.Equal(t, []string{"lucide:arrow-down", "lucide:arrow-down-up", "lucide:download"}, iconIds(search(IconSearch{Query: "arrow down"}).Items),
		"the exact phrase first, a match spread over name and tags last")
	assert.Equal(t, []string{"lucide:arrow-down-up"}, iconIds(search(IconSearch{Query: "down arrow up"}).Items), "every word must match")
	assert.Equal(t, []string{"lucide:github"}, iconIds(search(IconSearch{Query: "version"}).Items), "tags are searched")
	assert.Equal(t, []string{"simpleicons:cplusplus"}, iconIds(search(IconSearch{Query: "c++"}).Items), "the query is not a pattern")
	assert.Empty(t, search(IconSearch{Query: ".*"}).Items)

	assert.Equal(t, []string{"simpleicons:github"}, iconIds(search(IconSearch{Query: "github", Sources: []string{"simpleicons"}}).Items))
	assert.Len(t, search(IconSearch{Query: "github", Sources: []string{"simpleicons", "lucide"}}).Items, 2)
	assert.Equal(t, []string{"simpleicons:arrow"}, iconIds(search(IconSearch{Tag: "electronics"}).Items), "tags match without case")
	assert.Empty(t, search(IconSearch{Tag: "electro"}).Items, "but in full")

	page := search(IconSearch{Limit: 2, Offset: 5})
	assert.Equal(t, 7, page.Total)
	assert.Equal(t, []string{"lucide:github", "simpleicons:github"}, iconIds(page.Items))
	assert.Equal(t, []Icon{}, search(IconSearch{Offset: 50}).Items)
	assert.Equal(t, 0, search(IconSearch{Query: "nothing"}).Total)
}

func TestIconDatabaseFailures(t *testing.T) {
	testutil.BreakDB(t)
	_, err := GetAllIcons(ctx, &IconFilter{})
	assert.Error(t, err)
	_, err = CountIcons(ctx)
	assert.Error(t, err)
	assert.Error(t, ReplaceIcons(ctx, "lucide", iconFixtures))
	assert.Error(t, ReplaceIcons(ctx, "lucide", nil))
	_, err = SearchIcons(ctx, &IconSearch{Limit: 10})
	assert.Error(t, err)
}
