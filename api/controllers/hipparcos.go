package controllers

import (
	"net/http"
	"strconv"

	"github.com/Paracetamol56/dev-website/api/models"
	"github.com/gin-gonic/gin"
)

type HipparcosController struct {
}

func (controller *HipparcosController) GetHipparcosHR(c *gin.Context) {
	hip, err := models.GetAllHipparcosHR(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, hip)
}

func (controller *HipparcosController) GetHipparcosHRByHIP(c *gin.Context) {
	num, err := strconv.Atoi(c.Param("hip"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hip, err := models.GetHipparcosHRByHIP(c, num)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if hip == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
		return
	}
	c.JSON(http.StatusOK, hip[0])
}
