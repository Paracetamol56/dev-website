package admin

import (
	"context"
	"net/http/httptest"
	"testing"

	"dev/internal/testutil"
	users "dev/internal/user"

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
	group := engine.Group("/admin", testutil.Authenticated, RequireAdmin())
	group.GET("/users", controller.GetUsers)
	group.GET("/users/:id", controller.GetUser)
	group.PATCH("/users/:id/role", controller.PatchUserRole)
	group.DELETE("/users/:id", controller.DeleteUser)
	group.GET("/audit", controller.GetAudit)
	return engine
}

func do(t *testing.T, method string, path string, body any, options ...testutil.Option) *httptest.ResponseRecorder {
	t.Helper()
	return testutil.Do(t, newEngine(), method, path, body, options...)
}

func createUser(t *testing.T, user users.User) *users.User {
	t.Helper()
	if user.Email == "" {
		user.Email = primitive.NewObjectID().Hex() + "@example.test"
	}
	user.Flavour = "mocha"
	result, err := users.CreateUser(ctx, &user)
	require.NoError(t, err)
	return storedUser(t, result.InsertedID.(primitive.ObjectID))
}

func storedUser(t *testing.T, id primitive.ObjectID) *users.User {
	t.Helper()
	user, err := users.GetFullUserById(ctx, id)
	require.NoError(t, err)
	return user
}
