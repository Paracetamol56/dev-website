package slides

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyChangesMatchesCodeMirror(t *testing.T) {
	content, err := os.ReadFile("testdata/changesets.json")
	require.NoError(t, err)
	var cases []struct {
		Doc     string          `json:"doc"`
		Changes json.RawMessage `json:"changes"`
		Result  string          `json:"result"`
	}
	require.NoError(t, json.Unmarshal(content, &cases))
	require.NotEmpty(t, cases)

	for i, test := range cases {
		result, err := applyChanges(encodeText(test.Doc), test.Changes)
		require.NoError(t, err, "case %d: %s", i, test.Changes)
		assert.Equal(t, test.Result, decodeText(result), "case %d: %q with %s", i, test.Doc, test.Changes)
	}
}

func TestApplyChangesRefusesInvalidChanges(t *testing.T) {
	doc := encodeText("hello😀")
	for name, changes := range map[string]string{
		"not json":           `[`,
		"not a list":         `{"a": 1}`,
		"too short":          `[3]`,
		"too long":           `[8]`,
		"past the end":       `[5, [3]]`,
		"negative keep":      `[-1, 8]`,
		"negative delete":    `[[-1], 7]`,
		"empty change":       `[[], 7]`,
		"text as length":     `[["a"], 7]`,
		"number as text":     `[[0, 1], 7]`,
		"keep then too long": `[7, [0, "x"], 1]`,
	} {
		t.Run(name, func(t *testing.T) {
			_, err := applyChanges(doc, json.RawMessage(changes))
			assert.ErrorIs(t, err, ErrInvalidChanges)
		})
	}

	result, err := applyChanges(doc, json.RawMessage(`[5, [2, "!", ""]]`))
	require.NoError(t, err)
	assert.Equal(t, "hello!\n", decodeText(result), "an emoji is two code units")
}
