package chip

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
	engine.GET("/chips", controller.GetMicroprocessor)
	engine.GET("/chips/:id", controller.GetMicroprocessorById)
	return engine
}

func do(t *testing.T, method string, path string, body any, options ...testutil.Option) *httptest.ResponseRecorder {
	t.Helper()
	return testutil.Do(t, newEngine(), method, path, body, options...)
}
