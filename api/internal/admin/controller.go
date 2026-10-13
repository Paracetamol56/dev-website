package admin

import (
	"dev/internal/middlewares"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Controller struct{}

func userId(c *gin.Context) (primitive.ObjectID, bool) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
		return id, false
	}
	return id, true
}

// GetUsers godoc
//
//	@Summary		List users
//	@Description	List the users whose email or name contains the search, newest first
//	@Tags			admin
//	@Produce		json
//	@Param			search	query		string	false	"Part of an email or a name"
//	@Param			page	query		int		false	"Page, from 1"
//	@Param			limit	query		int		false	"Users per page, 25 by default, 100 at most"
//	@Success		200		{object}	UsersPage
//	@Failure		400		{object}	map[string]string
//	@Failure		403		{object}	map[string]string
//	@Security		Bearer
//	@Router			/admin/users [get]
func (controller *Controller) GetUsers(c *gin.Context) {
	var query UsersQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	page, err := ListUsers(c, query)
	if err != nil {
		middlewares.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, page)
}

// GetUser godoc
//
//	@Summary		Get a user
//	@Description	Get the whole account of any user
//	@Tags			admin
//	@Produce		json
//	@Param			id	path		string	true	"User ID"
//	@Success		200	{object}	user.User
//	@Failure		400	{object}	map[string]string
//	@Failure		403	{object}	map[string]string
//	@Failure		404	{object}	map[string]string
//	@Security		Bearer
//	@Router			/admin/users/{id} [get]
func (controller *Controller) GetUser(c *gin.Context) {
	id, ok := userId(c)
	if !ok {
		return
	}

	user, err := GetUser(c, id)
	if err != nil {
		middlewares.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

// PatchUserRole godoc
//
//	@Summary		Change the role of a user
//	@Description	Make a user an admin, or a regular user with an empty role. Admins cannot change their own role nor the one of the site owner.
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string		true	"User ID"
//	@Param			body	body		RoleBody	true	"Role"
//	@Success		200		{object}	user.User
//	@Failure		400		{object}	map[string]string
//	@Failure		403		{object}	map[string]string
//	@Failure		404		{object}	map[string]string
//	@Security		Bearer
//	@Router			/admin/users/{id}/role [patch]
func (controller *Controller) PatchUserRole(c *gin.Context) {
	id, ok := userId(c)
	if !ok {
		return
	}
	var body RoleBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := ChangeRole(c, c.MustGet("x-user-id").(primitive.ObjectID), id, body.Role)
	if err != nil {
		middlewares.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

// DeleteUser godoc
//
//	@Summary		Delete a user
//	@Description	Delete the account of a user, who is told by email. Admins cannot delete their own account here nor the one of the site owner.
//	@Tags			admin
//	@Param			id	path	string	true	"User ID"
//	@Success		204
//	@Failure		400	{object}	map[string]string
//	@Failure		403	{object}	map[string]string
//	@Failure		404	{object}	map[string]string
//	@Security		Bearer
//	@Router			/admin/users/{id} [delete]
func (controller *Controller) DeleteUser(c *gin.Context) {
	id, ok := userId(c)
	if !ok {
		return
	}

	if err := DeleteUser(c, c.MustGet("x-user-id").(primitive.ObjectID), id); err != nil {
		middlewares.RespondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// GetAudit godoc
//
//	@Summary		List audit entries
//	@Description	List what users did, newest first. With a user, the entries are the ones of that user and the ones about them.
//	@Tags			admin
//	@Produce		json
//	@Param			userId	query		string	false	"User ID"
//	@Param			action	query		string	false	"Action, such as auth.login"
//	@Param			page	query		int		false	"Page, from 1"
//	@Param			limit	query		int		false	"Entries per page, 25 by default, 100 at most"
//	@Success		200		{object}	AuditPage
//	@Failure		400		{object}	map[string]string
//	@Failure		403		{object}	map[string]string
//	@Security		Bearer
//	@Router			/admin/audit [get]
func (controller *Controller) GetAudit(c *gin.Context) {
	var query AuditQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	page, err := ListAudit(c, query)
	if err != nil {
		middlewares.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, page)
}
