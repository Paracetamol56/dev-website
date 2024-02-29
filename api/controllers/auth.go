package controllers

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Paracetamol56/dev-website/api/models"
	"github.com/Paracetamol56/dev-website/api/utils"
	"github.com/gin-gonic/gin"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuthController struct {
}

// SendVerificationEmail godoc
// Sends a verification email to the user with the provided url
func SendVerificationEmail(c *gin.Context, user *models.User, url string) error {
	from := mail.NewEmail("Matheo Galuba", os.Getenv("ADMIN_EMAIL"))
	subject := "Verify your email address"
	to := mail.NewEmail(user.Name, user.Email)
	htmlContent := `
		<h1>Verify your email address</h1>
		<p>
			👋 Hi,<br>
			Thanks for signing up to my website!<br>
			Please verify your email address by clicking the link below.
		</p>

		<a href="` + url + `">Verify your email address</a>

		<p>
			Thanks,<br>
			Mathéo
		</p>

		<br>

		<h4>🌱 Why is this email ugly?</h4>
		<p>
			This email is voluntarily ugly because it's lightweight, so the environment impact is reduced..<br>
			By the way, this email is single-use, so you can delete it to avoid keeping it on someone's else hard drive 😎.
		</p>

		<p><small>If you didn't sign up to my website, please ignore this email.</small></p>
	`

	mail := mail.NewSingleEmail(from, subject, to, "", htmlContent)
	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
	response, err := client.Send(mail)
	if err != nil || response.StatusCode != http.StatusAccepted {
		return err
	}
	return nil
}

// AddContact godoc
// Adds the user to the SendGrid contact list
func AddContact(c *gin.Context, user *models.User) error {
	host := "https://api.sendgrid.com"
	request := sendgrid.GetRequest(os.Getenv("SENDGRID_API_KEY"), "/v3/marketing/contacts", host)
	request.Method = "PUT"
	request.Body = []byte(`{
		"list_ids": [
			"` + os.Getenv("SENDGRID_CONTACT_LIST_ID") + `"
		],
		"contacts": [
			{
				"email": "` + user.Email + `"
			}
		]
	}`)

	response, err := sendgrid.API(request)
	if err != nil || response.StatusCode != http.StatusAccepted {
		return err
	}

	return nil
}

// SignTokenPair godoc
// Signs a refresh token and an access token for the user with the provided id
func SignTokenPair(c *gin.Context, userId string) (string, string, error) {
	refreshtoken, err := utils.SignRefreshToken(userId, 168)
	if err != nil {
		return "", "", err
	}
	accesstoken, err := utils.SignAccessToken(userId, 1)
	if err != nil {
		return "", "", err
	}
	return refreshtoken, accesstoken, nil
}

type LoginBody struct {
	Email string `json:"email" binding:"required,email"`
}

// PostLogin godoc
//
//	@Summary		Login or register a user
//	@Description	Login or register a user by email
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		LoginBody	true	"Email"
//	@Success		200		{object}	models.FullUser
//	@Failure		400
//	@Router			/auth/login [post]
func (controller *AuthController) PostLogin(c *gin.Context) {
	var login LoginBody
	if err := c.ShouldBindJSON(&login); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, _ := models.GetFullUserByEmail(c, login.Email)

	if user == nil {
		result, err := models.CreateUser(c, &models.User{
			Email:   login.Email,
			Flavour: "mocha",
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		userId, _ := primitive.ObjectIDFromHex(result.InsertedID.(primitive.ObjectID).Hex())
		user, err = models.GetFullUserById(c, userId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	verificationToken, err := utils.SignRefreshToken(user.Id.Hex(), 1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	link := c.GetHeader("Origin") + "/verify/email?token=" + verificationToken + "&redirect=" + c.GetHeader("Referer")
	if err := SendVerificationEmail(c, user, link); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
}

type VerifyBody struct {
	Token string `json:"token" binding:"required"`
}

// PostVerify godoc
//
//	@Summary		Verify a user's email
//	@Description	Verify a user's email by token
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		VerifyBody	true	"Token"
//	@Success		200		{object}	models.FullUser
//	@Failure		400
//	@Router			/auth/verify [post]
func (controller *AuthController) PostVerify(c *gin.Context) {
	var verify VerifyBody
	if err := c.ShouldBindJSON(&verify); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	userId, err := utils.ExtractID(verify.Token, os.Getenv("REFRESH_TOKEN_SECRET"))
	log.Println("userId", userId)
	if err != nil {
		log.Println("err", err)
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

	AddContact(c, user)

	refreshtoken, accesstoken, err := SignTokenPair(c, userId.Hex())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	user.LastLogin = time.Now()
	user.LastRefresh = time.Now()
	if _, err = models.UpdateUser(c, userId, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"refreshToken": refreshtoken,
		"accessToken":  accesstoken,
		"user":         user,
	})
}

// PostRefresh godoc
//
//	@Summary		Refresh a user's access token
//	@Description	Refresh a user's access token by refresh token
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		VerifyBody	true	"Refresh token"
//	@Success		200		{object}	map[string]string
//	@Failure		400
//	@Router			/auth/refresh [post]
func (controller *AuthController) PostRefresh(c *gin.Context) {
	verrify := VerifyBody{}
	if err := c.ShouldBindJSON(&verrify); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	authorized, err := utils.IsAuthorized(verrify.Token, os.Getenv("REFRESH_TOKEN_SECRET"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !authorized {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userId, err := utils.ExtractID(verrify.Token, os.Getenv("REFRESH_TOKEN_SECRET"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIdString := userId.Hex()
	accesstoken, err := utils.SignAccessToken(userIdString, 1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	refreshtoken, err := utils.SignRefreshToken(userIdString, 168)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accessToken":  accesstoken,
		"refreshToken": refreshtoken,
	})
}

type GithubLoginBody struct {
	Code string `json:"code" binding:"required"`
}

// PostGithubLogin godoc
//
//	@Summary		Login or register a user with GitHub
//	@Description	Login or register a user with GitHub by code
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		GithubLoginBody	true	"Code"
//	@Success		200		{object}	map[string]interface{}
//	@Failure		400
//	@Router			/auth/github/login [post]
func (controller *AuthController) PostGithubLogin(c *gin.Context) {
	body := GithubLoginBody{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Get access token
	accessToken, err := utils.GetGithubAccessToken(body.Code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Fetch the GitHub user
	githubUser, err := utils.GetGithubUser(accessToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if the user exists in the database
	user, err := models.GetFullUserByEmail(c, githubUser.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if user == nil {
		// If not, create it
		result, err := models.CreateUser(c, &models.User{
			Email:             githubUser.Email,
			Name:              githubUser.Name,
			Flavour:           "mocha",
			ProfilePicture:    githubUser.AvatarUrl,
			GitHubAccessToken: accessToken,
			Github:            githubUser,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		userId, _ := primitive.ObjectIDFromHex(result.InsertedID.(primitive.ObjectID).Hex())
		user, err = models.GetFullUserById(c, userId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else if user.Github == nil {
		// If it exists but doesn't have a github account, update it
		user.GitHubAccessToken = accessToken
		user.Github = githubUser
	}

	user.LastLogin = time.Now()
	user.LastRefresh = time.Now()
	if _, err := models.UpdateUser(c, user.Id, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	accesstoken, err := utils.SignAccessToken(user.Id.Hex(), 1)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	refreshtoken, err := utils.SignRefreshToken(user.Id.Hex(), 168)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accessToken":  accesstoken,
		"refreshToken": refreshtoken,
		"user":         user,
	})
}
