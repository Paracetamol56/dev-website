package hipparcos

import (
	"dev/internal/middlewares"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Controller struct{}

// GetHipparcosHR godoc
//
//	@Summary		Get all Hipparcos HR stars
//	@Description	Get all Hipparcos HR stars
//	@Tags			hipparcos
//	@Produce		json
//	@Success		200	{array}	HipparcosHR
//	@Router			/hipparcos/hr [get]
func (controller *Controller) GetHipparcosHR(c *gin.Context) {
	stars, err := ListStars(c)
	if err != nil {
		middlewares.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, stars)
}

// GetHipparcosHRByHIP godoc
//
//	@Summary		Get one Hipparcos HR star
//	@Description	Get one Hipparcos HR star by HIP number
//	@Tags			hipparcos
//	@Produce		json
//	@Param			hip	path		string	true	"HIP number"
//	@Success		200	{object}	HipparcosHR
//	@Failure		400
//	@Failure		404
//	@Router			/hipparcos/hr/{hip} [get]
func (controller *Controller) GetHipparcosHRByHIP(c *gin.Context) {
	hip, err := strconv.Atoi(c.Param("hip"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	star, err := GetStar(c, hip)
	if err != nil {
		middlewares.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, star)
}
