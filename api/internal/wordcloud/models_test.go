package wordcloud

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWordLimit(t *testing.T) {
	limit := func(n int) *int { return &n }
	assert.Equal(t, DefaultMaxWords, Settings{}.WordLimit())
	assert.Equal(t, 0, Settings{MaxWords: limit(0)}.WordLimit())
	assert.Equal(t, 5, Settings{MaxWords: limit(5)}.WordLimit())
}
