package utils

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// JwtClaims represents the claims of a JWT token.
type JwtClaims struct {
	UserId string `json:"userId"` // The ID of the user associated with the token.
	jwt.RegisteredClaims
}

// JwtRefreshClaims represents the claims for a JWT refresh token.
type JwtRefreshClaims struct {
	UserId string `json:"userId"` // The user ID associated with the refresh token.
	jwt.RegisteredClaims
}

// SignAccessToken generates a signed access token for the given user ID and expiry time.
// It uses the ACCESS_TOKEN_SECRET environment variable to sign the token.
// The token is signed using the HS256 signing method.
// The generated token is returned along with any error that occurred during the signing process.
func SignAccessToken(userId string, expiry int64) (accessToken string, err error) {
	exp := time.Now().Add(time.Hour * time.Duration(expiry))
	claims := &JwtClaims{
		UserId: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(os.Getenv("ACCESS_TOKEN_SECRET")))
	if err != nil {
		return "", err
	}
	return t, err
}

// SignRefreshToken generates a refresh token for the given user ID and expiry time.
// It creates a JWT token with the user ID as the payload and sets the expiration time.
// The token is then signed using the refresh token secret from the environment variables.
// The generated refresh token and any error encountered during the process are returned.
func SignRefreshToken(userId string, expiry int64) (refreshToken string, err error) {
	exp := time.Now().Add(time.Hour * time.Duration(expiry))
	claimsRefresh := &JwtRefreshClaims{
		UserId: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claimsRefresh)
	rt, err := token.SignedString([]byte(os.Getenv("REFRESH_TOKEN_SECRET")))
	if err != nil {
		return "", err
	}
	return rt, err
}

// IsAuthorized checks if the provided request token is authorized using the given secret.
// It parses the request token and verifies the signing method and secret.
// If the token is valid and authorized, it returns true. Otherwise, it returns false and an error.
func IsAuthorized(requestToken string, secret string) (bool, error) {
	_, err := jwt.Parse(requestToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

// ExtractID extracts the user ID from a JWT token.
// It takes the requestToken string and secret string as input parameters.
// It returns the extracted user ID as a primitive.ObjectID and an error if any.
func ExtractID(requestToken string, secret string) (primitive.ObjectID, error) {
	token, err := jwt.Parse(requestToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return primitive.NilObjectID, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok && !token.Valid {
		return primitive.NilObjectID, err
	}

	userId, err := primitive.ObjectIDFromHex(claims["userId"].(string))
	if err != nil {
		return primitive.NilObjectID, err
	}

	return userId, nil
}
