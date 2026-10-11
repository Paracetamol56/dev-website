package middlewares

import (
	"dev/internal/utils"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

var errorStatuses = map[utils.ErrorKind]int{
	utils.ErrInvalid:      http.StatusBadRequest,
	utils.ErrUnauthorized: http.StatusUnauthorized,
	utils.ErrForbidden:    http.StatusForbidden,
	utils.ErrNotFound:     http.StatusNotFound,
	utils.ErrConflict:     http.StatusConflict,
}

// RespondError answers with the status matching a service error, 500 for anything else
func RespondError(c *gin.Context, err error) {
	var failure *utils.Error
	if !errors.As(err, &failure) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if failure.Kind == utils.ErrTooManyRequests {
		AbortTooManyRequests(c, failure.RetryAfter)
		return
	}
	c.JSON(errorStatuses[failure.Kind], gin.H{"error": failure.Message})
}
