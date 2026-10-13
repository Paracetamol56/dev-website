package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRateLimiterWindow(t *testing.T) {
	limiter := NewRateLimiter(2, 50*time.Millisecond)
	allowed, _ := limiter.Allow("a")
	assert.True(t, allowed)
	allowed, _ = limiter.Allow("a")
	assert.True(t, allowed)
	allowed, retryAfter := limiter.Allow("a")
	assert.False(t, allowed)
	assert.Greater(t, retryAfter, time.Duration(0))
	assert.LessOrEqual(t, retryAfter, 50*time.Millisecond)
	allowed, _ = limiter.Allow("b")
	assert.True(t, allowed)

	time.Sleep(60 * time.Millisecond)
	allowed, _ = limiter.Allow("a")
	assert.True(t, allowed)
}

func TestRateLimiterPurgesExpiredWindows(t *testing.T) {
	limiter := NewRateLimiter(1, 20*time.Millisecond)
	limiter.Allow("stale")
	time.Sleep(30 * time.Millisecond)
	limiter.Allow("fresh")

	assert.NotContains(t, limiter.windows, "stale")
	assert.Contains(t, limiter.windows, "fresh")
}

func TestErrors(t *testing.T) {
	for kind, err := range map[ErrorKind]error{
		ErrInvalid: Invalid("message"), ErrUnauthorized: Unauthorized("message"), ErrForbidden: Forbidden("message"),
		ErrNotFound: NotFound("message"), ErrConflict: Conflict("message"),
	} {
		assert.Equal(t, &Error{Kind: kind, Message: "message"}, err)
		assert.EqualError(t, err, "message")
	}
	assert.Equal(t, &Error{Kind: ErrTooManyRequests, Message: "Too many requests, please try again later", RetryAfter: time.Minute},
		TooManyRequests(time.Minute))
}
