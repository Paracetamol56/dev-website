package auth

import (
	"dev/internal/user"
	"encoding/json"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const loginTokenLifetime = 15 * time.Minute

type loginToken struct {
	Hash      string             `bson:"_id"`
	UserId    primitive.ObjectID `bson:"userId"`
	ExpiresAt time.Time          `bson:"expiresAt"`
}

type passkeyChallenge struct {
	Id        string    `bson:"_id"`
	Session   []byte    `bson:"session"`
	ExpiresAt time.Time `bson:"expiresAt"`
}

type PasskeyBeginResponse struct {
	SessionId string `json:"sessionId"`
	Options   any    `json:"options"`
}

type PasskeyFinishBody struct {
	SessionId  string          `json:"sessionId" binding:"required"`
	Name       string          `json:"name" binding:"omitempty,max=50"`
	Credential json.RawMessage `json:"credential" binding:"required" swaggertype:"object"`
}

// Session is what a client receives once logged in
type Session struct {
	AccessToken  string      `json:"accessToken"`
	RefreshToken string      `json:"refreshToken"`
	User         *user.User `json:"user,omitempty"`
}
