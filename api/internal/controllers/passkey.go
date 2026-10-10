package controllers

import (
	"dev/internal/models"
	"dev/internal/utils"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PasskeyController struct{}

type PasskeyBeginResponse struct {
	SessionId string `json:"sessionId"`
	Options   any    `json:"options"`
}

type PasskeyFinishBody struct {
	SessionId  string          `json:"sessionId" binding:"required"`
	Name       string          `json:"name" binding:"omitempty,max=50"`
	Credential json.RawMessage `json:"credential" binding:"required" swaggertype:"object"`
}

func newWebAuthn(c *gin.Context) *webauthn.WebAuthn {
	webAuthn, err := utils.NewWebAuthn(c.GetHeader("Origin"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Passkeys are not available from this origin"})
		return nil
	}
	return webAuthn
}

func getOwnUser(c *gin.Context) *models.User {
	userId, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
		return nil
	}
	if userId != c.MustGet("x-user-id").(primitive.ObjectID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return nil
	}
	user, err := models.GetFullUserById(c, userId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return nil
	}
	return user
}

// PostLoginBegin godoc
//
//	@Summary		Start a passkey login
//	@Description	Get the options to pass to navigator.credentials.get(). No email is needed, the passkey identifies the user.
//	@Tags			auth
//	@Produce		json
//	@Success		200	{object}	PasskeyBeginResponse
//	@Failure		400	{object}	map[string]string
//	@Router			/auth/passkey/begin [post]
func (controller *PasskeyController) PostLoginBegin(c *gin.Context) {
	webAuthn := newWebAuthn(c)
	if webAuthn == nil {
		return
	}

	options, session, err := webAuthn.BeginDiscoverableLogin(webauthn.WithUserVerification(protocol.VerificationPreferred))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	sessionId, err := models.CreatePasskeyChallenge(c, session)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, PasskeyBeginResponse{SessionId: sessionId, Options: options})
}

// PostLoginFinish godoc
//
//	@Summary		Finish a passkey login
//	@Description	Verify the assertion returned by navigator.credentials.get() and log the owner of the passkey in
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		PasskeyFinishBody	true	"Session and assertion"
//	@Success		200		{object}	map[string]interface{}
//	@Failure		400		{object}	map[string]string
//	@Failure		401		{object}	map[string]string
//	@Router			/auth/passkey/finish [post]
func (controller *PasskeyController) PostLoginFinish(c *gin.Context) {
	webAuthn := newWebAuthn(c)
	if webAuthn == nil {
		return
	}

	var body PasskeyFinishBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	session, err := models.ConsumePasskeyChallenge(c, body.SessionId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	assertion, err := protocol.ParseCredentialRequestResponseBytes(body.Credential)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid passkey response"})
		return
	}

	findUser := func(rawID, userHandle []byte) (webauthn.User, error) {
		if len(userHandle) != len(primitive.NilObjectID) {
			return nil, errors.New("invalid user handle")
		}
		return models.GetFullUserById(c, primitive.ObjectID(userHandle))
	}
	owner, credential, err := webAuthn.ValidatePasskeyLogin(findUser, *session, assertion)
	if err != nil {
		log.Printf("Passkey login refused: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "This passkey is not recognized"})
		return
	}
	user := owner.(*models.User)

	passkey := user.GetPasskey(models.PasskeyId(credential))
	if passkey == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "This passkey is not recognized"})
		return
	}
	// The stored credential carries the signature counter used to detect cloned authenticators
	if passkey.Credential, err = json.Marshal(credential); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	now := time.Now()
	passkey.LastUsed = &now
	user.LastLogin = now
	user.LastRefresh = now
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

// PostRegisterBegin godoc
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
func (controller *PasskeyController) PostRegisterBegin(c *gin.Context) {
	webAuthn := newWebAuthn(c)
	if webAuthn == nil {
		return
	}
	user := getOwnUser(c)
	if user == nil {
		return
	}
	if len(user.Passkeys) >= models.MaxPasskeys {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Too many passkeys, remove one first"})
		return
	}

	options, session, err := webAuthn.BeginRegistration(user,
		webauthn.WithResidentKeyRequirement(protocol.ResidentKeyRequirementRequired),
		webauthn.WithExclusions(webauthn.Credentials(user.WebAuthnCredentials()).CredentialDescriptors()),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	sessionId, err := models.CreatePasskeyChallenge(c, session)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, PasskeyBeginResponse{SessionId: sessionId, Options: options})
}

// PostRegisterFinish godoc
//
//	@Summary		Finish adding a passkey
//	@Description	Verify the credential returned by navigator.credentials.create() and add it to the authenticated user
//	@Tags			user
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string				true	"User ID"
//	@Param			body	body		PasskeyFinishBody	true	"Session, passkey name and credential"
//	@Success		201		{object}	models.User
//	@Failure		400		{object}	map[string]string
//	@Failure		403		{object}	map[string]string
//	@Failure		404		{object}	map[string]string
//	@Security		Bearer
//	@Router			/users/{id}/passkeys/finish [post]
func (controller *PasskeyController) PostRegisterFinish(c *gin.Context) {
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
	session, err := models.ConsumePasskeyChallenge(c, body.SessionId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	attestation, err := protocol.ParseCredentialCreationResponseBytes(body.Credential)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid passkey response"})
		return
	}
	credential, err := webAuthn.CreateCredential(user, *session, attestation)
	if err != nil {
		log.Printf("Passkey registration refused: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "This passkey could not be verified"})
		return
	}

	if body.Name == "" {
		body.Name = "Passkey"
	}
	passkey, err := models.NewPasskey(body.Name, credential)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	user.Passkeys = append(user.Passkeys, *passkey)
	if _, err := models.UpdateUser(c, user.Id, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// DeletePasskey godoc
//
//	@Summary		Remove a passkey
//	@Description	Remove a passkey from the authenticated user
//	@Tags			user
//	@Produce		json
//	@Param			id			path		string	true	"User ID"
//	@Param			passkeyId	path		string	true	"Passkey ID"
//	@Success		200			{object}	models.User
//	@Failure		400			{object}	map[string]string
//	@Failure		403			{object}	map[string]string
//	@Failure		404			{object}	map[string]string
//	@Security		Bearer
//	@Router			/users/{id}/passkeys/{passkeyId} [delete]
func (controller *PasskeyController) DeletePasskey(c *gin.Context) {
	user := getOwnUser(c)
	if user == nil {
		return
	}
	if !user.RemovePasskey(c.Param("passkeyId")) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Passkey not found"})
		return
	}
	if _, err := models.UpdateUser(c, user.Id, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}
