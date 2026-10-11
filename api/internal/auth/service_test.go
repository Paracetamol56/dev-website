package auth

import (
	"context"
	"dev/internal/testutil"
	users "dev/internal/user"
	"dev/internal/utils"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestVerificationEmailEscapesTheLink(t *testing.T) {
	internet := testutil.FakeInternet(t)
	testutil.FakeBrevo(internet)

	require.NoError(t, SendVerificationEmail(testutil.Context(), &users.User{Email: "ada@example.test"}, `https://x/?a=1&b="><script>`))

	content := testutil.SentEmails(t, internet)[0].HTMLContent
	assert.Contains(t, content, `<a href="https://x/?a=1&amp;b=&#34;&gt;&lt;script&gt;">`)
	assert.NotContains(t, content, "<script>")
}

func TestSignTokenPair(t *testing.T) {
	userId := primitive.NewObjectID()
	refresh, access, err := SignTokenPair(testutil.Context(), userId.Hex())
	require.NoError(t, err)
	assertSession(t, tokenResponse{AccessToken: access, RefreshToken: refresh}, userId)
}

func TestVerificationLink(t *testing.T) {
	const origin = "https://dev.matheo-galuba.com"

	_, err := verificationLink("https://attacker.example", origin+"/tool/ormi", "tok")
	assert.Error(t, err, "unknown origins must not receive login links")
	_, err = verificationLink("", "", "tok")
	assert.Error(t, err)

	cases := map[string]string{
		origin + "/tool/ormi?tab=stats":  "%2Ftool%2Formi%3Ftab%3Dstats",
		"https://attacker.example/phish": "%2Fphish",
		"":                               "%2F",
		"https://x//attacker.example/a":  "%2F",
		`https://x/\attacker.example`:    "%2F%255Cattacker.example",
		`"><a href="https://evil">`:      "%2F",
		"javascript:alert(1)":            "%2F",
	}
	for referer, redirect := range cases {
		link, err := verificationLink(origin, referer, "a.b-c")
		assert.NoError(t, err)
		assert.Equal(t, origin+"/verify/email?redirect="+redirect+"&token=a.b-c", link, referer)
	}
}

func TestOAuthProviders(t *testing.T) {
	names := []string{}
	for _, provider := range GetOAuthProviders() {
		names = append(names, provider.Name)
	}
	assert.Equal(t, []string{"github", "google"}, names)

	github := GetOAuthProvider("github")
	require.NotNil(t, github)
	assert.Equal(t, "test-github_client_id", github.ClientId())
	assert.Nil(t, GetOAuthProvider("facebook"))

	t.Setenv("GOOGLE_CLIENT_SECRET", "")
	assert.Nil(t, GetOAuthProvider("google"), "a provider without credentials is disabled")
	assert.Len(t, GetOAuthProviders(), 1)
	t.Setenv("GITHUB_CLIENT_ID", "")
	assert.Empty(t, GetOAuthProviders())
}

func fakeGithubAPI(internet *testutil.Internet, emails string) {
	internet.JSON("github.com", http.StatusOK, `{"access_token":"gho_token"}`)
	internet.Handle("api.github.com", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer gho_token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		switch r.URL.Path {
		case "/user":
			io.WriteString(w, `{"id":42,"login":"ada","name":"Ada Lovelace","avatar_url":"https://avatars.example/ada","html_url":"https://github.com/ada"}`)
		case "/user/emails":
			io.WriteString(w, emails)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
}

func TestGithubAuthenticate(t *testing.T) {
	internet := testutil.FakeInternet(t)
	fakeGithubAPI(internet, `[
		{"email":"old@example.test","primary":false,"verified":true},
		{"email":"ada@example.test","primary":true,"verified":true}
	]`)

	identity, err := GetOAuthProvider("github").Authenticate(context.Background(), "the-code", "http://localhost:5173/verify/github")
	require.NoError(t, err)
	assert.Equal(t, &users.Identity{
		Provider:    "github",
		ProviderId:  "42",
		Email:       "ada@example.test",
		Username:    "ada",
		Name:        "Ada Lovelace",
		AvatarUrl:   "https://avatars.example/ada",
		ProfileUrl:  "https://github.com/ada",
		AccessToken: "gho_token",
	}, identity)

	exchange := internet.Requests("github.com")
	require.Len(t, exchange, 1)
	assert.Equal(t, "/login/oauth/access_token", exchange[0].Path)
	form, err := url.ParseQuery(exchange[0].Body)
	require.NoError(t, err)
	assert.Equal(t, url.Values{
		"client_id":     {"test-github_client_id"},
		"client_secret": {"test-github_client_secret"},
		"code":          {"the-code"},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {"http://localhost:5173/verify/github"},
	}, form)
}

func TestGithubAuthenticateFailures(t *testing.T) {
	ctx := context.Background()
	github := GetOAuthProvider("github")

	internet := testutil.FakeInternet(t)
	_, err := github.Authenticate(ctx, "code", "")
	assert.ErrorContains(t, err, "unexpected request", "token endpoint unreachable")

	internet.JSON("github.com", http.StatusOK, `{"error":"bad_verification_code","error_description":"The code is incorrect"}`)
	_, err = github.Authenticate(ctx, "code", "")
	assert.ErrorContains(t, err, "github refused the authorization code: bad_verification_code The code is incorrect")
	form, _ := url.ParseQuery(internet.Requests("github.com")[0].Body)
	assert.NotContains(t, form, "redirect_uri")

	internet.JSON("github.com", http.StatusOK, `<html>`)
	_, err = github.Authenticate(ctx, "code", "")
	assert.Error(t, err, "token endpoint not answering JSON")

	fakeGithubAPI(internet, `[{"email":"ada@example.test","primary":true,"verified":false}]`)
	_, err = github.Authenticate(ctx, "code", "")
	assert.EqualError(t, err, "this GitHub account has no verified primary email")

	fakeGithubAPI(internet, `{}`)
	_, err = github.Authenticate(ctx, "code", "")
	assert.Error(t, err, "emails not answering a list")

	internet.JSON("github.com", http.StatusOK, `{"access_token":"revoked"}`)
	_, err = github.Authenticate(ctx, "code", "")
	assert.ErrorContains(t, err, "GET https://api.github.com/user: status 401")

	internet.Handle("api.github.com", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/user" {
			io.WriteString(w, `{"id":42}`)
			return
		}
		w.WriteHeader(http.StatusForbidden)
	})
	_, err = github.Authenticate(ctx, "code", "")
	assert.ErrorContains(t, err, "GET https://api.github.com/user/emails: status 403")

	_, err = github.Authenticate(nil, "code", "")
	assert.Error(t, err, "nil context")
	assert.Error(t, getOAuthResource(nil, "https://api.github.com/user", "token", nil), "nil context")
	internet = testutil.FakeInternet(t)
	assert.ErrorContains(t, getOAuthResource(ctx, "https://api.github.com/user", "token", nil), "unexpected request")
}

func TestGoogleAuthenticate(t *testing.T) {
	ctx := context.Background()
	internet := testutil.FakeInternet(t)
	internet.JSON("oauth2.googleapis.com", http.StatusOK, `{"access_token":"ya29"}`)
	internet.JSON("openidconnect.googleapis.com", http.StatusOK,
		`{"sub":"1234567890","email":"ada@example.test","email_verified":true,"name":"Ada Lovelace","picture":"https://lh3.example/ada"}`)

	identity, err := GetOAuthProvider("google").Authenticate(ctx, "the-code", "http://localhost:5173/verify/google")
	require.NoError(t, err)
	assert.Equal(t, &users.Identity{
		Provider:    "google",
		ProviderId:  "1234567890",
		Email:       "ada@example.test",
		Name:        "Ada Lovelace",
		AvatarUrl:   "https://lh3.example/ada",
		AccessToken: "ya29",
	}, identity)
	userinfo := internet.Requests("openidconnect.googleapis.com")
	require.Len(t, userinfo, 1)
	assert.Equal(t, "/v1/userinfo", userinfo[0].Path)
	assert.Equal(t, "Bearer ya29", userinfo[0].Header.Get("Authorization"))

	internet.JSON("openidconnect.googleapis.com", http.StatusOK, `{"sub":"1","email":"ada@example.test","email_verified":false}`)
	_, err = GetOAuthProvider("google").Authenticate(ctx, "the-code", "")
	assert.EqualError(t, err, "this Google account has no verified email")

	internet.JSON("openidconnect.googleapis.com", http.StatusOK, `{"sub":"1","email_verified":true}`)
	_, err = GetOAuthProvider("google").Authenticate(ctx, "the-code", "")
	assert.EqualError(t, err, "this Google account has no verified email")

	internet.JSON("openidconnect.googleapis.com", http.StatusInternalServerError, ``)
	_, err = GetOAuthProvider("google").Authenticate(ctx, "the-code", "")
	assert.ErrorContains(t, err, "status 500")
}

func TestNewWebAuthn(t *testing.T) {
	webAuthn, err := NewWebAuthn("https://dev.matheo-galuba.com")
	assert.NoError(t, err)
	assert.Equal(t, "dev.matheo-galuba.com", webAuthn.Config.RPID)
	assert.Equal(t, []string{"https://dev.matheo-galuba.com"}, webAuthn.Config.RPOrigins)

	webAuthn, err = NewWebAuthn("http://localhost:5173")
	assert.NoError(t, err)
	assert.Equal(t, "localhost", webAuthn.Config.RPID)

	_, err = NewWebAuthn("https://attacker.example")
	assert.EqualError(t, err, "origin not allowed")
	_, err = NewWebAuthn("")
	assert.Error(t, err)

	utils.AllowedOrigins = append(utils.AllowedOrigins, "http://bad\x7f")
	defer func() { utils.AllowedOrigins = utils.AllowedOrigins[:len(utils.AllowedOrigins)-1] }()
	_, err = NewWebAuthn("http://bad\x7f")
	assert.Error(t, err, "unparsable origin")
}
