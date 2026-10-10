package controllers

import (
	"dev/internal/models"
	"dev/internal/utils"
	"errors"
	"html"
	"log"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuthController struct{}

// SendVerificationEmail godoc
// Sends a verification email to the user with the provided url
func SendVerificationEmail(c *gin.Context, user *models.User, url string) error {
	_, err := utils.SendEmail(c, utils.Email{
		To:      utils.EmailAddress{Name: user.Name, Email: user.Email},
		Subject: "Verify your email address",
		HTMLContent: `
		<h1>Verify your email address</h1>
		<p>
			👋 Hi,<br>
			Thanks for signing up to my website!<br>
			Please verify your email address by clicking the link below.
		</p>

		<a href="` + html.EscapeString(url) + `">Verify your email address</a>

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
	`,
	})
	return err
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

func verificationLink(origin string, referer string, token string) (string, error) {
	if !slices.Contains(utils.AllowedOrigins, origin) {
		return "", errors.New("origin not allowed")
	}

	redirect := "/"
	if parsed, err := url.Parse(referer); err == nil && parsed.Path != "" {
		redirect = parsed.RequestURI()
	}
	// "//host" and "/\host" are read by browsers as another host
	if !strings.HasPrefix(redirect, "/") || strings.HasPrefix(redirect, "//") || strings.HasPrefix(redirect, "/\\") {
		redirect = "/"
	}

	query := url.Values{"token": {token}, "redirect": {redirect}}
	return origin + "/verify/email?" + query.Encode(), nil
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
//	@Success		204
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

	verificationToken, err := models.CreateLoginToken(c, user.Id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	link, err := verificationLink(c.GetHeader("Origin"), c.GetHeader("Referer"), verificationToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := SendVerificationEmail(c, user, link); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
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
//	@Success		200		{object}	models.User
//	@Failure		400
//	@Router			/auth/verify [post]
func (controller *AuthController) PostVerify(c *gin.Context) {
	var verify VerifyBody
	if err := c.ShouldBindJSON(&verify); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	userId, err := models.ConsumeLoginToken(c, verify.Token)
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

	if err := utils.AddEmailContact(c, user.Email); err != nil {
		log.Printf("Failed to add %s to the contact list: %v", user.Email, err)
	}

	refreshtoken, accesstoken, err := SignTokenPair(c, userId.Hex())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	user.SetIdentity(models.Identity{
		Provider:   models.IdentityProviderEmail,
		ProviderId: user.Email,
		Email:      user.Email,
	})
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

type OAuthProviderResponse struct {
	Name         string `json:"name" example:"github"`
	ClientId     string `json:"clientId"`
	AuthorizeUrl string `json:"authorizeUrl" example:"https://github.com/login/oauth/authorize"`
	Scope        string `json:"scope" example:"read:user user:email"`
}

// GetOAuthProviders godoc
//
//	@Summary		List OAuth providers
//	@Description	List the identity providers configured on this server, with what a client needs to start their authorization flow
//	@Tags			auth
//	@Produce		json
//	@Success		200	{array}	OAuthProviderResponse
//	@Router			/auth/providers [get]
func (controller *AuthController) GetOAuthProviders(c *gin.Context) {
	providers := []OAuthProviderResponse{}
	for _, provider := range utils.GetOAuthProviders() {
		providers = append(providers, OAuthProviderResponse{
			Name:         provider.Name,
			ClientId:     provider.ClientId(),
			AuthorizeUrl: provider.AuthorizeURL,
			Scope:        provider.Scope,
		})
	}
	c.JSON(http.StatusOK, providers)
}

type OAuthLoginBody struct {
	Code        string `json:"code" binding:"required"`
	RedirectUri string `json:"redirectUri" binding:"omitempty,url"`
}

// PostOAuthLogin godoc
//
//	@Summary		Login, register or link an identity with an OAuth provider
//	@Description	Exchange an authorization code for the identity of its owner and attach it to the user having the same email, creating the user if needed. When called with a bearer token, the identity must have the email of the authenticated user.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			provider	path		string			true	"Identity provider"	Enums(github, google)
//	@Param			body		body		OAuthLoginBody	true	"Authorization code and the redirect URI it was issued for"
//	@Success		200			{object}	map[string]interface{}
//	@Failure		400			{object}	map[string]string
//	@Failure		401			{object}	map[string]string
//	@Failure		404			{object}	map[string]string
//	@Failure		409			{object}	map[string]string
//	@Router			/auth/{provider} [post]
func (controller *AuthController) PostOAuthLogin(c *gin.Context) {
	provider := utils.GetOAuthProvider(c.Param("provider"))
	if provider == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Unknown identity provider"})
		return
	}

	body := OAuthLoginBody{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user *models.User
	if authHeader := c.GetHeader("Authorization"); authHeader != "" {
		userId, err := utils.ExtractID(strings.TrimPrefix(authHeader, "Bearer "), os.Getenv("ACCESS_TOKEN_SECRET"))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		if user, err = models.GetFullUserById(c, userId); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
	}

	identity, err := provider.Authenticate(c, body.Code, body.RedirectUri)
	if err != nil {
		log.Printf("Failed to authenticate with %s: %v", provider.Name, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to authenticate with " + provider.Name})
		return
	}

	if user != nil {
		if !strings.EqualFold(user.Email, identity.Email) {
			c.JSON(http.StatusConflict, gin.H{"error": "This account uses another email address (" + identity.Email + ")"})
			return
		}
	} else if user, err = models.GetFullUserByEmail(c, identity.Email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if user == nil {
		result, err := models.CreateUser(c, &models.User{
			Email:   identity.Email,
			Flavour: "mocha",
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if user, err = models.GetFullUserById(c, result.InsertedID.(primitive.ObjectID)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if err := utils.AddEmailContact(c, user.Email); err != nil {
			log.Printf("Failed to add %s to the contact list: %v", user.Email, err)
		}
	}

	if err := models.DetachIdentity(c, identity.Provider, identity.ProviderId, user.Id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	user.SetIdentity(*identity)
	if user.Name == "" {
		user.Name = identity.Name
	}
	if user.ProfilePicture == "" {
		user.ProfilePicture = identity.AvatarUrl
	}
	user.LastLogin = time.Now()
	user.LastRefresh = time.Now()
	if _, err := models.UpdateUser(c, user.Id, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	refreshtoken, accesstoken, err := SignTokenPair(c, user.Id.Hex())
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
