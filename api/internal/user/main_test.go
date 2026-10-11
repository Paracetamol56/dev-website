package user

import (
	"context"
	"net/http/httptest"
	"testing"

	"dev/internal/testutil"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestMain(m *testing.M) {
	testutil.Main(m)
}

var ctx = context.Background()

func newEngine() *gin.Engine {
	controller := new(Controller)
	engine := gin.New()
	group := engine.Group("/users", testutil.Authenticated)
	group.GET("/:id", controller.GetUser)
	group.PATCH("/:id", controller.PatchUser)
	group.DELETE("/:id", controller.DeleteUser)
	group.GET("/:id/export", controller.GetExport)
	group.DELETE("/:id/identities/:provider", controller.DeleteIdentity)
	return engine
}

func do(t *testing.T, method string, path string, body any, options ...testutil.Option) *httptest.ResponseRecorder {
	t.Helper()
	return testutil.Do(t, newEngine(), method, path, body, options...)
}

func createUser(t *testing.T, user User) *User {
	t.Helper()
	if user.Email == "" {
		user.Email = primitive.NewObjectID().Hex() + "@example.test"
	}
	if user.Flavour == "" {
		user.Flavour = "mocha"
	}
	result, err := CreateUser(ctx, &user)
	require.NoError(t, err)
	if len(user.Identities) > 0 || len(user.Passkeys) > 0 {
		_, err = UpdateUser(ctx, result.InsertedID.(primitive.ObjectID), &user)
		require.NoError(t, err)
	}
	return storedUser(t, result.InsertedID.(primitive.ObjectID))
}

func storedUser(t *testing.T, id primitive.ObjectID) *User {
	t.Helper()
	user, err := GetFullUserById(ctx, id)
	require.NoError(t, err)
	return user
}
