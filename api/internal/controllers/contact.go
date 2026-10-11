package controllers

import (
	"dev/internal/models"
	"dev/internal/utils"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ContactController struct{}

type ContactBody struct {
	UserId  primitive.ObjectID `json:"userId" binding:""`
	Name    string             `json:"name" binding:"required,min=2,max=100"`
	Email   string             `json:"email" binding:"required,email"`
	Message string             `json:"message" binding:"required,min=10,max=1000"`
}

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
func (controller *ContactController) PostContact(c *gin.Context) {
	var contact ContactBody
	if err := c.ShouldBindJSON(&contact); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	mailId, err := utils.SendEmail(c, utils.Email{
		To:          utils.EmailAddress{Name: "Admin", Email: os.Getenv("ADMIN_EMAIL")},
		ReplyTo:     &utils.EmailAddress{Name: contact.Name, Email: contact.Email},
		Subject:     "New message from dev.matheo-galuba.com",
		TextContent: "User: " + contact.UserId.String() + "\nName: " + contact.Name + "\nEmail: " + contact.Email + "\nMessage: " + contact.Message,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while sending email"})
		return
	}

	if _, err := models.CreateContact(c, models.Contact{
		UserId:    contact.UserId,
		Name:      contact.Name,
		Email:     contact.Email,
		Message:   contact.Message,
		MailId:    mailId,
		CreatedAt: primitive.NewDateTimeFromTime(time.Now()),
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Message sent"})
}
