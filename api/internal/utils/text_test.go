package utils

import (
	"strings"
	"testing"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

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

func TestRateLimiterWindow(t *testing.T) {
	limiter := NewRateLimiter(2, 50*time.Millisecond)
	allowed, _ := limiter.Allow("a")
	assert.True(t, allowed)
	allowed, _ = limiter.Allow("a")
	assert.True(t, allowed)
	allowed, retryAfter := limiter.Allow("a")
	assert.False(t, allowed)
	assert.Greater(t, retryAfter, time.Duration(0))
	assert.LessOrEqual(t, retryAfter, 50*time.Millisecond)
	allowed, _ = limiter.Allow("b")
	assert.True(t, allowed)

	time.Sleep(60 * time.Millisecond)
	allowed, _ = limiter.Allow("a")
	assert.True(t, allowed)
}

func TestRateLimiterPurgesExpiredWindows(t *testing.T) {
	limiter := NewRateLimiter(1, 20*time.Millisecond)
	limiter.Allow("stale")
	time.Sleep(30 * time.Millisecond)
	limiter.Allow("fresh")

	assert.NotContains(t, limiter.windows, "stale")
	assert.Contains(t, limiter.windows, "fresh")
}

func TestNewWebAuthn(t *testing.T) {
	webAuthn, err := NewWebAuthn("https://dev.matheo-galuba.com")
	assert.NoError(t, err)
	assert.Equal(t, "dev.matheo-galuba.com", webAuthn.Config.RPID)
	assert.Equal(t, []string{"https://dev.matheo-galuba.com"}, webAuthn.Config.RPOrigins)

	webAuthn, err = NewWebAuthn("http://localhost:5173")
	assert.NoError(t, err)
	assert.Equal(t, "localhost", webAuthn.Config.RPID)

	_, err = NewWebAuthn("https://attacker.example")
	assert.EqualError(t, err, "origin not allowed")
	_, err = NewWebAuthn("")
	assert.Error(t, err)

	AllowedOrigins = append(AllowedOrigins, "http://bad\x7f")
	defer func() { AllowedOrigins = AllowedOrigins[:len(AllowedOrigins)-1] }()
	_, err = NewWebAuthn("http://bad\x7f")
	assert.Error(t, err, "unparsable origin")
}
