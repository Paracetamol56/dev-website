package models

import (
	"encoding/json"
	"testing"
	"time"

	"dev/internal/db"
	"dev/internal/testutil"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestLoginToken(t *testing.T) {
	testutil.Clean(t)
	userId := primitive.NewObjectID()

	token, err := CreateLoginToken(ctx, userId)
	require.NoError(t, err)
	assert.Len(t, token, 43, "32 random bytes")

	var stored bson.M
	require.NoError(t, db.GetDB().Collection("login_tokens").FindOne(ctx, bson.M{}).Decode(&stored))
	assert.NotEqual(t, token, stored["_id"], "only a hash is stored")
	assert.WithinDuration(t, time.Now().Add(15*time.Minute), stored["expiresAt"].(primitive.DateTime).Time(), 5*time.Second)

	consumed, err := ConsumeLoginToken(ctx, token)
	require.NoError(t, err)
	assert.Equal(t, userId, consumed)

	_, err = ConsumeLoginToken(ctx, token)
	assert.EqualError(t, err, "the login link is invalid or has expired, please try again", "a link works once")
	_, err = ConsumeLoginToken(ctx, "unknown")
	assert.Error(t, err)
}

func TestLoginTokenIsReplacedByANewOne(t *testing.T) {
	testutil.Clean(t)
	userId, otherUser := primitive.NewObjectID(), primitive.NewObjectID()

	first, err := CreateLoginToken(ctx, userId)
	require.NoError(t, err)
	others, err := CreateLoginToken(ctx, otherUser)
	require.NoError(t, err)
	second, err := CreateLoginToken(ctx, userId)
	require.NoError(t, err)
	assert.NotEqual(t, first, second)

	_, err = ConsumeLoginToken(ctx, first)
	assert.Error(t, err, "the previous link is revoked")
	consumed, err := ConsumeLoginToken(ctx, second)
	require.NoError(t, err)
	assert.Equal(t, userId, consumed)
	consumed, err = ConsumeLoginToken(ctx, others)
	require.NoError(t, err)
	assert.Equal(t, otherUser, consumed, "links of other users are kept")
}

func TestExpiredLoginToken(t *testing.T) {
	testutil.Clean(t)
	insertRaw(t, "login_tokens", loginToken{Hash: hashLoginToken("late"), UserId: primitive.NewObjectID(), ExpiresAt: time.Now().Add(-time.Second)})

	_, err := ConsumeLoginToken(ctx, "late")
	assert.Error(t, err)
	count, err := db.GetDB().Collection("login_tokens").CountDocuments(ctx, bson.M{})
	require.NoError(t, err)
	assert.Zero(t, count, "an expired link is consumed too")
}

func TestLoginTokenDatabaseFailures(t *testing.T) {
	testutil.BreakDB(t)
	_, err := CreateLoginToken(ctx, primitive.NewObjectID())
	assert.Error(t, err)
	_, err = ConsumeLoginToken(ctx, "token")
	assert.Error(t, err)
}

func TestPasskeys(t *testing.T) {
	credential := &webauthn.Credential{ID: []byte{0xfb, 0xff, 0x01}, PublicKey: []byte("key")}
	assert.Equal(t, "-_8B", PasskeyId(credential), "URL-safe, unpadded")

	passkey, err := NewPasskey("Laptop", credential)
	require.NoError(t, err)
	assert.Equal(t, "-_8B", passkey.Id)
	assert.Equal(t, "Laptop", passkey.Name)
	assert.WithinDuration(t, time.Now(), passkey.CreatedAt, time.Second)
	assert.Nil(t, passkey.LastUsed)

	encoded, err := json.Marshal(passkey)
	require.NoError(t, err)
	assert.NotContains(t, string(encoded), "credential", "the credential is never sent to clients")

	user := &User{
		Id:       primitive.NewObjectID(),
		Email:    "ada@example.test",
		Passkeys: []Passkey{*passkey, {Id: "corrupt", Credential: []byte("{")}, {Id: "other", Credential: []byte(`{"id":"AQID"}`)}},
	}
	assert.Equal(t, user.Id[:], user.WebAuthnID())
	assert.Equal(t, "ada@example.test", user.WebAuthnName())
	assert.Equal(t, "ada@example.test", user.WebAuthnDisplayName())
	user.Name = "Ada"
	assert.Equal(t, "Ada", user.WebAuthnDisplayName())

	credentials := user.WebAuthnCredentials()
	require.Len(t, credentials, 2, "unreadable credentials are skipped")
	assert.Equal(t, credential.ID, credentials[0].ID)
	assert.Equal(t, []byte{1, 2, 3}, credentials[1].ID)
	assert.Empty(t, (&User{}).WebAuthnCredentials())

	assert.Same(t, &user.Passkeys[2], user.GetPasskey("other"))
	assert.Nil(t, user.GetPasskey("unknown"))
	assert.False(t, user.RemovePasskey("unknown"))
	assert.True(t, user.RemovePasskey("corrupt"))
	assert.Len(t, user.Passkeys, 2)
	assert.Nil(t, user.GetPasskey("corrupt"))
}

func TestPasskeyChallenge(t *testing.T) {
	testutil.Clean(t)
	session := &webauthn.SessionData{Challenge: "challenge", UserID: []byte("user"), RelyingPartyID: "localhost"}

	id, err := CreatePasskeyChallenge(ctx, session)
	require.NoError(t, err)
	assert.Len(t, id, 64)
	other, err := CreatePasskeyChallenge(ctx, session)
	require.NoError(t, err)
	assert.NotEqual(t, id, other)

	consumed, err := ConsumePasskeyChallenge(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, session.Challenge, consumed.Challenge)
	assert.Equal(t, session.UserID, consumed.UserID)
	assert.Equal(t, "localhost", consumed.RelyingPartyID)

	_, err = ConsumePasskeyChallenge(ctx, id)
	assert.EqualError(t, err, "the passkey request has expired, please try again", "a ceremony cannot be replayed")
	_, err = ConsumePasskeyChallenge(ctx, "unknown")
	assert.Error(t, err)

	insertRaw(t, "passkey_challenges",
		passkeyChallenge{Id: "late", Session: []byte("{}"), ExpiresAt: time.Now().Add(-time.Second)},
		passkeyChallenge{Id: "corrupt", Session: []byte("{"), ExpiresAt: time.Now().Add(time.Minute)},
	)
	_, err = ConsumePasskeyChallenge(ctx, "late")
	assert.EqualError(t, err, "the passkey request has expired, please try again")
	_, err = ConsumePasskeyChallenge(ctx, "corrupt")
	assert.Error(t, err)

	testutil.BreakDB(t)
	_, err = CreatePasskeyChallenge(ctx, session)
	assert.Error(t, err)
	_, err = ConsumePasskeyChallenge(ctx, other)
	assert.Error(t, err)
}

func TestCreateContact(t *testing.T) {
	testutil.Clean(t)
	contact := Contact{
		UserId: primitive.NewObjectID(), Name: "Ada", Email: "ada@example.test", Message: "Hello there",
		MailId: "<id@brevo>", CreatedAt: primitive.NewDateTimeFromTime(time.Now()),
	}

	id, err := CreateContact(ginContext(), contact)
	require.NoError(t, err)

	var stored Contact
	require.NoError(t, db.GetDB().Collection("contacts").FindOne(ctx, bson.M{"_id": id}).Decode(&stored))
	contact.Id = *id
	assert.Equal(t, contact, stored)

	testutil.BreakDB(t)
	_, err = CreateContact(ginContext(), contact)
	assert.Error(t, err)
}
