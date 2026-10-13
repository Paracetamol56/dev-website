package middlewares

import (
	"dev/internal/utils"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
)

func RequestIdMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := uuid.NewV4().String()
		c.Set(utils.RequestIdKey, id)
		c.Writer.Header().Set(utils.RequestIdKey, id)
		c.Next()
	}
}
