package testutil

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"dev/internal/db"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	BrevoHost  = "api.brevo.com"
	userHeader = "X-Test-User"
)

// Authenticated stands in for the JWT middleware: the user is the one named by As, so
// handler tests do not depend on tokens.
func Authenticated(c *gin.Context) {
	userId, err := primitive.ObjectIDFromHex(c.GetHeader(userHeader))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	c.Set("x-user-id", userId)
}

type Option func(*http.Request)

func As(userId primitive.ObjectID) Option {
	return Header(userHeader, userId.Hex())
}

func Header(name string, value string) Option {
	return func(request *http.Request) { request.Header.Set(name, value) }
}

var FromApp = Header("Origin", Origin)

func AccessToken(t testing.TB, userId primitive.ObjectID) string {
	t.Helper()
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId": userId.Hex(),
		"exp":    time.Now().Add(time.Hour).Unix(),
	}).SignedString([]byte(AccessSecret))
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func Bearer(t testing.TB, userId primitive.ObjectID) Option {
	t.Helper()
	return Header("Authorization", "Bearer "+AccessToken(t, userId))
}

// Do sends a request to the engine; body is sent as is when it is a string, as JSON otherwise.
func Do(t testing.TB, engine *gin.Engine, method string, path string, body any, options ...Option) *httptest.ResponseRecorder {
	t.Helper()
	var content []byte
	switch value := body.(type) {
	case nil:
	case string:
		content = []byte(value)
	default:
		encoded, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		content = encoded
	}
	request := httptest.NewRequest(method, path, bytes.NewReader(content))
	request.Header.Set("Content-Type", "application/json")
	for _, option := range options {
		option(request)
	}
	recorder := httptest.NewRecorder()
	engine.ServeHTTP(recorder, request)
	return recorder
}

func Decode[T any](t testing.TB, recorder *httptest.ResponseRecorder) T {
	t.Helper()
	var value T
	if err := json.Unmarshal(recorder.Body.Bytes(), &value); err != nil {
		t.Fatalf("%v: %s", err, recorder.Body.String())
	}
	return value
}

func ErrorOf(t testing.TB, recorder *httptest.ResponseRecorder) string {
	t.Helper()
	return Decode[map[string]string](t, recorder)["error"]
}

func Context() *gin.Context {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	return c
}

func Count(t testing.TB, collection string) int {
	t.Helper()
	total, err := db.GetDB().Collection(collection).CountDocuments(context.Background(), bson.M{})
	if err != nil {
		t.Fatal(err)
	}
	return int(total)
}

func InsertRaw(t testing.TB, collection string, documents ...any) {
	t.Helper()
	if _, err := db.GetDB().Collection(collection).InsertMany(context.Background(), documents); err != nil {
		t.Fatal(err)
	}
}

// FakeBrevo accepts every email and contact
func FakeBrevo(internet *Internet) {
	internet.JSON(BrevoHost, http.StatusCreated, `{"messageId":"<message@brevo>"}`)
}

type SentEmail struct {
	To []struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"to"`
	ReplyTo *struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"replyTo"`
	Subject     string `json:"subject"`
	HTMLContent string `json:"htmlContent"`
	TextContent string `json:"textContent"`
}

func SentEmails(t testing.TB, internet *Internet) []SentEmail {
	t.Helper()
	emails := []SentEmail{}
	for _, request := range internet.Requests(BrevoHost) {
		if request.Path != "/v3/smtp/email" {
			continue
		}
		var email SentEmail
		if err := json.Unmarshal([]byte(request.Body), &email); err != nil {
			t.Fatal(err)
		}
		emails = append(emails, email)
	}
	return emails
}

func AddedContacts(internet *Internet) []string {
	contacts := []string{}
	for _, request := range internet.Requests(BrevoHost) {
		if request.Path == "/v3/contacts" {
			var body struct {
				Email string `json:"email"`
			}
			json.Unmarshal([]byte(request.Body), &body)
			contacts = append(contacts, body.Email)
		}
	}
	return contacts
}
