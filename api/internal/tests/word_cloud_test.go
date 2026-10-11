package controllers_test

import (
	"dev/internal/controllers"
	"dev/internal/models"
	"dev/internal/utils"
	"strings"
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestGenerateCodeUsesUnambiguousAlphabet(t *testing.T) {
	seen := map[string]bool{}
	for range 500 {
		code := controllers.GenerateCode()
		assert.Len(t, code, 5)
		assert.NoError(t, controllers.ValidateCode(code))
		assert.False(t, strings.ContainsAny(code, "0O1IL"), code)
		assert.Equal(t, strings.ToUpper(code), code)
		seen[code] = true
	}
	assert.Greater(t, len(seen), 490)
}

func TestValidateCode(t *testing.T) {
	assert.NoError(t, controllers.ValidateCode("aB3dE"))
	assert.Error(t, controllers.ValidateCode("ABCD"))
	assert.Error(t, controllers.ValidateCode("ABCDEF"))
	assert.Error(t, controllers.ValidateCode("AB-DE"))
}

func TestCleanText(t *testing.T) {
	assert.Equal(t, "ete", utils.CleanText("  Été "))
	assert.Equal(t, "hello-world", utils.CleanText("Hello-World!"))
	assert.Equal(t, "", utils.CleanText(" !? "))
	assert.Equal(t, "let go", utils.CleanText("Let   go!"))
	assert.Equal(t, "hello world", utils.CleanText(" hello ! world "))
	assert.Equal(t, "привет", utils.CleanText("Привет!"))
	assert.Equal(t, "信任", utils.CleanText("信任"))
}

func TestParticipantToken(t *testing.T) {
	t.Setenv("ACCESS_TOKEN_SECRET", "test-secret")
	participant := uuid.NewV4()
	token := utils.SignParticipant("session-a", participant)

	verified, err := utils.VerifyParticipant("session-a", token)
	assert.NoError(t, err)
	assert.Equal(t, participant, verified)

	_, err = utils.VerifyParticipant("session-b", token)
	assert.ErrorIs(t, err, utils.ErrInvalidParticipant, "token of another session")

	other := uuid.NewV4().String()
	_, signature, _ := strings.Cut(token, ".")
	_, err = utils.VerifyParticipant("session-a", other+"."+signature)
	assert.ErrorIs(t, err, utils.ErrInvalidParticipant, "signature of another participant")

	for _, malformed := range []string{"", "no-dot", participant.String() + ".", "not-a-uuid." + signature} {
		_, err = utils.VerifyParticipant("session-a", malformed)
		assert.ErrorIs(t, err, utils.ErrInvalidParticipant, malformed)
	}
}

func TestIsProfane(t *testing.T) {
	assert.True(t, utils.IsProfane("merde"))
	assert.True(t, utils.IsProfane(utils.CleanText("Enculé")))
	assert.True(t, utils.IsProfane(utils.CleanText("nique ta mère")))
	assert.True(t, utils.IsProfane("holy-fuck"))
	assert.True(t, utils.IsProfane("holy fuck"))
	assert.False(t, utils.IsProfane("let go"))
	assert.True(t, utils.IsProfane(utils.CleanText("Scheiße")))
	assert.True(t, utils.IsProfane(utils.CleanText("三级片")))
	assert.False(t, utils.IsProfane("trust"))
	assert.False(t, utils.IsProfane("scunthorpe"))
	assert.False(t, utils.IsProfane("classic"))
}

func TestFingerprint(t *testing.T) {
	t.Setenv("ACCESS_TOKEN_SECRET", "test-secret")
	assert.Equal(t, utils.Fingerprint("network", "203.0.113.7"), utils.Fingerprint("network", "203.0.113.7"))
	assert.NotEqual(t, utils.Fingerprint("network", "203.0.113.7"), utils.Fingerprint("network", "203.0.113.8"))
	assert.NotEqual(t, utils.Fingerprint("network", "x"), utils.Fingerprint("device", "x"))
	assert.Len(t, utils.Fingerprint("device", "x"), 16)
}

func TestSentiment(t *testing.T) {
	score, ok := utils.Sentiment("good")
	assert.True(t, ok)
	assert.Equal(t, 3, score)
	score, ok = utils.Sentiment("hate")
	assert.Equal(t, -3, score)
	score, ok = utils.Sentiment(utils.CleanText("can't stand"))
	assert.True(t, ok)
	assert.Less(t, score, 0)
	score, ok = utils.Sentiment("great team")
	assert.True(t, ok)
	assert.Equal(t, 3, score)
	score, ok = utils.Sentiment("fun-hate")
	assert.True(t, ok)
	assert.InDelta(t, 0, score, 1)
	score, ok = utils.Sentiment("mauvais")
	assert.True(t, ok)
	assert.Less(t, score, 0)
	score, ok = utils.Sentiment(utils.CleanText("Génial"))
	assert.True(t, ok)
	assert.Equal(t, 3, score)
	score, ok = utils.Sentiment("dobry")
	assert.True(t, ok)
	assert.Greater(t, score, 0)
	_, ok = utils.Sentiment("teamwork")
	assert.False(t, ok)
}

func TestWordLimit(t *testing.T) {
	limit := func(n int) *int { return &n }
	assert.Equal(t, models.DefaultMaxWords, models.Settings{}.WordLimit())
	assert.Equal(t, 0, models.Settings{MaxWords: limit(0)}.WordLimit())
	assert.Equal(t, 5, models.Settings{MaxWords: limit(5)}.WordLimit())
}
