package icon

import (
	"dev/internal/testutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIcons(t *testing.T) {
	testutil.Clean(t)
	require.NoError(t, ReplaceIcons(ctx, "seed", []Icon{
		{Id: "lucide:zap", Name: "zap", Title: "Zap", Source: "lucide", Tags: []string{"flash"}},
		{Id: "lucide:zap-off", Name: "zap-off", Title: "Zap Off", Source: "lucide", Tags: []string{}},
		{Id: "simpleicons:zapier", Name: "zapier", Title: "Zapier", Source: "simpleicons", Tags: []string{}, Hex: "FF4F00"},
	}))

	all := do(t, http.MethodGet, "/icons", nil)
	assert.Equal(t, http.StatusOK, all.Code)
	assert.Equal(t, "public, max-age=3600", all.Header().Get("Cache-Control"))
	assert.Len(t, testutil.Decode[[]Icon](t, all), 3)
	assert.Len(t, testutil.Decode[[]Icon](t, do(t, http.MethodGet, "/icons?source=simpleicons", nil)), 1)
	assert.Equal(t, http.StatusBadRequest, do(t, http.MethodGet, "/icons?source=fontawesome", nil).Code)

	found := do(t, http.MethodGet, "/icons/search?q=zap&source=lucide&source=simpleicons&limit=2", nil)
	assert.Equal(t, http.StatusOK, found.Code)
	assert.Equal(t, "public, max-age=3600", found.Header().Get("Cache-Control"))
	result := testutil.Decode[IconSearchResult](t, found)
	assert.Equal(t, 3, result.Total)
	require.Len(t, result.Items, 2)
	assert.Equal(t, "lucide:zap", result.Items[0].Id)

	assert.Equal(t, 3, testutil.Decode[IconSearchResult](t, do(t, http.MethodGet, "/icons/search", nil)).Total, "defaults")
	assert.Equal(t, 1, testutil.Decode[IconSearchResult](t, do(t, http.MethodGet, "/icons/search?tag=FLASH", nil)).Total)
	for _, query := range []string{"limit=0", "limit=501", "offset=-1", "source=fontawesome", "q=" + strings.Repeat("a", 101), "limit=many"} {
		assert.Equal(t, http.StatusBadRequest, do(t, http.MethodGet, "/icons/search?"+query, nil).Code, query)
	}

	testutil.BreakDB(t)
	assert.Equal(t, http.StatusInternalServerError, do(t, http.MethodGet, "/icons", nil).Code)
	assert.Equal(t, http.StatusInternalServerError, do(t, http.MethodGet, "/icons/search", nil).Code)
}
