package slides

import (
	"net/http/httptest"
	"testing"

	"dev/internal/testutil"

	"github.com/gin-gonic/gin"
)

func TestMain(m *testing.M) {
	testutil.Main(m)
}

func newEngine() *gin.Engine {
	controller := new(Controller)
	engine := gin.New()
	group := engine.Group("/slides", testutil.Authenticated)
	group.GET("", controller.GetDecks)
	group.POST("", controller.PostDeck)
	group.GET("/:id", controller.GetDeck)
	engine.GET("/slides/:id/ws", controller.WSDeck)
	group.POST("/:id/duplicate", controller.PostDeckDuplicate)
	group.DELETE("/:id", controller.DeleteDeck)
	return engine
}

func do(t *testing.T, method string, path string, body any, options ...testutil.Option) *httptest.ResponseRecorder {
	t.Helper()
	return testutil.Do(t, newEngine(), method, path, body, options...)
}
