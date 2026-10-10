package middlewares

import (
	"dev/internal/utils"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func RateLimitByIP(limit int, window time.Duration) gin.HandlerFunc {
	limiter := utils.NewRateLimiter(limit, window)
	return func(c *gin.Context) {
		if allowed, retryAfter := limiter.Allow(c.ClientIP()); !allowed {
			AbortTooManyRequests(c, retryAfter)
			return
		}
		c.Next()
	}
}

func AbortTooManyRequests(c *gin.Context, retryAfter time.Duration) {
	c.Header("Retry-After", strconv.Itoa(int(math.Ceil(retryAfter.Seconds()))))
	c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests, please try again later"})
}
