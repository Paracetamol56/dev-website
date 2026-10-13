//go:build e2e

package e2e

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (browser *browser) join(sessionId string) string {
	browser.t.Helper()
	return browser.post("/word-cloud/"+sessionId+"/participant", nil).expect(http.StatusCreated).object()["token"].(string)
}

func (browser *browser) say(sessionId string, token string, text string) *response {
	browser.t.Helper()
	return browser.post("/word-cloud/"+sessionId+"/word", map[string]any{
		"text": text, "token": token, "client": map[string]any{"timezone": "Europe/Paris", "languages": []string{"fr-FR"}},
	})
}

func wordsOf(session map[string]any) []string {
	texts := []string{}
	for _, word := range session["words"].([]any) {
		if text, plain := word.(string); plain {
			texts = append(texts, text)
		} else {
			texts = append(texts, word.(map[string]any)["text"].(string))
		}
	}
	return texts
}

func TestWordCloudSession(t *testing.T) {
	app := start(t)
	host := app.visitor()
	app.visitor().post("/word-cloud", map[string]any{"name": "Anonymous"}).expect(http.StatusUnauthorized)
	host.login(app.address("host"))

	id := host.post("/word-cloud", map[string]any{
		"name": "Retrospective", "description": "One word about this sprint",
		"settings": map[string]any{"showCloudToAudience": true, "filterProfanity": true, "allowUpvotes": true, "sentiment": true, "maxWords": 3},
	}).expect(http.StatusCreated).object()["id"].(string)
	mine := host.get("/word-cloud?user=" + host.userId).expect(http.StatusOK).list()
	require.Len(t, mine, 1)
	code := mine[0]["code"].(string)
	assert.Len(t, code, 5)

	ada, grace := app.visitor(), app.visitor()
	joined := ada.get("/word-cloud?code=" + code).expect(http.StatusOK).object()
	assert.Equal(t, id, joined["id"], "the audience joins with the code on screen")
	assert.Equal(t, "Retrospective", joined["name"])
	assert.Equal(t, 3.0, joined["settings"].(map[string]any)["maxWords"])
	ada.get("/word-cloud?code=ZZZZZ").expect(http.StatusNotFound)
	ada.get("/word-cloud?user=" + host.userId).expect(http.StatusUnauthorized)
	adaToken, graceToken := ada.join(id), grace.join(id)

	screen := host.socket("/word-cloud/" + id + "/ws")
	assert.Equal(t, map[string]any{"type": "ready", "owner": true}, screen.next())
	phone := ada.socket("/word-cloud/" + id + "/ws")
	assert.Equal(t, map[string]any{"type": "ready", "owner": false}, phone.next())

	assert.Equal(t, "great", ada.say(id, adaToken, " Great! ").expect(http.StatusCreated).object()["text"])
	onScreen := screen.next()
	assert.Equal(t, "word", onScreen["type"])
	assert.Equal(t, "great", onScreen["word"].(map[string]any)["text"])
	assert.Equal(t, 3.0, onScreen["word"].(map[string]any)["sentiment"])
	assert.Equal(t, map[string]any{"type": "word", "word": map[string]any{"text": "great"}}, phone.next())

	grace.say(id, graceToken, "GREAT").expect(http.StatusCreated)
	screen.next()
	phone.next()
	grace.say(id, graceToken, "great").expect(http.StatusTooManyRequests)
	time.Sleep(1100 * time.Millisecond)
	ada.say(id, adaToken, "great").expect(http.StatusConflict)
	ada.say(id, "forged-token", "sneaky").expect(http.StatusUnauthorized)
	app.visitor().say(id, app.visitor().join(id), "merde").expect(http.StatusCreated)
	assert.Equal(t, true, screen.next()["word"].(map[string]any)["hidden"], "profanity reaches the host only, hidden")
	app.visitor().say(id, app.visitor().join(id), "tiring").expect(http.StatusCreated)
	assert.Equal(t, "tiring", screen.next()["word"].(map[string]any)["text"])
	assert.Equal(t, "tiring", phone.next()["word"].(map[string]any)["text"], "the hidden word was skipped for the audience")

	audienceView := ada.get("/word-cloud/" + id).expect(http.StatusOK).object()
	assert.Equal(t, []string{"great", "great", "tiring"}, wordsOf(audienceView))
	assert.Equal(t, []any{"tiring", "great"}, audienceView["recent"])
	hostView := host.get("/word-cloud/" + id).expect(http.StatusOK).object()
	assert.Equal(t, []string{"great", "great", "merde", "tiring"}, wordsOf(hostView))
	assert.NotContains(t, string(host.get("/word-cloud/"+id).Body), "Europe/Paris")

	mallory := app.visitor()
	mallory.login(app.address("mallory"))
	tiring := hostView["words"].([]any)[3].(map[string]any)["id"].(string)
	mallory.patch("/word-cloud/"+id+"/words/"+tiring, map[string]any{"hidden": true}).expect(http.StatusForbidden)
	mallory.patch("/word-cloud/"+id, map[string]any{"open": false}).expect(http.StatusForbidden)
	mallory.delete("/word-cloud/" + id).expect(http.StatusForbidden)
	assert.Equal(t, []string{"great", "great", "tiring"}, wordsOf(mallory.get("/word-cloud/"+id).expect(http.StatusOK).object()),
		"another logged-in user is part of the audience")

	host.patch("/word-cloud/"+id+"/words/"+tiring, map[string]any{"hidden": true}).expect(http.StatusOK)
	assert.Equal(t, "session", screen.next()["type"])
	assert.Equal(t, []string{"great", "great"}, wordsOf(phone.next()["session"].(map[string]any)))

	merged := host.post("/word-cloud/"+id+"/words/merge", map[string]any{"from": []string{"great"}, "into": "Awesome"}).expect(http.StatusOK).object()
	assert.Equal(t, []string{"awesome", "awesome", "merde", "tiring"}, wordsOf(merged))
	screen.next()
	assert.Equal(t, []string{"awesome", "awesome"}, wordsOf(phone.next()["session"].(map[string]any)))

	host.patch("/word-cloud/"+id, map[string]any{"open": false}).expect(http.StatusOK)
	screen.next()
	assert.Equal(t, false, phone.next()["session"].(map[string]any)["open"])
	app.visitor().get("/word-cloud?code=" + code).expect(http.StatusNotFound)
	app.visitor().post("/word-cloud/"+id+"/participant", nil).expect(http.StatusConflict)
	assert.Empty(t, host.get("/word-cloud?user="+host.userId).expect(http.StatusOK).list())
	archived := host.get("/word-cloud?user=" + host.userId + "&status=closed").expect(http.StatusOK).list()
	require.Len(t, archived, 1)
	assert.Equal(t, 4.0, archived[0]["submissions"])

	copyId := host.post("/word-cloud/"+id+"/duplicate", nil).expect(http.StatusCreated).object()["id"].(string)
	copied := host.get("/word-cloud/" + copyId).expect(http.StatusOK).object()
	assert.Equal(t, "Retrospective", copied["name"])
	assert.Empty(t, copied["words"])
	assert.NotEqual(t, code, copied["code"])

	host.patch("/word-cloud/"+id, map[string]any{"open": true}).expect(http.StatusOK)
	screen.next()
	assert.Equal(t, true, phone.next()["session"].(map[string]any)["open"])
	app.visitor().get("/word-cloud?code=" + code).expect(http.StatusOK)

	host.delete("/word-cloud/" + id).expect(http.StatusNoContent)
	assert.Equal(t, map[string]any{"type": "session", "session": nil}, phone.next(), "the audience is told the session is gone")
	ada.get("/word-cloud/" + id).expect(http.StatusNotFound)
	assert.Len(t, host.get("/word-cloud?user="+host.userId).expect(http.StatusOK).list(), 1)
}
