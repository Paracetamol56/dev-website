package controllers

import (
	"net/http"

	"github.com/Paracetamol56/dev-website/api/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ExportController struct {
}

func (controller *ExportController) GetExport(c *gin.Context) {
	userId := c.MustGet("x-user-id").(primitive.ObjectID)

	user, err := models.GetFullUserById(c, userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}
