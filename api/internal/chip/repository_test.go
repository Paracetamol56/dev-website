package chip

import (
	"dev/internal/testutil"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

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
	testutil.InsertRaw(t, "microprocessors", bson.M{"_id": legacyId, "name": "Legacy", "type": "CPU", "vendor": "AMD", "freq": 3200.0, "tdp": 65.0})

	names := func(filter MicroprocessorFilter) []string {
		found, err := GetAllMicroprocessors(testutil.Context(), &filter)
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

	chip, err := GetMicroprocessorById(testutil.Context(), "wikipedia:gpu-b")
	require.NoError(t, err)
	assert.Equal(t, "B", chip.Name)
	assert.Equal(t, release, *chip.Release)
	assert.Equal(t, "TSMC", chip.Details["fab"], "source columns are stored next to the fields")
	encoded, err := json.Marshal(chip)
	require.NoError(t, err)
	assert.JSONEq(t, `{"id": "wikipedia:gpu-b", "name": "B", "type": "GPU", "vendor": "NVIDIA", "source": "wikipedia",
		"release": "1999-01-01T00:00:00Z", "density": 122000, "fab": "TSMC", "process": {"value": 220, "unit": "nm"}}`, string(encoded))

	legacy, err := GetMicroprocessorById(testutil.Context(), legacyId.Hex())
	require.NoError(t, err, "chips stored before the import have an ObjectID")
	assert.Equal(t, "Legacy", legacy.Name)
	assert.Equal(t, 3200.0, legacy.Frequency)
	_, err = GetMicroprocessorById(testutil.Context(), "unknown")
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
	_, err = GetAllMicroprocessors(testutil.Context(), &MicroprocessorFilter{})
	assert.Error(t, err)
	_, err = GetMicroprocessorById(testutil.Context(), "wikipedia:cpu-a")
	assert.Error(t, err)
	_, err = CountMicroprocessors(ctx, "")
	assert.Error(t, err)
	assert.Error(t, ReplaceMicroprocessors(ctx, "wikipedia", chips))
	assert.Error(t, ReplaceMicroprocessors(ctx, "wikipedia", nil))
}
