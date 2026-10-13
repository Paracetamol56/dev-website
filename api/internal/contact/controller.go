package contact

import (
	"dev/internal/middlewares"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Controller struct{}

// PostContact godoc
//
//	@Summary		Send a contact message
//	@Description	Send a contact message
//	@Tags			contact
//	@Accept			json
//	@Produce		json
//	@Param			contact	body	ContactBody	true	"Contact message"
//	@Success		201
//	@Failure		400
//	@Router			/contact [post]
func (controller *Controller) PostContact(c *gin.Context) {
	var body ContactBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := SendMessage(c, body); err != nil {
		middlewares.RespondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Message sent"})
}
