package utils

import (
	"embed"
	"math"
	"strconv"
	"strings"
)

// AFINN lists by Finn Årup Nielsen, github.com/fnielsen/afinn (Apache 2.0, see LICENSE there)
//
//go:embed data/sentiment/*.txt
var afinnFiles embed.FS

// Words listed more than once, in one language or several, get their average score
var afinn = func() map[string]int {
	totals, counts := map[string]int{}, map[string]int{}
	entries, _ := afinnFiles.ReadDir("data/sentiment")
	for _, entry := range entries {
		content, _ := afinnFiles.ReadFile("data/sentiment/" + entry.Name())
		for _, line := range strings.Split(string(content), "\n") {
			word, score, found := strings.Cut(line, "\t")
			value, err := strconv.Atoi(strings.TrimSpace(score))
			if !found || err != nil {
				continue
			}
			if key := CleanText(word); key != "" {
				totals[key] += value
				counts[key]++
			}
		}
	}
	scores := make(map[string]int, len(totals))
	for key, total := range totals {
		scores[key] = int(math.Round(float64(total) / float64(counts[key])))
	}
	return scores
}()

// Sentiment scores a cleaned submission from -5 to 5; ok is false when it is not rated.
// Unknown phrases are rated by the average of their rated words.
func Sentiment(word string) (score int, ok bool) {
	if score, ok := afinn[word]; ok {
		return score, true
	}
	total, rated := 0, 0
	for _, part := range wordParts(word) {
		if score, ok := afinn[part]; ok {
			total += score
			rated++
		}
	}
	if rated == 0 {
		return 0, false
	}
	return int(math.Round(float64(total) / float64(rated))), true
}
