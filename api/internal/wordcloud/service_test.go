package wordcloud

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestGenerateCodeUsesUnambiguousAlphabet(t *testing.T) {
	seen := map[string]bool{}
	for range 500 {
		code := GenerateCode()
		assert.Len(t, code, 5)
		assert.NoError(t, ValidateCode(code))
		assert.False(t, strings.ContainsAny(code, "0O1IL"), code)
		assert.Equal(t, strings.ToUpper(code), code)
		seen[code] = true
	}
	assert.Greater(t, len(seen), 490)
}

func TestValidateCode(t *testing.T) {
	assert.NoError(t, ValidateCode("aB3dE"))
	assert.Error(t, ValidateCode("ABCD"))
	assert.Error(t, ValidateCode("ABCDEF"))
	assert.Error(t, ValidateCode("AB-DE"))
}

func TestRecentDistinct(t *testing.T) {
	assert.Equal(t, []string{"c", "a", "b"}, recentDistinct([]string{"a", "b", "a", "c"}, 8))
	assert.Equal(t, []string{"c", "a"}, recentDistinct([]string{"a", "b", "a", "c"}, 2))
	assert.Equal(t, []string{}, recentDistinct(nil, 8))
}

func TestWithSentiment(t *testing.T) {
	words := []Word{{Text: "good"}, {Text: "hate"}, {Text: "teamwork"}}
	rated := withSentiment(&WordCloud{Settings: Settings{Sentiment: true}}, words)
	assert.Equal(t, 3, *rated[0].Sentiment)
	assert.Equal(t, -3, *rated[1].Sentiment)
	assert.Nil(t, rated[2].Sentiment)
	assert.Nil(t, words[0].Sentiment, "the given words are left untouched")

	assert.Nil(t, withSentiment(&WordCloud{}, rated)[0].Sentiment, "sentiment turned off")
	assert.NotNil(t, rated[0].Sentiment)
}

func TestBroadcastDropsBrokenSockets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		conn.Close()
		sessionId := primitive.NewObjectID()
		socketMutex.Lock()
		socketClients[sessionId] = []*socketClient{{conn: conn, owner: true}}
		socketMutex.Unlock()
		broadcast(sessionId, gin.H{"type": "word"}, nil)
		socketMutex.Lock()
		delete(socketClients, sessionId)
		socketMutex.Unlock()
	}))
	defer server.Close()

	conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http"), nil)
	require.NoError(t, err)
	defer conn.Close()
	_, _, err = conn.ReadMessage()
	assert.Error(t, err)
}

func TestCleanText(t *testing.T) {
	assert.Equal(t, "ete", CleanText("  Été "))
	assert.Equal(t, "hello-world", CleanText("Hello-World!"))
	assert.Equal(t, "", CleanText(" !? "))
	assert.Equal(t, "let go", CleanText("Let   go!"))
	assert.Equal(t, "hello world", CleanText(" hello ! world "))
	assert.Equal(t, "привет", CleanText("Привет!"))
	assert.Equal(t, "信任", CleanText("信任"))
	assert.Equal(t, "snake_case 42", CleanText("snake_case 42"))
}

func TestWordParts(t *testing.T) {
	assert.Equal(t, []string{"a", "b", "c", "d"}, wordParts("a b-c_d"))
	assert.Empty(t, wordParts(""))
}

func TestIsProfane(t *testing.T) {
	assert.True(t, IsProfane("merde"))
	assert.True(t, IsProfane(CleanText("Enculé")))
	assert.True(t, IsProfane(CleanText("nique ta mère")))
	assert.True(t, IsProfane("holy-fuck"))
	assert.True(t, IsProfane("holy fuck"))
	assert.False(t, IsProfane("let go"))
	assert.True(t, IsProfane(CleanText("Scheiße")))
	assert.True(t, IsProfane(CleanText("三级片")))
	assert.False(t, IsProfane("trust"))
	assert.False(t, IsProfane("scunthorpe"))
	assert.False(t, IsProfane("classic"))
}

func TestSentiment(t *testing.T) {
	score, ok := Sentiment("good")
	assert.True(t, ok)
	assert.Equal(t, 3, score)
	score, _ = Sentiment("hate")
	assert.Equal(t, -3, score)
	score, ok = Sentiment(CleanText("can't stand"))
	assert.True(t, ok)
	assert.Less(t, score, 0)
	score, ok = Sentiment("great team")
	assert.True(t, ok)
	assert.Equal(t, 3, score)
	score, ok = Sentiment("fun-hate")
	assert.True(t, ok)
	assert.InDelta(t, 0, score, 1)
	score, ok = Sentiment("mauvais")
	assert.True(t, ok)
	assert.Less(t, score, 0)
	score, ok = Sentiment(CleanText("Génial"))
	assert.True(t, ok)
	assert.Equal(t, 3, score)
	score, ok = Sentiment("dobry")
	assert.True(t, ok)
	assert.Greater(t, score, 0)
	_, ok = Sentiment("teamwork")
	assert.False(t, ok)
}

func TestParticipantToken(t *testing.T) {
	participant := uuid.NewV4()
	token := SignParticipant("session-a", participant)

	verified, err := VerifyParticipant("session-a", token)
	assert.NoError(t, err)
	assert.Equal(t, participant, verified)

	_, err = VerifyParticipant("session-b", token)
	assert.ErrorIs(t, err, ErrInvalidParticipant, "token of another session")

	other := uuid.NewV4().String()
	_, signature, _ := strings.Cut(token, ".")
	_, err = VerifyParticipant("session-a", other+"."+signature)
	assert.ErrorIs(t, err, ErrInvalidParticipant, "signature of another participant")

	for _, malformed := range []string{"", "no-dot", participant.String() + ".", participant.String() + ".!!!", "not-a-uuid." + signature} {
		_, err = VerifyParticipant("session-a", malformed)
		assert.ErrorIs(t, err, ErrInvalidParticipant, malformed)
	}

	t.Setenv("ACCESS_TOKEN_SECRET", "another-secret")
	_, err = VerifyParticipant("session-a", token)
	assert.ErrorIs(t, err, ErrInvalidParticipant, "token signed with another secret")
}

func TestFingerprint(t *testing.T) {
	assert.Equal(t, Fingerprint("network", "203.0.113.7"), Fingerprint("network", "203.0.113.7"))
	assert.NotEqual(t, Fingerprint("network", "203.0.113.7"), Fingerprint("network", "203.0.113.8"))
	assert.NotEqual(t, Fingerprint("network", "x"), Fingerprint("device", "x"))
	assert.Len(t, Fingerprint("device", "x"), 16)
}
