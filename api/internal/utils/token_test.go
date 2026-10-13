package utils

import (
	"testing"
	"time"

	"dev/internal/testutil"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func signed(t *testing.T, method jwt.SigningMethod, key any, claims jwt.Claims) string {
	t.Helper()
	token, err := jwt.NewWithClaims(method, claims).SignedString(key)
	require.NoError(t, err)
	return token
}

func TestAccessToken(t *testing.T) {
	userId := primitive.NewObjectID()
	token, err := SignAccessToken(userId.Hex(), 1)
	require.NoError(t, err)

	authorized, err := IsAuthorized(token, testutil.AccessSecret)
	assert.NoError(t, err)
	assert.True(t, authorized)
	extracted, err := ExtractID(token, testutil.AccessSecret)
	assert.NoError(t, err)
	assert.Equal(t, userId, extracted)

	claims := JwtClaims{}
	_, err = jwt.ParseWithClaims(token, &claims, func(*jwt.Token) (any, error) { return []byte(testutil.AccessSecret), nil })
	require.NoError(t, err)
	assert.WithinDuration(t, time.Now().Add(time.Hour), claims.ExpiresAt.Time, time.Minute)
}

func TestRefreshToken(t *testing.T) {
	userId := primitive.NewObjectID()
	token, err := SignRefreshToken(userId.Hex(), 168)
	require.NoError(t, err)

	extracted, err := ExtractID(token, testutil.RefreshSecret)
	assert.NoError(t, err)
	assert.Equal(t, userId, extracted)

	authorized, err := IsAuthorized(token, testutil.AccessSecret)
	assert.Error(t, err, "a refresh token is not an access token")
	assert.False(t, authorized)
}

func TestTokenRejections(t *testing.T) {
	userId := primitive.NewObjectID().Hex()
	secret := []byte(testutil.AccessSecret)

	expired, err := SignAccessToken(userId, -1)
	require.NoError(t, err)
	unsigned := signed(t, jwt.SigningMethodNone, jwt.UnsafeAllowNoneSignatureType, JwtClaims{UserId: userId})

	for name, token := range map[string]string{
		"expired":      expired,
		"none alg":     unsigned,
		"garbage":      "not.a.token",
		"empty":        "",
		"other secret": signed(t, jwt.SigningMethodHS256, []byte("other"), JwtClaims{UserId: userId}),
	} {
		authorized, err := IsAuthorized(token, testutil.AccessSecret)
		assert.Error(t, err, name)
		assert.False(t, authorized, name)
		_, err = ExtractID(token, testutil.AccessSecret)
		assert.Error(t, err, name)
	}

	notAnId := signed(t, jwt.SigningMethodHS256, secret, JwtClaims{UserId: "nope"})
	authorized, err := IsAuthorized(notAnId, testutil.AccessSecret)
	assert.NoError(t, err)
	assert.True(t, authorized)
	extracted, err := ExtractID(notAnId, testutil.AccessSecret)
	assert.Error(t, err)
	assert.Equal(t, primitive.NilObjectID, extracted)
}
