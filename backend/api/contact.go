package api

import (
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Contact struct {
	UserId  primitive.ObjectID `json:"userId" validate:""`
	Name    string             `json:"name" validate:"required,min=2,max=100"`
	Email   string             `json:"email" validate:"required,email"`
	Message string             `json:"message" validate:"required,min=10,max=1000"`
}

func PostContact(c *gin.Context) {
	var contact Contact
	validate := validator.New()
	if err := c.ShouldBindJSON(&contact); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err := validate.Struct(contact); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var userId interface{}
	if contact.UserId.IsZero() {
		userId = nil
	} else {
		userId = contact.UserId
	}

	from := mail.NewEmail(contact.Name, "matheo.galu56@gmail.com")
	subject := "New message from dev.matheo-galuba.com"
	to := mail.NewEmail("Admin", os.Getenv("ADMIN_EMAIL"))
	plainTextContent := "User: " + contact.UserId.String() + "\nName: " + contact.Name + "\nEmail: " + contact.Email + "\nMessage: " + contact.Message
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, "")
	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
	response, err := client.Send(message)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	if response.StatusCode != 202 {
		c.JSON(500, gin.H{"error": "Error while sending email"})
		return
	}

	collection := mongoClient.Database("dev").Collection("contacts")
	_, err = collection.InsertOne(c, bson.M{
		"userId":    userId,
		"name":      contact.Name,
		"email":     contact.Email,
		"message":   contact.Message,
		"mailId":    response.Headers["X-Message-Id"][0],
		"createdAt": time.Now(),
	})
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"response": response})

}
