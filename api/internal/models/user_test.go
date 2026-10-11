package models

import (
	"testing"
	"time"

	"dev/internal/db"
	"dev/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func createUser(t *testing.T, user User) *User {
	t.Helper()
	if user.Email == "" {
		user.Email = primitive.NewObjectID().Hex() + "@example.test"
	}
	result, err := CreateUser(ginContext(), &user)
	require.NoError(t, err)
	stored, err := GetFullUserById(ginContext(), result.InsertedID.(primitive.ObjectID))
	require.NoError(t, err)
	return stored
}

func TestIdentities(t *testing.T) {
	user := User{}
	assert.Nil(t, user.GetIdentity("github"))
	assert.False(t, user.RemoveIdentity("github"))

	user.SetIdentity(Identity{Provider: "email", Email: "ada@example.test"})
	user.SetIdentity(Identity{Provider: "github", ProviderId: "42", Username: "ada"})
	require.Len(t, user.Identities, 2)
	github := user.GetIdentity("github")
	require.NotNil(t, github)
	assert.WithinDuration(t, time.Now(), github.LinkedAt, time.Second)
	assert.Equal(t, github.LinkedAt, github.LastLogin)

	linkedAt := time.Now().Add(-48 * time.Hour)
	github.LinkedAt = linkedAt
	user.SetIdentity(Identity{Provider: "github", ProviderId: "42", Username: "countess"})
	require.Len(t, user.Identities, 2, "an identity is replaced, not duplicated")
	github = user.GetIdentity("github")
	assert.Equal(t, "countess", github.Username)
	assert.Equal(t, linkedAt, github.LinkedAt, "the link date survives a new login")
	assert.WithinDuration(t, time.Now(), github.LastLogin, time.Second)

	assert.True(t, user.RemoveIdentity("email"))
	assert.Nil(t, user.GetIdentity("email"))
	assert.Len(t, user.Identities, 1)
}

func TestCreateAndGetUser(t *testing.T) {
	testutil.Clean(t)
	user := createUser(t, User{Name: "Ada", Email: "Ada@Example.test", Flavour: "mocha", ProfilePicture: "https://example.test/ada.png"})

	assert.False(t, user.Id.IsZero())
	assert.WithinDuration(t, time.Now(), user.CreatedAt, 2*time.Second)

	light, err := GetUserById(ginContext(), user.Id)
	require.NoError(t, err)
	assert.Equal(t, &UserLight{Id: user.Id, Name: "Ada", ProfilePicture: "https://example.test/ada.png"}, light)

	byEmail, err := GetFullUserByEmail(ginContext(), "ada@example.TEST")
	require.NoError(t, err)
	require.NotNil(t, byEmail, "emails are matched without case")
	assert.Equal(t, user.Id, byEmail.Id)
	assert.Equal(t, "Ada@Example.test", byEmail.Email)

	missing, err := GetFullUserByEmail(ginContext(), "nobody@example.test")
	assert.NoError(t, err)
	assert.Nil(t, missing)
	_, err = GetFullUserById(ginContext(), primitive.NewObjectID())
	assert.ErrorIs(t, err, mongo.ErrNoDocuments)
	_, err = GetUserById(ginContext(), primitive.NewObjectID())
	assert.ErrorIs(t, err, mongo.ErrNoDocuments)
}

func TestCreateUserKeepsAGivenCreationDate(t *testing.T) {
	testutil.Clean(t)
	createdAt := time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)
	user := createUser(t, User{CreatedAt: createdAt})
	assert.True(t, createdAt.Equal(user.CreatedAt))
}

func TestUserWithoutCreationDateFallsBackToItsId(t *testing.T) {
	testutil.Clean(t)
	id := primitive.NewObjectIDFromTimestamp(time.Date(2019, 5, 6, 0, 0, 0, 0, time.UTC))
	insertRaw(t, "users", bson.M{"_id": id, "email": "old@example.test"})

	byId, err := GetFullUserById(ginContext(), id)
	require.NoError(t, err)
	assert.Equal(t, 2019, byId.CreatedAt.Year())
	byEmail, err := GetFullUserByEmail(ginContext(), "old@example.test")
	require.NoError(t, err)
	assert.Equal(t, 2019, byEmail.CreatedAt.Year())
}

func TestUpdateUser(t *testing.T) {
	testutil.Clean(t)
	user := createUser(t, User{Email: "ada@example.test", Flavour: "mocha"})
	createdAt := user.CreatedAt

	lastLogin := time.Now().Truncate(time.Millisecond)
	user.Name = "Ada"
	user.Flavour = "latte"
	user.LastLogin = lastLogin
	user.SetIdentity(Identity{Provider: "github", ProviderId: "42", AccessToken: "secret"})
	user.Passkeys = []Passkey{{Id: "key", Name: "Laptop", Credential: []byte("{}")}}
	user.CreatedAt = time.Time{}
	result, err := UpdateUser(ginContext(), user.Id, user)
	require.NoError(t, err)
	assert.EqualValues(t, 1, result.MatchedCount)

	updated, err := GetFullUserById(ginContext(), user.Id)
	require.NoError(t, err)
	assert.Equal(t, "Ada", updated.Name)
	assert.Equal(t, "latte", updated.Flavour)
	assert.True(t, lastLogin.Equal(updated.LastLogin))
	assert.True(t, createdAt.Equal(updated.CreatedAt), "the creation date is not rewritten")
	require.Len(t, updated.Identities, 1)
	assert.Equal(t, "secret", updated.Identities[0].AccessToken)
	require.Len(t, updated.Passkeys, 1)
	assert.Equal(t, "Laptop", updated.Passkeys[0].Name)

	result, err = UpdateUser(ginContext(), primitive.NewObjectID(), user)
	require.NoError(t, err)
	assert.Zero(t, result.MatchedCount)
}

func TestDeleteUser(t *testing.T) {
	testutil.Clean(t)
	user := createUser(t, User{Email: "ada@example.test"})

	result, err := DeleteUser(ginContext(), user.Id)
	require.NoError(t, err)
	assert.EqualValues(t, 1, result.MatchedCount)

	_, err = GetFullUserById(ginContext(), user.Id)
	assert.ErrorIs(t, err, mongo.ErrNoDocuments, "a deleted user is hidden")
	_, err = GetUserById(ginContext(), user.Id)
	assert.ErrorIs(t, err, mongo.ErrNoDocuments)
	byEmail, err := GetFullUserByEmail(ginContext(), "ada@example.test")
	assert.NoError(t, err)
	assert.Nil(t, byEmail)

	count, err := db.GetDB().Collection("users").CountDocuments(ctx, bson.M{})
	require.NoError(t, err)
	assert.EqualValues(t, 1, count, "but kept for 30 days")
}

func TestDeleteOldUsers(t *testing.T) {
	testutil.Clean(t)
	active, recent, old := primitive.NewObjectID(), primitive.NewObjectID(), primitive.NewObjectID()
	insertRaw(t, "users",
		bson.M{"_id": active, "email": "active@example.test"},
		bson.M{"_id": recent, "email": "recent@example.test", "deletedAt": time.Now().AddDate(0, 0, -29)},
		bson.M{"_id": old, "email": "old@example.test", "deletedAt": time.Now().AddDate(0, 0, -31)},
	)

	result, err := DeleteOldUsers(ctx)
	require.NoError(t, err)
	assert.EqualValues(t, 1, result.DeletedCount)

	var remaining []struct {
		Id primitive.ObjectID `bson:"_id"`
	}
	cursor, err := db.GetDB().Collection("users").Find(ctx, bson.M{})
	require.NoError(t, err)
	require.NoError(t, cursor.All(ctx, &remaining))
	ids := []primitive.ObjectID{}
	for _, user := range remaining {
		ids = append(ids, user.Id)
	}
	assert.ElementsMatch(t, []primitive.ObjectID{active, recent}, ids)
}

func TestDetachIdentity(t *testing.T) {
	testutil.Clean(t)
	link := func(email string, identities ...Identity) *User {
		user := createUser(t, User{Email: email})
		user.Identities = identities
		_, err := UpdateUser(ginContext(), user.Id, user)
		require.NoError(t, err)
		return user
	}
	account := Identity{Provider: "github", ProviderId: "42"}
	owner := link("owner@example.test", account)
	former := link("former@example.test", account, Identity{Provider: "email", ProviderId: "former@example.test"})
	other := link("other@example.test", Identity{Provider: "github", ProviderId: "43"}, Identity{Provider: "google", ProviderId: "42"})

	require.NoError(t, DetachIdentity(ctx, "github", "42", owner.Id))

	providers := func(id primitive.ObjectID) []string {
		user, err := GetFullUserById(ginContext(), id)
		require.NoError(t, err)
		names := []string{}
		for _, identity := range user.Identities {
			names = append(names, identity.Provider+":"+identity.ProviderId)
		}
		return names
	}
	assert.Equal(t, []string{"github:42"}, providers(owner.Id))
	assert.Equal(t, []string{"email:former@example.test"}, providers(former.Id))
	assert.Equal(t, []string{"github:43", "google:42"}, providers(other.Id))
}

func TestMigrateUserIdentities(t *testing.T) {
	testutil.Clean(t)
	lastLogin := time.Date(2024, 3, 1, 12, 0, 0, 0, time.UTC)
	legacy, linked, cleared, modern := primitive.NewObjectID(), primitive.NewObjectID(), primitive.NewObjectID(), primitive.NewObjectID()
	insertRaw(t, "users",
		bson.M{"_id": legacy, "email": "legacy@example.test", "lastLogin": lastLogin, "githubAccessToken": "gho_old",
			"github": bson.M{"id": 42, "login": "ada", "name": "Ada", "avatar_url": "https://avatars.example/ada", "html_url": "https://github.com/ada"}},
		bson.M{"_id": linked, "email": "linked@example.test", "github": bson.M{"id": 7, "login": "old-login"},
			"identities": bson.A{bson.M{"provider": "github", "providerId": "7", "username": "new-login"}}},
		bson.M{"_id": cleared, "email": "cleared@example.test", "github": nil},
		bson.M{"_id": modern, "email": "modern@example.test", "identities": bson.A{}},
	)

	migrated, err := MigrateUserIdentities(ctx)
	require.NoError(t, err)
	assert.Equal(t, 3, migrated)

	user, err := GetFullUserById(ginContext(), legacy)
	require.NoError(t, err)
	require.Len(t, user.Identities, 1)
	assert.Equal(t, Identity{
		Provider: "github", ProviderId: "42", Email: "legacy@example.test", Username: "ada", Name: "Ada",
		AvatarUrl: "https://avatars.example/ada", ProfileUrl: "https://github.com/ada", AccessToken: "gho_old",
		LinkedAt: lastLogin, LastLogin: lastLogin,
	}, normalized(user.Identities[0]))

	user, err = GetFullUserById(ginContext(), linked)
	require.NoError(t, err)
	require.Len(t, user.Identities, 1, "an already linked account is not added twice")
	assert.Equal(t, "new-login", user.Identities[0].Username)

	user, err = GetFullUserById(ginContext(), cleared)
	require.NoError(t, err)
	assert.Equal(t, []Identity{}, user.Identities)

	legacyFields, err := db.GetDB().Collection("users").CountDocuments(ctx, bson.M{"$or": bson.A{
		bson.M{"github": bson.M{"$exists": true}},
		bson.M{"githubAccessToken": bson.M{"$exists": true}},
	}})
	require.NoError(t, err)
	assert.Zero(t, legacyFields)

	migrated, err = MigrateUserIdentities(ctx)
	require.NoError(t, err)
	assert.Zero(t, migrated, "the migration only runs once")
}

func normalized(identity Identity) Identity {
	identity.LinkedAt = identity.LinkedAt.UTC()
	identity.LastLogin = identity.LastLogin.UTC()
	return identity
}

func TestUserDatabaseFailures(t *testing.T) {
	testutil.BreakDB(t)
	id := primitive.NewObjectID()

	_, err := CreateUser(ginContext(), &User{})
	assert.Error(t, err)
	_, err = GetUserById(ginContext(), id)
	assert.Error(t, err)
	_, err = GetFullUserById(ginContext(), id)
	assert.Error(t, err)
	_, err = GetFullUserByEmail(ginContext(), "ada@example.test")
	assert.Error(t, err)
	_, err = UpdateUser(ginContext(), id, &User{})
	assert.Error(t, err)
	_, err = DeleteUser(ginContext(), id)
	assert.Error(t, err)
	_, err = DeleteOldUsers(ctx)
	assert.Error(t, err)
	assert.Error(t, DetachIdentity(ctx, "github", "42", id))
	_, err = MigrateUserIdentities(ctx)
	assert.Error(t, err)
}
