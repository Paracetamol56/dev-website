package controllers

import (
	"net/http"
	"testing"

	"dev/internal/db"
	"dev/internal/models"
	"dev/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestGetUser(t *testing.T) {
	testutil.Clean(t)
	ada := createUser(t, models.User{Email: "ada@example.test", Name: "Ada", ProfilePicture: "https://example.test/ada.png",
		Identities: []models.Identity{{Provider: "github", ProviderId: "42", AccessToken: "gho_token"}}})
	grace := createUser(t, models.User{Email: "grace@example.test", Name: "Grace"})

	own := do(t, http.MethodGet, "/users/"+ada.Id.Hex(), nil, as(ada.Id))
	require.Equal(t, http.StatusOK, own.Code)
	full := decode[map[string]any](t, own)
	assert.Equal(t, "ada@example.test", full["email"])
	assert.Len(t, full["identities"], 1)
	assert.NotContains(t, own.Body.String(), "gho_token")

	other := do(t, http.MethodGet, "/users/"+ada.Id.Hex(), nil, as(grace.Id))
	require.Equal(t, http.StatusOK, other.Code)
	assert.JSONEq(t, `{"id": "`+ada.Id.Hex()+`", "name": "Ada", "profilePicture": "https://example.test/ada.png"}`, other.Body.String(),
		"other users only see the public profile")

	assert.Equal(t, http.StatusBadRequest, do(t, http.MethodGet, "/users/me", nil, as(ada.Id)).Code)
	assert.Equal(t, http.StatusUnauthorized, do(t, http.MethodGet, "/users/"+ada.Id.Hex(), nil).Code)
	// Unknown users are reported as a server error rather than a 404
	assert.Equal(t, http.StatusInternalServerError, do(t, http.MethodGet, "/users/"+primitive.NewObjectID().Hex(), nil, as(ada.Id)).Code)
	ghost := primitive.NewObjectID()
	assert.Equal(t, http.StatusInternalServerError, do(t, http.MethodGet, "/users/"+ghost.Hex(), nil, as(ghost)).Code)
}

func TestPatchUser(t *testing.T) {
	testutil.Clean(t)
	user := createUser(t, models.User{Email: "ada@example.test", Name: "Ada", Flavour: "mocha"})
	path := "/users/" + user.Id.Hex()

	response := do(t, http.MethodPatch, path, map[string]string{"flavour": "latte"}, as(user.Id))
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	assert.Equal(t, "latte", decode[models.User](t, response).Flavour)
	stored := storedUser(t, user.Id)
	assert.Equal(t, "latte", stored.Flavour)
	assert.Equal(t, "Ada", stored.Name, "fields left out are kept")

	response = do(t, http.MethodPatch, path, map[string]string{"name": "Countess", "profilePicture": "https://example.test/c.png"}, as(user.Id))
	require.Equal(t, http.StatusOK, response.Code)
	stored = storedUser(t, user.Id)
	assert.Equal(t, "Countess", stored.Name)
	assert.Equal(t, "https://example.test/c.png", stored.ProfilePicture)
	assert.Equal(t, "ada@example.test", stored.Email)

	assert.Equal(t, http.StatusOK, do(t, http.MethodPatch, path, map[string]string{"email": "evil@example.test"}, as(user.Id)).Code)
	assert.Equal(t, "ada@example.test", storedUser(t, user.Id).Email, "the email cannot be changed this way")

	for name, body := range map[string]any{
		"not json":       "not json",
		"short name":     map[string]string{"name": "A"},
		"long name":      map[string]string{"name": repeat("a", 101)},
		"unknown theme":  map[string]string{"flavour": "espresso"},
		"picture no url": map[string]string{"profilePicture": "me.png"},
	} {
		assert.Equal(t, http.StatusBadRequest, do(t, http.MethodPatch, path, body, as(user.Id)).Code, name)
	}
	assert.Equal(t, http.StatusBadRequest, do(t, http.MethodPatch, "/users/me", map[string]string{}, as(user.Id)).Code)
	assert.Equal(t, http.StatusForbidden, do(t, http.MethodPatch, path, map[string]string{"name": "Mallory"}, as(primitive.NewObjectID())).Code)
	assert.Equal(t, "Countess", storedUser(t, user.Id).Name)

	testutil.BreakDB(t)
	assert.Equal(t, http.StatusInternalServerError, do(t, http.MethodPatch, path, map[string]string{"name": "Countess"}, as(user.Id)).Code)
}

func TestDeleteIdentity(t *testing.T) {
	testutil.Clean(t)
	user := createUser(t, models.User{Identities: []models.Identity{
		{Provider: "email", ProviderId: "ada@example.test"}, {Provider: "github", ProviderId: "42"}, {Provider: "google", ProviderId: "g-1"},
	}})
	path := "/users/" + user.Id.Hex() + "/identities/"

	response := do(t, http.MethodDelete, path+"github", nil, as(user.Id))
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	assert.Len(t, decode[models.User](t, response).Identities, 2)
	assert.Nil(t, storedUser(t, user.Id).GetIdentity("github"))
	assert.NotNil(t, storedUser(t, user.Id).GetIdentity("google"))

	again := do(t, http.MethodDelete, path+"github", nil, as(user.Id))
	assert.Equal(t, http.StatusNotFound, again.Code)
	assert.Equal(t, "Identity not found", errorOf(t, again))

	email := do(t, http.MethodDelete, path+"email", nil, as(user.Id))
	assert.Equal(t, http.StatusBadRequest, email.Code)
	assert.Equal(t, "The email identity cannot be removed", errorOf(t, email))
	assert.NotNil(t, storedUser(t, user.Id).GetIdentity("email"))

	assert.Equal(t, http.StatusBadRequest, do(t, http.MethodDelete, "/users/me/identities/google", nil, as(user.Id)).Code)
	assert.Equal(t, http.StatusForbidden, do(t, http.MethodDelete, path+"google", nil, as(primitive.NewObjectID())).Code)
	ghost := primitive.NewObjectID()
	assert.Equal(t, http.StatusNotFound, do(t, http.MethodDelete, "/users/"+ghost.Hex()+"/identities/google", nil, as(ghost)).Code)
	assert.NotNil(t, storedUser(t, user.Id).GetIdentity("google"))
}

func TestDeleteUser(t *testing.T) {
	testutil.Clean(t)
	internet := testutil.FakeInternet(t)
	fakeBrevo(internet)
	user := createUser(t, models.User{Email: "ada@example.test", Name: "Ada"})
	path := "/users/" + user.Id.Hex()

	assert.Equal(t, http.StatusBadRequest, do(t, http.MethodDelete, "/users/me", nil, as(user.Id)).Code)
	assert.Equal(t, http.StatusForbidden, do(t, http.MethodDelete, path, nil, as(primitive.NewObjectID())).Code)
	assert.Empty(t, sentEmails(t, internet))

	response := do(t, http.MethodDelete, path, nil, as(user.Id))
	assert.Equal(t, http.StatusNoContent, response.Code)

	emails := sentEmails(t, internet)
	require.Len(t, emails, 1)
	assert.Equal(t, "ada@example.test", emails[0].To[0].Email)
	assert.Equal(t, "Your account has been deleted", emails[0].Subject)
	assert.Contains(t, emails[0].HTMLContent, "permanently deleted in 30 days")

	_, err := models.GetFullUserById(ginContext(), user.Id)
	assert.Error(t, err, "the account is no longer reachable")
	var kept bson.M
	require.NoError(t, db.GetDB().Collection("users").FindOne(ctx, bson.M{"_id": user.Id}).Decode(&kept))
	assert.Contains(t, kept, "deletedAt", "but kept until the purge")

	// An account that is already gone is reported as a server error rather than a 404
	assert.Equal(t, http.StatusInternalServerError, do(t, http.MethodDelete, path, nil, as(user.Id)).Code)
	assert.Len(t, sentEmails(t, internet), 1)
}

func TestDeleteUserWhenTheEmailFails(t *testing.T) {
	testutil.Clean(t)
	internet := testutil.FakeInternet(t)
	internet.JSON(brevoHost, http.StatusInternalServerError, `{}`)
	user := createUser(t, models.User{})

	assert.Equal(t, http.StatusInternalServerError, do(t, http.MethodDelete, "/users/"+user.Id.Hex(), nil, as(user.Id)).Code)
	_, err := models.GetFullUserById(ginContext(), user.Id)
	assert.Error(t, err, "the account is deleted even though the request reports a failure")
}

func TestGetExport(t *testing.T) {
	testutil.Clean(t)
	user := createUser(t, models.User{Email: "ada@example.test", Name: "Ada",
		Identities: []models.Identity{{Provider: "github", ProviderId: "42", AccessToken: "gho_token"}},
		Passkeys:   []models.Passkey{{Id: "key", Name: "Laptop", Credential: []byte(`{"secret":true}`)}}})
	path := "/users/" + user.Id.Hex() + "/export"

	response := do(t, http.MethodGet, path, nil, as(user.Id))
	require.Equal(t, http.StatusOK, response.Code)
	export := decode[models.User](t, response)
	assert.Equal(t, "ada@example.test", export.Email)
	assert.Len(t, export.Identities, 1)
	assert.Len(t, export.Passkeys, 1)
	assert.NotContains(t, response.Body.String(), "gho_token")
	assert.NotContains(t, response.Body.String(), "secret")

	assert.Equal(t, http.StatusBadRequest, do(t, http.MethodGet, "/users/me/export", nil, as(user.Id)).Code)
	assert.Equal(t, http.StatusForbidden, do(t, http.MethodGet, path, nil, as(primitive.NewObjectID())).Code)
	testutil.BreakDB(t)
	assert.Equal(t, http.StatusInternalServerError, do(t, http.MethodGet, path, nil, as(user.Id)).Code)
}
