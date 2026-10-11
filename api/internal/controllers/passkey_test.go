package controllers

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"dev/internal/models"
	"dev/internal/testutil"

	"github.com/descope/virtualwebauthn"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var relyingParty = virtualwebauthn.RelyingParty{ID: "localhost", Name: "Mathéo Galuba - Dev", Origin: testutil.Origin}

type passkeyBegin struct {
	SessionId string          `json:"sessionId"`
	Options   json.RawMessage `json:"options"`
}

// device is a software authenticator holding one passkey
type device struct {
	authenticator virtualwebauthn.Authenticator
	credential    virtualwebauthn.Credential
}

func newDevice(userId primitive.ObjectID) *device {
	return &device{
		authenticator: virtualwebauthn.NewAuthenticatorWithOptions(virtualwebauthn.AuthenticatorOptions{UserHandle: userId[:]}),
		credential:    virtualwebauthn.NewCredential(virtualwebauthn.KeyTypeEC2),
	}
}

func (device *device) attest(t *testing.T, begin passkeyBegin) json.RawMessage {
	t.Helper()
	options, err := virtualwebauthn.ParseAttestationOptions(string(begin.Options))
	require.NoError(t, err)
	return json.RawMessage(virtualwebauthn.CreateAttestationResponse(relyingParty, device.authenticator, device.credential, *options))
}

func (device *device) assert(t *testing.T, begin passkeyBegin) json.RawMessage {
	t.Helper()
	options, err := virtualwebauthn.ParseAssertionOptions(string(begin.Options))
	require.NoError(t, err)
	return json.RawMessage(virtualwebauthn.CreateAssertionResponse(relyingParty, device.authenticator, device.credential, *options))
}

func (device *device) register(t *testing.T, userId primitive.ObjectID, name string) {
	t.Helper()
	path := "/users/" + userId.Hex() + "/passkeys"
	begin := do(t, http.MethodPost, path+"/begin", nil, as(userId), fromApp)
	require.Equal(t, http.StatusOK, begin.Code, begin.Body.String())
	session := decode[passkeyBegin](t, begin)
	finish := do(t, http.MethodPost, path+"/finish",
		PasskeyFinishBody{SessionId: session.SessionId, Name: name, Credential: device.attest(t, session)}, as(userId), fromApp)
	require.Equal(t, http.StatusCreated, finish.Code, finish.Body.String())
}

func beginLogin(t *testing.T) passkeyBegin {
	t.Helper()
	begin := do(t, http.MethodPost, "/auth/passkey/begin", nil, fromApp)
	require.Equal(t, http.StatusOK, begin.Code, begin.Body.String())
	return decode[passkeyBegin](t, begin)
}

func TestPasskeyRegistration(t *testing.T) {
	testutil.Clean(t)
	user := createUser(t, models.User{Email: "ada@example.test", Name: "Ada"})
	path := "/users/" + user.Id.Hex() + "/passkeys"
	laptop := newDevice(user.Id)

	begin := do(t, http.MethodPost, path+"/begin", nil, as(user.Id), fromApp)
	require.Equal(t, http.StatusOK, begin.Code, begin.Body.String())
	session := decode[passkeyBegin](t, begin)
	var options struct {
		PublicKey struct {
			Rp   struct{ Id string }
			User struct {
				Name        string
				DisplayName string
			}
			AuthenticatorSelection struct{ ResidentKey string }
			ExcludeCredentials     []any
		}
	}
	require.NoError(t, json.Unmarshal(session.Options, &options))
	assert.Equal(t, "localhost", options.PublicKey.Rp.Id)
	assert.Equal(t, "ada@example.test", options.PublicKey.User.Name)
	assert.Equal(t, "Ada", options.PublicKey.User.DisplayName)
	assert.Equal(t, "required", options.PublicKey.AuthenticatorSelection.ResidentKey)
	assert.Empty(t, options.PublicKey.ExcludeCredentials)

	finish := do(t, http.MethodPost, path+"/finish",
		PasskeyFinishBody{SessionId: session.SessionId, Name: "Laptop", Credential: laptop.attest(t, session)}, as(user.Id), fromApp)
	require.Equal(t, http.StatusCreated, finish.Code, finish.Body.String())
	assert.NotContains(t, finish.Body.String(), "credential")
	returned := decode[models.User](t, finish)
	require.Len(t, returned.Passkeys, 1)
	assert.Equal(t, "Laptop", returned.Passkeys[0].Name)

	stored := storedUser(t, user.Id)
	require.Len(t, stored.Passkeys, 1)
	assert.Equal(t, models.PasskeyId(&webauthn.Credential{ID: laptop.credential.ID}), stored.Passkeys[0].Id)
	assert.WithinDuration(t, time.Now(), stored.Passkeys[0].CreatedAt, 5*time.Second)
	assert.Nil(t, stored.Passkeys[0].LastUsed)
	require.Len(t, stored.WebAuthnCredentials(), 1)

	replayed := do(t, http.MethodPost, path+"/finish",
		PasskeyFinishBody{SessionId: session.SessionId, Credential: laptop.attest(t, session)}, as(user.Id), fromApp)
	assert.Equal(t, http.StatusBadRequest, replayed.Code, "a ceremony cannot be replayed")

	phone := newDevice(user.Id)
	phone.register(t, user.Id, "")
	stored = storedUser(t, user.Id)
	require.Len(t, stored.Passkeys, 2)
	assert.Equal(t, "Passkey", stored.Passkeys[1].Name, "default name")

	second := decode[passkeyBegin](t, do(t, http.MethodPost, path+"/begin", nil, as(user.Id), fromApp))
	require.NoError(t, json.Unmarshal(second.Options, &options))
	assert.Len(t, options.PublicKey.ExcludeCredentials, 2, "a device is not registered twice")
}

func TestPasskeyRegistrationRejections(t *testing.T) {
	testutil.Clean(t)
	user := createUser(t, models.User{})
	path := "/users/" + user.Id.Hex() + "/passkeys"
	stranger := primitive.NewObjectID()

	for _, step := range []string{"/begin", "/finish"} {
		noOrigin := do(t, http.MethodPost, path+step, nil, as(user.Id))
		assert.Equal(t, http.StatusBadRequest, noOrigin.Code)
		assert.Equal(t, "Passkeys are not available from this origin", errorOf(t, noOrigin))
		assert.Equal(t, http.StatusBadRequest, do(t, http.MethodPost, path+step, nil, as(user.Id), header("Origin", "https://attacker.example")).Code)
		assert.Equal(t, http.StatusBadRequest, do(t, http.MethodPost, "/users/me/passkeys"+step, nil, as(user.Id), fromApp).Code)
		assert.Equal(t, http.StatusForbidden, do(t, http.MethodPost, path+step, nil, as(stranger), fromApp).Code)
		assert.Equal(t, http.StatusNotFound, do(t, http.MethodPost, "/users/"+stranger.Hex()+"/passkeys"+step, nil, as(stranger), fromApp).Code)
	}

	session := decode[passkeyBegin](t, do(t, http.MethodPost, path+"/begin", nil, as(user.Id), fromApp))
	device := newDevice(user.Id)
	finish := func(body any) int {
		return do(t, http.MethodPost, path+"/finish", body, as(user.Id), fromApp).Code
	}
	assert.Equal(t, http.StatusBadRequest, finish("not json"))
	assert.Equal(t, http.StatusBadRequest, finish(map[string]any{"sessionId": session.SessionId}), "no credential")
	assert.Equal(t, http.StatusBadRequest, finish(PasskeyFinishBody{SessionId: session.SessionId, Name: repeat("a", 51), Credential: device.attest(t, session)}))
	assert.Equal(t, http.StatusBadRequest, finish(PasskeyFinishBody{SessionId: "unknown", Credential: device.attest(t, session)}))
	assert.Equal(t, http.StatusBadRequest, finish(PasskeyFinishBody{SessionId: session.SessionId, Credential: json.RawMessage(`{"id":"x"}`)}),
		"not a WebAuthn response")

	other := decode[passkeyBegin](t, do(t, http.MethodPost, path+"/begin", nil, as(user.Id), fromApp))
	third := decode[passkeyBegin](t, do(t, http.MethodPost, path+"/begin", nil, as(user.Id), fromApp))
	mismatch := do(t, http.MethodPost, path+"/finish", PasskeyFinishBody{SessionId: other.SessionId, Credential: device.attest(t, third)}, as(user.Id), fromApp)
	assert.Equal(t, http.StatusBadRequest, mismatch.Code)
	assert.Equal(t, "This passkey could not be verified", errorOf(t, mismatch), "answer to another challenge")
	assert.Empty(t, storedUser(t, user.Id).Passkeys)

	full := createUser(t, models.User{Passkeys: make([]models.Passkey, models.MaxPasskeys)})
	tooMany := do(t, http.MethodPost, "/users/"+full.Id.Hex()+"/passkeys/begin", nil, as(full.Id), fromApp)
	assert.Equal(t, http.StatusBadRequest, tooMany.Code)
	assert.Equal(t, "Too many passkeys, remove one first", errorOf(t, tooMany))
}

func TestPasskeyLogin(t *testing.T) {
	testutil.Clean(t)
	user := createUser(t, models.User{Email: "ada@example.test"})
	laptop := newDevice(user.Id)
	laptop.register(t, user.Id, "Laptop")
	phone := newDevice(user.Id)
	phone.register(t, user.Id, "Phone")

	session := beginLogin(t)
	var options struct {
		PublicKey struct {
			RpId             string
			UserVerification string
			AllowCredentials []any
		}
	}
	require.NoError(t, json.Unmarshal(session.Options, &options))
	assert.Equal(t, "localhost", options.PublicKey.RpId)
	assert.Equal(t, "preferred", options.PublicKey.UserVerification)
	assert.Empty(t, options.PublicKey.AllowCredentials, "the passkey names the user, no email is asked")

	finish := do(t, http.MethodPost, "/auth/passkey/finish", PasskeyFinishBody{SessionId: session.SessionId, Credential: phone.assert(t, session)}, fromApp)
	require.Equal(t, http.StatusOK, finish.Code, finish.Body.String())
	tokens := decode[tokenResponse](t, finish)
	assertSession(t, tokens, user.Id)
	assert.Equal(t, user.Id, tokens.User.Id)

	stored := storedUser(t, user.Id)
	assert.WithinDuration(t, time.Now(), stored.LastLogin, 5*time.Second)
	assert.Nil(t, stored.GetPasskey(stored.Passkeys[0].Id).LastUsed)
	require.NotNil(t, stored.Passkeys[1].LastUsed, "the passkey that was used is dated")
	assert.WithinDuration(t, time.Now(), *stored.Passkeys[1].LastUsed, 5*time.Second)

	replayed := do(t, http.MethodPost, "/auth/passkey/finish", PasskeyFinishBody{SessionId: session.SessionId, Credential: phone.assert(t, session)}, fromApp)
	assert.Equal(t, http.StatusBadRequest, replayed.Code, "a ceremony cannot be replayed")
}

func TestPasskeyLoginRejections(t *testing.T) {
	testutil.Clean(t)
	user := createUser(t, models.User{})
	laptop := newDevice(user.Id)
	laptop.register(t, user.Id, "Laptop")
	finish := func(body any, options ...option) int {
		return do(t, http.MethodPost, "/auth/passkey/finish", body, options...).Code
	}

	assert.Equal(t, http.StatusBadRequest, do(t, http.MethodPost, "/auth/passkey/begin", nil).Code, "no origin")
	assert.Equal(t, http.StatusBadRequest, finish(nil), "no origin")
	assert.Equal(t, http.StatusBadRequest, finish("not json", fromApp))
	assert.Equal(t, http.StatusBadRequest, finish(PasskeyFinishBody{SessionId: "unknown", Credential: json.RawMessage(`{}`)}, fromApp))
	session := beginLogin(t)
	assert.Equal(t, http.StatusBadRequest, finish(PasskeyFinishBody{SessionId: session.SessionId, Credential: json.RawMessage(`{"id":"x"}`)}, fromApp))

	unknown := newDevice(user.Id)
	session = beginLogin(t)
	refused := do(t, http.MethodPost, "/auth/passkey/finish", PasskeyFinishBody{SessionId: session.SessionId, Credential: unknown.assert(t, session)}, fromApp)
	assert.Equal(t, http.StatusUnauthorized, refused.Code)
	assert.Equal(t, "This passkey is not recognized", errorOf(t, refused), "a key that was never registered")

	orphan := newDevice(primitive.NewObjectID())
	session = beginLogin(t)
	assert.Equal(t, http.StatusUnauthorized, finish(PasskeyFinishBody{SessionId: session.SessionId, Credential: orphan.assert(t, session)}, fromApp),
		"user that does not exist")

	malformed := &device{
		authenticator: virtualwebauthn.NewAuthenticatorWithOptions(virtualwebauthn.AuthenticatorOptions{UserHandle: []byte("short")}),
		credential:    laptop.credential,
	}
	session = beginLogin(t)
	assert.Equal(t, http.StatusUnauthorized, finish(PasskeyFinishBody{SessionId: session.SessionId, Credential: malformed.assert(t, session)}, fromApp),
		"user handle that is not an id")

	first, second := beginLogin(t), beginLogin(t)
	assert.Equal(t, http.StatusUnauthorized, finish(PasskeyFinishBody{SessionId: first.SessionId, Credential: laptop.assert(t, second)}, fromApp),
		"answer to another challenge")

	removed := do(t, http.MethodDelete, "/users/"+user.Id.Hex()+"/passkeys/"+storedUser(t, user.Id).Passkeys[0].Id, nil, as(user.Id))
	require.Equal(t, http.StatusOK, removed.Code)
	session = beginLogin(t)
	assert.Equal(t, http.StatusUnauthorized, finish(PasskeyFinishBody{SessionId: session.SessionId, Credential: laptop.assert(t, session)}, fromApp),
		"a removed passkey no longer logs in")

	testutil.BreakDB(t)
	assert.Equal(t, http.StatusInternalServerError, do(t, http.MethodPost, "/auth/passkey/begin", nil, fromApp).Code)
}

func TestDeletePasskey(t *testing.T) {
	testutil.Clean(t)
	user := createUser(t, models.User{Passkeys: []models.Passkey{{Id: "laptop", Name: "Laptop"}, {Id: "phone", Name: "Phone"}}})
	path := "/users/" + user.Id.Hex() + "/passkeys/"

	response := do(t, http.MethodDelete, path+"laptop", nil, as(user.Id))
	require.Equal(t, http.StatusOK, response.Code)
	remaining := decode[models.User](t, response).Passkeys
	require.Len(t, remaining, 1)
	assert.Equal(t, "phone", remaining[0].Id)
	assert.Len(t, storedUser(t, user.Id).Passkeys, 1)

	missing := do(t, http.MethodDelete, path+"laptop", nil, as(user.Id))
	assert.Equal(t, http.StatusNotFound, missing.Code)
	assert.Equal(t, "Passkey not found", errorOf(t, missing))
	assert.Equal(t, http.StatusForbidden, do(t, http.MethodDelete, path+"phone", nil, as(primitive.NewObjectID())).Code)
	assert.Equal(t, http.StatusBadRequest, do(t, http.MethodDelete, "/users/me/passkeys/phone", nil, as(user.Id)).Code)
	assert.Len(t, storedUser(t, user.Id).Passkeys, 1)
}
