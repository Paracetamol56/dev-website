package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthController struct {
}

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
func (controller *HealthController) GetHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
