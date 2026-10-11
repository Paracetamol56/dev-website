package utils

import (
	"embed"
	"strings"
)

// Lists from github.com/LDNOOBW/List-of-Dirty-Naughty-Obscene-and-Otherwise-Bad-Words (CC BY 4.0, see LICENSE there)
//
//go:embed data/profanity/*.txt
var profanityFiles embed.FS

var profanities = func() map[string]bool {
	words := map[string]bool{}
	entries, _ := profanityFiles.ReadDir("data/profanity")
	for _, entry := range entries {
		content, _ := profanityFiles.ReadFile("data/profanity/" + entry.Name())
		for _, line := range strings.Split(string(content), "\n") {
			// Entries with symbols would lose characters when cleaned and become too broad
			lower := strings.Join(strings.Fields(removeDiacritics(strings.ToLower(line))), " ")
			if cleaned := CleanText(line); cleaned != "" && cleaned == lower {
				words[cleaned] = true
			}
		}
	}
	return words
}()

// IsProfane expects a cleaned submission; each of its words is checked too.
func IsProfane(word string) bool {
	if profanities[word] {
		return true
	}
	for _, part := range wordParts(word) {
		if profanities[part] {
			return true
		}
	}
	return false
}
