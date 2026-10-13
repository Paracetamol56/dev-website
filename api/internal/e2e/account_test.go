//go:build e2e

package e2e

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"dev/internal/testutil"

	"github.com/descope/virtualwebauthn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func providers(user map[string]any) []string {
	names := []string{}
	for _, identity := range user["identities"].([]any) {
		names = append(names, identity.(map[string]any)["provider"].(string))
	}
	return names
}

func TestMagicLinkAccountLifecycle(t *testing.T) {
	app := start(t)
	ada := app.visitor()

	ada.get("/users/000000000000000000000000").expect(http.StatusUnauthorized)
	ada.post("/auth/verify", map[string]string{"token": "guessed"}).expect(http.StatusBadRequest)

	ada.login(app.address("ada"))
	profile := ada.get("/users/" + ada.userId).expect(http.StatusOK).object()
	assert.Equal(t, app.address("ada"), profile["email"])
	assert.Equal(t, "mocha", profile["flavour"])
	assert.Equal(t, []string{"email"}, providers(profile))

	updated := ada.patch("/users/"+ada.userId, map[string]string{"name": "Ada Lovelace", "flavour": "latte"}).expect(http.StatusOK).object()
	assert.Equal(t, "Ada Lovelace", updated["name"])
	assert.Equal(t, "latte", ada.get("/users/" + ada.userId).expect(http.StatusOK).object()["flavour"])

	grace := app.visitor()
	grace.login(app.address("grace"))
	assert.NotEqual(t, ada.userId, grace.userId)
	seenByGrace := grace.get("/users/" + ada.userId).expect(http.StatusOK).object()
	assert.Equal(t, map[string]any{"id": ada.userId, "name": "Ada Lovelace", "profilePicture": ""}, seenByGrace, "others only see a public profile")
	grace.patch("/users/"+ada.userId, map[string]string{"name": "Hacked"}).expect(http.StatusForbidden)
	grace.get("/users/" + ada.userId + "/export").expect(http.StatusForbidden)
	grace.delete("/users/" + ada.userId).expect(http.StatusForbidden)

	previous := ada.accessToken
	refreshed := ada.post("/auth/refresh", map[string]string{"token": ada.refreshToken}).expect(http.StatusOK).object()
	ada.accessToken = refreshed["accessToken"].(string)
	ada.get("/users/" + ada.userId).expect(http.StatusOK)
	ada.post("/auth/refresh", map[string]string{"token": previous}).expect(http.StatusBadRequest)

	laptop := app.visitor()
	laptop.login(strings.ToUpper(app.address("ada")))
	assert.Equal(t, ada.userId, laptop.userId, "logging in again from another device reaches the same account")
	assert.Equal(t, "Ada Lovelace", laptop.get("/users/" + laptop.userId).expect(http.StatusOK).object()["name"])

	export := ada.get("/users/" + ada.userId + "/export").expect(http.StatusOK).object()
	assert.Equal(t, app.address("ada"), export["email"])

	ada.delete("/users/" + ada.userId).expect(http.StatusNoContent)
	notices := app.inbox(app.address("ada"))
	assert.Equal(t, "Your account has been deleted", notices[len(notices)-1].Subject)
	ada.get("/users/" + ada.userId).expect(http.StatusNotFound)
	grace.get("/users/" + ada.userId).expect(http.StatusNotFound)
	ada.delete("/users/" + ada.userId).expect(http.StatusNotFound)

	reborn := app.visitor()
	reborn.login(app.address("ada"))
	assert.NotEqual(t, ada.userId, reborn.userId, "the address can sign up again, as a new account")
	assert.Equal(t, "", reborn.get("/users/" + reborn.userId).expect(http.StatusOK).object()["name"])
}

func TestLoginLinksAreSingleUseAndOnlySentToTheApp(t *testing.T) {
	app := start(t)
	ada := app.visitor()
	ada.login(app.address("ada"))

	emails := app.inbox(app.address("ada"))
	require.Len(t, emails, 1)
	token := linkPattern.FindStringSubmatch(emails[0].HTMLContent)[1]
	token = token[len(token)-43:]
	replay := app.visitor().post("/auth/verify", map[string]string{"token": token}).expect(http.StatusBadRequest)
	assert.Contains(t, replay.error(), "invalid or has expired")

	request, err := http.NewRequest(http.MethodPost, app.server.URL+"/api/auth/login", strings.NewReader(`{"email": "`+app.address("ada")+`"}`))
	require.NoError(t, err)
	request.Header.Set("Origin", "https://attacker.example")
	result, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	result.Body.Close()
	assert.Equal(t, http.StatusForbidden, result.StatusCode)
	assert.Len(t, app.inbox(app.address("ada")), 1, "no link for a login started from another site")
}

func TestLoginEmailsAreRateLimited(t *testing.T) {
	app := start(t)
	visitor := app.visitor()

	for range 3 {
		visitor.post("/auth/login", map[string]string{"email": app.address("flood")}).expect(http.StatusNoContent)
	}
	limited := visitor.post("/auth/login", map[string]string{"email": app.address("flood")}).expect(http.StatusTooManyRequests)
	assert.NotEmpty(t, limited.Header.Get("Retry-After"))
	assert.Len(t, app.inbox(app.address("flood")), 3)

	for i := range 6 {
		visitor.post("/auth/login", map[string]string{"email": app.address(string(rune('a' + i)))}).expect(http.StatusNoContent)
	}
	visitor.post("/auth/login", map[string]string{"email": app.address("one-too-many")}).expect(http.StatusTooManyRequests)
	assert.Empty(t, app.inbox(app.address("one-too-many")), "an address cannot be used to get around the limit per visitor")
}

func fakeProviders(app *app, githubEmail string, googleEmail string) {
	app.internet.JSON("github.com", http.StatusOK, `{"access_token":"gho_token"}`)
	app.internet.Handle("api.github.com", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/user" {
			io.WriteString(w, `{"id":42,"login":"ada","name":"Ada Lovelace","avatar_url":"https://avatars.example/ada","html_url":"https://github.com/ada"}`)
			return
		}
		io.WriteString(w, `[{"email":"`+githubEmail+`","primary":true,"verified":true}]`)
	})
	app.internet.JSON("oauth2.googleapis.com", http.StatusOK, `{"access_token":"ya29"}`)
	app.internet.JSON("openidconnect.googleapis.com", http.StatusOK,
		`{"sub":"g-1","email":"`+googleEmail+`","email_verified":true,"name":"Ada","picture":"https://lh3.example/ada"}`)
}

func TestOAuthIdentities(t *testing.T) {
	app := start(t)
	fakeProviders(app, app.address("ada"), app.address("ada"))
	ada := app.visitor()

	offered := ada.get("/auth/providers").expect(http.StatusOK).list()
	require.Len(t, offered, 2)
	assert.Equal(t, "github", offered[0]["name"])
	assert.Equal(t, "https://github.com/login/oauth/authorize", offered[0]["authorizeUrl"])

	code := map[string]string{"code": "code-from-the-provider", "redirectUri": testutil.Origin + "/verify/github"}
	ada.keepSession(ada.post("/auth/github", code).expect(http.StatusOK))
	profile := ada.get("/users/" + ada.userId).expect(http.StatusOK)
	assert.Equal(t, "Ada Lovelace", profile.object()["name"])
	assert.Equal(t, "https://avatars.example/ada", profile.object()["profilePicture"])
	assert.Equal(t, []string{"github"}, providers(profile.object()))
	assert.NotContains(t, string(profile.Body), "gho_token")

	phone := app.visitor()
	phone.keepSession(phone.post("/auth/google", code).expect(http.StatusOK))
	assert.Equal(t, ada.userId, phone.userId, "same email, same account")
	phone.login(app.address("ada"))
	assert.Equal(t, ada.userId, phone.userId)
	assert.Equal(t, []string{"github", "google", "email"}, providers(ada.get("/users/"+ada.userId).expect(http.StatusOK).object()))

	unlinked := ada.delete("/users/" + ada.userId + "/identities/github").expect(http.StatusOK).object()
	assert.Equal(t, []string{"google", "email"}, providers(unlinked))
	ada.delete("/users/" + ada.userId + "/identities/github").expect(http.StatusNotFound)
	ada.delete("/users/" + ada.userId + "/identities/email").expect(http.StatusBadRequest)

	fakeProviders(app, "someone-else@example.test", app.address("ada"))
	conflict := ada.post("/auth/github", code).expect(http.StatusConflict)
	assert.Contains(t, conflict.error(), "someone-else@example.test")
	fakeProviders(app, app.address("ada"), app.address("ada"))
	ada.post("/auth/github", code).expect(http.StatusOK)
	assert.Equal(t, []string{"google", "email", "github"}, providers(ada.get("/users/"+ada.userId).expect(http.StatusOK).object()))

	ada.post("/auth/facebook", code).expect(http.StatusNotFound)
	app.internet.JSON("github.com", http.StatusOK, `{"error":"bad_verification_code"}`)
	app.visitor().post("/auth/github", code).expect(http.StatusBadRequest)
}

func TestPasskeys(t *testing.T) {
	app := start(t)
	ada := app.visitor()
	ada.login(app.address("ada"))
	userId, err := primitive.ObjectIDFromHex(ada.userId)
	require.NoError(t, err)

	relyingParty := virtualwebauthn.RelyingParty{ID: "localhost", Name: "Mathéo Galuba - Dev", Origin: testutil.Origin}
	authenticator := virtualwebauthn.NewAuthenticatorWithOptions(virtualwebauthn.AuthenticatorOptions{UserHandle: userId[:]})
	credential := virtualwebauthn.NewCredential(virtualwebauthn.KeyTypeEC2)
	type ceremony struct {
		SessionId string          `json:"sessionId"`
		Options   json.RawMessage `json:"options"`
	}

	var registration ceremony
	ada.post("/users/"+ada.userId+"/passkeys/begin", nil).expect(http.StatusOK).into(&registration)
	creation, err := virtualwebauthn.ParseAttestationOptions(string(registration.Options))
	require.NoError(t, err)
	attestation := virtualwebauthn.CreateAttestationResponse(relyingParty, authenticator, credential, *creation)
	withPasskey := ada.post("/users/"+ada.userId+"/passkeys/finish",
		map[string]any{"sessionId": registration.SessionId, "name": "Security key", "credential": json.RawMessage(attestation)}).expect(http.StatusCreated).object()
	passkeys := withPasskey["passkeys"].([]any)
	require.Len(t, passkeys, 1)
	passkeyId := passkeys[0].(map[string]any)["id"].(string)
	assert.Equal(t, "Security key", passkeys[0].(map[string]any)["name"])

	signIn := func() *response {
		visitor := app.visitor()
		var login ceremony
		visitor.post("/auth/passkey/begin", nil).expect(http.StatusOK).into(&login)
		request, err := virtualwebauthn.ParseAssertionOptions(string(login.Options))
		require.NoError(t, err)
		assertion := virtualwebauthn.CreateAssertionResponse(relyingParty, authenticator, credential, *request)
		return visitor.post("/auth/passkey/finish", map[string]any{"sessionId": login.SessionId, "credential": json.RawMessage(assertion)})
	}

	other := app.visitor()
	other.keepSession(signIn().expect(http.StatusOK))
	assert.Equal(t, ada.userId, other.userId, "the passkey alone identifies the user")
	used := other.get("/users/" + other.userId).expect(http.StatusOK).object()["passkeys"].([]any)[0].(map[string]any)
	assert.NotEmpty(t, used["lastUsed"])
	other.keepSession(signIn().expect(http.StatusOK))

	mallory := app.visitor()
	mallory.login(app.address("mallory"))
	mallory.delete("/users/" + ada.userId + "/passkeys/" + passkeyId).expect(http.StatusForbidden)
	mallory.post("/users/"+ada.userId+"/passkeys/begin", nil).expect(http.StatusForbidden)

	remaining := ada.delete("/users/" + ada.userId + "/passkeys/" + passkeyId).expect(http.StatusOK).object()
	assert.Empty(t, remaining["passkeys"])
	rejected := signIn().expect(http.StatusUnauthorized)
	assert.Equal(t, "This passkey is not recognized", rejected.error())
}
