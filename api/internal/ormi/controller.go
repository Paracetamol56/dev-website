package ormi

import (
	"dev/internal/middlewares"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Controller struct{}

// GetTodos godoc
//
//	@Summary		List todos
//	@Description	List a page of the todos of the authenticated user, filtered by state, labels and date of last state change, along with the total number of matches
//	@Tags			ormi
//	@Produce		json
//	@Param			state	query		[]string	false	"Only return todos in these states (repeat the parameter for several)"	Enums(TODO, IN_PROGRESS, STANDBY, DONE, CANCELLED)	collectionFormat(multi)
//	@Param			label	query		[]string	false	"Only return todos having all these labels (repeat the parameter for several)"	collectionFormat(multi)
//	@Param			from	query		string		false	"Only return todos whose last state change is at or after this date"	Format(date-time)
//	@Param			to		query		string		false	"Only return todos whose last state change is before this date"			Format(date-time)
//	@Param			sort	query		string		false	"Custom order, due date (todos without one last), creation date or last state change"	Enums(position, due, created, updated)	default(position)
//	@Param			order	query		string		false	"Sort direction, defaults to desc for created and updated, asc otherwise"	Enums(asc, desc)
//	@Param			limit	query		int			false	"Maximum number of results"	default(50)	minimum(1)	maximum(200)
//	@Param			offset	query		int			false	"Number of results to skip"	default(0)	minimum(0)
//	@Success		200		{object}	TodoPage
//	@Failure		400		{object}	map[string]string
//	@Failure		401		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Security		Bearer
//	@Router			/ormi [get]
func (controller *Controller) GetTodos(c *gin.Context) {
	var filter TodoFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	todos, err := ListTodos(c, c.MustGet("x-user-id").(primitive.ObjectID), &filter)
	if err != nil {
		middlewares.RespondError(c, err)
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
//	@Success		200	{object}	Todo
//	@Failure		400	{object}	map[string]string
//	@Failure		401	{object}	map[string]string
//	@Failure		404	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Security		Bearer
//	@Router			/ormi/{id} [get]
func (controller *Controller) GetTodo(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	todo, err := GetTodo(c, c.MustGet("x-user-id").(primitive.ObjectID), id)
	if err != nil {
		middlewares.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, todo)
}

// PostTodo godoc
//
//	@Summary		Create a todo
//	@Description	Create a todo in the TODO state, placed at the end of the list
//	@Tags			ormi
//	@Accept			json
//	@Produce		json
//	@Param			todo	body		TodoBody	true	"Todo data"
//	@Success		201		{object}	Todo
//	@Failure		400		{object}	map[string]string
//	@Failure		401		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Security		Bearer
//	@Router			/ormi [post]
func (controller *Controller) PostTodo(c *gin.Context) {
	var body TodoBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	todo, err := AddTodo(c, c.MustGet("x-user-id").(primitive.ObjectID), body)
	if err != nil {
		middlewares.RespondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, todo)
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
//	@Success		200		{object}	Todo
//	@Failure		400		{object}	map[string]string
//	@Failure		401		{object}	map[string]string
//	@Failure		404		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Security		Bearer
//	@Router			/ormi/{id} [patch]
func (controller *Controller) PatchTodo(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var body PatchTodoBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	todo, err := ChangeTodo(c, c.MustGet("x-user-id").(primitive.ObjectID), id, body)
	if err != nil {
		middlewares.RespondError(c, err)
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
func (controller *Controller) DeleteTodo(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := RemoveTodo(c, c.MustGet("x-user-id").(primitive.ObjectID), id); err != nil {
		middlewares.RespondError(c, err)
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
func (controller *Controller) PutTodoOrder(c *gin.Context) {
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

	if err := OrderTodos(c, c.MustGet("x-user-id").(primitive.ObjectID), ids); err != nil {
		middlewares.RespondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

type TodoStatsQuery struct {
	Timezone string `form:"tz,default=UTC"`
}

// GetTodoStats godoc
//
//	@Summary		Get todo stats
//	@Description	Todos completed per day over the last year (days without completion are omitted), completion streaks and counts of open todos
//	@Tags			ormi
//	@Produce		json
//	@Param			tz	query		string	false	"IANA timezone used to split days"	default(UTC)
//	@Success		200	{object}	TodoStats
//	@Failure		400	{object}	map[string]string
//	@Failure		401	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Security		Bearer
//	@Router			/ormi/stats [get]
func (controller *Controller) GetTodoStats(c *gin.Context) {
	var query TodoStatsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	stats, err := Stats(c, c.MustGet("x-user-id").(primitive.ObjectID), query.Timezone)
	if err != nil {
		middlewares.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetTodoYearStats godoc
//
//	@Summary		Get completions of a year
//	@Description	Count the todos completed per day of a calendar year, days without completion are omitted
//	@Tags			ormi
//	@Produce		json
//	@Param			year	path		int		true	"Calendar year"
//	@Param			tz		query		string	false	"IANA timezone used to split days"	default(UTC)
//	@Success		200		{array}		TodoDayCount
//	@Failure		400		{object}	map[string]string
//	@Failure		401		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Security		Bearer
//	@Router			/ormi/stats/{year} [get]
func (controller *Controller) GetTodoYearStats(c *gin.Context) {
	year, err := strconv.Atoi(c.Param("year"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year"})
		return
	}
	var query TodoStatsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	days, err := YearStats(c, c.MustGet("x-user-id").(primitive.ObjectID), year, query.Timezone)
	if err != nil {
		middlewares.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, days)
}

type TodoLabelsQuery struct {
	Search string `form:"q" binding:"omitempty,max=50"`
}

// GetTodoLabels godoc
//
//	@Summary		List used labels
//	@Description	List the labels already used by the authenticated user, most used first, optionally restricted to those containing a text (case-insensitive)
//	@Tags			ormi
//	@Produce		json
//	@Param			q	query		string	false	"Text the labels must contain"
//	@Success		200	{array}		string
//	@Failure		400	{object}	map[string]string
//	@Failure		401	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Security		Bearer
//	@Router			/ormi/labels [get]
func (controller *Controller) GetTodoLabels(c *gin.Context) {
	var query TodoLabelsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	labels, err := SuggestLabels(c, c.MustGet("x-user-id").(primitive.ObjectID), query.Search)
	if err != nil {
		middlewares.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, labels)
}
