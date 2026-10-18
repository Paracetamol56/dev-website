package slides

import (
	"context"
	"dev/internal/testutil"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type socket struct {
	t    *testing.T
	conn *websocket.Conn
}

func serve(t *testing.T) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(newEngine())
	t.Cleanup(server.Close)
	return server
}

func dial(t *testing.T, server *httptest.Server, id primitive.ObjectID, origin string) (*websocket.Conn, error) {
	t.Helper()
	header := http.Header{}
	header.Set("Origin", origin)
	conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http")+"/slides/"+id.Hex()+"/ws", header)
	if err == nil {
		t.Cleanup(func() { conn.Close() })
	}
	return conn, err
}

// open joins a deck as userId, having version, and reads the state message
func open(t *testing.T, server *httptest.Server, id primitive.ObjectID, userId primitive.ObjectID, version int64) *socket {
	t.Helper()
	conn, err := dial(t, server, id, testutil.Origin)
	require.NoError(t, err)
	opened := &socket{t: t, conn: conn}
	opened.send(map[string]any{"type": "auth", "token": testutil.AccessToken(t, userId), "version": version})
	assert.Equal(t, "state", opened.next()["type"])
	return opened
}

func (s *socket) send(message any) {
	s.t.Helper()
	require.NoError(s.t, s.conn.WriteJSON(message))
}

func (s *socket) next() map[string]any {
	s.t.Helper()
	s.conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	var message map[string]any
	require.NoError(s.t, s.conn.ReadJSON(&message))
	return message
}

// expect skips the notices about other clients joining or leaving, unless they are expected
func (s *socket) expect(kind string) map[string]any {
	s.t.Helper()
	for {
		message := s.next()
		if message["type"] != kind && (message["type"] == "joined" || message["type"] == "left") {
			continue
		}
		require.Equal(s.t, kind, message["type"], "%v", message)
		return message
	}
}

func update(clientId string, changes string) map[string]any {
	return map[string]any{"clientID": clientId, "changes": json.RawMessage(changes)}
}

func push(version int64, updates ...map[string]any) map[string]any {
	return map[string]any{"type": "push", "version": version, "updates": updates}
}

func storedDeck(t *testing.T, id primitive.ObjectID) *Deck {
	t.Helper()
	deck, err := GetDeckById(context.Background(), id)
	require.NoError(t, err)
	return deck
}

// closeAll disconnects the sockets and waits for the hub to save and close
func closeAll(t *testing.T, id primitive.ObjectID, sockets ...*socket) {
	t.Helper()
	for _, s := range sockets {
		s.conn.Close()
	}
	require.Eventually(t, func() bool {
		hubs.mutex.Lock()
		defer hubs.mutex.Unlock()
		return hubs.hubs[id] == nil && hubs.closing[id] == nil
	}, 3*time.Second, 10*time.Millisecond)
}

func TestSocketRefusesStrangers(t *testing.T) {
	testutil.Clean(t)
	server := serve(t)
	owner := primitive.NewObjectID()
	deck := postDeck(t, owner, map[string]any{"title": "Private"})

	_, err := dial(t, server, deck.Id, "https://evil.example")
	assert.Error(t, err, "another site cannot open the socket")

	for name, auth := range map[string]any{
		"no auth":        map[string]any{"type": "push"},
		"invalid token":  map[string]any{"type": "auth", "token": "nope"},
		"someone else's": map[string]any{"type": "auth", "token": testutil.AccessToken(t, primitive.NewObjectID())},
	} {
		t.Run(name, func(t *testing.T) {
			conn, err := dial(t, server, deck.Id, testutil.Origin)
			require.NoError(t, err)
			require.NoError(t, conn.WriteJSON(auth))
			conn.SetReadDeadline(time.Now().Add(3 * time.Second))
			_, _, err = conn.ReadMessage()
			var closed *websocket.CloseError
			require.ErrorAs(t, err, &closed)
			assert.Contains(t, []int{closeAuth, closeNotFound}, closed.Code)
		})
	}
}

func TestSocketJoin(t *testing.T) {
	testutil.Clean(t)
	server := serve(t)
	owner := primitive.NewObjectID()
	deck := postDeck(t, owner, map[string]any{"title": "Live"})

	s := open(t, server, deck.Id, owner, 0)
	caughtUp := s.expect("updates")
	assert.EqualValues(t, 0, caughtUp["version"])
	assert.Equal(t, []any{}, caughtUp["updates"], "an empty list, not null")

	ahead := open(t, server, deck.Id, owner, 5)
	reset := ahead.expect("reset")
	assert.EqualValues(t, 0, reset["version"])
	assert.Equal(t, defaultContent, reset["doc"], "a client ahead of the server starts again from its document")
}

func TestSocketPush(t *testing.T) {
	testutil.Clean(t)
	server := serve(t)
	owner := primitive.NewObjectID()
	deck := postDeck(t, owner, map[string]any{"title": "Live", "content": "hello"})

	alice := open(t, server, deck.Id, owner, 0)
	alice.expect("updates")
	bob := open(t, server, deck.Id, owner, 0)
	bob.expect("updates")

	alice.send(push(0, update("alice", `[5, [0, " world"]]`)))
	received := alice.expect("updates")
	assert.EqualValues(t, 0, received["version"])
	assert.Len(t, received["updates"], 1)
	assert.Equal(t, true, alice.expect("pushed")["ok"])
	broadcast := bob.expect("updates")
	assert.Equal(t, received["updates"], broadcast["updates"], "the other clients get the same update")

	bob.send(push(0, update("bob", `[[5, "bye"]]`)))
	assert.Equal(t, false, bob.expect("pushed")["ok"], "an update from an old version is refused")

	bob.send(push(1, update("bob", `[11, [0, "!"]]`), update("bob", `[[0, "> "], 12]`)))
	assert.Len(t, bob.expect("updates")["updates"], 2)
	assert.Equal(t, true, bob.expect("pushed")["ok"])
	assert.EqualValues(t, 1, alice.expect("updates")["version"])

	alice.send(push(3, update("alice", `[3]`)))
	assert.Contains(t, alice.expect("error")["error"], "do not match")
	assert.Equal(t, false, alice.expect("pushed")["ok"])

	alice.send(map[string]any{"type": "pull", "version": 1})
	pulled := alice.expect("updates")
	assert.EqualValues(t, 1, pulled["version"])
	assert.Len(t, pulled["updates"], 2)

	closeAll(t, deck.Id, alice, bob)
	stored := storedDeck(t, deck.Id)
	assert.Equal(t, "> hello world!", stored.Content, "the hub saves when the last client leaves")
	assert.EqualValues(t, 3, stored.Revision)

	again := open(t, server, deck.Id, owner, 3)
	assert.Empty(t, again.expect("updates")["updates"], "a new hub starts from the saved version")
}

func TestSocketConcurrentPushes(t *testing.T) {
	testutil.Clean(t)
	server := serve(t)
	owner := primitive.NewObjectID()
	deck := postDeck(t, owner, map[string]any{"title": "Race", "content": "ab"})

	alice := open(t, server, deck.Id, owner, 0)
	alice.expect("updates")
	bob := open(t, server, deck.Id, owner, 0)
	bob.expect("updates")

	alice.send(push(0, update("alice", `[1, [0, "A"], 1]`)))
	bob.send(push(0, update("bob", `[2, [0, "B"]]`)))

	results := map[string]bool{}
	for _, s := range []*socket{alice, bob} {
		for {
			message := s.next()
			if message["type"] == "pushed" {
				results[s.conn.LocalAddr().String()] = message["ok"].(bool)
				break
			}
		}
	}
	accepted := 0
	for _, ok := range results {
		if ok {
			accepted++
		}
	}
	assert.Equal(t, 1, accepted, "only one of two pushes from the same version is accepted")

	closeAll(t, deck.Id, alice, bob)
	assert.Contains(t, []string{"aAb", "abB"}, storedDeck(t, deck.Id).Content)
}

func TestSocketKeepsRecentUpdates(t *testing.T) {
	testutil.Clean(t)
	server := serve(t)
	owner := primitive.NewObjectID()
	deck := postDeck(t, owner, map[string]any{"title": "Long", "content": ""})

	s := open(t, server, deck.Id, owner, 0)
	s.expect("updates")
	total := int64(keptUpdates + 5)
	for version := int64(0); version < total; version++ {
		s.send(push(version, update("a", `[[0, "x"], `+strconv.FormatInt(version, 10)+`]`)))
		s.expect("updates")
		require.Equal(t, true, s.expect("pushed")["ok"])
	}

	late := open(t, server, deck.Id, owner, total-1)
	assert.Len(t, late.expect("updates")["updates"], 1)
	early := open(t, server, deck.Id, owner, 1)
	reset := early.expect("reset")
	assert.EqualValues(t, total, reset["version"], "updates older than the kept ones are replaced by the document")
	assert.Equal(t, strings.Repeat("x", int(total)), reset["doc"])
}

func TestSocketTitleAndSettings(t *testing.T) {
	testutil.Clean(t)
	server := serve(t)
	owner := primitive.NewObjectID()
	deck := postDeck(t, owner, map[string]any{"title": "Draft"})

	alice := open(t, server, deck.Id, owner, 0)
	alice.expect("updates")
	bob := open(t, server, deck.Id, owner, 0)
	bob.expect("updates")

	alice.send(map[string]any{"type": "title", "title": "  Final  "})
	assert.Equal(t, "Final", alice.expect("title")["title"])
	assert.Equal(t, "Final", bob.expect("title")["title"])
	assert.Equal(t, "Final", storedDeck(t, deck.Id).Title)

	alice.send(map[string]any{"type": "title", "title": " "})
	assert.Contains(t, alice.expect("error")["error"], "title")

	settings := DefaultSettings()
	settings.Theme = "moon"
	settings.Width, settings.Height = 1280, 720
	bob.send(map[string]any{"type": "settings", "settings": settings})
	assert.Equal(t, "moon", bob.expect("settings")["settings"].(map[string]any)["theme"])
	assert.Equal(t, "moon", alice.expect("settings")["settings"].(map[string]any)["theme"])
	assert.Equal(t, settings, storedDeck(t, deck.Id).Settings)

	invalid := DefaultSettings()
	invalid.Theme = "neon"
	bob.send(map[string]any{"type": "settings", "settings": invalid})
	assert.Contains(t, bob.expect("error")["error"], "settings")
	assert.Equal(t, "moon", storedDeck(t, deck.Id).Settings.Theme)

	bob.send(map[string]any{"type": "dance"})
	assert.Contains(t, bob.expect("error")["error"], "Unknown")

	closeAll(t, deck.Id, alice, bob)
	assert.Equal(t, []string{"slides.created", "slides.renamed"}, testutil.AuditActions(t))
	entry := testutil.Audit(t)[1]
	assert.Equal(t, owner, entry.UserId)
	assert.Equal(t, "Draft", entry.Meta["from"])
}

func TestSocketDeckDeleted(t *testing.T) {
	testutil.Clean(t)
	server := serve(t)
	owner := primitive.NewObjectID()
	deck := postDeck(t, owner, map[string]any{"title": "Gone soon"})

	s := open(t, server, deck.Id, owner, 0)
	s.expect("updates")
	s.send(push(0, update("a", `[[0, "x"], `+strconv.Itoa(len(encodeText(defaultContent)))+`]`)))
	s.expect("updates")
	s.expect("pushed")

	require.Equal(t, http.StatusNoContent, do(t, http.MethodDelete, "/slides/"+deck.Id.Hex(), nil, testutil.As(owner)).Code)
	s.expect("deleted")
	s.conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, _, err := s.conn.ReadMessage()
	assert.Error(t, err, "the socket is closed")

	closeAll(t, deck.Id)
	assert.Zero(t, testutil.Count(t, "slides"), "the hub does not save a deleted deck back")
}

func TestSaveAll(t *testing.T) {
	testutil.Clean(t)
	server := serve(t)
	owner := primitive.NewObjectID()
	deck := postDeck(t, owner, map[string]any{"title": "Stopping", "content": ""})

	s := open(t, server, deck.Id, owner, 0)
	s.expect("updates")
	s.send(push(0, update("a", `[[0, "kept"]]`)))
	s.expect("updates")
	s.expect("pushed")
	assert.Equal(t, "", storedDeck(t, deck.Id).Content, "not saved yet")

	SaveAll(context.Background())
	assert.Equal(t, "kept", storedDeck(t, deck.Id).Content)
	assert.EqualValues(t, 1, storedDeck(t, deck.Id).Revision)
	closeAll(t, deck.Id, s)
}

func TestSocketPresence(t *testing.T) {
	testutil.Clean(t)
	server := serve(t)
	owner := primitive.NewObjectID()
	testutil.InsertRaw(t, "users", bson.M{"_id": owner, "email": "speaker@example.test", "name": "Ada"})
	deck := postDeck(t, owner, map[string]any{"title": "Shared", "content": "hello world"})

	alice := open(t, server, deck.Id, owner, 0)
	alice.expect("updates")
	bob := open(t, server, deck.Id, owner, 0)
	bob.expect("updates")
	joined := alice.expect("joined")
	assert.Equal(t, "Ada", joined["name"])
	bobPeer := joined["peer"]

	alice.send(map[string]any{"type": "presence", "version": 0, "ranges": []Range{{Anchor: 0, Head: 5}, {Anchor: 8, Head: 8}}})
	presence := bob.expect("presence")
	assert.NotEqual(t, bobPeer, presence["peer"])
	assert.Equal(t, "Ada", presence["name"])
	assert.EqualValues(t, 0, presence["version"])
	assert.Equal(t, []any{
		map[string]any{"anchor": float64(0), "head": float64(5)},
		map[string]any{"anchor": float64(8), "head": float64(8)},
	}, presence["ranges"])

	for name, ranges := range map[string][]Range{
		"no range":          {},
		"negative position": {{Anchor: -1, Head: 0}},
		"too far":           {{Anchor: 0, Head: maxContentLength + 1}},
		"too many ranges":   make([]Range, maxRanges+1),
	} {
		t.Run(name, func(t *testing.T) {
			alice.send(map[string]any{"type": "presence", "version": 0, "ranges": ranges})
			assert.Contains(t, alice.expect("error")["error"], "selection")
		})
	}

	alice.send(push(0, update("alice", `[11, [0, "!"]]`)))
	alice.expect("updates")
	alice.expect("pushed")
	assert.Equal(t, "updates", bob.next()["type"], "bob gets nothing about alice's own selection errors")

	bob.conn.Close()
	assert.Equal(t, bobPeer, alice.expect("left")["peer"])

	closeAll(t, deck.Id, alice)
	assert.Equal(t, "hello world!", storedDeck(t, deck.Id).Content, "selections are not saved")
}
