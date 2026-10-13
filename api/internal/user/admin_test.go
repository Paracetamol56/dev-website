package user

import (
	"net/http"
	"testing"
	"time"

	"dev/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestSetUserRole(t *testing.T) {
	testutil.Clean(t)
	user := createUser(t, User{Name: "Ada"})
	assert.False(t, user.IsAdmin())

	require.NoError(t, SetUserRole(ctx, user.Id, RoleAdmin))
	assert.True(t, storedUser(t, user.Id).IsAdmin())

	changed := storedUser(t, user.Id)
	changed.Name = "Countess"
	changed.Role = ""
	_, err := UpdateUser(ctx, user.Id, changed)
	require.NoError(t, err)
	assert.True(t, storedUser(t, user.Id).IsAdmin(), "a profile update never touches the role")

	require.NoError(t, SetUserRole(ctx, user.Id, ""))
	assert.False(t, storedUser(t, user.Id).IsAdmin())

	assert.ErrorIs(t, SetUserRole(ctx, primitive.NewObjectID(), RoleAdmin), mongo.ErrNoDocuments)
	_, err = DeleteUser(ctx, user.Id)
	require.NoError(t, err)
	assert.ErrorIs(t, SetUserRole(ctx, user.Id, RoleAdmin), mongo.ErrNoDocuments, "a deleted account")

	testutil.BreakDB(t)
	assert.Error(t, SetUserRole(testutil.Context(), user.Id, RoleAdmin))
}

func TestListUsers(t *testing.T) {
	testutil.Clean(t)
	now := time.Now()
	ada := createUser(t, User{Id: primitive.NewObjectIDFromTimestamp(now.Add(-3 * time.Hour)), Email: "ada@example.test", Name: "Ada Lovelace"})
	grace := createUser(t, User{Id: primitive.NewObjectIDFromTimestamp(now.Add(-2 * time.Hour)), Email: "grace@example.test", Name: "Grace (Hopper)"})
	alan := createUser(t, User{Id: primitive.NewObjectIDFromTimestamp(now.Add(-time.Hour)), Email: "alan@other.test", Name: "Alan", Role: RoleAdmin})
	deleted := createUser(t, User{Email: "gone@example.test"})
	_, err := DeleteUser(ctx, deleted.Id)
	require.NoError(t, err)
	ids := func(users []User) []primitive.ObjectID {
		found := []primitive.ObjectID{}
		for _, user := range users {
			found = append(found, user.Id)
		}
		return found
	}

	users, total, err := ListUsers(ctx, "", 1, 10)
	require.NoError(t, err)
	assert.EqualValues(t, 3, total)
	assert.Equal(t, []primitive.ObjectID{alan.Id, grace.Id, ada.Id}, ids(users), "newest first, without deleted accounts")
	assert.True(t, users[0].IsAdmin())
	assert.False(t, users[0].CreatedAt.IsZero())

	users, total, err = ListUsers(ctx, "", 2, 2)
	require.NoError(t, err)
	assert.EqualValues(t, 3, total)
	assert.Equal(t, []primitive.ObjectID{ada.Id}, ids(users))

	for search, expected := range map[string][]primitive.ObjectID{
		"EXAMPLE.test": {grace.Id, ada.Id},
		"lovelace":     {ada.Id},
		"(hopper)":     {grace.Id},
		".*":           {},
		"gone":         {},
	} {
		users, total, err = ListUsers(ctx, search, 1, 10)
		require.NoError(t, err)
		assert.Equal(t, expected, ids(users), search)
		assert.EqualValues(t, len(expected), total, search)
	}

	testutil.BreakDB(t)
	_, _, err = ListUsers(testutil.Context(), "", 1, 10)
	assert.Error(t, err)
}

func TestGetUsersByIds(t *testing.T) {
	testutil.Clean(t)
	ada := createUser(t, User{Email: "ada@example.test"})
	deleted := createUser(t, User{Email: "gone@example.test"})
	createUser(t, User{})
	_, err := DeleteUser(ctx, deleted.Id)
	require.NoError(t, err)

	users, err := GetUsersByIds(ctx, []primitive.ObjectID{ada.Id, deleted.Id, ada.Id, primitive.NewObjectID()})
	require.NoError(t, err)
	emails := []string{}
	for _, user := range users {
		emails = append(emails, user.Email)
	}
	assert.ElementsMatch(t, []string{"ada@example.test", "gone@example.test"}, emails)

	users, err = GetUsersByIds(ctx, nil)
	require.NoError(t, err)
	assert.Empty(t, users)

	testutil.BreakDB(t)
	_, err = GetUsersByIds(testutil.Context(), []primitive.ObjectID{ada.Id})
	assert.Error(t, err)
}

func TestRootAdmin(t *testing.T) {
	testutil.Clean(t)
	assert.True(t, IsRootAdmin("ADMIN@example.test"))
	assert.False(t, IsRootAdmin("ada@example.test"))
	assert.False(t, IsRootAdmin(""))

	require.NoError(t, EnsureRootAdmin(ctx), "the owner has no account yet")
	other := createUser(t, User{Email: "ada@example.test"})
	owner := createUser(t, User{Email: "Admin@Example.test"})

	require.NoError(t, EnsureRootAdmin(ctx))
	require.NoError(t, EnsureRootAdmin(ctx))
	assert.True(t, storedUser(t, owner.Id).IsAdmin())
	assert.False(t, storedUser(t, other.Id).IsAdmin())
	entries := testutil.Audit(t)
	require.Len(t, entries, 1, "promoted once")
	assert.Equal(t, "admin.role_changed", entries[0].Action)
	assert.Equal(t, owner.Id.Hex(), entries[0].Target.Id)

	require.NoError(t, PromoteRootAdmin(ctx, nil))
	require.NoError(t, PromoteRootAdmin(ctx, other))
	assert.False(t, other.IsAdmin())

	t.Setenv("ADMIN_EMAIL", "")
	assert.False(t, IsRootAdmin(""))
	require.NoError(t, EnsureRootAdmin(ctx))

	t.Setenv("ADMIN_EMAIL", testutil.AdminEmail)
	require.NoError(t, SetUserRole(ctx, owner.Id, ""))
	testutil.BreakDB(t)
	assert.Error(t, EnsureRootAdmin(testutil.Context()))
	assert.Error(t, PromoteRootAdmin(testutil.Context(), owner))
}

func TestTheRoleIsOnlyShownToItsOwner(t *testing.T) {
	testutil.Clean(t)
	admin := createUser(t, User{Name: "Ada", Role: RoleAdmin})
	other := createUser(t, User{})
	path := "/users/" + admin.Id.Hex()

	own := do(t, http.MethodGet, path, nil, testutil.As(admin.Id))
	assert.Equal(t, RoleAdmin, testutil.Decode[User](t, own).Role)
	assert.NotContains(t, do(t, http.MethodGet, path, nil, testutil.As(other.Id)).Body.String(), "role")
	assert.NotContains(t, do(t, http.MethodGet, "/users/"+other.Id.Hex(), nil, testutil.As(other.Id)).Body.String(), "role")

	patched := do(t, http.MethodPatch, "/users/"+other.Id.Hex(), map[string]string{"name": "Mallory", "role": RoleAdmin}, testutil.As(other.Id))
	require.Equal(t, http.StatusOK, patched.Code)
	assert.False(t, storedUser(t, other.Id).IsAdmin(), "nobody can give themselves a role")
}
