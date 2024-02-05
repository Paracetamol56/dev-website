package controllers

import (
	"net/http"

	"github.com/Paracetamol56/dev-website/api/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserController struct {
}

// GetUsers godoc
//	@Summary		Get one user
//	@Description	Get one user by id
//	@Note			Returns a full user if the user is the same as the one making the request
//	@Tags			user
//	@Produce		json
//	@Param			id	path		string	true	"User ID"
//	@Success		200	{object}	models.FullUser
//	@Failure		400
//	@Failure		404
//	@Security		Bearer
//	@Router			/user/{id} [get]
func (controller *UserController) GetUser(c *gin.Context) {
	userId, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
		return
	}

	if userId == c.MustGet("x-user-id").(primitive.ObjectID) {
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

type PatchUserBody struct {
	Name           string `json:"name" binding:"omitempty,min=2,max=100"`
	Flavour        string `json:"flavour" binding:"omitempty,oneof=latte frappe macchiato mocha"`
	ProfilePicture string `json:"profilePicture" binding:"omitempty,url"`
}

// PatchUser godoc
//	@Summary		Update a user
//	@Description	Update a user by id
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Param			id			path		string			true	"User ID"
//	@Param			patchUser	body		PatchUserBody	true	"User data"
//	@Success		200			{object}	models.FullUser
//	@Failure		400
//	@Failure		403
//	@Failure		404
//	@Security		Bearer
//	@Router			/user/{id} [patch]
func (controller *UserController) PatchUser(c *gin.Context) {
	userId, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
		return
	}

	var patchUser PatchUserBody
	if err := c.ShouldBindJSON(&patchUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user is authorized
	if userId != c.MustGet("x-user-id").(primitive.ObjectID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	user, err := models.GetFullUserById(c, userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if patchUser.Name != "" {
		user.Name = patchUser.Name
	}
	if patchUser.Flavour != "" {
		user.Flavour = patchUser.Flavour
	}
	if patchUser.ProfilePicture != "" {
		user.ProfilePicture = patchUser.ProfilePicture
	}

	if _, err := models.UpdateUser(c, userId, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// DeleteUser godoc
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
//	@Router			/user/{id} [delete]
func (controller *UserController) DeleteUser(c *gin.Context) {
	userIdString := c.Param("id")
	userId, err := primitive.ObjectIDFromHex(userIdString)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
		return
	}

	// Check if user is authorized
	if userId != c.MustGet("x-user-id").(primitive.ObjectID) {
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

// GetExport godoc
//	@Summary		Get user export
//	@Description	Get user export by id
//	@Tags			user
//	@Produce		json
//	@Param			id	path		string	true	"User ID"
//	@Success		200	{object}	models.FullUser
//	@Failure		400
//	@Failure		403
//	@Failure		404
//	@Security		Bearer
//	@Router			/user/{id}/export [get]
func (controller *UserController) GetExport(c *gin.Context) {
	userIdString := c.Param("id")
	userId, err := primitive.ObjectIDFromHex(userIdString)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
		return
	}

	// Check if user is authorized
	if userId != c.MustGet("x-user-id").(primitive.ObjectID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	user, err := models.GetFullUserById(c, userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}
