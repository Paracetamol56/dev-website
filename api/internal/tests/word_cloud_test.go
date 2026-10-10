package controllers_test

import (
	"dev/internal/controllers"
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
	assert.Equal(t, "ete", controllers.CleanText("  Été "))
	assert.Equal(t, "hello-world", controllers.CleanText("Hello-World!"))
	assert.Equal(t, "", controllers.CleanText(" !? "))
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
