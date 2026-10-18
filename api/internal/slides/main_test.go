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
	group.POST("/join", controller.PostDeckJoin)
	group.PUT("/:id/links/:role", controller.PutDeckLink)
	group.DELETE("/:id/links/:role", controller.DeleteDeckLink)
	group.POST("/:id/links/:role/email", controller.PostDeckLinkEmail)
	group.PATCH("/:id/members/:userId", controller.PatchDeckMember)
	group.DELETE("/:id/members/:userId", controller.DeleteDeckMember)
	return engine
}

func do(t *testing.T, method string, path string, body any, options ...testutil.Option) *httptest.ResponseRecorder {
	t.Helper()
	return testutil.Do(t, newEngine(), method, path, body, options...)
}
