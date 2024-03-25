package controllers

import (
	"net/http"
	"time"

	"github.com/Paracetamol56/dev-website/api/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrmiController struct {
}

func (controller *OrmiController) GetTodos(c *gin.Context) {
	userId := c.MustGet("x-user-id").(primitive.ObjectID)

	todos, err := models.GetTodosByUserId(c, userId)
	println(todos)
	println(err)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, todos)
}

type TodoBody struct {
	Title       string   `json:"title" binding:"required,min=2,max=100"`
	Description string   `json:"description" binding:"max=1000"`
	Labels      []string `json:"labels" binding:"omitempty"`
	GitURL      string   `json:"gitURL" binding:"omitempty,url"`
	GitIssue    uint32   `json:"gitIssue" binding:"omitempty,numeric,min=1"`
}

func (controller *OrmiController) PostTodo(c *gin.Context) {
	userId := c.MustGet("x-user-id").(primitive.ObjectID)

	var body TodoBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := models.CreateTodo(c, &models.Todo{
		User:        userId,
		Title:       body.Title,
		Description: body.Description,
		Labels:      body.Labels,
		GitURL:      body.GitURL,
		GitIssue:    body.GitIssue,
		CreatedAt:   primitive.NewDateTimeFromTime(time.Now()),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	todo, err := models.GetTodoById(c, result.InsertedID.(primitive.ObjectID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, todo)
}
