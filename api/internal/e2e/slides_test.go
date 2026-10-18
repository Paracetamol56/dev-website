//go:build e2e

package e2e

import (
	"dev/internal/testutil"
	"net/http"
	"regexp"
	"testing"
	"time"

	"github.com/gorilla/websocket"
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

var joinPattern = regexp.MustCompile(`href="` + regexp.QuoteMeta(testutil.Origin) + `/tools/slides/join/([A-Za-z0-9_-]+)"`)

func TestSharedSlideDeck(t *testing.T) {
	app := start(t)
	owner := app.visitor()
	owner.login(app.address("owner"))
	deck := owner.post("/slides", map[string]any{"title": "Team talk", "content": "hello"}).expect(http.StatusCreated).object()
	id := deck["id"].(string)

	guestAddress := app.address("guest")
	owner.post("/slides/"+id+"/links/editor/email", map[string]string{"email": guestAddress}).expect(http.StatusNoContent)
	emails := app.inbox(guestAddress)
	require.Len(t, emails, 1)
	match := joinPattern.FindStringSubmatch(emails[0].HTMLContent)
	require.NotNil(t, match, "the email links to the app")

	guest := app.visitor()
	guest.login(guestAddress)
	guest.get("/slides/" + id).expect(http.StatusNotFound)
	joined := guest.post("/slides/join", map[string]string{"token": match[1]}).expect(http.StatusOK).object()
	assert.Equal(t, id, joined["id"])
	assert.Equal(t, "editor", joined["role"])
	shared := guest.get("/slides").expect(http.StatusOK).object()
	require.EqualValues(t, 1, shared["total"])
	assert.Contains(t, shared["items"].([]any)[0].(map[string]any)["ownerName"], "owner")

	ownerLive := owner.socket("/slides/" + id + "/ws")
	assert.Equal(t, "owner", ownerLive.next()["role"])
	ownerLive.next()
	guestLive := guest.socket("/slides/" + id + "/ws")
	assert.Equal(t, "editor", guestLive.next()["role"])
	guestLive.next()
	assert.Equal(t, "joined", ownerLive.next()["type"])

	require.NoError(t, guestLive.conn.WriteJSON(map[string]any{"type": "push", "version": 0, "updates": []map[string]any{
		{"clientID": "guest", "changes": []any{5, []any{0, " from the guest"}}},
	}}))
	assert.Equal(t, "updates", ownerLive.next()["type"], "the owner sees the guest typing")

	members := owner.get("/slides/" + id).expect(http.StatusOK).object()["members"].([]any)
	require.Len(t, members, 1)
	guestId := members[0].(map[string]any)["user"].(string)
	owner.delete("/slides/" + id + "/members/" + guestId).expect(http.StatusNoContent)
	for {
		guestLive.conn.SetReadDeadline(time.Now().Add(3 * time.Second))
		var message map[string]any
		err := guestLive.conn.ReadJSON(&message)
		if err != nil {
			closed, ok := err.(*websocket.CloseError)
			require.True(t, ok, "%v", err)
			assert.Equal(t, 4403, closed.Code, "the removed guest is disconnected")
			break
		}
	}
	guest.get("/slides/" + id).expect(http.StatusNotFound)
	guestLive.conn.Close()
	ownerLive.conn.Close()
	require.Eventually(t, func() bool {
		return owner.get("/slides/" + id).expect(http.StatusOK).object()["content"] == "hello from the guest"
	}, 5*time.Second, 50*time.Millisecond)
}
