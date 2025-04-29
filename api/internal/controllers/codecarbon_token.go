package controllers

import (
	"dev/internal/models"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func generateToken() string {
	charSet := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 28)
	for i := range b {
		b[i] = charSet[rand.Intn(len(charSet))]
	}
	return "cc_" + string(b[:28])
}

// GetCodeCarbonProjectTokens godoc
//
// @Summary		Get codecarbon project tokens
// @Description	Get all codecarbon project tokens of the authenticated user
// @Tags			codecarbon token
// @Produce		json
// @Success	200	{array}	models.CodeCarbonToken
// @Failure	403
// @Failure	404
// @Router	/codecarbon/projects/{id}/tokens [get]
// @Security	Bearer
func (controller *CodeCarbonController) GetCodeCarbonProjectTokens(c *gin.Context) {
	projectId, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
		return
	}

	userId := c.MustGet("x-user-id").(primitive.ObjectID)

	if _, err := controller.projectRepo.GetCodeCarbonProjectById(c, userId, projectId); err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get codecarbon project"})
		}
		return
	}

	projectTokens, err := models.GetCodeCarbonTokensByProject(c, userId, projectId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get codecarbon project tokens"})
		return
	}

	c.JSON(http.StatusOK, projectTokens)
}

type PostCodeCarbonProjectTokenBody struct {
	Name string `json:"name" binding:"required,min=1,max=100"`
}

// PostCodeCarbonProjectToken godoc
//
// @Summary		Create codecarbon project token
// @Description	Create codecarbon project token
// @Tags			codecarbon token
// @Accept			json
// @Produce		json
// @Param			id	path		string	true	"Project ID"
// @Param			body	body		models.CodeCarbonToken	true	"Token object"
// @Success	200	{object}	models.CodeCarbonToken
// @Failure	400
// @Failure	403
// @Failure	404
// @Router	/codecarbon/projects/{id}/tokens [post]
// @Security	Bearer
func (controller *CodeCarbonController) PostCodeCarbonProjectToken(c *gin.Context) {
	projectId, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
		return
	}

	userId := c.MustGet("x-user-id").(primitive.ObjectID)

	var body PostCodeCarbonProjectTokenBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	project, err := controller.projectRepo.GetCodeCarbonProjectById(c, userId, projectId)
	if err != nil || project == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	token := models.CodeCarbonToken{
		Id:        primitive.NewObjectID(),
		Name:      body.Name,
		Token:     generateToken(),
		CreatedAt: time.Now(),
	}

	_, err = models.CreateCodeCarbonToken(c, userId, projectId, &token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create codecarbon project token"})
		return
	}

	c.JSON(http.StatusOK, token)
}

// DeleteCodeCarbonProjectToken godoc
//
// @Summary		Delete codecarbon project token
// @Description	Delete codecarbon project token
// @Tags			codecarbon token
// @Produce		json
// @Param			id	path		string	true	"Project ID"
// @Param			tokenId	path		string	true	"Token ID"
// @Success	200	{object}	models.CodeCarbonToken
// @Failure	400
// @Failure	403
// @Router	/codecarbon/projects/{id}/tokens/{tokenId} [delete]
// @Security	Bearer
func (controller *CodeCarbonController) DeleteCodeCarbonProjectToken(c *gin.Context) {
	projectId, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
		return
	}

	userId := c.MustGet("x-user-id").(primitive.ObjectID)

	tokenId, err := primitive.ObjectIDFromHex(c.Param("token_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token id"})
		return
	}

	_, err = models.RevokeToken(c, userId, projectId, tokenId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete codecarbon project token"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
