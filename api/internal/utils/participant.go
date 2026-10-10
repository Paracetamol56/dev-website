package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"os"
	"strings"

	uuid "github.com/satori/go.uuid"
)

var ErrInvalidParticipant = errors.New("invalid participant token")

// Tokens are "<uuid>.<hmac of session and uuid>", so they cannot be forged or moved to another session.
func participantSignature(sessionId string, participant uuid.UUID) []byte {
	mac := hmac.New(sha256.New, []byte("word-cloud:"+os.Getenv("ACCESS_TOKEN_SECRET")))
	mac.Write([]byte(sessionId + ":" + participant.String()))
	return mac.Sum(nil)
}

func SignParticipant(sessionId string, participant uuid.UUID) string {
	return participant.String() + "." + base64.RawURLEncoding.EncodeToString(participantSignature(sessionId, participant))
}

func VerifyParticipant(sessionId string, token string) (uuid.UUID, error) {
	id, signature, found := strings.Cut(token, ".")
	if !found {
		return uuid.Nil, ErrInvalidParticipant
	}
	participant, err := uuid.FromString(id)
	if err != nil {
		return uuid.Nil, ErrInvalidParticipant
	}
	decoded, err := base64.RawURLEncoding.DecodeString(signature)
	if err != nil || !hmac.Equal(decoded, participantSignature(sessionId, participant)) {
		return uuid.Nil, ErrInvalidParticipant
	}
	return participant, nil
}
