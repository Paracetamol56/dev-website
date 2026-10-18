package slides

import (
	"encoding/json"
	"errors"
	"strings"
	"unicode/utf16"
)

var ErrInvalidChanges = errors.New("invalid changes")

// Update is one CodeMirror collab update: Changes is a ChangeSet in its JSON form, where a number
// keeps that many characters and [deleted, ...insertedLines] replaces them.
type Update struct {
	ClientId string          `json:"clientID"`
	Changes  json.RawMessage `json:"changes" swaggertype:"array,object"`
}

// applyChanges works on UTF-16 code units, the positions CodeMirror counts in
func applyChanges(doc []uint16, raw json.RawMessage) ([]uint16, error) {
	var parts []json.RawMessage
	if err := json.Unmarshal(raw, &parts); err != nil {
		return nil, ErrInvalidChanges
	}

	result := make([]uint16, 0, len(doc))
	position := 0
	for _, part := range parts {
		var kept int
		if json.Unmarshal(part, &kept) == nil {
			if kept < 0 || position+kept > len(doc) {
				return nil, ErrInvalidChanges
			}
			result = append(result, doc[position:position+kept]...)
			position += kept
			continue
		}

		var change []json.RawMessage
		var deleted int
		if json.Unmarshal(part, &change) != nil || len(change) == 0 || json.Unmarshal(change[0], &deleted) != nil {
			return nil, ErrInvalidChanges
		}
		if deleted < 0 || position+deleted > len(doc) {
			return nil, ErrInvalidChanges
		}
		lines := make([]string, len(change)-1)
		for i, line := range change[1:] {
			if json.Unmarshal(line, &lines[i]) != nil {
				return nil, ErrInvalidChanges
			}
		}
		result = append(result, utf16.Encode([]rune(strings.Join(lines, "\n")))...)
		position += deleted
	}
	if position != len(doc) {
		return nil, ErrInvalidChanges
	}
	return result, nil
}

func encodeText(text string) []uint16 {
	return utf16.Encode([]rune(text))
}

func decodeText(units []uint16) string {
	return string(utf16.Decode(units))
}
