package models

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"dev/internal/db"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const loginTokenLifetime = 15 * time.Minute

type loginToken struct {
	Hash      string             `bson:"_id"`
	UserId    primitive.ObjectID `bson:"userId"`
	ExpiresAt time.Time          `bson:"expiresAt"`
}

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
