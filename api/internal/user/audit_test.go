package user

import (
	"net/http"
	"testing"

	"dev/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAccountActionsAreAudited(t *testing.T) {
	testutil.Clean(t)
	internet := testutil.FakeInternet(t)
	testutil.FakeBrevo(internet)
	user := createUser(t, User{Name: "Ada", Identities: []Identity{
		{Provider: IdentityProviderEmail, ProviderId: "ada@example.test"},
		{Provider: "github", ProviderId: "42"},
	}})
	other := createUser(t, User{})
	path := "/users/" + user.Id.Hex()

	require.Equal(t, http.StatusOK, do(t, http.MethodPatch, path, map[string]string{"flavour": "latte"}, testutil.As(user.Id)).Code)
	require.Equal(t, http.StatusOK, do(t, http.MethodPatch, path, map[string]string{"name": "Ada"}, testutil.As(user.Id)).Code)
	assert.Empty(t, testutil.Audit(t), "a theme or an unchanged name is not worth an entry")

	require.Equal(t, http.StatusOK, do(t, http.MethodPatch, path,
		map[string]string{"name": "Countess", "profilePicture": "https://example.test/me.png"}, testutil.As(user.Id)).Code)
	require.Equal(t, http.StatusOK, do(t, http.MethodGet, path, nil, testutil.As(user.Id)).Code)
	require.Equal(t, http.StatusOK, do(t, http.MethodGet, path+"/export", nil, testutil.As(user.Id)).Code)
	require.Equal(t, http.StatusOK, do(t, http.MethodDelete, path+"/identities/github", nil, testutil.As(user.Id)).Code)
	require.Equal(t, http.StatusNotFound, do(t, http.MethodDelete, path+"/identities/github", nil, testutil.As(user.Id)).Code)

	entries := testutil.Audit(t)
	require.Equal(t, []string{"user.profile_updated", "user.data_exported", "user.identity_unlinked"}, testutil.AuditActions(t))
	assert.ElementsMatch(t, []any{"name", "profilePicture"}, entries[0].Meta["fields"])
	assert.Equal(t, "github", entries[2].Meta["provider"])
	for _, entry := range entries {
		assert.Equal(t, user.Id, entry.UserId)
	}

	testutil.Clean(t)
	user = createUser(t, User{})
	path = "/users/" + user.Id.Hex()
	for _, request := range [][2]string{
		{http.MethodPatch, path}, {http.MethodGet, path + "/export"}, {http.MethodDelete, path + "/identities/github"}, {http.MethodDelete, path},
	} {
		require.Equal(t, http.StatusForbidden, do(t, request[0], request[1], map[string]string{"name": "Mallory"}, testutil.As(other.Id)).Code, request)
	}
	entries = testutil.Audit(t)
	require.Equal(t, []string{"access.denied", "access.denied", "access.denied", "access.denied"}, testutil.AuditActions(t))
	assert.Equal(t, other.Id, entries[0].UserId)
	assert.Equal(t, "user", entries[0].Target.Type)
	assert.Equal(t, user.Id.Hex(), entries[0].Target.Id)

	require.Equal(t, http.StatusNoContent, do(t, http.MethodDelete, path, nil, testutil.As(user.Id)).Code)
	entries = testutil.Audit(t)
	deleted := entries[len(entries)-1]
	assert.Equal(t, "user.deleted", deleted.Action)
	assert.Equal(t, user.Id, deleted.UserId, "the entry outlives the account")
	assert.Equal(t, "user", deleted.Meta["by"])
}
