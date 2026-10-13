package admin

import (
	"net/http"
	"strconv"
	"testing"
	"time"

	"dev/internal/audit"
	"dev/internal/testutil"
	users "dev/internal/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestOnlyAdminsGetIn(t *testing.T) {
	testutil.Clean(t)
	admin := createUser(t, users.User{Role: users.RoleAdmin})
	visitor := createUser(t, users.User{})
	id := visitor.Id.Hex()
	routes := [][2]string{
		{http.MethodGet, "/admin/users"}, {http.MethodGet, "/admin/users/" + id}, {http.MethodPatch, "/admin/users/" + id + "/role"},
		{http.MethodDelete, "/admin/users/" + id}, {http.MethodGet, "/admin/audit"},
	}

	for _, route := range routes {
		assert.Equal(t, http.StatusUnauthorized, do(t, route[0], route[1], nil).Code, route)
		assert.Equal(t, http.StatusForbidden, do(t, route[0], route[1], map[string]string{"role": "admin"}, testutil.As(visitor.Id)).Code, route)
		assert.Equal(t, http.StatusForbidden, do(t, route[0], route[1], nil, testutil.As(primitive.NewObjectID())).Code, route)
	}
	assert.False(t, storedUser(t, visitor.Id).IsAdmin())
	entries := testutil.Audit(t)
	require.Len(t, entries, 2*len(routes))
	assert.Equal(t, "access.denied", entries[0].Action)
	assert.Equal(t, visitor.Id, entries[0].UserId)
	assert.Equal(t, "admin", entries[0].Target.Type)
	assert.Equal(t, "/admin/users", entries[0].Target.Id)

	assert.Equal(t, http.StatusOK, do(t, http.MethodGet, "/admin/users", nil, testutil.As(admin.Id)).Code)
	require.NoError(t, users.SetUserRole(ctx, admin.Id, ""))
	assert.Equal(t, http.StatusForbidden, do(t, http.MethodGet, "/admin/users", nil, testutil.As(admin.Id)).Code, "a demotion applies at once")

	testutil.BreakDB(t)
	assert.Equal(t, http.StatusForbidden, do(t, http.MethodGet, "/admin/users", nil, testutil.As(admin.Id)).Code)
}

func TestGetUsers(t *testing.T) {
	testutil.Clean(t)
	admin := createUser(t, users.User{Email: "root@example.test", Role: users.RoleAdmin})
	for i := range 30 {
		createUser(t, users.User{Email: "user" + strconv.Itoa(i) + "@crowd.test", Name: "User " + strconv.Itoa(i)})
	}
	get := func(query string) UsersPage {
		response := do(t, http.MethodGet, "/admin/users"+query, nil, testutil.As(admin.Id))
		require.Equal(t, http.StatusOK, response.Code, response.Body.String())
		return testutil.Decode[UsersPage](t, response)
	}

	page := get("")
	assert.EqualValues(t, 31, page.Total)
	assert.Equal(t, 1, page.Page)
	assert.Equal(t, DefaultLimit, page.Limit)
	assert.Len(t, page.Users, DefaultLimit)

	page = get("?page=2&limit=20")
	assert.Len(t, page.Users, 11)
	assert.Equal(t, 2, page.Page)

	page = get("?search=ROOT")
	require.Len(t, page.Users, 1)
	assert.Equal(t, admin.Id, page.Users[0].Id)
	assert.Equal(t, users.RoleAdmin, page.Users[0].Role)
	assert.EqualValues(t, 1, page.Total)

	assert.Empty(t, get("?search=nobody").Users)
	assert.Contains(t, do(t, http.MethodGet, "/admin/users?search=nobody", nil, testutil.As(admin.Id)).Body.String(), `"users":[]`)

	for _, query := range []string{"?page=-1", "?page=x", "?limit=101", "?limit=-1"} {
		assert.Equal(t, http.StatusBadRequest, do(t, http.MethodGet, "/admin/users"+query, nil, testutil.As(admin.Id)).Code, query)
	}
}

func TestGetUser(t *testing.T) {
	testutil.Clean(t)
	admin := createUser(t, users.User{Role: users.RoleAdmin})
	visitor := createUser(t, users.User{Email: "ada@example.test", Name: "Ada"})
	_, err := users.UpdateUser(ctx, visitor.Id, &users.User{Email: "ada@example.test", Name: "Ada", Flavour: "mocha",
		Identities: []users.Identity{{Provider: "github", ProviderId: "42", AccessToken: "gho_secret"}}})
	require.NoError(t, err)

	response := do(t, http.MethodGet, "/admin/users/"+visitor.Id.Hex(), nil, testutil.As(admin.Id))
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	found := testutil.Decode[users.User](t, response)
	assert.Equal(t, "ada@example.test", found.Email)
	require.Len(t, found.Identities, 1)
	assert.NotContains(t, response.Body.String(), "gho_secret", "provider tokens stay on the server")

	assert.Equal(t, http.StatusNotFound, do(t, http.MethodGet, "/admin/users/"+primitive.NewObjectID().Hex(), nil, testutil.As(admin.Id)).Code)
	assert.Equal(t, http.StatusBadRequest, do(t, http.MethodGet, "/admin/users/nope", nil, testutil.As(admin.Id)).Code)
}

func TestPatchUserRole(t *testing.T) {
	testutil.Clean(t)
	admin := createUser(t, users.User{Role: users.RoleAdmin})
	owner := createUser(t, users.User{Email: testutil.AdminEmail, Role: users.RoleAdmin})
	visitor := createUser(t, users.User{})
	patch := func(id primitive.ObjectID, body any) int {
		return do(t, http.MethodPatch, "/admin/users/"+id.Hex()+"/role", body, testutil.As(admin.Id)).Code
	}

	response := do(t, http.MethodPatch, "/admin/users/"+visitor.Id.Hex()+"/role", map[string]string{"role": "admin"}, testutil.As(admin.Id))
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	assert.Equal(t, users.RoleAdmin, testutil.Decode[users.User](t, response).Role)
	assert.True(t, storedUser(t, visitor.Id).IsAdmin())
	assert.Equal(t, http.StatusOK, patch(visitor.Id, map[string]string{"role": "admin"}))

	entries := testutil.Audit(t)
	require.Len(t, entries, 1, "an unchanged role is not recorded")
	assert.Equal(t, "admin.role_changed", entries[0].Action)
	assert.Equal(t, admin.Id, entries[0].UserId)
	assert.Equal(t, visitor.Id.Hex(), entries[0].Target.Id)
	assert.Equal(t, "admin", entries[0].Meta["role"])

	assert.Equal(t, http.StatusOK, patch(visitor.Id, map[string]string{"role": ""}))
	assert.False(t, storedUser(t, visitor.Id).IsAdmin())
	assert.Equal(t, "", testutil.Audit(t)[1].Meta["role"])

	assert.Equal(t, http.StatusBadRequest, patch(admin.Id, map[string]string{"role": ""}), "not oneself")
	assert.Equal(t, http.StatusForbidden, patch(owner.Id, map[string]string{"role": ""}), "not the site owner")
	assert.Equal(t, http.StatusBadRequest, patch(visitor.Id, map[string]string{"role": "superuser"}))
	assert.Equal(t, http.StatusBadRequest, patch(visitor.Id, "not json"))
	assert.Equal(t, http.StatusNotFound, patch(primitive.NewObjectID(), map[string]string{"role": "admin"}))
	assert.Equal(t, http.StatusBadRequest, do(t, http.MethodPatch, "/admin/users/nope/role", nil, testutil.As(admin.Id)).Code)
	assert.True(t, storedUser(t, admin.Id).IsAdmin())
	assert.True(t, storedUser(t, owner.Id).IsAdmin())
	assert.Len(t, testutil.Audit(t), 2)
}

func TestDeleteUser(t *testing.T) {
	testutil.Clean(t)
	internet := testutil.FakeInternet(t)
	testutil.FakeBrevo(internet)
	admin := createUser(t, users.User{Role: users.RoleAdmin})
	owner := createUser(t, users.User{Email: testutil.AdminEmail, Role: users.RoleAdmin})
	visitor := createUser(t, users.User{Email: "ada@example.test"})
	remove := func(id primitive.ObjectID) int {
		return do(t, http.MethodDelete, "/admin/users/"+id.Hex(), nil, testutil.As(admin.Id)).Code
	}

	assert.Equal(t, http.StatusBadRequest, remove(admin.Id), "not oneself")
	assert.Equal(t, http.StatusForbidden, remove(owner.Id), "not the site owner")
	assert.Equal(t, http.StatusNotFound, remove(primitive.NewObjectID()))
	assert.Equal(t, http.StatusBadRequest, do(t, http.MethodDelete, "/admin/users/nope", nil, testutil.As(admin.Id)).Code)
	assert.Empty(t, testutil.SentEmails(t, internet))
	assert.Empty(t, testutil.Audit(t))

	assert.Equal(t, http.StatusNoContent, remove(visitor.Id))
	_, err := users.GetFullUserById(ctx, visitor.Id)
	assert.Error(t, err)
	emails := testutil.SentEmails(t, internet)
	require.Len(t, emails, 1)
	assert.Equal(t, "ada@example.test", emails[0].To[0].Email)
	entries := testutil.Audit(t)
	require.Len(t, entries, 1)
	assert.Equal(t, "user.deleted", entries[0].Action)
	assert.Equal(t, admin.Id, entries[0].UserId)
	assert.Equal(t, visitor.Id.Hex(), entries[0].Target.Id)
	assert.Equal(t, "admin", entries[0].Meta["by"])

	assert.Equal(t, http.StatusNotFound, remove(visitor.Id))

	internet.JSON(testutil.BrevoHost, http.StatusInternalServerError, `{}`)
	other := createUser(t, users.User{})
	assert.Equal(t, http.StatusNoContent, remove(other.Id), "a notice that cannot be sent does not undo the deletion")
}

func TestGetAudit(t *testing.T) {
	testutil.Clean(t)
	admin := createUser(t, users.User{Role: users.RoleAdmin})
	ada, grace := createUser(t, users.User{Email: "ada@example.test", Name: "Ada"}).Id, primitive.NewObjectID()
	_, err := users.DeleteUser(ctx, ada)
	require.NoError(t, err)
	now := time.Now()
	for i, entry := range []audit.Entry{
		{Action: audit.AuthLogin, UserId: &ada},
		{Action: audit.UserProfileUpdated, UserId: &ada},
		{Action: audit.AuthLogin, UserId: &grace},
		{Action: audit.AuthLoginFailed},
		{Action: audit.AdminRoleChanged, UserId: &grace, Target: &audit.Target{Type: audit.TargetUser, Id: ada.Hex()}},
	} {
		entry.CreatedAt = now.Add(time.Duration(i) * time.Minute)
		require.NoError(t, audit.CreateEntry(ctx, &entry))
	}
	get := func(query string) AuditPage {
		response := do(t, http.MethodGet, "/admin/audit"+query, nil, testutil.As(admin.Id))
		require.Equal(t, http.StatusOK, response.Code, response.Body.String())
		return testutil.Decode[AuditPage](t, response)
	}
	actions := func(page AuditPage) []audit.Action {
		found := []audit.Action{}
		for _, entry := range page.Entries {
			found = append(found, entry.Action)
		}
		return found
	}

	page := get("")
	assert.EqualValues(t, 5, page.Total)
	assert.Equal(t, []audit.Action{audit.AdminRoleChanged, audit.AuthLoginFailed, audit.AuthLogin, audit.UserProfileUpdated, audit.AuthLogin},
		actions(page), "newest first")
	assert.Equal(t, map[string]UserSummary{ada.Hex(): {Name: "Ada", Email: "ada@example.test"}}, page.Users,
		"known users are named, deleted ones included")

	page = get("?limit=2&page=2")
	assert.Equal(t, []audit.Action{audit.AuthLogin, audit.UserProfileUpdated}, actions(page))
	assert.EqualValues(t, 5, page.Total)

	page = get("?userId=" + ada.Hex())
	assert.Equal(t, []audit.Action{audit.AdminRoleChanged, audit.UserProfileUpdated, audit.AuthLogin}, actions(page),
		"what the user did and what was done to them")

	page = get("?action=auth.login")
	assert.EqualValues(t, 2, page.Total)
	page = get("?action=auth.login&userId=" + grace.Hex())
	assert.Equal(t, []audit.Action{audit.AuthLogin}, actions(page))
	assert.Empty(t, get("?action=nothing").Entries)

	for _, query := range []string{"?userId=nope", "?page=-1", "?limit=500"} {
		assert.Equal(t, http.StatusBadRequest, do(t, http.MethodGet, "/admin/audit"+query, nil, testutil.As(admin.Id)).Code, query)
	}
}

func TestAdminDatabaseFailures(t *testing.T) {
	testutil.Clean(t)
	visitor := createUser(t, users.User{})
	requester := primitive.NewObjectID()
	testutil.BreakDB(t)
	c := testutil.Context()

	_, err := ListUsers(c, UsersQuery{})
	assert.Error(t, err)
	_, err = GetUser(c, visitor.Id)
	assert.Error(t, err)
	_, err = ChangeRole(c, requester, visitor.Id, users.RoleAdmin)
	assert.Error(t, err)
	assert.Error(t, DeleteUser(c, requester, visitor.Id))
	_, err = ListAudit(c, AuditQuery{})
	assert.Error(t, err)
}
