package chip

import (
	"dev/internal/testutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestChips(t *testing.T) {
	testutil.Clean(t)
	require.NoError(t, ReplaceMicroprocessors(ctx, "wikipedia", []Microprocessor{
		{Id: "wikipedia:cpu-intel-4004-1971", Name: "Intel 4004", Type: "CPU", Vendor: "Intel", Transistors: 0.00225,
			Details: map[string]any{"process": map[string]any{"value": 10000.0, "unit": "nm"}}},
		{Id: "wikipedia:gpu-geforce-256-1999", Name: "GeForce 256", Type: "GPU", Vendor: "NVIDIA"},
	}))

	assert.Len(t, testutil.Decode[[]map[string]any](t, do(t, http.MethodGet, "/chips", nil)), 2)
	gpus := testutil.Decode[[]map[string]any](t, do(t, http.MethodGet, "/chips?type=GPU", nil))
	require.Len(t, gpus, 1)
	assert.Equal(t, "GeForce 256", gpus[0]["name"])
	assert.Len(t, testutil.Decode[[]map[string]any](t, do(t, http.MethodGet, "/chips?vendor=Intel", nil)), 1)
	assert.JSONEq(t, `[]`, do(t, http.MethodGet, "/chips?vendor=Intel&type=GPU", nil).Body.String())

	one := do(t, http.MethodGet, "/chips/wikipedia:cpu-intel-4004-1971", nil)
	assert.Equal(t, http.StatusOK, one.Code)
	assert.JSONEq(t, `{"id": "wikipedia:cpu-intel-4004-1971", "name": "Intel 4004", "type": "CPU", "vendor": "Intel",
		"transistors": 0.00225, "process": {"value": 10000, "unit": "nm"}}`, one.Body.String())

	missing := do(t, http.MethodGet, "/chips/"+primitive.NewObjectID().Hex(), nil)
	assert.Equal(t, http.StatusNotFound, missing.Code)
	assert.Equal(t, "Chip not found", testutil.ErrorOf(t, missing))

	testutil.BreakDB(t)
	assert.Equal(t, http.StatusInternalServerError, do(t, http.MethodGet, "/chips", nil).Code)
	assert.Equal(t, http.StatusInternalServerError, do(t, http.MethodGet, "/chips/x", nil).Code)
}
