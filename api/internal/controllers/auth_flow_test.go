package controllers

import (
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"testing"
	"time"

	"dev/internal/models"
	"dev/internal/testutil"
	"dev/internal/utils"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var linkPattern = regexp.MustCompile(`<a href="([^"]+)">Verify your email address</a>`)

// magicLink returns the link of the last verification email and its token
func magicLink(t *testing.T, internet *testutil.Internet) (*url.URL, string) {
	t.Helper()
	emails := sentEmails(t, internet)
	require.NotEmpty(t, emails)
	match := linkPattern.FindStringSubmatch(emails[len(emails)-1].HTMLContent)
	require.NotNil(t, match, "no link in the email")
	link, err := url.Parse(html.UnescapeString(match[1]))
	require.NoError(t, err)
	return link, link.Query().Get("token")
}

type tokenResponse struct {
	AccessToken  string      `json:"accessToken"`
	RefreshToken string      `json:"refreshToken"`
	User         models.User `json:"user"`
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
	fakeBrevo(internet)

	response := do(t, http.MethodPost, "/auth/login", map[string]string{"email": "Ada@Example.test"},
		fromApp, header("Referer", testutil.Origin+"/tool/ormi?tab=stats"))
	assert.Equal(t, http.StatusNoContent, response.Code)
	assert.Empty(t, response.Body.String())

	user, err := models.GetFullUserByEmail(ginContext(), "ada@example.test")
	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, "Ada@Example.test", user.Email)
	assert.Equal(t, "mocha", user.Flavour)
	assert.Empty(t, user.Identities, "nothing is verified yet")

	emails := sentEmails(t, internet)
	require.Len(t, emails, 1)
	assert.Equal(t, "Ada@Example.test", emails[0].To[0].Email)
	assert.Equal(t, "Verify your email address", emails[0].Subject)
	link, token := magicLink(t, internet)
	assert.Equal(t, testutil.Origin+"/verify/email", link.Scheme+"://"+link.Host+link.Path)
	assert.Equal(t, "/tool/ormi?tab=stats", link.Query().Get("redirect"))

	consumed, err := models.ConsumeLoginToken(ctx, token)
	require.NoError(t, err)
	assert.Equal(t, user.Id, consumed)
}

func TestPostLoginReusesTheUser(t *testing.T) {
	testutil.Clean(t)
	resetLimiters()
	internet := testutil.FakeInternet(t)
	fakeBrevo(internet)
	user := createUser(t, models.User{Email: "ada@example.test", Name: "Ada"})

	assert.Equal(t, http.StatusNoContent, do(t, http.MethodPost, "/auth/login", map[string]string{"email": "ADA@example.test"}, fromApp).Code)
	_, first := magicLink(t, internet)
	assert.Equal(t, http.StatusNoContent, do(t, http.MethodPost, "/auth/login", map[string]string{"email": "ada@example.test"}, fromApp).Code)
	_, second := magicLink(t, internet)

	assert.Equal(t, 1, count(t, "users"))
	assert.Equal(t, "Ada", sentEmails(t, internet)[0].To[0].Name)
	_, err := models.ConsumeLoginToken(ctx, first)
	assert.Error(t, err, "asking for a new link revokes the previous one")
	consumed, err := models.ConsumeLoginToken(ctx, second)
	require.NoError(t, err)
	assert.Equal(t, user.Id, consumed)
}

func TestPostLoginRejections(t *testing.T) {
	testutil.Clean(t)
	resetLimiters()
	internet := testutil.FakeInternet(t)
	fakeBrevo(internet)

	for _, body := range []any{"not json", map[string]string{}, map[string]string{"email": "not-an-email"}} {
		assert.Equal(t, http.StatusBadRequest, do(t, http.MethodPost, "/auth/login", body, fromApp).Code)
	}
	assert.Zero(t, count(t, "users"))

	for _, origin := range []string{"", "https://attacker.example"} {
		refused := do(t, http.MethodPost, "/auth/login", map[string]string{"email": "ada@example.test"}, header("Origin", origin))
		assert.Equal(t, http.StatusBadRequest, refused.Code)
		assert.Equal(t, "origin not allowed", errorOf(t, refused))
	}
	assert.Empty(t, sentEmails(t, internet), "no link is ever sent to another site")

	internet.JSON(brevoHost, http.StatusInternalServerError, `{}`)
	assert.Equal(t, http.StatusInternalServerError, do(t, http.MethodPost, "/auth/login", map[string]string{"email": "bob@example.test"}, fromApp).Code)

	testutil.BreakDB(t)
	assert.Equal(t, http.StatusInternalServerError, do(t, http.MethodPost, "/auth/login", map[string]string{"email": "eve@example.test"}, fromApp).Code)
}

func TestPostLoginIsLimitedPerEmail(t *testing.T) {
	testutil.Clean(t)
	resetLimiters()
	internet := testutil.FakeInternet(t)
	fakeBrevo(internet)
	login := func(email string) int {
		return do(t, http.MethodPost, "/auth/login", map[string]string{"email": email}, fromApp).Code
	}

	for range 3 {
		assert.Equal(t, http.StatusNoContent, login("ada@example.test"))
	}
	limited := do(t, http.MethodPost, "/auth/login", map[string]string{"email": "ADA@example.test"}, fromApp)
	assert.Equal(t, http.StatusTooManyRequests, limited.Code)
	assert.NotEmpty(t, limited.Header().Get("Retry-After"))
	assert.Len(t, sentEmails(t, internet), 3)
	assert.Equal(t, http.StatusNoContent, login("grace@example.test"), "other addresses are not affected")
}

func TestPostVerify(t *testing.T) {
	testutil.Clean(t)
	internet := testutil.FakeInternet(t)
	fakeBrevo(internet)
	user := createUser(t, models.User{Email: "ada@example.test"})
	token, err := models.CreateLoginToken(ctx, user.Id)
	require.NoError(t, err)

	response := do(t, http.MethodPost, "/auth/verify", map[string]string{"token": token})
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	tokens := decode[tokenResponse](t, response)
	assertSession(t, tokens, user.Id)
	assert.Equal(t, user.Id, tokens.User.Id)
	assert.Equal(t, []string{"ada@example.test"}, addedContacts(internet))

	stored := storedUser(t, user.Id)
	require.Len(t, stored.Identities, 1)
	assert.Equal(t, "email", stored.Identities[0].Provider)
	assert.Equal(t, "ada@example.test", stored.Identities[0].ProviderId)
	assert.WithinDuration(t, time.Now(), stored.LastLogin, 5*time.Second)
	assert.WithinDuration(t, time.Now(), stored.LastRefresh, 5*time.Second)

	replayed := do(t, http.MethodPost, "/auth/verify", map[string]string{"token": token})
	assert.Equal(t, http.StatusBadRequest, replayed.Code, "a link works once")
	assert.Contains(t, errorOf(t, replayed), "invalid or has expired")
}

func TestPostVerifyRejections(t *testing.T) {
	testutil.Clean(t)
	internet := testutil.FakeInternet(t)
	internet.JSON(brevoHost, http.StatusInternalServerError, `{}`)

	for _, body := range []any{"not json", map[string]string{}, map[string]string{"token": "unknown"}} {
		assert.Equal(t, http.StatusBadRequest, do(t, http.MethodPost, "/auth/verify", body).Code)
	}

	gone, err := models.CreateLoginToken(ctx, primitive.NewObjectID())
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, do(t, http.MethodPost, "/auth/verify", map[string]string{"token": gone}).Code, "the user no longer exists")

	user := createUser(t, models.User{})
	token, err := models.CreateLoginToken(ctx, user.Id)
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
	assertSession(t, decode[tokenResponse](t, response), userId)

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
	fakeBrevo(internet)
	fakeGithub(internet, "42", "ada@example.test")

	response := do(t, http.MethodPost, "/auth/github", oauthBody)
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	tokens := decode[tokenResponse](t, response)
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
	assert.Equal(t, []string{"ada@example.test"}, addedContacts(internet))
	assert.Contains(t, internet.Requests("github.com")[0].Body, "redirect_uri="+url.QueryEscape(testutil.Origin+"/verify/github"))
}

func TestPostOAuthLoginAttachesToTheUserWithTheSameEmail(t *testing.T) {
	testutil.Clean(t)
	internet := testutil.FakeInternet(t)
	fakeBrevo(internet)
	fakeGithub(internet, "42", "ADA@example.test")
	fakeGoogle(internet, "g-1", "ada@example.test")
	existing := createUser(t, models.User{Email: "ada@example.test", Name: "Countess", ProfilePicture: "https://example.test/me.png",
		Identities: []models.Identity{{Provider: "email", ProviderId: "ada@example.test"}}})

	response := do(t, http.MethodPost, "/auth/github", map[string]string{"code": "the-code"})
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	assert.Equal(t, existing.Id, decode[tokenResponse](t, response).User.Id)
	require.Equal(t, http.StatusOK, do(t, http.MethodPost, "/auth/google", map[string]string{"code": "the-code"}).Code)

	assert.Equal(t, 1, count(t, "users"))
	user := storedUser(t, existing.Id)
	assert.Equal(t, "Countess", user.Name, "a chosen name is kept")
	assert.Equal(t, "https://example.test/me.png", user.ProfilePicture)
	providers := []string{}
	for _, identity := range user.Identities {
		providers = append(providers, identity.Provider)
	}
	assert.Equal(t, []string{"email", "github", "google"}, providers)
	assert.Empty(t, addedContacts(internet), "only new users join the mailing list")
}

func TestPostOAuthLoginLinksAnIdentityToTheLoggedInUser(t *testing.T) {
	testutil.Clean(t)
	internet := testutil.FakeInternet(t)
	fakeBrevo(internet)
	user := createUser(t, models.User{Email: "ada@example.test"})

	fakeGithub(internet, "42", "someone-else@example.test")
	conflict := do(t, http.MethodPost, "/auth/github", oauthBody, bearer(t, user.Id))
	assert.Equal(t, http.StatusConflict, conflict.Code)
	assert.Equal(t, "This account uses another email address (someone-else@example.test)", errorOf(t, conflict))
	assert.Empty(t, storedUser(t, user.Id).Identities)
	assert.Equal(t, 1, count(t, "users"), "no account is created for the other address")

	fakeGithub(internet, "42", "Ada@Example.test")
	linked := do(t, http.MethodPost, "/auth/github", oauthBody, bearer(t, user.Id))
	require.Equal(t, http.StatusOK, linked.Code, linked.Body.String())
	assert.Equal(t, user.Id, decode[tokenResponse](t, linked).User.Id)
	assert.Len(t, storedUser(t, user.Id).Identities, 1)
}

func TestPostOAuthLoginMovesAnAccountWhoseEmailChanged(t *testing.T) {
	testutil.Clean(t)
	internet := testutil.FakeInternet(t)
	fakeBrevo(internet)
	former := createUser(t, models.User{Email: "old@example.test", Identities: []models.Identity{
		{Provider: "email", ProviderId: "old@example.test"}, {Provider: "github", ProviderId: "42"},
	}})
	fakeGithub(internet, "42", "new@example.test")

	response := do(t, http.MethodPost, "/auth/github", oauthBody)
	require.Equal(t, http.StatusOK, response.Code)
	owner := decode[tokenResponse](t, response).User

	assert.NotEqual(t, former.Id, owner.Id)
	assert.Equal(t, "new@example.test", owner.Email)
	remaining := storedUser(t, former.Id).Identities
	require.Len(t, remaining, 1, "the account is linked to one user only")
	assert.Equal(t, "email", remaining[0].Provider)
}

func TestPostOAuthLoginRejections(t *testing.T) {
	testutil.Clean(t)
	internet := testutil.FakeInternet(t)
	fakeBrevo(internet)
	fakeGithub(internet, "42", "ada@example.test")

	unknown := do(t, http.MethodPost, "/auth/facebook", oauthBody)
	assert.Equal(t, http.StatusNotFound, unknown.Code)
	assert.Equal(t, "Unknown identity provider", errorOf(t, unknown))

	for _, body := range []any{"not json", map[string]string{}, map[string]string{"code": "c", "redirectUri": "not a url"}} {
		assert.Equal(t, http.StatusBadRequest, do(t, http.MethodPost, "/auth/github", body).Code)
	}

	assert.Equal(t, http.StatusUnauthorized, do(t, http.MethodPost, "/auth/github", oauthBody, header("Authorization", "Bearer garbage")).Code)
	assert.Equal(t, http.StatusUnauthorized, do(t, http.MethodPost, "/auth/github", oauthBody, bearer(t, primitive.NewObjectID())).Code,
		"token of a user that no longer exists")

	internet.JSON("github.com", http.StatusOK, `{"error":"bad_verification_code"}`)
	refused := do(t, http.MethodPost, "/auth/github", oauthBody)
	assert.Equal(t, http.StatusBadRequest, refused.Code)
	assert.Equal(t, "Failed to authenticate with github", errorOf(t, refused), "provider errors are not leaked")
	assert.Zero(t, count(t, "users"))

	t.Setenv("GOOGLE_CLIENT_SECRET", "")
	assert.Equal(t, http.StatusNotFound, do(t, http.MethodPost, "/auth/google", oauthBody).Code, "a provider without credentials is off")

	fakeGithub(internet, "42", "ada@example.test")
	testutil.BreakDB(t)
	assert.Equal(t, http.StatusInternalServerError, do(t, http.MethodPost, "/auth/github", oauthBody).Code)
}

func TestVerificationEmailEscapesTheLink(t *testing.T) {
	internet := testutil.FakeInternet(t)
	fakeBrevo(internet)

	require.NoError(t, SendVerificationEmail(ginContext(), &models.User{Email: "ada@example.test"}, `https://x/?a=1&b="><script>`))

	content := sentEmails(t, internet)[0].HTMLContent
	assert.Contains(t, content, `<a href="https://x/?a=1&amp;b=&#34;&gt;&lt;script&gt;">`)
	assert.NotContains(t, content, "<script>")
}

func TestSignTokenPair(t *testing.T) {
	userId := primitive.NewObjectID()
	refresh, access, err := SignTokenPair(ginContext(), userId.Hex())
	require.NoError(t, err)
	assertSession(t, tokenResponse{AccessToken: access, RefreshToken: refresh}, userId)
}
