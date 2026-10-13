package middlewares

import (
	"dev/internal/audit"
	"dev/internal/utils"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// A flood of refused requests leaves one audit entry per address and route, not one each
var rateLimitRecords = utils.NewRateLimiter(1, 10*time.Minute)

func RateLimitByIP(limit int, window time.Duration) gin.HandlerFunc {
	limiter := utils.NewRateLimiter(limit, window)
	return func(c *gin.Context) {
		if allowed, retryAfter := limiter.Allow(c.ClientIP()); !allowed {
			if first, _ := rateLimitRecords.Allow(c.ClientIP() + " " + c.FullPath()); first {
				audit.Record(c, audit.Entry{Action: audit.AccessRateLimited, Meta: audit.Meta{"route": c.Request.Method + " " + c.FullPath()}})
			}
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
