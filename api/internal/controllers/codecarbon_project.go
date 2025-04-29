package controllers

import (
	"dev/internal/models"
	"dev/internal/utils"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetCodeCarbonProjects godoc
//
// @Summary		Get codecarbon projects
// @Description	Get all codecarbon projects of the authenticated user
// @Tags			codecarbon
// @Query		limit	int		false	"Limit the number of results"	default(10)
// @Query		page	int		false	"Page number"	default(1)
// @Query		skip_total	bool	false	"Skip total count"	default(false)
// @Produce		json
// @Success	200	{object}	utils.PaginatedArrayResult[models.CodeCarbonProject]
// @Failure	403
// @Security		Bearer
// @Router /codecarbon/projects [get]
func (controller *CodeCarbonController) GetCodeCarbonProjects(c *gin.Context) {
	userId := c.MustGet("x-user-id").(primitive.ObjectID)
	var pagination utils.Pagination
	if err := c.ShouldBindQuery(&pagination); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	codeCarbonProjects, err := controller.projectRepo.GetCodeCarbonProjectsByUser(c, userId, &pagination)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get codecarbon projects"})
		return
	}

	c.JSON(http.StatusOK, codeCarbonProjects.ToPaginatedArrayResult(&pagination))
}

// GetCodeCarbonProjectById godoc
//
// @Summary		Get codecarbon project by id
// @Description	Get codecarbon project by id
// @Tags			codecarbon
// @Produce		json
// @Param			id	path		string	true	"Project ID"
// @Success	200	{object}	models.CodeCarbonProject
// @Failure	403
// @Failure	404
// @Security		Bearer
// @Router	/codecarbon/projects/{id} [get]
func (controller *CodeCarbonController) GetCodeCarbonProjectById(c *gin.Context) {
	projectId, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invalid id"})
		return
	}

	userId := c.MustGet("x-user-id").(primitive.ObjectID)

	codeCarbonProject, err := controller.projectRepo.GetCodeCarbonProjectById(c, userId, projectId)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get codecarbon project"})
		return
	}

	c.JSON(http.StatusOK, codeCarbonProject)
}

type PostCodeCarbonProjectBody struct {
	Name        string `json:"name" binding:"required,max=100"`
	Description string `json:"description" binding:"omitempty,max=1000"`
}

// PostCodeCarbonProject godoc
//
// @Summary		Create codecarbon project
// @Description	Create a new codecarbon project
// @Tags			codecarbon
// @Accept			json
// @Produce		json
// @Param			postProject body		PostCodeCarbonProjectBody	true	"Project data"
// @Success	201	{object}	models.CodeCarbonProject
// @Failure	400
// @Failure	403
// @Security		Bearer
// @Router	/codecarbon/projects [post]
func (controller *CodeCarbonController) PostCodeCarbonProject(c *gin.Context) {
	var body PostCodeCarbonProjectBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userId := c.MustGet("x-user-id").(primitive.ObjectID)

	codeCarbonProject := models.CodeCarbonProject{
		Id:          primitive.NewObjectID(),
		Name:        body.Name,
		Description: body.Description,
		Experiments: make([]models.CodeCarbonExperiment, 0),
		Tokens:      make([]models.CodeCarbonToken, 0),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	_, err := controller.projectRepo.CreateCodeCarbonProject(c, userId, &codeCarbonProject)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, codeCarbonProject)
}

type PatchCodeCarbonProjectBody struct {
	Name        string `json:"name" binding:"omitempty,max=100"`
	Description string `json:"description" binding:"omitempty,max=1000"`
}

// PatchCodeCarbonProject godoc
//
// @Summary		Update codecarbon project
// @Description	Update a codecarbon project by id
// @Tags			codecarbon
// @Accept			json
// @Produce		json
// @Param			id	path		string	true	"Project ID"
// @Param			patchProject body		PatchCodeCarbonProjectBody	true	"Project data"
// @Success	200	{object}	models.CodeCarbonProject
// @Failure	400
// @Failure	403
// @Failure	404
// @Security		Bearer
// @Router	/codecarbon/projects/{id} [patch]
func (controller *CodeCarbonController) PatchCodeCarbonProject(c *gin.Context) {
	projectId, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
		return
	}

	var body PatchCodeCarbonProjectBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userId := c.MustGet("x-user-id").(primitive.ObjectID)

	codeCarbonProject, err := controller.projectRepo.GetCodeCarbonProjectById(c, userId, projectId)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
			return
		}
		log.Println("Error getting codecarbon project:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get codecarbon project"})
		return
	}

	if body.Name != "" {
		codeCarbonProject.Name = body.Name
	}
	if body.Description != "" {
		codeCarbonProject.Description = body.Description
	}

	codeCarbonProject.UpdatedAt = time.Now()

	if _, err = controller.projectRepo.UpdateCodeCarbonProject(c, userId, codeCarbonProject); err != nil {
		log.Println("Error updating codecarbon project:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update codecarbon project"})
		return
	}

	c.JSON(http.StatusOK, codeCarbonProject)
}

// DeleteCodeCarbonProject godoc
//
// @Summary		Delete codecarbon project
// @Description	Delete a codecarbon project by id
// @Tags			codecarbon
// @Param			id	path		string	true	"Project ID"
// @Success	204
// @Failure	400
// @Failure	403
// @Failure	404
// @Security		Bearer
// @Router	/codecarbon/projects/{id} [delete]
func (controller *CodeCarbonController) DeleteCodeCarbonProject(c *gin.Context) {
	projectId, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
		return
	}

	userId := c.MustGet("x-user-id").(primitive.ObjectID)

	_, err = controller.projectRepo.DeleteCodeCarbonProject(c, userId, projectId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete codecarbon project"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
