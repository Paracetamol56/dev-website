package controllers

import (
	"net/http"
	"time"

	"github.com/Paracetamol56/dev-website/api/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type OrmiController struct {
}

type GetTodoQuery struct {
	State string `form:"state" binding:"omitempty" validate:"oneof=TODO IN_PROGRESS DONE CANCELLED"`
}

func (controller *OrmiController) GetTodos(c *gin.Context) {
	var query GetTodoQuery
	if err := c.ShouldBind(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userId := c.MustGet("x-user-id").(primitive.ObjectID)

	var todos []models.Todo
	var err error
	if query.State == "" {
		todos, err = models.GetTodosByUserId(c, userId)
	} else {
		todos, err = models.GetTodosByUserIdAndState(c, userId, query.State)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, todos)
}

func (controller *OrmiController) GetTodo(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	userId := c.MustGet("x-user-id").(primitive.ObjectID)

	todo, err := models.GetTodoByUserIdAndId(c, userId, id)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, todo)
}

type TodoBody struct {
	Title       string             `json:"title" binding:"required,min=2,max=100"`
	Description string             `json:"description" binding:"omitempty,max=1000"`
	DueDate     primitive.DateTime `json:"dueDate" binding:"omitempty"`
	Labels      []string           `json:"labels" binding:"omitempty"`
	GitURL      string             `json:"gitURL" binding:"omitempty,url"`
	GitIssue    uint32             `json:"gitIssue" binding:"omitempty,numeric,min=1"`
}

func (controller *OrmiController) PostTodo(c *gin.Context) {
	userId := c.MustGet("x-user-id").(primitive.ObjectID)

	var body TodoBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Due date should be in the future if provided
	if body.DueDate != 0 && body.DueDate.Time().Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Due date should be in the future"})
		return
	}

	result, err := models.CreateTodo(c, &models.Todo{
		User:        userId,
		Title:       body.Title,
		Description: body.Description,
		DueDate:     body.DueDate,
		Labels:      body.Labels,
		History: []models.OrmiEvent{{
			PreviousState: "",
			NewState:      "TODO",
			UpdatedAt:     primitive.NewDateTimeFromTime(time.Now()),
		}},
		GitURL:    body.GitURL,
		GitIssue:  body.GitIssue,
		CreatedAt: primitive.NewDateTimeFromTime(time.Now()),
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

type PatchTodoBody struct {
	Title       *string             `json:"title" binding:"omitempty,min=2,max=100"`
	Description *string             `json:"description" binding:"omitempty,max=1000"`
	DueDate     *primitive.DateTime `json:"dueDate" binding:"omitempty"`
	Labels      *[]string           `json:"labels" binding:"omitempty"`
	GitURL      *string             `json:"gitURL" binding:"omitempty,url"`
	GitIssue    *uint32             `json:"gitIssue" binding:"omitempty,numeric,min=1"`
	State       *string             `json:"state" binding:"omitempty,oneof=TODO IN_PROGRESS DONE CANCELLED"`
}

func (controller *OrmiController) PatchTodo(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	userId := c.MustGet("x-user-id").(primitive.ObjectID)

	var body PatchTodoBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.DueDate != nil && body.DueDate.Time().Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Due date should be in the future"})
		return
	}

	todo, err := models.GetTodoByUserIdAndId(c, userId, id)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if body.Title != nil {
		todo.Title = *body.Title
	}
	if body.Description != nil {
		todo.Description = *body.Description
	}
	if body.DueDate != nil {
		todo.DueDate = *body.DueDate
	}
	if body.Labels != nil {
		todo.Labels = *body.Labels
	}
	if body.GitURL != nil {
		todo.GitURL = *body.GitURL
	}
	if body.GitIssue != nil {
		todo.GitIssue = *body.GitIssue
	}
	if body.State != nil {
		if todo.State != *body.State {
			todo.History = append(todo.History, models.OrmiEvent{
				PreviousState: todo.State,
				NewState:      *body.State,
				UpdatedAt:     primitive.NewDateTimeFromTime(time.Now()),
			})
			todo.State = *body.State
		}
	}

	if _, err := models.UpdateTodo(c, todo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, todo)
}
