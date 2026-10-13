package auth

import (
	"dev/internal/testutil"
	users "dev/internal/user"
	"dev/internal/utils"
	"encoding/json"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/descope/virtualwebauthn"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var linkPattern = regexp.MustCompile(`<a href="([^"]+)">Verify your email address</a>`)

// magicLink returns the link of the last verification email and its token
func magicLink(t *testing.T, internet *testutil.Internet) (*url.URL, string) {
	t.Helper()
	emails := testutil.SentEmails(t, internet)
	require.NotEmpty(t, emails)
	match := linkPattern.FindStringSubmatch(emails[len(emails)-1].HTMLContent)
	require.NotNil(t, match, "no link in the email")
	link, err := url.Parse(html.UnescapeString(match[1]))
	require.NoError(t, err)
	return link, link.Query().Get("token")
}

type tokenResponse struct {
	AccessToken  string     `json:"accessToken"`
	RefreshToken string     `json:"refreshToken"`
	User         users.User `json:"user"`
}

func assertSession(t *testing.T, tokens tokenResponse, userId primitive.ObjectID) {
	t.Helper()
	fromAccess, err := utils.ExtractID(tokens.AccessToken, testutil.AccessSecret)
	require.NoError(t, err)
	assert.Equal(t, userId, fromAccess)
	fromRefresh, err := utils.ExtractID(tokens.RefreshToken, testutil.RefreshSecret)
	require.NoError(t, err)
	assert.Equal(t, userId, fromRefresh)
}

func TestPostLoginCreatesTheUserAndSendsALink(t *testing.T) {
	testutil.Clean(t)
	resetLimiters()
	internet := testutil.FakeInternet(t)
	testutil.FakeBrevo(internet)

	response := do(t, http.MethodPost, "/auth/login", map[string]string{"email": "Ada@Example.test"},
		testutil.FromApp, testutil.Header("Referer", testutil.Origin+"/tool/ormi?tab=stats"))
	assert.Equal(t, http.StatusNoContent, response.Code)
	assert.Empty(t, response.Body.String())

	user, err := users.GetFullUserByEmail(testutil.Context(), "ada@example.test")
	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, "Ada@Example.test", user.Email)
	assert.Equal(t, "mocha", user.Flavour)
	assert.Empty(t, user.Identities, "nothing is verified yet")

	emails := testutil.SentEmails(t, internet)
	require.Len(t, emails, 1)
	assert.Equal(t, "Ada@Example.test", emails[0].To[0].Email)
	assert.Equal(t, "Verify your email address", emails[0].Subject)
	link, token := magicLink(t, internet)
	assert.Equal(t, testutil.Origin+"/verify/email", link.Scheme+"://"+link.Host+link.Path)
	assert.Equal(t, "/tool/ormi?tab=stats", link.Query().Get("redirect"))

	consumed, err := ConsumeLoginToken(ctx, token)
	require.NoError(t, err)
	assert.Equal(t, user.Id, consumed)
}

func TestPostLoginReusesTheUser(t *testing.T) {
	testutil.Clean(t)
	resetLimiters()
	internet := testutil.FakeInternet(t)
	testutil.FakeBrevo(internet)
	user := createUser(t, users.User{Email: "ada@example.test", Name: "Ada"})

	assert.Equal(t, http.StatusNoContent, do(t, http.MethodPost, "/auth/login", map[string]string{"email": "ADA@example.test"}, testutil.FromApp).Code)
	_, first := magicLink(t, internet)
	assert.Equal(t, http.StatusNoContent, do(t, http.MethodPost, "/auth/login", map[string]string{"email": "ada@example.test"}, testutil.FromApp).Code)
	_, second := magicLink(t, internet)

	assert.Equal(t, 1, testutil.Count(t, "users"))
	assert.Equal(t, "Ada", testutil.SentEmails(t, internet)[0].To[0].Name)
	_, err := ConsumeLoginToken(ctx, first)
	assert.Error(t, err, "asking for a new link revokes the previous one")
	consumed, err := ConsumeLoginToken(ctx, second)
	require.NoError(t, err)
	assert.Equal(t, user.Id, consumed)
}

func TestPostLoginRejections(t *testing.T) {
	testutil.Clean(t)
	resetLimiters()
	internet := testutil.FakeInternet(t)
	testutil.FakeBrevo(internet)

	for _, body := range []any{"not json", map[string]string{}, map[string]string{"email": "not-an-email"}} {
		assert.Equal(t, http.StatusBadRequest, do(t, http.MethodPost, "/auth/login", body, testutil.FromApp).Code)
	}
	assert.Zero(t, testutil.Count(t, "users"))

	for _, origin := range []string{"", "https://attacker.example"} {
		refused := do(t, http.MethodPost, "/auth/login", map[string]string{"email": "ada@example.test"}, testutil.Header("Origin", origin))
		assert.Equal(t, http.StatusBadRequest, refused.Code)
		assert.Equal(t, "origin not allowed", testutil.ErrorOf(t, refused))
	}
	assert.Empty(t, testutil.SentEmails(t, internet), "no link is ever sent to another site")

	internet.JSON(testutil.BrevoHost, http.StatusInternalServerError, `{}`)
	assert.Equal(t, http.StatusInternalServerError, do(t, http.MethodPost, "/auth/login", map[string]string{"email": "bob@example.test"}, testutil.FromApp).Code)

	testutil.BreakDB(t)
	assert.Equal(t, http.StatusInternalServerError, do(t, http.MethodPost, "/auth/login", map[string]string{"email": "eve@example.test"}, testutil.FromApp).Code)
}

func TestPostLoginIsLimitedPerEmail(t *testing.T) {
	testutil.Clean(t)
	resetLimiters()
	internet := testutil.FakeInternet(t)
	testutil.FakeBrevo(internet)
	login := func(email string) int {
		return do(t, http.MethodPost, "/auth/login", map[string]string{"email": email}, testutil.FromApp).Code
	}

	for range 3 {
		assert.Equal(t, http.StatusNoContent, login("ada@example.test"))
	}
	limited := do(t, http.MethodPost, "/auth/login", map[string]string{"email": "ADA@example.test"}, testutil.FromApp)
	assert.Equal(t, http.StatusTooManyRequests, limited.Code)
	assert.NotEmpty(t, limited.Header().Get("Retry-After"))
	assert.Len(t, testutil.SentEmails(t, internet), 3)
	assert.Equal(t, http.StatusNoContent, login("grace@example.test"), "other addresses are not affected")
}

func TestPostVerify(t *testing.T) {
	testutil.Clean(t)
	internet := testutil.FakeInternet(t)
	testutil.FakeBrevo(internet)
	user := createUser(t, users.User{Email: "ada@example.test"})
	token, err := CreateLoginToken(ctx, user.Id)
	require.NoError(t, err)

	response := do(t, http.MethodPost, "/auth/verify", map[string]string{"token": token})
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	tokens := testutil.Decode[tokenResponse](t, response)
	assertSession(t, tokens, user.Id)
	assert.Equal(t, user.Id, tokens.User.Id)
	assert.Equal(t, []string{"ada@example.test"}, testutil.AddedContacts(internet))

	stored := storedUser(t, user.Id)
	require.Len(t, stored.Identities, 1)
	assert.Equal(t, "email", stored.Identities[0].Provider)
	assert.Equal(t, "ada@example.test", stored.Identities[0].ProviderId)
	assert.WithinDuration(t, time.Now(), stored.LastLogin, 5*time.Second)
	assert.WithinDuration(t, time.Now(), stored.LastRefresh, 5*time.Second)

	replayed := do(t, http.MethodPost, "/auth/verify", map[string]string{"token": token})
	assert.Equal(t, http.StatusBadRequest, replayed.Code, "a link works once")
	assert.Contains(t, testutil.ErrorOf(t, replayed), "invalid or has expired")
}

func TestPostVerifyRejections(t *testing.T) {
	testutil.Clean(t)
	internet := testutil.FakeInternet(t)
	internet.JSON(testutil.BrevoHost, http.StatusInternalServerError, `{}`)

	for _, body := range []any{"not json", map[string]string{}, map[string]string{"token": "unknown"}} {
		assert.Equal(t, http.StatusBadRequest, do(t, http.MethodPost, "/auth/verify", body).Code)
	}

	gone, err := CreateLoginToken(ctx, primitive.NewObjectID())
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, do(t, http.MethodPost, "/auth/verify", map[string]string{"token": gone}).Code, "the user no longer exists")

	user := createUser(t, users.User{})
	token, err := CreateLoginToken(ctx, user.Id)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, do(t, http.MethodPost, "/auth/verify", map[string]string{"token": token}).Code,
		"the mailing list being unreachable does not prevent logging in")

	testutil.BreakDB(t)
	assert.Equal(t, http.StatusBadRequest, do(t, http.MethodPost, "/auth/verify", map[string]string{"token": token}).Code)
}

func TestPostRefresh(t *testing.T) {
	userId := primitive.NewObjectID()
	refresh, err := utils.SignRefreshToken(userId.Hex(), 168)
	require.NoError(t, err)

	response := do(t, http.MethodPost, "/auth/refresh", map[string]string{"token": refresh})
	require.Equal(t, http.StatusOK, response.Code)
	assertSession(t, testutil.Decode[tokenResponse](t, response), userId)

	access, err := utils.SignAccessToken(userId.Hex(), 1)
	require.NoError(t, err)
	expired, err := utils.SignRefreshToken(userId.Hex(), -1)
	require.NoError(t, err)
	notAnId, err := jwt.NewWithClaims(jwt.SigningMethodHS256, utils.JwtRefreshClaims{UserId: "nope"}).SignedString([]byte(testutil.RefreshSecret))
	require.NoError(t, err)
	for name, body := range map[string]any{
		"not json":          "not json",
		"no token":          map[string]string{},
		"garbage":           map[string]string{"token": "garbage"},
		"access token":      map[string]string{"token": access},
		"expired":           map[string]string{"token": expired},
		"user id not an id": map[string]string{"token": notAnId},
	} {
		assert.Equal(t, http.StatusBadRequest, do(t, http.MethodPost, "/auth/refresh", body).Code, name)
	}
}

func TestGetOAuthProviders(t *testing.T) {
	response := do(t, http.MethodGet, "/auth/providers", nil)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.JSONEq(t, `[
		{"name": "github", "clientId": "test-github_client_id", "authorizeUrl": "https://github.com/login/oauth/authorize", "scope": "read:user user:email"},
		{"name": "google", "clientId": "test-google_client_id", "authorizeUrl": "https://accounts.google.com/o/oauth2/v2/auth", "scope": "openid email profile"}
	]`, response.Body.String())
	assert.NotContains(t, response.Body.String(), "secret")

	t.Setenv("GITHUB_CLIENT_SECRET", "")
	t.Setenv("GOOGLE_CLIENT_ID", "")
	assert.JSONEq(t, `[]`, do(t, http.MethodGet, "/auth/providers", nil).Body.String())
}

func fakeGithub(internet *testutil.Internet, id string, email string) {
	internet.JSON("github.com", http.StatusOK, `{"access_token":"gho_token"}`)
	internet.Handle("api.github.com", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/user" {
			io.WriteString(w, `{"id":`+id+`,"login":"ada","name":"Ada Lovelace","avatar_url":"https://avatars.example/ada","html_url":"https://github.com/ada"}`)
			return
		}
		io.WriteString(w, `[{"email":"`+email+`","primary":true,"verified":true}]`)
	})
}

func fakeGoogle(internet *testutil.Internet, id string, email string) {
	internet.JSON("oauth2.googleapis.com", http.StatusOK, `{"access_token":"ya29"}`)
	internet.JSON("openidconnect.googleapis.com", http.StatusOK,
		`{"sub":"`+id+`","email":"`+email+`","email_verified":true,"name":"Ada from Google","picture":"https://lh3.example/ada"}`)
}

var oauthBody = map[string]string{"code": "the-code", "redirectUri": testutil.Origin + "/verify/github"}

func TestPostOAuthLoginCreatesTheUser(t *testing.T) {
	testutil.Clean(t)
	internet := testutil.FakeInternet(t)
	testutil.FakeBrevo(internet)
	fakeGithub(internet, "42", "ada@example.test")

	response := do(t, http.MethodPost, "/auth/github", oauthBody)
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	tokens := testutil.Decode[tokenResponse](t, response)
	assertSession(t, tokens, tokens.User.Id)
	assert.NotContains(t, response.Body.String(), "gho_token", "provider tokens never leave the server")

	user := storedUser(t, tokens.User.Id)
	assert.Equal(t, "ada@example.test", user.Email)
	assert.Equal(t, "Ada Lovelace", user.Name)
	assert.Equal(t, "https://avatars.example/ada", user.ProfilePicture)
	assert.Equal(t, "mocha", user.Flavour)
	assert.WithinDuration(t, time.Now(), user.LastLogin, 5*time.Second)
	require.Len(t, user.Identities, 1)
	assert.Equal(t, "github", user.Identities[0].Provider)
	assert.Equal(t, "42", user.Identities[0].ProviderId)
	assert.Equal(t, "ada", user.Identities[0].Username)
	assert.Equal(t, "gho_token", user.Identities[0].AccessToken)
	assert.Equal(t, []string{"ada@example.test"}, testutil.AddedContacts(internet))
	assert.Contains(t, internet.Requests("github.com")[0].Body, "redirect_uri="+url.QueryEscape(testutil.Origin+"/verify/github"))
}

func TestPostOAuthLoginAttachesToTheUserWithTheSameEmail(t *testing.T) {
	testutil.Clean(t)
	internet := testutil.FakeInternet(t)
	testutil.FakeBrevo(internet)
	fakeGithub(internet, "42", "ADA@example.test")
	fakeGoogle(internet, "g-1", "ada@example.test")
	existing := createUser(t, users.User{Email: "ada@example.test", Name: "Countess", ProfilePicture: "https://example.test/me.png",
		Identities: []users.Identity{{Provider: "email", ProviderId: "ada@example.test"}}})

	response := do(t, http.MethodPost, "/auth/github", map[string]string{"code": "the-code"})
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	assert.Equal(t, existing.Id, testutil.Decode[tokenResponse](t, response).User.Id)
	require.Equal(t, http.StatusOK, do(t, http.MethodPost, "/auth/google", map[string]string{"code": "the-code"}).Code)

	assert.Equal(t, 1, testutil.Count(t, "users"))
	user := storedUser(t, existing.Id)
	assert.Equal(t, "Countess", user.Name, "a chosen name is kept")
	assert.Equal(t, "https://example.test/me.png", user.ProfilePicture)
	providers := []string{}
	for _, identity := range user.Identities {
		providers = append(providers, identity.Provider)
	}
	assert.Equal(t, []string{"email", "github", "google"}, providers)
	assert.Empty(t, testutil.AddedContacts(internet), "only new users join the mailing list")
}

func TestPostOAuthLoginLinksAnIdentityToTheLoggedInUser(t *testing.T) {
	testutil.Clean(t)
	internet := testutil.FakeInternet(t)
	testutil.FakeBrevo(internet)
	user := createUser(t, users.User{Email: "ada@example.test"})

	fakeGithub(internet, "42", "someone-else@example.test")
	conflict := do(t, http.MethodPost, "/auth/github", oauthBody, testutil.Bearer(t, user.Id))
	assert.Equal(t, http.StatusConflict, conflict.Code)
	assert.Equal(t, "This account uses another email address (someone-else@example.test)", testutil.ErrorOf(t, conflict))
	assert.Empty(t, storedUser(t, user.Id).Identities)
	assert.Equal(t, 1, testutil.Count(t, "users"), "no account is created for the other address")

	fakeGithub(internet, "42", "Ada@Example.test")
	linked := do(t, http.MethodPost, "/auth/github", oauthBody, testutil.Bearer(t, user.Id))
	require.Equal(t, http.StatusOK, linked.Code, linked.Body.String())
	assert.Equal(t, user.Id, testutil.Decode[tokenResponse](t, linked).User.Id)
	assert.Len(t, storedUser(t, user.Id).Identities, 1)
}

func TestPostOAuthLoginMovesAnAccountWhoseEmailChanged(t *testing.T) {
	testutil.Clean(t)
	internet := testutil.FakeInternet(t)
	testutil.FakeBrevo(internet)
	former := createUser(t, users.User{Email: "old@example.test", Identities: []users.Identity{
		{Provider: "email", ProviderId: "old@example.test"}, {Provider: "github", ProviderId: "42"},
	}})
	fakeGithub(internet, "42", "new@example.test")

	response := do(t, http.MethodPost, "/auth/github", oauthBody)
	require.Equal(t, http.StatusOK, response.Code)
	owner := testutil.Decode[tokenResponse](t, response).User

	assert.NotEqual(t, former.Id, owner.Id)
	assert.Equal(t, "new@example.test", owner.Email)
	remaining := storedUser(t, former.Id).Identities
	require.Len(t, remaining, 1, "the account is linked to one user only")
	assert.Equal(t, "email", remaining[0].Provider)
}

func TestPostOAuthLoginRejections(t *testing.T) {
	testutil.Clean(t)
	internet := testutil.FakeInternet(t)
	testutil.FakeBrevo(internet)
	fakeGithub(internet, "42", "ada@example.test")

	unknown := do(t, http.MethodPost, "/auth/facebook", oauthBody)
	assert.Equal(t, http.StatusNotFound, unknown.Code)
	assert.Equal(t, "Unknown identity provider", testutil.ErrorOf(t, unknown))

	for _, body := range []any{"not json", map[string]string{}, map[string]string{"code": "c", "redirectUri": "not a url"}} {
		assert.Equal(t, http.StatusBadRequest, do(t, http.MethodPost, "/auth/github", body).Code)
	}

	assert.Equal(t, http.StatusUnauthorized, do(t, http.MethodPost, "/auth/github", oauthBody, testutil.Header("Authorization", "Bearer garbage")).Code)
	assert.Equal(t, http.StatusUnauthorized, do(t, http.MethodPost, "/auth/github", oauthBody, testutil.Bearer(t, primitive.NewObjectID())).Code,
		"token of a user that no longer exists")

	internet.JSON("github.com", http.StatusOK, `{"error":"bad_verification_code"}`)
	refused := do(t, http.MethodPost, "/auth/github", oauthBody)
	assert.Equal(t, http.StatusBadRequest, refused.Code)
	assert.Equal(t, "Failed to authenticate with github", testutil.ErrorOf(t, refused), "provider errors are not leaked")
	assert.Zero(t, testutil.Count(t, "users"))

	t.Setenv("GOOGLE_CLIENT_SECRET", "")
	assert.Equal(t, http.StatusNotFound, do(t, http.MethodPost, "/auth/google", oauthBody).Code, "a provider without credentials is off")

	fakeGithub(internet, "42", "ada@example.test")
	testutil.BreakDB(t)
	assert.Equal(t, http.StatusInternalServerError, do(t, http.MethodPost, "/auth/github", oauthBody).Code)
}

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
	begin := do(t, http.MethodPost, path+"/begin", nil, testutil.As(userId), testutil.FromApp)
	require.Equal(t, http.StatusOK, begin.Code, begin.Body.String())
	session := testutil.Decode[passkeyBegin](t, begin)
	finish := do(t, http.MethodPost, path+"/finish",
		PasskeyFinishBody{SessionId: session.SessionId, Name: name, Credential: device.attest(t, session)}, testutil.As(userId), testutil.FromApp)
	require.Equal(t, http.StatusCreated, finish.Code, finish.Body.String())
}

func beginLogin(t *testing.T) passkeyBegin {
	t.Helper()
	begin := do(t, http.MethodPost, "/auth/passkey/begin", nil, testutil.FromApp)
	require.Equal(t, http.StatusOK, begin.Code, begin.Body.String())
	return testutil.Decode[passkeyBegin](t, begin)
}

func TestPasskeyRegistration(t *testing.T) {
	testutil.Clean(t)
	user := createUser(t, users.User{Email: "ada@example.test", Name: "Ada"})
	path := "/users/" + user.Id.Hex() + "/passkeys"
	laptop := newDevice(user.Id)

	begin := do(t, http.MethodPost, path+"/begin", nil, testutil.As(user.Id), testutil.FromApp)
	require.Equal(t, http.StatusOK, begin.Code, begin.Body.String())
	session := testutil.Decode[passkeyBegin](t, begin)
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
		PasskeyFinishBody{SessionId: session.SessionId, Name: "Laptop", Credential: laptop.attest(t, session)}, testutil.As(user.Id), testutil.FromApp)
	require.Equal(t, http.StatusCreated, finish.Code, finish.Body.String())
	assert.NotContains(t, finish.Body.String(), "credential")
	returned := testutil.Decode[users.User](t, finish)
	require.Len(t, returned.Passkeys, 1)
	assert.Equal(t, "Laptop", returned.Passkeys[0].Name)

	stored := storedUser(t, user.Id)
	require.Len(t, stored.Passkeys, 1)
	assert.Equal(t, users.PasskeyId(&webauthn.Credential{ID: laptop.credential.ID}), stored.Passkeys[0].Id)
	assert.WithinDuration(t, time.Now(), stored.Passkeys[0].CreatedAt, 5*time.Second)
	assert.Nil(t, stored.Passkeys[0].LastUsed)
	require.Len(t, stored.WebAuthnCredentials(), 1)

	replayed := do(t, http.MethodPost, path+"/finish",
		PasskeyFinishBody{SessionId: session.SessionId, Credential: laptop.attest(t, session)}, testutil.As(user.Id), testutil.FromApp)
	assert.Equal(t, http.StatusBadRequest, replayed.Code, "a ceremony cannot be replayed")

	phone := newDevice(user.Id)
	phone.register(t, user.Id, "")
	stored = storedUser(t, user.Id)
	require.Len(t, stored.Passkeys, 2)
	assert.Equal(t, "Passkey", stored.Passkeys[1].Name, "default name")

	second := testutil.Decode[passkeyBegin](t, do(t, http.MethodPost, path+"/begin", nil, testutil.As(user.Id), testutil.FromApp))
	require.NoError(t, json.Unmarshal(second.Options, &options))
	assert.Len(t, options.PublicKey.ExcludeCredentials, 2, "a device is not registered twice")
}

func TestPasskeyRegistrationRejections(t *testing.T) {
	testutil.Clean(t)
	user := createUser(t, users.User{})
	path := "/users/" + user.Id.Hex() + "/passkeys"
	stranger := primitive.NewObjectID()

	for _, step := range []string{"/begin", "/finish"} {
		noOrigin := do(t, http.MethodPost, path+step, nil, testutil.As(user.Id))
		assert.Equal(t, http.StatusBadRequest, noOrigin.Code)
		assert.Equal(t, "Passkeys are not available from this origin", testutil.ErrorOf(t, noOrigin))
		assert.Equal(t, http.StatusBadRequest, do(t, http.MethodPost, path+step, nil, testutil.As(user.Id), testutil.Header("Origin", "https://attacker.example")).Code)
		assert.Equal(t, http.StatusBadRequest, do(t, http.MethodPost, "/users/me/passkeys"+step, nil, testutil.As(user.Id), testutil.FromApp).Code)
		assert.Equal(t, http.StatusForbidden, do(t, http.MethodPost, path+step, nil, testutil.As(stranger), testutil.FromApp).Code)
		assert.Equal(t, http.StatusNotFound, do(t, http.MethodPost, "/users/"+stranger.Hex()+"/passkeys"+step, nil, testutil.As(stranger), testutil.FromApp).Code)
	}

	session := testutil.Decode[passkeyBegin](t, do(t, http.MethodPost, path+"/begin", nil, testutil.As(user.Id), testutil.FromApp))
	device := newDevice(user.Id)
	finish := func(body any) int {
		return do(t, http.MethodPost, path+"/finish", body, testutil.As(user.Id), testutil.FromApp).Code
	}
	assert.Equal(t, http.StatusBadRequest, finish("not json"))
	assert.Equal(t, http.StatusBadRequest, finish(map[string]any{"sessionId": session.SessionId}), "no credential")
	assert.Equal(t, http.StatusBadRequest, finish(PasskeyFinishBody{SessionId: session.SessionId, Name: strings.Repeat("a", 51), Credential: device.attest(t, session)}))
	assert.Equal(t, http.StatusBadRequest, finish(PasskeyFinishBody{SessionId: "unknown", Credential: device.attest(t, session)}))
	assert.Equal(t, http.StatusBadRequest, finish(PasskeyFinishBody{SessionId: session.SessionId, Credential: json.RawMessage(`{"id":"x"}`)}),
		"not a WebAuthn response")

	other := testutil.Decode[passkeyBegin](t, do(t, http.MethodPost, path+"/begin", nil, testutil.As(user.Id), testutil.FromApp))
	third := testutil.Decode[passkeyBegin](t, do(t, http.MethodPost, path+"/begin", nil, testutil.As(user.Id), testutil.FromApp))
	mismatch := do(t, http.MethodPost, path+"/finish", PasskeyFinishBody{SessionId: other.SessionId, Credential: device.attest(t, third)}, testutil.As(user.Id), testutil.FromApp)
	assert.Equal(t, http.StatusBadRequest, mismatch.Code)
	assert.Equal(t, "This passkey could not be verified", testutil.ErrorOf(t, mismatch), "answer to another challenge")
	assert.Empty(t, storedUser(t, user.Id).Passkeys)

	full := createUser(t, users.User{Passkeys: make([]users.Passkey, users.MaxPasskeys)})
	tooMany := do(t, http.MethodPost, "/users/"+full.Id.Hex()+"/passkeys/begin", nil, testutil.As(full.Id), testutil.FromApp)
	assert.Equal(t, http.StatusBadRequest, tooMany.Code)
	assert.Equal(t, "Too many passkeys, remove one first", testutil.ErrorOf(t, tooMany))
}

func TestPasskeyLogin(t *testing.T) {
	testutil.Clean(t)
	user := createUser(t, users.User{Email: "ada@example.test"})
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

	finish := do(t, http.MethodPost, "/auth/passkey/finish", PasskeyFinishBody{SessionId: session.SessionId, Credential: phone.assert(t, session)}, testutil.FromApp)
	require.Equal(t, http.StatusOK, finish.Code, finish.Body.String())
	tokens := testutil.Decode[tokenResponse](t, finish)
	assertSession(t, tokens, user.Id)
	assert.Equal(t, user.Id, tokens.User.Id)

	stored := storedUser(t, user.Id)
	assert.WithinDuration(t, time.Now(), stored.LastLogin, 5*time.Second)
	assert.Nil(t, stored.GetPasskey(stored.Passkeys[0].Id).LastUsed)
	require.NotNil(t, stored.Passkeys[1].LastUsed, "the passkey that was used is dated")
	assert.WithinDuration(t, time.Now(), *stored.Passkeys[1].LastUsed, 5*time.Second)

	replayed := do(t, http.MethodPost, "/auth/passkey/finish", PasskeyFinishBody{SessionId: session.SessionId, Credential: phone.assert(t, session)}, testutil.FromApp)
	assert.Equal(t, http.StatusBadRequest, replayed.Code, "a ceremony cannot be replayed")
}

func TestPasskeyLoginRejections(t *testing.T) {
	testutil.Clean(t)
	user := createUser(t, users.User{})
	laptop := newDevice(user.Id)
	laptop.register(t, user.Id, "Laptop")
	finish := func(body any, options ...testutil.Option) int {
		return do(t, http.MethodPost, "/auth/passkey/finish", body, options...).Code
	}

	assert.Equal(t, http.StatusBadRequest, do(t, http.MethodPost, "/auth/passkey/begin", nil).Code, "no origin")
	assert.Equal(t, http.StatusBadRequest, finish(nil), "no origin")
	assert.Equal(t, http.StatusBadRequest, finish("not json", testutil.FromApp))
	assert.Equal(t, http.StatusBadRequest, finish(PasskeyFinishBody{SessionId: "unknown", Credential: json.RawMessage(`{}`)}, testutil.FromApp))
	session := beginLogin(t)
	assert.Equal(t, http.StatusBadRequest, finish(PasskeyFinishBody{SessionId: session.SessionId, Credential: json.RawMessage(`{"id":"x"}`)}, testutil.FromApp))

	unknown := newDevice(user.Id)
	session = beginLogin(t)
	refused := do(t, http.MethodPost, "/auth/passkey/finish", PasskeyFinishBody{SessionId: session.SessionId, Credential: unknown.assert(t, session)}, testutil.FromApp)
	assert.Equal(t, http.StatusUnauthorized, refused.Code)
	assert.Equal(t, "This passkey is not recognized", testutil.ErrorOf(t, refused), "a key that was never registered")

	orphan := newDevice(primitive.NewObjectID())
	session = beginLogin(t)
	assert.Equal(t, http.StatusUnauthorized, finish(PasskeyFinishBody{SessionId: session.SessionId, Credential: orphan.assert(t, session)}, testutil.FromApp),
		"user that does not exist")

	malformed := &device{
		authenticator: virtualwebauthn.NewAuthenticatorWithOptions(virtualwebauthn.AuthenticatorOptions{UserHandle: []byte("short")}),
		credential:    laptop.credential,
	}
	session = beginLogin(t)
	assert.Equal(t, http.StatusUnauthorized, finish(PasskeyFinishBody{SessionId: session.SessionId, Credential: malformed.assert(t, session)}, testutil.FromApp),
		"user handle that is not an id")

	first, second := beginLogin(t), beginLogin(t)
	assert.Equal(t, http.StatusUnauthorized, finish(PasskeyFinishBody{SessionId: first.SessionId, Credential: laptop.assert(t, second)}, testutil.FromApp),
		"answer to another challenge")

	removed := do(t, http.MethodDelete, "/users/"+user.Id.Hex()+"/passkeys/"+storedUser(t, user.Id).Passkeys[0].Id, nil, testutil.As(user.Id))
	require.Equal(t, http.StatusOK, removed.Code)
	session = beginLogin(t)
	assert.Equal(t, http.StatusUnauthorized, finish(PasskeyFinishBody{SessionId: session.SessionId, Credential: laptop.assert(t, session)}, testutil.FromApp),
		"a removed passkey no longer logs in")

	testutil.BreakDB(t)
	assert.Equal(t, http.StatusInternalServerError, do(t, http.MethodPost, "/auth/passkey/begin", nil, testutil.FromApp).Code)
}

func TestDeletePasskey(t *testing.T) {
	testutil.Clean(t)
	user := createUser(t, users.User{Passkeys: []users.Passkey{{Id: "laptop", Name: "Laptop"}, {Id: "phone", Name: "Phone"}}})
	path := "/users/" + user.Id.Hex() + "/passkeys/"

	response := do(t, http.MethodDelete, path+"laptop", nil, testutil.As(user.Id))
	require.Equal(t, http.StatusOK, response.Code)
	remaining := testutil.Decode[users.User](t, response).Passkeys
	require.Len(t, remaining, 1)
	assert.Equal(t, "phone", remaining[0].Id)
	assert.Len(t, storedUser(t, user.Id).Passkeys, 1)

	missing := do(t, http.MethodDelete, path+"laptop", nil, testutil.As(user.Id))
	assert.Equal(t, http.StatusNotFound, missing.Code)
	assert.Equal(t, "Passkey not found", testutil.ErrorOf(t, missing))
	assert.Equal(t, http.StatusForbidden, do(t, http.MethodDelete, path+"phone", nil, testutil.As(primitive.NewObjectID())).Code)
	assert.Equal(t, http.StatusBadRequest, do(t, http.MethodDelete, "/users/me/passkeys/phone", nil, testutil.As(user.Id)).Code)
	assert.Len(t, storedUser(t, user.Id).Passkeys, 1)
}
