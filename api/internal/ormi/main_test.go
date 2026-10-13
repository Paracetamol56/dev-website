package ormi

import (
	"context"
	"net/http/httptest"
	"testing"

	"dev/internal/testutil"

	"github.com/gin-gonic/gin"
)

func TestMain(m *testing.M) {
	testutil.Main(m)
}

var ctx = context.Background()

func newEngine() *gin.Engine {
	controller := new(Controller)
	engine := gin.New()
	group := engine.Group("/ormi", testutil.Authenticated)
	group.GET("", controller.GetTodos)
	group.GET("/stats", controller.GetTodoStats)
	group.GET("/stats/:year", controller.GetTodoYearStats)
	group.GET("/labels", controller.GetTodoLabels)
	group.PUT("/order", controller.PutTodoOrder)
	group.GET("/:id", controller.GetTodo)
	group.POST("", controller.PostTodo)
	group.PATCH("/:id", controller.PatchTodo)
	group.DELETE("/:id", controller.DeleteTodo)
	return engine
}

func do(t *testing.T, method string, path string, body any, options ...testutil.Option) *httptest.ResponseRecorder {
	t.Helper()
	return testutil.Do(t, newEngine(), method, path, body, options...)
}
