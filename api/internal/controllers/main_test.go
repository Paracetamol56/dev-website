package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"dev/internal/db"
	"dev/internal/models"
	"dev/internal/testutil"
	"dev/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestMain(m *testing.M) {
	testutil.Main(m)
}

const (
	brevoHost  = "api.brevo.com"
	testUserId = "X-Test-User"
)

var ctx = context.Background()

// The handlers are mounted like in the server, except for authentication: the user is the
// one named by the X-Test-User header, so these tests do not depend on tokens.
func newEngine() *gin.Engine {
	authenticated := func(c *gin.Context) {
		userId, err := primitive.ObjectIDFromHex(c.GetHeader(testUserId))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Set("x-user-id", userId)
	}

	auth := new(AuthController)
	contact := new(ContactController)
	health := new(HealthController)
	hipparcos := new(HipparcosController)
	icon := new(IconController)
	wordCloud := new(WordCloudController)
	microprocessor := new(MicroprocessorController)
	ormi := new(OrmiController)
	passkey := new(PasskeyController)
	user := new(UserController)

	engine := gin.New()
	engine.GET("/health", health.GetHealth)
	engine.POST("/contact", contact.PostContact)
	engine.POST("/auth/login", auth.PostLogin)
	engine.POST("/auth/verify", auth.PostVerify)
	engine.POST("/auth/refresh", auth.PostRefresh)
	engine.GET("/auth/providers", auth.GetOAuthProviders)
	engine.POST("/auth/:provider", auth.PostOAuthLogin)
	engine.POST("/auth/passkey/begin", passkey.PostLoginBegin)
	engine.POST("/auth/passkey/finish", passkey.PostLoginFinish)
	engine.GET("/hipparcos", hipparcos.GetHipparcosHR)
	engine.GET("/hipparcos/:hip", hipparcos.GetHipparcosHRByHIP)
	engine.GET("/icons", icon.GetIcons)
	engine.GET("/icons/search", icon.SearchIcons)
	engine.GET("/chips", microprocessor.GetMicroprocessor)
	engine.GET("/chips/:id", microprocessor.GetMicroprocessorById)

	engine.GET("/word-cloud", wordCloud.GetWordCloud)
	engine.GET("/word-cloud/:id", wordCloud.GetWordCloudById)
	engine.GET("/word-cloud/:id/ws", wordCloud.WSWordCloud)
	engine.POST("/word-cloud/:id/participant", wordCloud.PostWordCloudParticipant)
	engine.POST("/word-cloud/:id/word", wordCloud.PostWordCloudWord)
	engine.POST("/word-cloud", authenticated, wordCloud.PostWordCloud)
	engine.PATCH("/word-cloud/:id", authenticated, wordCloud.PatchWordCloud)
	engine.POST("/word-cloud/:id/duplicate", authenticated, wordCloud.PostWordCloudDuplicate)
	engine.DELETE("/word-cloud/:id", authenticated, wordCloud.DeleteWordCloud)
	engine.PATCH("/word-cloud/:id/words/:wordId", authenticated, wordCloud.PatchWord)
	engine.POST("/word-cloud/:id/words/merge", authenticated, wordCloud.PostMergeWords)

	ormiGroup := engine.Group("/ormi", authenticated)
	ormiGroup.GET("", ormi.GetTodos)
	ormiGroup.GET("/stats", ormi.GetTodoStats)
	ormiGroup.GET("/stats/:year", ormi.GetTodoYearStats)
	ormiGroup.GET("/labels", ormi.GetTodoLabels)
	ormiGroup.PUT("/order", ormi.PutTodoOrder)
	ormiGroup.GET("/:id", ormi.GetTodo)
	ormiGroup.POST("", ormi.PostTodo)
	ormiGroup.PATCH("/:id", ormi.PatchTodo)
	ormiGroup.DELETE("/:id", ormi.DeleteTodo)

	userGroup := engine.Group("/users", authenticated)
	userGroup.GET("/:id", user.GetUser)
	userGroup.PATCH("/:id", user.PatchUser)
	userGroup.DELETE("/:id", user.DeleteUser)
	userGroup.GET("/:id/export", user.GetExport)
	userGroup.DELETE("/:id/identities/:provider", user.DeleteIdentity)
	userGroup.POST("/:id/passkeys/begin", passkey.PostRegisterBegin)
	userGroup.POST("/:id/passkeys/finish", passkey.PostRegisterFinish)
	userGroup.DELETE("/:id/passkeys/:passkeyId", passkey.DeletePasskey)
	return engine
}

type option func(*http.Request)

func as(userId primitive.ObjectID) option {
	return func(request *http.Request) { request.Header.Set(testUserId, userId.Hex()) }
}

func header(name string, value string) option {
	return func(request *http.Request) { request.Header.Set(name, value) }
}

func bearer(t *testing.T, userId primitive.ObjectID) option {
	t.Helper()
	token, err := utils.SignAccessToken(userId.Hex(), 1)
	require.NoError(t, err)
	return header("Authorization", "Bearer "+token)
}

var fromApp = header("Origin", testutil.Origin)

// do sends a request to a fresh engine; body is sent as is when it is a string, as JSON otherwise.
func do(t *testing.T, method string, path string, body any, options ...option) *httptest.ResponseRecorder {
	t.Helper()
	var reader *bytes.Reader
	switch value := body.(type) {
	case nil:
		reader = bytes.NewReader(nil)
	case string:
		reader = bytes.NewReader([]byte(value))
	default:
		encoded, err := json.Marshal(value)
		require.NoError(t, err)
		reader = bytes.NewReader(encoded)
	}
	request := httptest.NewRequest(method, path, reader)
	request.Header.Set("Content-Type", "application/json")
	for _, option := range options {
		option(request)
	}
	recorder := httptest.NewRecorder()
	newEngine().ServeHTTP(recorder, request)
	return recorder
}

func decode[T any](t *testing.T, recorder *httptest.ResponseRecorder) T {
	t.Helper()
	var value T
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &value), recorder.Body.String())
	return value
}

func errorOf(t *testing.T, recorder *httptest.ResponseRecorder) string {
	t.Helper()
	return decode[map[string]string](t, recorder)["error"]
}

func ginContext() *gin.Context {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	return c
}

func createUser(t *testing.T, user models.User) *models.User {
	t.Helper()
	if user.Email == "" {
		user.Email = primitive.NewObjectID().Hex() + "@example.test"
	}
	if user.Flavour == "" {
		user.Flavour = "mocha"
	}
	result, err := models.CreateUser(ginContext(), &user)
	require.NoError(t, err)
	if len(user.Identities) > 0 || len(user.Passkeys) > 0 {
		_, err = models.UpdateUser(ginContext(), result.InsertedID.(primitive.ObjectID), &user)
		require.NoError(t, err)
	}
	return storedUser(t, result.InsertedID.(primitive.ObjectID))
}

func storedUser(t *testing.T, id primitive.ObjectID) *models.User {
	t.Helper()
	user, err := models.GetFullUserById(ginContext(), id)
	require.NoError(t, err)
	return user
}

func count(t *testing.T, collection string) int {
	t.Helper()
	total, err := db.GetDB().Collection(collection).CountDocuments(ctx, bson.M{})
	require.NoError(t, err)
	return int(total)
}

// fakeBrevo accepts every email and contact
func fakeBrevo(internet *testutil.Internet) {
	internet.JSON(brevoHost, http.StatusCreated, `{"messageId":"<message@brevo>"}`)
}

type sentEmail struct {
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

func sentEmails(t *testing.T, internet *testutil.Internet) []sentEmail {
	t.Helper()
	emails := []sentEmail{}
	for _, request := range internet.Requests(brevoHost) {
		if request.Path != "/v3/smtp/email" {
			continue
		}
		var email sentEmail
		require.NoError(t, json.Unmarshal([]byte(request.Body), &email))
		emails = append(emails, email)
	}
	return emails
}

func addedContacts(internet *testutil.Internet) []string {
	contacts := []string{}
	for _, request := range internet.Requests(brevoHost) {
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

func resetLimiters() {
	loginEmailLimiter = utils.NewRateLimiter(3, 15*time.Minute)
	participantLimiter = utils.NewRateLimiter(1, wordCooldown)
}

func repeat(text string, count int) string {
	return strings.Repeat(text, count)
}
