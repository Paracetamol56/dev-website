package controllers

import (
	"net/http"
	"os"

	"github.com/Paracetamol56/dev-website/api/models"
	"github.com/gin-gonic/gin"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuthController struct {
}

type LoginBody struct {
	Email string `json:"email" binding:"required,email"`
}

func (controller *AuthController) PostLogin(c *gin.Context) {
	var login LoginBody
	if err := c.ShouldBindJSON(&login); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	user, err := models.GetFullUserByEmail(c, login.Email)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if user == nil {
		result, err := models.CreateUser(c, &models.FullUser{
			Email:   login.Email,
			Flavour: "mocha",
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		user, err = models.GetFullUserById(c, result.InsertedID.(primitive.ObjectID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	verificationToken, err := models.SignRefreshToken(&user.Id, 1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	link := c.GetHeader("Origin") + "/verify?token=" + verificationToken + "&redirect=" + c.GetHeader("Referer")
	if err := SendVerificationEmail(c, user, link); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
}

func SendVerificationEmail(c *gin.Context, user *models.FullUser, url string) error {
	from := mail.NewEmail("Matheo Galuba", os.Getenv("ADMIN_EMAIL"))
	subject := "Verify your email address"
	to := mail.NewEmail(user.Name, user.Email)
	plainTextContent := "Please verify your email address by clicking the link below:\n" + url

	mail := mail.NewSingleEmail(from, subject, to, plainTextContent, "")
	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
	response, err := client.Send(mail)
	if err != nil || response.StatusCode != http.StatusAccepted {
		return err
	}
	return nil
}

func (controller *AuthController) GetVerify(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing token"})
		return
	}

	userId, err := models.ExtractID(token, os.Getenv("REFRESH_TOKEN_SECRET"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := models.GetFullUserById(c, userId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if user == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
		return
	}

	refreshtoken, err := models.SignRefreshToken(&user.Id, 168)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	accesstoken, err := models.SignAccessToken(&user.Id, 1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"refreshToken": refreshtoken,
		"accessToken":  accesstoken,
		"user":         user,
	})
}
