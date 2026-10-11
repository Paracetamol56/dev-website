package controllers_test

import (
	"dev/internal/server"
	"dev/internal/utils"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func postContact(router *gin.Engine, remoteAddr string, forwardedFor string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/contact", strings.NewReader("{}"))
	req.RemoteAddr = remoteAddr
	if forwardedFor != "" {
		req.Header.Set("X-Forwarded-For", forwardedFor)
	}
	router.ServeHTTP(w, req)
	return w
}

func TestRateLimitIgnoresSpoofedForwardedFor(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := server.InitRouter()

	for i := range 5 {
		assert.Equal(t, http.StatusBadRequest, postContact(router, "203.0.113.7:4000", "198.51.100."+strconv.Itoa(i)).Code)
	}
	limited := postContact(router, "203.0.113.7:4000", "198.51.100.99")
	assert.Equal(t, http.StatusTooManyRequests, limited.Code)
	assert.NotEmpty(t, limited.Header().Get("Retry-After"))
}

func TestRateLimitTrustsPrivateProxy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := server.InitRouter()

	for range 5 {
		assert.Equal(t, http.StatusBadRequest, postContact(router, "172.18.0.2:4000", "198.51.100.1").Code)
	}
	assert.Equal(t, http.StatusTooManyRequests, postContact(router, "172.18.0.2:4000", "198.51.100.1").Code)
	assert.Equal(t, http.StatusBadRequest, postContact(router, "172.18.0.2:4000", "198.51.100.2").Code)
}

func TestRateLimiterWindow(t *testing.T) {
	limiter := utils.NewRateLimiter(2, 50*time.Millisecond)
	allowed, _ := limiter.Allow("a")
	assert.True(t, allowed)
	allowed, _ = limiter.Allow("a")
	assert.True(t, allowed)
	allowed, retryAfter := limiter.Allow("a")
	assert.False(t, allowed)
	assert.Greater(t, retryAfter, time.Duration(0))
	allowed, _ = limiter.Allow("b")
	assert.True(t, allowed)

	time.Sleep(60 * time.Millisecond)
	allowed, _ = limiter.Allow("a")
	assert.True(t, allowed)
}
