package middlewares

import (
	"net/http"
	"os"
	"strings"

	"github.com/Paracetamol56/dev-website/api/utils"
	"github.com/gin-gonic/gin"
)

// JwtAuthMiddleware is a middleware function that performs JWT authentication.
// It extracts the JWT token from the "Authorization" header, verifies its validity,
// and sets the user ID in the context if the token is authorized.
// If the token is not provided or invalid, it returns an "unauthorized" error response.
func JwtAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		t := strings.Split(authHeader, " ")
		if len(t) != 2 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}
		authToken := t[1]
		authorized, err := utils.IsAuthorized(authToken, os.Getenv("ACCESS_TOKEN_SECRET"))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}
		if authorized {
			userID, err := utils.ExtractID(authToken, os.Getenv("ACCESS_TOKEN_SECRET"))
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
				c.Abort()
				return
			}
			c.Set("x-user-id", userID)
			c.Next()
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		c.Abort()
	}
}
