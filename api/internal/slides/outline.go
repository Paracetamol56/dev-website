package slides

import (
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	"golang.org/x/net/html"
)

// OutlineEntry describes a horizontal slide; From and To are UTF-16 offsets, like editor positions
type OutlineEntry struct {
	Title    string `json:"title" bson:"title"`
	From     int    `json:"from" bson:"from"`
	To       int    `json:"to" bson:"to"`
	Vertical int    `json:"vertical" bson:"vertical"`
}

// utf16Offsets converts increasing byte offsets of a text to UTF-16 offsets
type utf16Offsets struct {
	text  string
	byte  int
	units int
}

func (offsets *utf16Offsets) at(position int) int {
	for offsets.byte < position {
		r, size := utf8.DecodeRuneInString(offsets.text[offsets.byte:])
		offsets.units += utf16.RuneLen(r)
		offsets.byte += size
	}
	return offsets.units
}

func headingLevel(tag string) int {
	if len(tag) == 2 && tag[0] == 'h' && tag[1] >= '1' && tag[1] <= '6' {
		return int(tag[1] - '0')
	}
	return 0
}

// BuildOutline lists the top-level sections of a reveal.js deck. A slide is titled by its
// highest-level heading, the first one when there are several.
func BuildOutline(content string) []OutlineEntry {
	outline := []OutlineEntry{}
	tokenizer := html.NewTokenizer(strings.NewReader(content))
	offsets := &utf16Offsets{text: content}

	var (
		position, depth int
		current         *OutlineEntry
		bestLevel       int
		heading         string
		headingLevelNow int
		headingText     strings.Builder
	)

	for {
		kind := tokenizer.Next()
		if kind == html.ErrorToken {
			break
		}
		start := position
		position += len(tokenizer.Raw())
		token := tokenizer.Token()

		switch kind {
		case html.StartTagToken, html.SelfClosingTagToken:
			if token.Data == "section" {
				if depth == 0 {
					outline = append(outline, OutlineEntry{From: offsets.at(start), To: -1})
					current = &outline[len(outline)-1]
					bestLevel = 0
				} else if depth == 1 {
					current.Vertical++
				}
				depth++
			} else if level := headingLevel(token.Data); level > 0 && depth > 0 && headingLevelNow == 0 {
				headingLevelNow = level
				heading = token.Data
				headingText.Reset()
			}
		case html.EndTagToken:
			if token.Data == "section" && depth > 0 {
				depth--
				if depth == 0 {
					current.To = offsets.at(position)
					current = nil
					headingLevelNow = 0
				}
			} else if headingLevelNow > 0 && token.Data == heading {
				if bestLevel == 0 || headingLevelNow < bestLevel {
					bestLevel = headingLevelNow
					current.Title = strings.Join(strings.Fields(headingText.String()), " ")
				}
				headingLevelNow = 0
			}
		case html.TextToken:
			if headingLevelNow > 0 {
				headingText.WriteString(token.Data)
			}
		}
	}

	end := offsets.at(len(content))
	for i := range outline {
		if outline[i].To < 0 {
			outline[i].To = end
		}
	}
	return outline
}
