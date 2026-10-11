package utils

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

var nonWordRegex = regexp.MustCompile(`[^\p{L}\p{N}\s_-]+`)

func removeDiacritics(text string) string {
	return strings.Map(func(r rune) rune {
		if unicode.Is(unicode.Mn, r) {
			return -1
		}
		return r
	}, norm.NFD.String(text))
}

// CleanText normalises a submission: lowercase, without diacritics, keeping letters and digits
// of any script, underscores, hyphens and single spaces between words.
func CleanText(text string) string {
	return strings.Join(strings.Fields(nonWordRegex.ReplaceAllString(removeDiacritics(strings.ToLower(text)), "")), " ")
}

// wordParts splits a cleaned submission into its words, hyphenated parts included.
func wordParts(text string) []string {
	return strings.FieldsFunc(text, func(r rune) bool { return r == ' ' || r == '-' || r == '_' })
}
