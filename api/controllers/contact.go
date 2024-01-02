package controllers

import (
	"net/http"
	"os"
	"time"

	"github.com/Paracetamol56/dev-website/api/models"
	"github.com/gin-gonic/gin"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ContactController struct {
}

type ContactBody struct {
	UserId  primitive.ObjectID `json:"userId" binding:""`
	Name    string             `json:"name" binding:"required,min=2,max=100"`
	Email   string             `json:"email" binding:"required,email"`
	Message string             `json:"message" binding:"required,min=10,max=1000"`
}

func (controller *ContactController) PostContact(c *gin.Context) {
	var contact ContactBody
	if err := c.ShouldBindJSON(&contact); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	from := mail.NewEmail(contact.Name, "matheo.galu56@gmail.com")
	subject := "New message from dev.matheo-galuba.com"
	to := mail.NewEmail("Admin", os.Getenv("ADMIN_EMAIL"))
	plainTextContent := "User: " + contact.UserId.String() + "\nName: " + contact.Name + "\nEmail: " + contact.Email + "\nMessage: " + contact.Message

	mail := mail.NewSingleEmail(from, subject, to, plainTextContent, "")
	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
	response, err := client.Send(mail)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if response.StatusCode != http.StatusAccepted {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while sending email"})
		return
	}

	if _, err := models.CreateContact(c, models.Contact{
		UserId:    contact.UserId,
		Name:      contact.Name,
		Email:     contact.Email,
		Message:   contact.Message,
		MailId:    response.Headers["X-Message-Id"][0],
		CreatedAt: primitive.NewDateTimeFromTime(time.Now()),
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Message sent"})
}
