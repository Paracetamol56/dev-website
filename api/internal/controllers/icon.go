package controllers

import (
	"dev/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type IconController struct{}

// GetIcons godoc
//
//	@Summary		List icons
//	@Description	List every available lucide.dev and simpleicons.org icon, refreshed weekly
//	@Tags			icons
//	@Produce		json
//	@Param			source	query		string	false	"Only return icons from this source"	Enums(lucide, simpleicons)
//	@Success		200		{array}		models.Icon
//	@Failure		400		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/icons [get]
func (controller *IconController) GetIcons(c *gin.Context) {
	var filter models.IconFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	data, err := models.GetAllIcons(c, &filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Cache-Control", "public, max-age=3600")
	c.JSON(http.StatusOK, data)
}

// SearchIcons godoc
//
//	@Summary		Search icons
//	@Description	Search lucide.dev and simpleicons.org icons by text (every word must match the name, title or a tag), source and tag. Results are ranked by relevance.
//	@Tags			icons
//	@Produce		json
//	@Param			q		query		string	false	"Text to search for"
//	@Param			source	query		[]string	false	"Only return icons from these sources (repeat the parameter for several)"	Enums(lucide, simpleicons)	collectionFormat(multi)
//	@Param			tag		query		string	false	"Only return icons having this tag (case-insensitive)"
//	@Param			limit	query		int		false	"Maximum number of results"	default(50)	minimum(1)	maximum(500)
//	@Param			offset	query		int		false	"Number of results to skip"	default(0)	minimum(0)
//	@Success		200		{object}	models.IconSearchResult
//	@Failure		400		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/icons/search [get]
func (controller *IconController) SearchIcons(c *gin.Context) {
	var search models.IconSearch
	if err := c.ShouldBindQuery(&search); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	data, err := models.SearchIcons(c, &search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Cache-Control", "public, max-age=3600")
	c.JSON(http.StatusOK, data)
}
