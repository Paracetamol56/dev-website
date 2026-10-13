package hipparcos

import (
	"dev/internal/testutil"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

func TestHipparcos(t *testing.T) {
	testutil.Clean(t)
	testutil.InsertRaw(t, "hipparcos",
		bson.M{"HIP": 1, "Vmag": 9.1, "Plx": 3.54, "B-V": 0.482},
		bson.M{"HIP": 2, "Vmag": 5.0, "Plx": 100.0, "B-V": -0.4},
		bson.M{"HIP": 3, "Vmag": 7.0, "Plx": 10.0, "B-V": 2.5},
		bson.M{"HIP": 4, "Vmag": 7.0, "Plx": 0.05, "B-V": 1.0},
		bson.M{"HIP": 5, "Vmag": 7.0, "Plx": -1.2, "B-V": 1.0},
	)

	stars, err := GetAllHipparcosHR(testutil.Context())
	require.NoError(t, err)
	require.Len(t, stars, 2, "stars outside the colour range or without a usable parallax are left out")
	byHip := map[int]HipparcosHR{}
	for _, star := range stars {
		byHip[star.HIP] = star
	}
	assert.InDelta(t, 9.1-5*math.Log10(1000/3.54/10), byHip[1].Amag, 1e-9)
	assert.Equal(t, 0.482, byHip[1].BV)
	assert.InDelta(t, 5.0, byHip[2].Amag, 1e-9, "at 10 parsecs the absolute magnitude is the apparent one")

	one, err := GetHipparcosHRByHIP(testutil.Context(), 2)
	require.NoError(t, err)
	require.Len(t, one, 1)
	assert.Equal(t, 2, one[0].HIP)

	for _, hip := range []int{3, 99} {
		none, err := GetHipparcosHRByHIP(testutil.Context(), hip)
		require.NoError(t, err)
		assert.Nil(t, none, hip)
	}

	testutil.BreakDB(t)
	_, err = GetAllHipparcosHR(testutil.Context())
	assert.Error(t, err)
	_, err = GetHipparcosHRByHIP(testutil.Context(), 1)
	assert.Error(t, err)
}
