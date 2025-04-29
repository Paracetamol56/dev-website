package middlewares

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func CORPMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.Contains(c.Request.URL.Path, "barnes-hut") {
			// c.Writer.Header().Set("Cross-Origin-Resource-Policy", "cross-origin")
			c.Writer.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
			c.Writer.Header().Set("Cross-Origin-Embedder-Policy", "require-corp")
		}
		c.Next()
	}
}
