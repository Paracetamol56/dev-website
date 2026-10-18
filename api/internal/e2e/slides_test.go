//go:build e2e

package e2e

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSlideDecks(t *testing.T) {
	app := start(t)
	speaker := app.visitor()
	speaker.get("/slides").expect(http.StatusUnauthorized)
	speaker.login(app.address("speaker"))

	deck := speaker.post("/slides", map[string]string{"title": "Conference"}).expect(http.StatusCreated).object()
	id := deck["id"].(string)
	assert.Contains(t, deck["content"], "<section>")

	page := speaker.get("/slides").expect(http.StatusOK).object()
	assert.EqualValues(t, 1, page["total"])

	live := speaker.socket("/slides/" + id + "/ws")
	assert.Equal(t, "state", live.next()["type"])
	assert.Equal(t, "updates", live.next()["type"])
	length := len([]rune(deck["content"].(string)))
	require.NoError(t, live.conn.WriteJSON(map[string]any{"type": "push", "version": 0, "updates": []map[string]any{
		{"clientID": "speaker", "changes": []any{[]any{length, "<section>Hello</section>"}}},
	}}))
	assert.Equal(t, "updates", live.next()["type"])
	assert.Equal(t, true, live.next()["ok"])
	require.NoError(t, live.conn.WriteJSON(map[string]any{"type": "title", "title": "Keynote"}))
	assert.Equal(t, "Keynote", live.next()["title"])
	live.conn.Close()
	require.Eventually(t, func() bool {
		return speaker.get("/slides/" + id).expect(http.StatusOK).object()["content"] == "<section>Hello</section>"
	}, 5*time.Second, 50*time.Millisecond)
	assert.Equal(t, "Keynote", speaker.get("/slides/" + id).object()["title"])

	stranger := app.visitor()
	stranger.login(app.address("stranger"))
	stranger.get("/slides/" + id).expect(http.StatusNotFound)
	stranger.delete("/slides/" + id).expect(http.StatusNotFound)

	copied := speaker.post("/slides/"+id+"/duplicate", nil).expect(http.StatusCreated).object()
	speaker.delete("/slides/" + id).expect(http.StatusNoContent)
	page = speaker.get("/slides").expect(http.StatusOK).object()
	require.EqualValues(t, 1, page["total"])
	assert.Equal(t, copied["id"], page["items"].([]any)[0].(map[string]any)["id"])
}
