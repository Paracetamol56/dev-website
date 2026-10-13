package middlewares

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"dev/internal/testutil"
	"dev/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestMain(m *testing.M) {
	testutil.Main(m)
}

func serve(handler gin.HandlerFunc, configure func(*http.Request)) (*httptest.ResponseRecorder, *gin.Context) {
	recorder := httptest.NewRecorder()
	var reached *gin.Context
	engine := gin.New()
	engine.Use(handler)
	engine.GET("/*path", func(c *gin.Context) {
		reached = c
		c.String(http.StatusOK, "reached")
	})
	request := httptest.NewRequest(http.MethodGet, "/resource", nil)
	if configure != nil {
		configure(request)
	}
	engine.ServeHTTP(recorder, request)
	return recorder, reached
}

func withAuthorization(value string) func(*http.Request) {
	return func(request *http.Request) { request.Header.Set("Authorization", value) }
}

func TestJwtAuthMiddleware(t *testing.T) {
	userId := primitive.NewObjectID()
	token, err := utils.SignAccessToken(userId.Hex(), 1)
	require.NoError(t, err)

	recorder, c := serve(JwtAuthMiddleware(), withAuthorization("Bearer "+token))
	assert.Equal(t, http.StatusOK, recorder.Code)
	require.NotNil(t, c)
	assert.Equal(t, userId, c.MustGet("x-user-id"))
}

func TestJwtAuthMiddlewareRejections(t *testing.T) {
	userId := primitive.NewObjectID().Hex()
	expired, err := utils.SignAccessToken(userId, -1)
	require.NoError(t, err)
	refresh, err := utils.SignRefreshToken(userId, 1)
	require.NoError(t, err)
	notAnId, err := jwt.NewWithClaims(jwt.SigningMethodHS256, utils.JwtClaims{UserId: "nope"}).SignedString([]byte("test-access-secret"))
	require.NoError(t, err)
	valid, err := utils.SignAccessToken(userId, 1)
	require.NoError(t, err)

	for name, header := range map[string]string{
		"missing":           "",
		"token only":        valid,
		"too many parts":    "Bearer " + valid + " extra",
		"garbage":           "Bearer garbage",
		"expired":           "Bearer " + expired,
		"refresh token":     "Bearer " + refresh,
		"user id not an id": "Bearer " + notAnId,
	} {
		recorder, c := serve(JwtAuthMiddleware(), withAuthorization(header))
		assert.Equal(t, http.StatusUnauthorized, recorder.Code, name)
		assert.Contains(t, recorder.Body.String(), `"error"`, name)
		assert.Nil(t, c, name+": the handler must not run")
	}
}

func TestCORPMiddleware(t *testing.T) {
	recorder, _ := serve(CORPMiddleware(), func(request *http.Request) { request.URL.Path = "/page/barnes-hut" })
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "same-origin", recorder.Header().Get("Cross-Origin-Opener-Policy"))
	assert.Equal(t, "require-corp", recorder.Header().Get("Cross-Origin-Embedder-Policy"))

	recorder, _ = serve(CORPMiddleware(), nil)
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Empty(t, recorder.Header().Get("Cross-Origin-Opener-Policy"))
	assert.Empty(t, recorder.Header().Get("Cross-Origin-Embedder-Policy"))
}

func TestRequestIdMiddleware(t *testing.T) {
	first, c := serve(RequestIdMiddleware(), nil)
	second, _ := serve(RequestIdMiddleware(), nil)

	assert.Equal(t, first.Header().Get("x-request-id"), utils.RequestId(c))

	id, err := uuid.FromString(first.Header().Get("x-request-id"))
	assert.NoError(t, err)
	assert.EqualValues(t, 4, id.Version())
	assert.NotEqual(t, first.Header().Get("x-request-id"), second.Header().Get("x-request-id"))
}

func TestRateLimitByIP(t *testing.T) {
	testutil.Clean(t)
	rateLimitRecords = utils.NewRateLimiter(1, 10*time.Minute)
	engine := gin.New()
	engine.GET("/", RateLimitByIP(2, time.Hour), func(c *gin.Context) { c.Status(http.StatusNoContent) })
	get := func(remoteAddr string) *httptest.ResponseRecorder {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		request.RemoteAddr = remoteAddr
		engine.ServeHTTP(recorder, request)
		return recorder
	}

	assert.Equal(t, http.StatusNoContent, get("203.0.113.7:1000").Code)
	assert.Equal(t, http.StatusNoContent, get("203.0.113.7:2000").Code)
	limited := get("203.0.113.7:3000")
	assert.Equal(t, http.StatusTooManyRequests, limited.Code)
	assert.Equal(t, "3600", limited.Header().Get("Retry-After"))
	assert.JSONEq(t, `{"error": "Too many requests, please try again later"}`, limited.Body.String())
	assert.Equal(t, http.StatusNoContent, get("203.0.113.8:1000").Code, "another address has its own budget")

	assert.Equal(t, http.StatusTooManyRequests, get("203.0.113.7:4000").Code)
	entries := testutil.Audit(t)
	require.Len(t, entries, 1, "repeated refusals leave a single entry")
	assert.Equal(t, "access.rate_limited", entries[0].Action)
	assert.Equal(t, "203.0.113.7", entries[0].Ip)
	assert.Equal(t, "GET /", entries[0].Meta["route"])
}

func TestAbortTooManyRequests(t *testing.T) {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)

	AbortTooManyRequests(c, 1500*time.Millisecond)

	assert.True(t, c.IsAborted())
	assert.Equal(t, http.StatusTooManyRequests, recorder.Code)
	assert.Equal(t, "2", recorder.Header().Get("Retry-After"), "rounded up to whole seconds")
}

func TestRespondError(t *testing.T) {
	respond := func(err error) *httptest.ResponseRecorder {
		recorder := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(recorder)
		RespondError(c, err)
		return recorder
	}

	for status, err := range map[int]error{
		http.StatusBadRequest:          utils.Invalid("message"),
		http.StatusUnauthorized:        utils.Unauthorized("message"),
		http.StatusForbidden:           utils.Forbidden("message"),
		http.StatusNotFound:            utils.NotFound("message"),
		http.StatusConflict:            utils.Conflict("message"),
		http.StatusInternalServerError: errors.New("message"),
	} {
		response := respond(err)
		assert.Equal(t, status, response.Code)
		assert.JSONEq(t, `{"error": "message"}`, response.Body.String())
	}
	assert.Equal(t, http.StatusNotFound, respond(fmt.Errorf("lookup: %w", utils.NotFound("message"))).Code, "wrapped errors keep their meaning")

	limited := respond(utils.TooManyRequests(90 * time.Second))
	assert.Equal(t, http.StatusTooManyRequests, limited.Code)
	assert.Equal(t, "90", limited.Header().Get("Retry-After"))
}
