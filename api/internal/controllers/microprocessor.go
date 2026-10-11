package controllers

import (
	"dev/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

type MicroprocessorController struct{}

// GetMicroprocessor godoc
//
//	@Summary		List chips
//	@Description	List CPUs and GPUs with their transistor count (in millions), process node (nm), die area (mm²) and density (transistors per mm²). Documents are flat and do not all have the same keys: besides the fields documented here, each one carries every column of its source table (for example processor, designer, fab, year), and fields without a value are omitted. A column holding a quantity is an object with its value and unit, such as {"value": 8000, "unit": "nm"}, plus the original text when it says more; other columns are text. Chips imported from Wikipedia are refreshed weekly and licensed under CC BY-SA 4.0, see their sourceUrl.
//	@Tags			chips
//	@Produce		json
//	@Param			type	query		string	false	"Only return chips of this type"	Enums(CPU, GPU)
//	@Param			vendor	query		string	false	"Only return chips of this vendor"
//	@Success		200		{array}		models.Microprocessor
//	@Failure		400		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/chips [get]
func (controller *MicroprocessorController) GetMicroprocessor(c *gin.Context) {
	// Apply optional filters to the query
	// Get query parameters in the struct models.MicroprocessorFilter
	var filter models.MicroprocessorFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	data, err := models.GetAllMicroprocessors(c, &filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, data)
}

// GetMicroprocessorById godoc
//
//	@Summary		Get one chip
//	@Description	Get a chip by id
//	@Tags			chips
//	@Produce		json
//	@Param			id	path		string	true	"Chip ID"
//	@Success		200	{object}	models.Microprocessor
//	@Failure		404	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Router			/chips/{id} [get]
func (controller *MicroprocessorController) GetMicroprocessorById(c *gin.Context) {
	data, err := models.GetMicroprocessorById(c, c.Param("id"))
	if err == mongo.ErrNoDocuments {
		c.JSON(http.StatusNotFound, gin.H{"error": "Chip not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, data)
}
