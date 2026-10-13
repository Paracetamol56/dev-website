package chip

import (
	"encoding/json"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

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
