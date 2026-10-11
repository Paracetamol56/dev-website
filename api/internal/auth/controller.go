package auth

import (
	"dev/internal/middlewares"
	users "dev/internal/user"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/webauthn"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Controller struct{}

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
func (controller *Controller) PostLogin(c *gin.Context) {
	var body LoginBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := RequestLoginLink(c, body.Email, c.GetHeader("Origin"), c.GetHeader("Referer")); err != nil {
		middlewares.RespondError(c, err)
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
//	@Success		200		{object}	Session
//	@Failure		400
//	@Router			/auth/verify [post]
func (controller *Controller) PostVerify(c *gin.Context) {
	var body VerifyBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	session, err := VerifyLoginLink(c, body.Token)
	if err != nil {
		middlewares.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, session)
}

// PostRefresh godoc
//
//	@Summary		Refresh a user's access token
//	@Description	Refresh a user's access token by refresh token
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		VerifyBody	true	"Refresh token"
//	@Success		200		{object}	Session
//	@Failure		400
//	@Router			/auth/refresh [post]
func (controller *Controller) PostRefresh(c *gin.Context) {
	var body VerifyBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	session, err := RefreshSession(c, body.Token)
	if err != nil {
		middlewares.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, session)
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
func (controller *Controller) GetOAuthProviders(c *gin.Context) {
	providers := []OAuthProviderResponse{}
	for _, provider := range GetOAuthProviders() {
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
//	@Success		200			{object}	Session
//	@Failure		400			{object}	map[string]string
//	@Failure		401			{object}	map[string]string
//	@Failure		404			{object}	map[string]string
//	@Failure		409			{object}	map[string]string
//	@Router			/auth/{provider} [post]
func (controller *Controller) PostOAuthLogin(c *gin.Context) {
	provider, err := FindOAuthProvider(c.Param("provider"))
	if err != nil {
		middlewares.RespondError(c, err)
		return
	}

	var body OAuthLoginBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accessToken := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
	session, err := LoginWithProvider(c, provider, body.Code, body.RedirectUri, accessToken)
	if err != nil {
		middlewares.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, session)
}

func newWebAuthn(c *gin.Context) *webauthn.WebAuthn {
	webAuthn, err := WebAuthnFor(c.GetHeader("Origin"))
	if err != nil {
		middlewares.RespondError(c, err)
		return nil
	}
	return webAuthn
}

func getOwnUser(c *gin.Context) *users.User {
	userId, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
		return nil
	}
	user, err := users.GetOwnUser(c, c.MustGet("x-user-id").(primitive.ObjectID), userId)
	if err != nil {
		middlewares.RespondError(c, err)
		return nil
	}
	return user
}

// PostPasskeyLoginBegin godoc
//
//	@Summary		Start a passkey login
//	@Description	Get the options to pass to navigator.credentials.get(). No email is needed, the passkey identifies the user.
//	@Tags			auth
//	@Produce		json
//	@Success		200	{object}	PasskeyBeginResponse
//	@Failure		400	{object}	map[string]string
//	@Router			/auth/passkey/begin [post]
func (controller *Controller) PostPasskeyLoginBegin(c *gin.Context) {
	webAuthn := newWebAuthn(c)
	if webAuthn == nil {
		return
	}

	response, err := BeginPasskeyLogin(c, webAuthn)
	if err != nil {
		middlewares.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// PostPasskeyLoginFinish godoc
//
//	@Summary		Finish a passkey login
//	@Description	Verify the assertion returned by navigator.credentials.get() and log the owner of the passkey in
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		PasskeyFinishBody	true	"Session and assertion"
//	@Success		200		{object}	Session
//	@Failure		400		{object}	map[string]string
//	@Failure		401		{object}	map[string]string
//	@Router			/auth/passkey/finish [post]
func (controller *Controller) PostPasskeyLoginFinish(c *gin.Context) {
	webAuthn := newWebAuthn(c)
	if webAuthn == nil {
		return
	}
	var body PasskeyFinishBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	session, err := FinishPasskeyLogin(c, webAuthn, body)
	if err != nil {
		middlewares.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, session)
}

// PostPasskeyRegisterBegin godoc
//
//	@Summary		Start adding a passkey
//	@Description	Get the options to pass to navigator.credentials.create() to add a passkey to the authenticated user
//	@Tags			user
//	@Produce		json
//	@Param			id	path		string	true	"User ID"
//	@Success		200	{object}	PasskeyBeginResponse
//	@Failure		400	{object}	map[string]string
//	@Failure		403	{object}	map[string]string
//	@Failure		404	{object}	map[string]string
//	@Security		Bearer
//	@Router			/users/{id}/passkeys/begin [post]
func (controller *Controller) PostPasskeyRegisterBegin(c *gin.Context) {
	webAuthn := newWebAuthn(c)
	if webAuthn == nil {
		return
	}
	user := getOwnUser(c)
	if user == nil {
		return
	}

	response, err := BeginPasskeyRegistration(c, webAuthn, user)
	if err != nil {
		middlewares.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// PostPasskeyRegisterFinish godoc
//
//	@Summary		Finish adding a passkey
//	@Description	Verify the credential returned by navigator.credentials.create() and add it to the authenticated user
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string				true	"User ID"
//	@Param			body	body		PasskeyFinishBody	true	"Session, passkey name and credential"
//	@Success		201		{object}	user.User
//	@Failure		400		{object}	map[string]string
//	@Failure		403		{object}	map[string]string
//	@Failure		404		{object}	map[string]string
//	@Security		Bearer
//	@Router			/users/{id}/passkeys/finish [post]
func (controller *Controller) PostPasskeyRegisterFinish(c *gin.Context) {
	webAuthn := newWebAuthn(c)
	if webAuthn == nil {
		return
	}
	user := getOwnUser(c)
	if user == nil {
		return
	}
	var body PasskeyFinishBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := FinishPasskeyRegistration(c, webAuthn, user, body)
	if err != nil {
		middlewares.RespondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, updated)
}

// DeletePasskey godoc
//
//	@Summary		Remove a passkey
//	@Description	Remove a passkey from the authenticated user
//	@Tags			user
//	@Produce		json
//	@Param			id			path		string	true	"User ID"
//	@Param			passkeyId	path		string	true	"Passkey ID"
//	@Success		200			{object}	user.User
//	@Failure		400			{object}	map[string]string
//	@Failure		403			{object}	map[string]string
//	@Failure		404			{object}	map[string]string
//	@Security		Bearer
//	@Router			/users/{id}/passkeys/{passkeyId} [delete]
func (controller *Controller) DeletePasskey(c *gin.Context) {
	user := getOwnUser(c)
	if user == nil {
		return
	}

	updated, err := RemovePasskey(c, user, c.Param("passkeyId"))
	if err != nil {
		middlewares.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, updated)
}
