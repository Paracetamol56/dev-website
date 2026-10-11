package hipparcos

import (
	"dev/internal/db"
	"dev/internal/testutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

func TestHipparcosEndpoints(t *testing.T) {
	testutil.Clean(t)
	_, err := db.GetDB().Collection("hipparcos").InsertMany(ctx, []any{
		bson.M{"HIP": 1, "Vmag": 5.0, "Plx": 100.0, "B-V": 0.5},
		bson.M{"HIP": 2, "Vmag": 6.0, "Plx": 100.0, "B-V": 1.5},
		bson.M{"HIP": 3, "Vmag": 6.0, "Plx": 100.0, "B-V": 9.0},
	})
	require.NoError(t, err)

	all := do(t, http.MethodGet, "/hipparcos", nil)
	assert.Equal(t, http.StatusOK, all.Code)
	assert.Len(t, testutil.Decode[[]HipparcosHR](t, all), 2)

	one := do(t, http.MethodGet, "/hipparcos/2", nil)
	assert.Equal(t, http.StatusOK, one.Code)
	assert.JSONEq(t, `{"hip": 2, "Amag": 6, "BV": 1.5}`, one.Body.String())

	assert.Equal(t, http.StatusNotFound, do(t, http.MethodGet, "/hipparcos/3", nil).Code, "outside the diagram")
	assert.Equal(t, http.StatusNotFound, do(t, http.MethodGet, "/hipparcos/99", nil).Code)
	assert.Equal(t, http.StatusBadRequest, do(t, http.MethodGet, "/hipparcos/sirius", nil).Code)

	testutil.BreakDB(t)
	assert.Equal(t, http.StatusInternalServerError, do(t, http.MethodGet, "/hipparcos", nil).Code)
	assert.Equal(t, http.StatusInternalServerError, do(t, http.MethodGet, "/hipparcos/1", nil).Code)
}
