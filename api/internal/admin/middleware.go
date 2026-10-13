package admin

import (
	"dev/internal/audit"
	"dev/internal/middlewares"
	users "dev/internal/user"
	"dev/internal/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RequireAdmin reads the role from the database on every request: a demotion applies at once,
// whatever the tokens still in circulation
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.MustGet("x-user-id").(primitive.ObjectID)
		user, err := users.GetFullUserById(c, userId)
		if err != nil || !user.IsAdmin() {
			audit.Record(c, audit.Entry{
				Action: audit.AccessDenied,
				UserId: &userId,
				Target: &audit.Target{Type: audit.TargetAdmin, Id: c.FullPath()},
			})
			middlewares.RespondError(c, utils.Forbidden("Forbidden"))
			c.Abort()
			return
		}
		c.Next()
	}
}
