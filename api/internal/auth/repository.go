package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"dev/internal/db"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var loginTokenIndex sync.Once

func hashLoginToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// CreateLoginToken replaces the pending login links of the user
func CreateLoginToken(ctx context.Context, userId primitive.ObjectID) (string, error) {
	collection := db.GetDB().Collection("login_tokens")
	loginTokenIndex.Do(func() {
		collection.Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys:    bson.D{{Key: "expiresAt", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(0),
		})
	})

	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return "", err
	}
	token := base64.RawURLEncoding.EncodeToString(secret)

	if _, err := collection.DeleteMany(ctx, bson.M{"userId": userId}); err != nil {
		return "", err
	}
	if _, err := collection.InsertOne(ctx, loginToken{
		Hash:      hashLoginToken(token),
		UserId:    userId,
		ExpiresAt: time.Now().Add(loginTokenLifetime),
	}); err != nil {
		return "", err
	}
	return token, nil
}

func ConsumeLoginToken(ctx context.Context, token string) (primitive.ObjectID, error) {
	var stored loginToken
	err := db.GetDB().Collection("login_tokens").FindOneAndDelete(ctx, bson.M{"_id": hashLoginToken(token)}).Decode(&stored)
	if err != nil || time.Now().After(stored.ExpiresAt) {
		return primitive.NilObjectID, errors.New("the login link is invalid or has expired, please try again")
	}
	return stored.UserId, nil
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
