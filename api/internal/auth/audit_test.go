package auth

import (
	"net/http"
	"testing"

	"dev/internal/testutil"
	users "dev/internal/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestMagicLinkLoginIsAudited(t *testing.T) {
	testutil.Clean(t)
	resetLimiters()
	internet := testutil.FakeInternet(t)
	testutil.FakeBrevo(internet)
	login := func() int {
		return do(t, http.MethodPost, "/auth/login", map[string]string{"email": "ada@example.test"}, testutil.FromApp).Code
	}

	require.Equal(t, http.StatusNoContent, login())
	user, err := users.GetFullUserByEmail(ctx, "ada@example.test")
	require.NoError(t, err)
	entries := testutil.Audit(t)
	require.Equal(t, []string{"user.created", "auth.link_requested"}, testutil.AuditActions(t))
	assert.Equal(t, user.Id, entries[0].UserId)
	assert.Equal(t, "email", entries[0].Meta["via"])
	assert.Equal(t, user.Id, entries[1].UserId)
	assert.NotEmpty(t, entries[1].Ip)

	_, token := magicLink(t, internet)
	require.Equal(t, http.StatusOK, do(t, http.MethodPost, "/auth/verify", map[string]string{"token": token}).Code)
	require.Equal(t, http.StatusBadRequest, do(t, http.MethodPost, "/auth/verify", map[string]string{"token": token}).Code)
	entries = testutil.Audit(t)
	require.Equal(t, []string{"user.created", "auth.link_requested", "auth.login", "auth.login_failed"}, testutil.AuditActions(t))
	assert.Equal(t, user.Id, entries[2].UserId)
	assert.Equal(t, "email", entries[2].Meta["method"])
	assert.True(t, entries[3].UserId.IsZero())
	assert.Equal(t, "invalid_link", entries[3].Meta["reason"])

	require.Equal(t, http.StatusNoContent, login())
	require.Equal(t, http.StatusNoContent, login())
	require.Equal(t, http.StatusTooManyRequests, login())
	entries = testutil.Audit(t)
	limited := entries[len(entries)-1]
	assert.Equal(t, "access.rate_limited", limited.Action)
	assert.Equal(t, "ada@example.test", limited.Meta["email"])
	assert.Equal(t, 1, testutil.Count(t, "users"), "the user is only created once")
}

func TestOAuthLoginIsAudited(t *testing.T) {
	testutil.Clean(t)
	internet := testutil.FakeInternet(t)
	testutil.FakeBrevo(internet)
	fakeGithub(internet, "42", "ada@example.test")
	fakeGoogle(internet, "g-1", "ada@example.test")

	response := do(t, http.MethodPost, "/auth/github", oauthBody)
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	userId := testutil.Decode[tokenResponse](t, response).User.Id
	entries := testutil.Audit(t)
	require.Equal(t, []string{"user.created", "auth.login"}, testutil.AuditActions(t))
	assert.Equal(t, "github", entries[0].Meta["via"])
	assert.Equal(t, userId, entries[1].UserId)
	assert.Equal(t, "github", entries[1].Meta["method"])

	require.Equal(t, http.StatusOK, do(t, http.MethodPost, "/auth/github", oauthBody).Code)
	assert.Equal(t, []string{"user.created", "auth.login", "auth.login"}, testutil.AuditActions(t), "a known identity is not linked again")

}

func TestOAuthLinkingIsAudited(t *testing.T) {
	testutil.Clean(t)
	internet := testutil.FakeInternet(t)
	testutil.FakeBrevo(internet)
	fakeGithub(internet, "42", "ada@example.test")
	fakeGoogle(internet, "g-1", "other@example.test")
	user := createUser(t, users.User{Email: "ada@example.test"})

	require.Equal(t, http.StatusOK, do(t, http.MethodPost, "/auth/github", oauthBody, testutil.Bearer(t, user.Id)).Code)
	entries := testutil.Audit(t)
	require.Equal(t, []string{"user.identity_linked"}, testutil.AuditActions(t), "linking is not a login")
	assert.Equal(t, user.Id, entries[0].UserId)
	assert.Equal(t, "github", entries[0].Meta["provider"])

	require.Equal(t, http.StatusConflict, do(t, http.MethodPost, "/auth/google", map[string]string{"code": "the-code"}, testutil.Bearer(t, user.Id)).Code)
	entries = testutil.Audit(t)
	require.Len(t, entries, 2)
	assert.Equal(t, "auth.login_failed", entries[1].Action)
	assert.Equal(t, user.Id, entries[1].UserId)
	assert.Equal(t, "email_mismatch", entries[1].Meta["reason"])
	assert.Equal(t, "google", entries[1].Meta["method"])

	internet.JSON("github.com", http.StatusOK, `{"error":"bad_verification_code"}`)
	require.Equal(t, http.StatusBadRequest, do(t, http.MethodPost, "/auth/github", oauthBody).Code)
	entries = testutil.Audit(t)
	require.Len(t, entries, 3)
	assert.Equal(t, "auth.login_failed", entries[2].Action)
	assert.Equal(t, "provider_refused", entries[2].Meta["reason"])
	assert.True(t, entries[2].UserId.IsZero())
}

func TestPasskeysAreAudited(t *testing.T) {
	testutil.Clean(t)
	user := createUser(t, users.User{})
	laptop := newDevice(user.Id)

	laptop.register(t, user.Id, "Laptop")
	passkeyId := storedUser(t, user.Id).Passkeys[0].Id
	entries := testutil.Audit(t)
	require.Equal(t, []string{"user.passkey_added"}, testutil.AuditActions(t))
	assert.Equal(t, user.Id, entries[0].UserId)
	assert.Equal(t, passkeyId, entries[0].Meta["passkeyId"])
	assert.Equal(t, "Laptop", entries[0].Meta["name"])

	session := beginLogin(t)
	body := PasskeyFinishBody{SessionId: session.SessionId, Credential: laptop.assert(t, session)}
	require.Equal(t, http.StatusOK, do(t, http.MethodPost, "/auth/passkey/finish", body, testutil.FromApp).Code)
	require.Equal(t, http.StatusBadRequest, do(t, http.MethodPost, "/auth/passkey/finish", body, testutil.FromApp).Code)
	entries = testutil.Audit(t)
	require.Equal(t, []string{"user.passkey_added", "auth.login", "auth.login_failed"}, testutil.AuditActions(t))
	assert.Equal(t, user.Id, entries[1].UserId)
	assert.Equal(t, "passkey", entries[1].Meta["method"])
	assert.Equal(t, passkeyId, entries[1].Meta["passkeyId"])
	assert.Equal(t, "invalid_challenge", entries[2].Meta["reason"])

	stranger := newDevice(primitive.NewObjectID())
	session = beginLogin(t)
	body = PasskeyFinishBody{SessionId: session.SessionId, Credential: stranger.assert(t, session)}
	require.Equal(t, http.StatusUnauthorized, do(t, http.MethodPost, "/auth/passkey/finish", body, testutil.FromApp).Code)
	entries = testutil.Audit(t)
	assert.Equal(t, "not_recognized", entries[len(entries)-1].Meta["reason"])

	path := "/users/" + user.Id.Hex() + "/passkeys/" + passkeyId
	require.Equal(t, http.StatusOK, do(t, http.MethodDelete, path, nil, testutil.As(user.Id)).Code)
	entries = testutil.Audit(t)
	removed := entries[len(entries)-1]
	assert.Equal(t, "user.passkey_removed", removed.Action)
	assert.Equal(t, passkeyId, removed.Meta["passkeyId"])

	before := len(entries)
	require.Equal(t, http.StatusNotFound, do(t, http.MethodDelete, path, nil, testutil.As(user.Id)).Code)
	assert.Len(t, testutil.Audit(t), before, "nothing was removed")
}
