package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthController struct {
}

func (controller *HealthController) GetHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
