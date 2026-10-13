package user

import (
	"dev/internal/middlewares"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Controller struct{}

// GetUsers godoc
//
//	@Summary		Get one user
//	@Description	Get one user by id
//	@Note			Returns a full user if the user is the same as the one making the request
//	@Tags			user
//	@Produce		json
//	@Param			id	path		string	true	"User ID"
//	@Success		200	{object}	User
//	@Failure		400
//	@Failure		404
//	@Security		Bearer
//	@Router			/users/{id} [get]
func (controller *Controller) GetUser(c *gin.Context) {
	userId, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
		return
	}

	user, err := GetProfile(c, c.MustGet("x-user-id").(primitive.ObjectID), userId)
	if err != nil {
		middlewares.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

// PatchUser godoc
//
//	@Summary		Update a user
//	@Description	Update a user by id
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Param			id			path		string			true	"User ID"
//	@Param			patchUser	body		PatchUserBody	true	"User data"
//	@Success		200			{object}	User
//	@Failure		400
//	@Failure		403
//	@Failure		404
//	@Security		Bearer
//	@Router			/users/{id} [patch]
func (controller *Controller) PatchUser(c *gin.Context) {
	userId, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
		return
	}

	var body PatchUserBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := UpdateProfile(c, c.MustGet("x-user-id").(primitive.ObjectID), userId, body)
	if err != nil {
		middlewares.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

// DeleteIdentity godoc
//
//	@Summary		Unlink an identity
//	@Description	Unlink an identity provider from a user. The email identity cannot be removed, so the user can always log in with a magic link.
//	@Tags			user
//	@Produce		json
//	@Param			id			path		string	true	"User ID"
//	@Param			provider	path		string	true	"Identity provider"	Enums(github, google)
//	@Success		200			{object}	User
//	@Failure		400			{object}	map[string]string
//	@Failure		403			{object}	map[string]string
//	@Failure		404			{object}	map[string]string
//	@Security		Bearer
//	@Router			/users/{id}/identities/{provider} [delete]
func (controller *Controller) DeleteIdentity(c *gin.Context) {
	userId, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
		return
	}

	user, err := UnlinkIdentity(c, c.MustGet("x-user-id").(primitive.ObjectID), userId, c.Param("provider"))
	if err != nil {
		middlewares.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

// DeleteUser godoc
//
//	@Summary		Delete a user
//	@Description	Delete a user by id
//	@Tags			user
//	@Produce		json
//	@Param			id	path	string	true	"User ID"
//	@Success		204
//	@Failure		400
//	@Failure		403
//	@Failure		404
//	@Security		Bearer
//	@Router			/users/{id} [delete]
func (controller *Controller) DeleteUser(c *gin.Context) {
	userId, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
		return
	}

	if err := DeleteAccount(c, c.MustGet("x-user-id").(primitive.ObjectID), userId); err != nil {
		middlewares.RespondError(c, err)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// GetExport godoc
//
//	@Summary		Get user export
//	@Description	Get user export by id
//	@Tags			user
//	@Produce		json
//	@Param			id	path		string	true	"User ID"
//	@Success		200	{object}	User
//	@Failure		400
//	@Failure		403
//	@Failure		404
//	@Security		Bearer
//	@Router			/users/{id}/export [get]
func (controller *Controller) GetExport(c *gin.Context) {
	userId, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
		return
	}

	user, err := ExportData(c, c.MustGet("x-user-id").(primitive.ObjectID), userId)
	if err != nil {
		middlewares.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}
