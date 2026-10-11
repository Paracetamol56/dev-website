package models

import (
	"encoding/json"
	"math"
	"testing"
	"time"

	"dev/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestHipparcos(t *testing.T) {
	testutil.Clean(t)
	insertRaw(t, "hipparcos",
		bson.M{"HIP": 1, "Vmag": 9.1, "Plx": 3.54, "B-V": 0.482},
		bson.M{"HIP": 2, "Vmag": 5.0, "Plx": 100.0, "B-V": -0.4},
		bson.M{"HIP": 3, "Vmag": 7.0, "Plx": 10.0, "B-V": 2.5},
		bson.M{"HIP": 4, "Vmag": 7.0, "Plx": 0.05, "B-V": 1.0},
		bson.M{"HIP": 5, "Vmag": 7.0, "Plx": -1.2, "B-V": 1.0},
	)

	stars, err := GetAllHipparcosHR(ginContext())
	require.NoError(t, err)
	require.Len(t, stars, 2, "stars outside the colour range or without a usable parallax are left out")
	byHip := map[int]HipparcosHR{}
	for _, star := range stars {
		byHip[star.HIP] = star
	}
	assert.InDelta(t, 9.1-5*math.Log10(1000/3.54/10), byHip[1].Amag, 1e-9)
	assert.Equal(t, 0.482, byHip[1].BV)
	assert.InDelta(t, 5.0, byHip[2].Amag, 1e-9, "at 10 parsecs the absolute magnitude is the apparent one")

	one, err := GetHipparcosHRByHIP(ginContext(), 2)
	require.NoError(t, err)
	require.Len(t, one, 1)
	assert.Equal(t, 2, one[0].HIP)

	for _, hip := range []int{3, 99} {
		none, err := GetHipparcosHRByHIP(ginContext(), hip)
		require.NoError(t, err)
		assert.Nil(t, none, hip)
	}

	testutil.BreakDB(t)
	_, err = GetAllHipparcosHR(ginContext())
	assert.Error(t, err)
	_, err = GetHipparcosHRByHIP(ginContext(), 1)
	assert.Error(t, err)
}

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

func TestMicroprocessorDetails(t *testing.T) {
	chip := Microprocessor{}
	chip.SetDetail("name", "ignored")
	chip.SetDetail("source_url", "ignored")
	assert.Nil(t, chip.Details, "documented fields cannot be shadowed")

	chip.SetDetail("processor", "Intel 4004")
	chip.SetDetail("process", map[string]any{"value": 10000.0, "unit": "nm"})
	assert.Equal(t, map[string]any{"processor": "Intel 4004", "process": map[string]any{"value": 10000.0, "unit": "nm"}}, chip.Details)
}

func TestMicroprocessorJSON(t *testing.T) {
	release := primitive.NewDateTimeFromTime(time.Date(1971, 1, 1, 0, 0, 0, 0, time.UTC))
	encoded, err := json.Marshal(Microprocessor{
		Id: "wikipedia:cpu-intel-4004-1971", Name: "Intel 4004", Type: "CPU", Release: &release, GateSize: 10000, Transistors: 0.00225,
		Details: map[string]any{
			"processor": "Intel 4004",
			"process":   primitive.D{{Key: "value", Value: 10000}, {Key: "unit", Value: "nm"}},
			"name":      "must not replace the field",
		},
	})
	require.NoError(t, err)
	assert.JSONEq(t, `{
		"id": "wikipedia:cpu-intel-4004-1971",
		"name": "Intel 4004",
		"type": "CPU",
		"release": "1971-01-01T00:00:00Z",
		"gateSize": 10000,
		"transistors": 0.00225,
		"processor": "Intel 4004",
		"process": {"value": 10000, "unit": "nm"}
	}`, string(encoded))

	encoded, err = json.Marshal([]Microprocessor{{Id: "x", Name: "Bare", Type: "GPU"}})
	require.NoError(t, err)
	assert.JSONEq(t, `[{"id": "x", "name": "Bare", "type": "GPU"}]`, string(encoded), "fields without a value are omitted")

	_, err = json.Marshal(Microprocessor{Details: map[string]any{"bad": make(chan int)}})
	assert.Error(t, err)
	_, err = json.Marshal(Microprocessor{GateSize: math.NaN()})
	assert.Error(t, err)
}

func TestMicroprocessors(t *testing.T) {
	testutil.Clean(t)
	release := primitive.NewDateTimeFromTime(time.Date(1999, 1, 1, 0, 0, 0, 0, time.UTC))
	chips := []Microprocessor{
		{Id: "wikipedia:cpu-a", Name: "A", Type: "CPU", Vendor: "Intel", Source: "wikipedia"},
		{Id: "wikipedia:gpu-b", Name: "B", Type: "GPU", Vendor: "NVIDIA", Source: "wikipedia", Release: &release, Density: 122000,
			Details: map[string]any{"fab": "TSMC", "process": map[string]any{"value": 220.0, "unit": "nm"}}},
		{Id: "wikipedia:gpu-c", Name: "C", Type: "GPU", Vendor: "Intel", Source: "wikipedia"},
	}
	require.NoError(t, ReplaceMicroprocessors(ctx, "wikipedia", chips))
	legacyId := primitive.NewObjectID()
	insertRaw(t, "microprocessors", bson.M{"_id": legacyId, "name": "Legacy", "type": "CPU", "vendor": "AMD", "freq": 3200.0, "tdp": 65.0})

	names := func(filter MicroprocessorFilter) []string {
		found, err := GetAllMicroprocessors(ginContext(), &filter)
		require.NoError(t, err)
		result := []string{}
		for _, chip := range found {
			result = append(result, chip.Name)
		}
		return result
	}
	assert.ElementsMatch(t, []string{"A", "B", "C", "Legacy"}, names(MicroprocessorFilter{}))
	assert.ElementsMatch(t, []string{"B", "C"}, names(MicroprocessorFilter{Type: "GPU"}))
	assert.ElementsMatch(t, []string{"A", "C"}, names(MicroprocessorFilter{Vendor: "Intel"}))
	assert.ElementsMatch(t, []string{"C"}, names(MicroprocessorFilter{Type: "GPU", Vendor: "Intel"}))
	assert.Equal(t, []string{}, names(MicroprocessorFilter{Type: "FPGA"}))

	chip, err := GetMicroprocessorById(ginContext(), "wikipedia:gpu-b")
	require.NoError(t, err)
	assert.Equal(t, "B", chip.Name)
	assert.Equal(t, release, *chip.Release)
	assert.Equal(t, "TSMC", chip.Details["fab"], "source columns are stored next to the fields")
	encoded, err := json.Marshal(chip)
	require.NoError(t, err)
	assert.JSONEq(t, `{"id": "wikipedia:gpu-b", "name": "B", "type": "GPU", "vendor": "NVIDIA", "source": "wikipedia",
		"release": "1999-01-01T00:00:00Z", "density": 122000, "fab": "TSMC", "process": {"value": 220, "unit": "nm"}}`, string(encoded))

	legacy, err := GetMicroprocessorById(ginContext(), legacyId.Hex())
	require.NoError(t, err, "chips stored before the import have an ObjectID")
	assert.Equal(t, "Legacy", legacy.Name)
	assert.Equal(t, 3200.0, legacy.Frequency)
	_, err = GetMicroprocessorById(ginContext(), "unknown")
	assert.ErrorIs(t, err, mongo.ErrNoDocuments)

	total, err := CountMicroprocessors(ctx, "")
	require.NoError(t, err)
	assert.EqualValues(t, 4, total)
	imported, err := CountMicroprocessors(ctx, "wikipedia")
	require.NoError(t, err)
	assert.EqualValues(t, 3, imported)

	require.NoError(t, ReplaceMicroprocessors(ctx, "wikipedia", chips[:1]))
	assert.ElementsMatch(t, []string{"A", "Legacy"}, names(MicroprocessorFilter{}), "only chips of the source are replaced")
	require.NoError(t, ReplaceMicroprocessors(ctx, "wikipedia", nil))
	assert.Equal(t, []string{"Legacy"}, names(MicroprocessorFilter{}))

	testutil.BreakDB(t)
	_, err = GetAllMicroprocessors(ginContext(), &MicroprocessorFilter{})
	assert.Error(t, err)
	_, err = GetMicroprocessorById(ginContext(), "wikipedia:cpu-a")
	assert.Error(t, err)
	_, err = CountMicroprocessors(ctx, "")
	assert.Error(t, err)
	assert.Error(t, ReplaceMicroprocessors(ctx, "wikipedia", chips))
	assert.Error(t, ReplaceMicroprocessors(ctx, "wikipedia", nil))
}
