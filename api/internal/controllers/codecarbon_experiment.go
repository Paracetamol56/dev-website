package controllers

import (
	"dev/internal/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetCodeCarbonProjectExperiments godoc
//
// @Summary		Get codecarbon project experiments
// @Description 	Get all codecarbon project experiments of the authenticated user
// @Tags			codecarbon experiment
// @Produce		json
// @Success	200	{array}	models.CodeCarbonExperiment
// @Failure	403
// @Failure	404
// @Router	/codecarbon/projects/{id}/experiments [get]
// @Security	Bearer
func (controller *CodeCarbonController) GetCodeCarbonProjectExperiments(c *gin.Context) {
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

	projectExperiments, err := models.GetCodeCarbonExperimentsByProject(c, userId, projectId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get codecarbon project tokens"})
		return
	}

	c.JSON(http.StatusOK, projectExperiments)
}

type PostCodeCarbonExperimentBody struct {
	Name           string `json:"name" binding:"required,min=1,max=100"`
	Description    string `json:"description" binding:"min=1,max=1000"`
	CountryIsoCode string `json:"country_iso_code" binding:"required,min=2,max=2"`
	Region         string `json:"region" binding:"required,min=1,max=100"`
	OnCloud        bool   `json:"on_cloud" binding:"required"`
	CloudProvider  string `json:"cloud_provider" binding:"omitempty,min=1,max=100"`
	CloudRegion    string `json:"cloud_region" binding:"omitempty,min=1,max=100"`
}

// PostCodeCarbonProjectExperiment godoc
//
// @Summary		Create a new codecarbon project experiment
// @Description	Create a new codecarbon project experiment for the authenticated user
// @Tags			codecarbon experiment
// @Accept		json
// @Produce		json
// @Param		id	path		string	true	"Project ID"
// @Param		body	body		models.CodeCarbonExperiment	true	"Experiment object"
// @Success	201	{object}	models.CodeCarbonExperiment
// @Failure	400
// @Failure	403
// @Failure	404
// @Router	/codecarbon/projects/{id}/experiments [post]
// @Security	Bearer
func (controller *CodeCarbonController) PostCodeCarbonProjectExperiment(c *gin.Context) {
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

	var body PostCodeCarbonExperimentBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	experiment := models.CodeCarbonExperiment{
		Id:             primitive.NewObjectID(),
		Name:           body.Name,
		Description:    body.Description,
		CountryIsoCode: body.CountryIsoCode,
		Region:         body.Region,
		OnCloud:        body.OnCloud,
		CloudProvider:  body.CloudProvider,
		CloudRegion:    body.CloudRegion,
		Runs:           make([]models.CodeCarbonRun, 0),
		CreatedAt:      time.Now(),
	}

	if result, err := models.CreateCodeCarbonExperiment(c, userId, projectId, &experiment); err != nil || result.ModifiedCount != 1 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create codecarbon experiment"})
		return
	}

	c.JSON(http.StatusCreated, experiment)
}

type PatchCodeCarbonExperimentBody struct {
	Name        *string `json:"name,omitempty" binding:"omitempty,min=1,max=100"`
	Description *string `json:"description,omitempty" binding:"omitempty,min=1,max=1000"`
}

// PatchCodeCarbonProjectExperiment godoc
//
// @Summary		Update a codecarbon project experiment
// @Description	Update a codecarbon project experiment for the authenticated user
// @Tags			codecarbon experiment
// @Accept		json
// @Produce		json
// @Param		id	path		string	true	"Project ID"
// @Param		experiment_id	path	string	true	"Experiment ID"
// @Param		body	body		models.CodeCarbonExperiment	true	"Experiment object"
// @Success	200	{object}	models.CodeCarbonExperiment
// @Failure	400
// @Failure	403
// @Failure	404
// @Router	/codecarbon/projects/{id}/experiments/{experiment_id} [patch]
// @Security	Bearer
func (controller *CodeCarbonController) PatchCodeCarbonProjectExperiment(c *gin.Context) {
	projectId, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
		return
	}

	var body PatchCodeCarbonExperimentBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
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

	experimentId, err := primitive.ObjectIDFromHex(c.Param("experiment_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid experiment id"})
		return
	}

	experiment, err := models.GetCodeCarbonExperimentsByProjectAndId(c, userId, projectId, experimentId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Experiment not found"})
		return
	}

	if body.Name != nil {
		experiment.Name = *body.Name
	}
	if body.Description != nil {
		experiment.Description = *body.Description
	}

	if _, err := models.UpdateCodeCarbonExperiment(c, userId, projectId, experiment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update codecarbon experiment"})
		return
	}

	c.JSON(http.StatusOK, experiment)
}

// DeleteCodeCarbonProjectExperiment godoc
//
// @Summary		Delete a codecarbon project experiment
// @Description	Delete a codecarbon project experiment for the authenticated user
// @Tags			codecarbon experiment
// @Accept		json
// @Produce		json
// @Param		id	path		string	true	"Project ID"
// @Param		experiment_id	path	string	true	"Experiment ID"
// @Param		body	body		models.CodeCarbonExperiment	true	"Experiment object"
// @Success	204
// @Failure	400
// @Failure	403
// @Failure	404
// @Router	/codecarbon/projects/{id}/experiments/{experiment_id} [delete]
// @Security	Bearer
func (controller *CodeCarbonController) DeleteCodeCarbonProjectExperiment(c *gin.Context) {
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

	experimentId, err := primitive.ObjectIDFromHex(c.Param("experiment_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid experiment id"})
		return
	}

	_, err = models.DeleteCodeCarbonExperiment(c, userId, projectId, experimentId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete codecarbon experiment"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
