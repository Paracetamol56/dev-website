package auth

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"dev/internal/testutil"
	users "dev/internal/user"
	"dev/internal/utils"

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
	engine.POST("/auth/login", controller.PostLogin)
	engine.POST("/auth/verify", controller.PostVerify)
	engine.POST("/auth/refresh", controller.PostRefresh)
	engine.GET("/auth/providers", controller.GetOAuthProviders)
	engine.POST("/auth/:provider", controller.PostOAuthLogin)
	engine.POST("/auth/passkey/begin", controller.PostPasskeyLoginBegin)
	engine.POST("/auth/passkey/finish", controller.PostPasskeyLoginFinish)
	group := engine.Group("/users", testutil.Authenticated)
	group.POST("/:id/passkeys/begin", controller.PostPasskeyRegisterBegin)
	group.POST("/:id/passkeys/finish", controller.PostPasskeyRegisterFinish)
	group.DELETE("/:id/passkeys/:passkeyId", controller.DeletePasskey)
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
	if user.Flavour == "" {
		user.Flavour = "mocha"
	}
	result, err := users.CreateUser(ctx, &user)
	require.NoError(t, err)
	if len(user.Identities) > 0 || len(user.Passkeys) > 0 {
		_, err = users.UpdateUser(ctx, result.InsertedID.(primitive.ObjectID), &user)
		require.NoError(t, err)
	}
	return storedUser(t, result.InsertedID.(primitive.ObjectID))
}

func storedUser(t *testing.T, id primitive.ObjectID) *users.User {
	t.Helper()
	user, err := users.GetFullUserById(ctx, id)
	require.NoError(t, err)
	return user
}

func resetLimiters() {
	loginEmailLimiter = utils.NewRateLimiter(3, 15*time.Minute)
}
