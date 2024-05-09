package controllers

import (
	"net/http"

	"github.com/Paracetamol56/dev-website/api/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MicroprocessorController struct {
}

func (controller *MicroprocessorController) GetMicroprocessor(c *gin.Context) {
	// Apply optional filters to the query
	// Get query parameters in the struct models.MicroprocessorFilter
	var filter models.MicroprocessorFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	println(filter.Type, filter.Vendor)

	data, err := models.GetAllMicroprocessors(c, &filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, data)
}

func (controller *MicroprocessorController) GetMicroprocessorById(c *gin.Context) {
	microprocessorId, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
		return
	}

	data, err := models.GetMicroprocessorById(c, microprocessorId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, data)
}
