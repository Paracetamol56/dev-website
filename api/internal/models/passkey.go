package models

import (
	"context"
	"crypto/rand"
	"dev/internal/db"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const MaxPasskeys = 10

type Passkey struct {
	Id         string     `json:"id" bson:"id"`
	Name       string     `json:"name" bson:"name"`
	CreatedAt  time.Time  `json:"createdAt" bson:"createdAt"`
	LastUsed   *time.Time `json:"lastUsed,omitempty" bson:"lastUsed,omitempty"`
	Credential []byte     `json:"-" bson:"credential"`
}

func PasskeyId(credential *webauthn.Credential) string {
	return base64.RawURLEncoding.EncodeToString(credential.ID)
}

func NewPasskey(name string, credential *webauthn.Credential) (*Passkey, error) {
	encoded, err := json.Marshal(credential)
	if err != nil {
		return nil, err
	}
	return &Passkey{
		Id:         PasskeyId(credential),
		Name:       name,
		CreatedAt:  time.Now(),
		Credential: encoded,
	}, nil
}

func (user *User) WebAuthnID() []byte {
	return user.Id[:]
}

func (user *User) WebAuthnName() string {
	return user.Email
}

func (user *User) WebAuthnDisplayName() string {
	if user.Name != "" {
		return user.Name
	}
	return user.Email
}

func (user *User) WebAuthnCredentials() []webauthn.Credential {
	credentials := []webauthn.Credential{}
	for _, passkey := range user.Passkeys {
		var credential webauthn.Credential
		if err := json.Unmarshal(passkey.Credential, &credential); err == nil {
			credentials = append(credentials, credential)
		}
	}
	return credentials
}

func (user *User) GetPasskey(id string) *Passkey {
	for i := range user.Passkeys {
		if user.Passkeys[i].Id == id {
			return &user.Passkeys[i]
		}
	}
	return nil
}

func (user *User) RemovePasskey(id string) bool {
	for i := range user.Passkeys {
		if user.Passkeys[i].Id == id {
			user.Passkeys = append(user.Passkeys[:i], user.Passkeys[i+1:]...)
			return true
		}
	}
	return false
}

type passkeyChallenge struct {
	Id        string    `bson:"_id"`
	Session   []byte    `bson:"session"`
	ExpiresAt time.Time `bson:"expiresAt"`
}

var passkeyChallengeIndex sync.Once

func CreatePasskeyChallenge(ctx context.Context, session *webauthn.SessionData) (string, error) {
	collection := db.GetDB().Collection("passkey_challenges")
	passkeyChallengeIndex.Do(func() {
		collection.Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys:    bson.D{{Key: "expiresAt", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(0),
		})
	})

	id := make([]byte, 32)
	if _, err := rand.Read(id); err != nil {
		return "", err
	}
	encoded, err := json.Marshal(session)
	if err != nil {
		return "", err
	}

	challenge := passkeyChallenge{
		Id:        hex.EncodeToString(id),
		Session:   encoded,
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}
	if _, err := collection.InsertOne(ctx, challenge); err != nil {
		return "", err
	}
	return challenge.Id, nil
}

// ConsumePasskeyChallenge deletes the challenge it returns, so that a ceremony cannot be replayed.
func ConsumePasskeyChallenge(ctx context.Context, id string) (*webauthn.SessionData, error) {
	var challenge passkeyChallenge
	err := db.GetDB().Collection("passkey_challenges").FindOneAndDelete(ctx, bson.M{"_id": id}).Decode(&challenge)
	if err != nil || time.Now().After(challenge.ExpiresAt) {
		return nil, errors.New("the passkey request has expired, please try again")
	}

	var session webauthn.SessionData
	if err := json.Unmarshal(challenge.Session, &session); err != nil {
		return nil, err
	}
	return &session, nil
}
