package auth

import (
	"net/http"
	"testing"

	"dev/internal/testutil"
	users "dev/internal/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTheSiteOwnerBecomesAdminOnceLoggedIn(t *testing.T) {
	testutil.Clean(t)
	resetLimiters()
	internet := testutil.FakeInternet(t)
	testutil.FakeBrevo(internet)
	login := func(email string) *users.User {
		require.Equal(t, http.StatusNoContent, do(t, http.MethodPost, "/auth/login", map[string]string{"email": email}, testutil.FromApp).Code)
		user, err := users.GetFullUserByEmail(ctx, email)
		require.NoError(t, err)
		return user
	}

	owner := login("Admin@Example.test")
	assert.False(t, owner.IsAdmin(), "asking for a link proves nothing")

	_, token := magicLink(t, internet)
	response := do(t, http.MethodPost, "/auth/verify", map[string]string{"token": token})
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	assert.Equal(t, users.RoleAdmin, testutil.Decode[tokenResponse](t, response).User.Role, "the session tells the app")
	assert.True(t, storedUser(t, owner.Id).IsAdmin())
	assert.Contains(t, testutil.AuditActions(t), "admin.role_changed")

	visitor := login("ada@example.test")
	_, token = magicLink(t, internet)
	response = do(t, http.MethodPost, "/auth/verify", map[string]string{"token": token})
	require.Equal(t, http.StatusOK, response.Code)
	assert.Empty(t, testutil.Decode[tokenResponse](t, response).User.Role)
	assert.False(t, storedUser(t, visitor.Id).IsAdmin())
}

func TestTheSiteOwnerBecomesAdminWithAProviderOrAPasskey(t *testing.T) {
	testutil.Clean(t)
	internet := testutil.FakeInternet(t)
	testutil.FakeBrevo(internet)
	fakeGithub(internet, "42", testutil.AdminEmail)

	response := do(t, http.MethodPost, "/auth/github", oauthBody)
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	owner := testutil.Decode[tokenResponse](t, response).User
	assert.Equal(t, users.RoleAdmin, owner.Role)

	require.NoError(t, users.SetUserRole(ctx, owner.Id, ""))
	phone := newDevice(owner.Id)
	phone.register(t, owner.Id, "Phone")
	session := beginLogin(t)
	finish := do(t, http.MethodPost, "/auth/passkey/finish", PasskeyFinishBody{SessionId: session.SessionId, Credential: phone.assert(t, session)}, testutil.FromApp)
	require.Equal(t, http.StatusOK, finish.Code, finish.Body.String())
	assert.True(t, storedUser(t, owner.Id).IsAdmin())
	assert.Len(t, storedUser(t, owner.Id).Passkeys, 1, "the role is set without touching the rest of the account")
}
