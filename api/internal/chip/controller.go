package chip

import (
	"dev/internal/middlewares"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Controller struct{}

// GetMicroprocessor godoc
//
//	@Summary		List chips
//	@Description	List CPUs and GPUs with their transistor count (in millions), process node (nm), die area (mm²) and density (transistors per mm²). Documents are flat and do not all have the same keys: besides the fields documented here, each one carries every column of its source table (for example processor, designer, fab, year), and fields without a value are omitted. A column holding a quantity is an object with its value and unit, such as {"value": 8000, "unit": "nm"}, plus the original text when it says more; other columns are text. Chips imported from Wikipedia are refreshed weekly and licensed under CC BY-SA 4.0, see their sourceUrl.
//	@Tags			chips
//	@Produce		json
//	@Param			type	query		string	false	"Only return chips of this type"	Enums(CPU, GPU)
//	@Param			vendor	query		string	false	"Only return chips of this vendor"
//	@Success		200		{array}		Microprocessor
//	@Failure		400		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Router			/chips [get]
func (controller *Controller) GetMicroprocessor(c *gin.Context) {
	var filter MicroprocessorFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	chips, err := ListChips(c, &filter)
	if err != nil {
		middlewares.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, chips)
}

// GetMicroprocessorById godoc
//
//	@Summary		Get one chip
//	@Description	Get a chip by id
//	@Tags			chips
//	@Produce		json
//	@Param			id	path		string	true	"Chip ID"
//	@Success		200	{object}	Microprocessor
//	@Failure		404	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Router			/chips/{id} [get]
func (controller *Controller) GetMicroprocessorById(c *gin.Context) {
	chip, err := GetChip(c, c.Param("id"))
	if err != nil {
		middlewares.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, chip)
}
