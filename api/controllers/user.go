package controllers

import (
	"net/http"

	"github.com/Paracetamol56/dev-website/api/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserController struct {
}

func (controller *UserController) GetUser(c *gin.Context) {
	userIdString := c.Param("id")
	userId, err := primitive.ObjectIDFromHex(userIdString)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
		return
	}

	if userIdString == c.MustGet("x-user-id").(string) {
		user, err := models.GetFullUserById(c, userId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, user)
	} else {
		user, err := models.GetUserById(c, userId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, user)
	}
}

func (controller *UserController) PatchUser(c *gin.Context) {
	userIdString := c.Param("id")
	_, err := primitive.ObjectIDFromHex(userIdString)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
		return
	}

	// Check if user is authorized
	if userIdString != c.MustGet("x-user-id").(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (controller *UserController) DeleteUser(c *gin.Context) {
	userIdString := c.Param("id")
	userId, err := primitive.ObjectIDFromHex(userIdString)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
		return
	}

	// Check if user is authorized
	if userIdString != c.MustGet("x-user-id").(string) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	result, err := models.DeleteUser(c, userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
