package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Controller struct{}

type HealthResponse struct {
	Status string `json:"status" example:"ok"`
}

// GetHealth godoc
//
//	@Summary		Get health status
//	@Description	Get health status
//	@Tags			health
//	@Produce		json
//	@Success		200 {object} HealthResponse
//	@Router			/health [get]
func (controller *Controller) GetHealth(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{Status: "ok"})
}
