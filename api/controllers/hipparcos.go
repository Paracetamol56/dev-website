package controllers

import (
	"net/http"
	"strconv"

	"github.com/Paracetamol56/dev-website/api/models"
	"github.com/gin-gonic/gin"
)

type HipparcosController struct {
}

// GetHipparcosHR godoc
//	@Summary		Get all Hipparcos HR stars
//	@Description	Get all Hipparcos HR stars
//	@Tags			hipparcos
//	@Produce		json
//	@Success		200	{array}	models.HipparcosHR
//	@Router			/hipparcos/hr [get]
func (controller *HipparcosController) GetHipparcosHR(c *gin.Context) {
	hip, err := models.GetAllHipparcosHR(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, hip)
}

// GetHipparcosHRByHIP godoc
//	@Summary		Get one Hipparcos HR star
//	@Description	Get one Hipparcos HR star by HIP number
//	@Tags			hipparcos
//	@Produce		json
//	@Param			hip	path		string	true	"HIP number"
//	@Success		200	{object}	models.HipparcosHR
//	@Failure		400
//	@Failure		404
//	@Router			/hipparcos/hr/{hip} [get]
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
