package slides

import (
	"context"
	"dev/internal/testutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestBuildOutline(t *testing.T) {
	for name, test := range map[string]struct {
		content string
		outline []OutlineEntry
	}{
		"empty":        {"", []OutlineEntry{}},
		"no section":   {"<p>Text</p>", []OutlineEntry{}},
		"one slide":    {"<section><h1>Hi</h1></section>", []OutlineEntry{{Title: "Hi", From: 0, To: 30}}},
		"text outside": {"intro<section><h2>A</h2></section>outro", []OutlineEntry{{Title: "A", From: 5, To: 34}}},
		"attributes and case": {`<SECTION data-background="red"><H2 class="x">Big</H2></SECTION>`,
			[]OutlineEntry{{Title: "Big", From: 0, To: 63}}},
		"highest heading wins": {"<section><h3>Small</h3><h1>Main</h1><h1>Other</h1></section>",
			[]OutlineEntry{{Title: "Main", From: 0, To: 60}}},
		"no heading": {"<section><p>Only text</p></section>", []OutlineEntry{{From: 0, To: 35}}},
		"tags and spaces in heading": {"<section><h1>  Hello <em>big</em>\n world </h1></section>",
			[]OutlineEntry{{Title: "Hello big world", From: 0, To: 56}}},
		"vertical stack": {"<section><section><h2>A</h2></section><section><h2>B</h2></section></section><section></section>",
			[]OutlineEntry{{Title: "A", From: 0, To: 77, Vertical: 2}, {From: 77, To: 96}}},
		"unclosed": {"<section><h1>Draft</h1><p>still typing", []OutlineEntry{{Title: "Draft", From: 0, To: 38}}},
		"unclosed heading": {"<section><h1>Draft</section><section><h2>Next</h2></section>",
			[]OutlineEntry{{From: 0, To: 28}, {Title: "Next", From: 28, To: 60}}},
		"stray closing tag": {"</section><section><h1>A</h1></section>", []OutlineEntry{{Title: "A", From: 10, To: 39}}},
		"not a section":     {"<sections></sections><section></section>", []OutlineEntry{{From: 21, To: 40}}},
		"in a comment":      {"<!-- <section> --><section></section>", []OutlineEntry{{From: 18, To: 37}}},
		"in a script":       {"<script>'<section>'</script><section></section>", []OutlineEntry{{From: 28, To: 47}}},
		"utf-16 offsets":    {"é😀<section><h1>Café 😀</h1></section>", []OutlineEntry{{Title: "Café 😀", From: 3, To: 38}}},
	} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.outline, BuildOutline(test.content))
		})
	}
}

func TestBackfillOutlines(t *testing.T) {
	testutil.Clean(t)
	old := primitive.NewObjectID()
	testutil.InsertRaw(t, "slides", bson.M{"_id": old, "user": primitive.NewObjectID(), "title": "Old", "content": "<section><h1>A</h1></section>", "slides": 1})
	postDeck(t, primitive.NewObjectID(), map[string]any{"title": "New"})

	filled, err := BackfillOutlines(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, filled)

	var stored bson.M
	require.NoError(t, collection().FindOne(context.Background(), bson.M{"_id": old}).Decode(&stored))
	assert.NotContains(t, stored, "slides")
	deck, err := GetDeckByUserIdAndId(context.Background(), stored["user"].(primitive.ObjectID), old)
	require.NoError(t, err)
	assert.Equal(t, []OutlineEntry{{Title: "A", From: 0, To: 29}}, deck.Outline)

	filled, err = BackfillOutlines(context.Background())
	require.NoError(t, err)
	assert.Zero(t, filled)
}
