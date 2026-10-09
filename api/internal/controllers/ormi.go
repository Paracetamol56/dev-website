package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"dev/internal/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type OrmiController struct {
}

type GetTodoQuery struct {
	State string `form:"state" binding:"omitempty,oneof=TODO IN_PROGRESS DONE CANCELLED"`
}

// GetTodos godoc
//
//	@Summary		List todos
//	@Description	List the todos of the authenticated user, in their custom order
//	@Tags			ormi
//	@Produce		json
//	@Param			state	query		string	false	"Only return todos in this state"	Enums(TODO, IN_PROGRESS, DONE, CANCELLED)
//	@Success		200		{array}		models.Todo
//	@Failure		400		{object}	map[string]string
//	@Failure		401		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Security		Bearer
//	@Router			/ormi [get]
func (controller *OrmiController) GetTodos(c *gin.Context) {
	var query GetTodoQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userId := c.MustGet("x-user-id").(primitive.ObjectID)

	todos, err := models.GetTodosByUserId(c, userId, query.State)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, todos)
}

// GetTodo godoc
//
//	@Summary		Get one todo
//	@Description	Get a todo of the authenticated user by id
//	@Tags			ormi
//	@Produce		json
//	@Param			id	path		string	true	"Todo ID"
//	@Success		200	{object}	models.Todo
//	@Failure		400	{object}	map[string]string
//	@Failure		401	{object}	map[string]string
//	@Failure		404	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Security		Bearer
//	@Router			/ormi/{id} [get]
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
	DueDate     primitive.DateTime `json:"dueDate" binding:"omitempty" swaggertype:"string" format:"date-time"`
	Labels      []string           `json:"labels" binding:"omitempty"`
	GitURL      string             `json:"gitURL" binding:"omitempty,url"`
	GitIssue    uint32             `json:"gitIssue" binding:"omitempty,numeric,min=1"`
}

// PostTodo godoc
//
//	@Summary		Create a todo
//	@Description	Create a todo in the TODO state, placed at the end of the list
//	@Tags			ormi
//	@Accept			json
//	@Produce		json
//	@Param			todo	body		TodoBody	true	"Todo data"
//	@Success		201		{object}	models.Todo
//	@Failure		400		{object}	map[string]string
//	@Failure		401		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Security		Bearer
//	@Router			/ormi [post]
func (controller *OrmiController) PostTodo(c *gin.Context) {
	userId := c.MustGet("x-user-id").(primitive.ObjectID)

	var body TodoBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.DueDate != 0 && isPastDueDate(body.DueDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Due date should not be in the past"})
		return
	}

	position, err := models.CountTodosByUserId(c, userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result, err := models.CreateTodo(c, &models.Todo{
		User:        userId,
		Title:       body.Title,
		Description: body.Description,
		DueDate:     body.DueDate,
		Labels:      body.Labels,
		State:       models.TodoStateTodo,
		Position:    int(position),
		History: []models.OrmiEvent{{
			PreviousState: "",
			NewState:      models.TodoStateTodo,
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

	todo, err := models.GetTodoByUserIdAndId(c, userId, result.InsertedID.(primitive.ObjectID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, todo)
}

// Clients send midnight in their own timezone, hence the day of slack
func isPastDueDate(dueDate primitive.DateTime) bool {
	return dueDate.Time().Before(time.Now().Add(-24 * time.Hour))
}

type OptionalDate struct {
	Set   bool
	Value primitive.DateTime
}

func (d *OptionalDate) UnmarshalJSON(data []byte) error {
	d.Set = true
	if bytes.Equal(data, []byte("null")) {
		d.Value = 0
		return nil
	}
	return json.Unmarshal(data, &d.Value)
}

type PatchTodoBody struct {
	Title       *string      `json:"title" binding:"omitempty,min=2,max=100"`
	Description *string      `json:"description" binding:"omitempty,max=1000"`
	DueDate     OptionalDate `json:"dueDate" swaggertype:"string" format:"date-time"`
	Labels      *[]string    `json:"labels" binding:"omitempty"`
	GitURL      *string      `json:"gitURL" binding:"omitempty,url"`
	GitIssue    *uint32      `json:"gitIssue" binding:"omitempty,numeric,min=1"`
	State       *string      `json:"state" binding:"omitempty,oneof=TODO IN_PROGRESS DONE CANCELLED"`
}

// PatchTodo godoc
//
//	@Summary		Update a todo
//	@Description	Update the given fields of a todo. A null dueDate removes the due date, a new state is appended to the history.
//	@Tags			ormi
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string			true	"Todo ID"
//	@Param			todo	body		PatchTodoBody	true	"Fields to update"
//	@Success		200		{object}	models.Todo
//	@Failure		400		{object}	map[string]string
//	@Failure		401		{object}	map[string]string
//	@Failure		404		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Security		Bearer
//	@Router			/ormi/{id} [patch]
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
	if body.DueDate.Set && body.DueDate.Value != todo.DueDate {
		if body.DueDate.Value != 0 && isPastDueDate(body.DueDate.Value) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Due date should not be in the past"})
			return
		}
		todo.DueDate = body.DueDate.Value
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

// DeleteTodo godoc
//
//	@Summary		Delete a todo
//	@Description	Delete a todo of the authenticated user by id
//	@Tags			ormi
//	@Param			id	path	string	true	"Todo ID"
//	@Success		204
//	@Failure		400	{object}	map[string]string
//	@Failure		401	{object}	map[string]string
//	@Failure		404	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Security		Bearer
//	@Router			/ormi/{id} [delete]
func (controller *OrmiController) DeleteTodo(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	userId := c.MustGet("x-user-id").(primitive.ObjectID)

	result, err := models.DeleteTodo(c, userId, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
		return
	}

	c.Status(http.StatusNoContent)
}

type TodoOrderBody struct {
	Ids []string `json:"ids" binding:"required,min=1,max=1000,dive,len=24,hexadecimal"`
}

// PutTodoOrder godoc
//
//	@Summary		Reorder todos
//	@Description	Set the position of the given todos to their index in the list
//	@Tags			ormi
//	@Accept			json
//	@Param			order	body	TodoOrderBody	true	"Todo IDs in their new order"
//	@Success		204
//	@Failure		400	{object}	map[string]string
//	@Failure		401	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Security		Bearer
//	@Router			/ormi/order [put]
func (controller *OrmiController) PutTodoOrder(c *gin.Context) {
	userId := c.MustGet("x-user-id").(primitive.ObjectID)

	var body TodoOrderBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ids := make([]primitive.ObjectID, len(body.Ids))
	for i, hex := range body.Ids {
		id, err := primitive.ObjectIDFromHex(hex)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}
		ids[i] = id
	}

	if err := models.ReorderTodos(c, userId, ids); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

type TodoStatsQuery struct {
	Timezone string `form:"tz,default=UTC"`
}

// GetTodoStats godoc
//
//	@Summary		Get completion stats
//	@Description	Count the todos completed per day over the last year, days without completion are omitted
//	@Tags			ormi
//	@Produce		json
//	@Param			tz	query		string	false	"IANA timezone used to split days"	default(UTC)
//	@Success		200	{array}		models.TodoDayCount
//	@Failure		400	{object}	map[string]string
//	@Failure		401	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Security		Bearer
//	@Router			/ormi/stats [get]
func (controller *OrmiController) GetTodoStats(c *gin.Context) {
	userId := c.MustGet("x-user-id").(primitive.ObjectID)

	var query TodoStatsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if _, err := time.LoadLocation(query.Timezone); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid timezone"})
		return
	}

	days, err := models.GetCompletedTodosPerDay(c, userId, time.Now().AddDate(-1, 0, -7), query.Timezone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, days)
}
