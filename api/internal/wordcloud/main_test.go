package wordcloud

import (
	"context"
	"net/http/httptest"
	"testing"

	"dev/internal/testutil"
	"dev/internal/utils"

	"github.com/gin-gonic/gin"
)

func TestMain(m *testing.M) {
	testutil.Main(m)
}

var ctx = context.Background()

func newEngine() *gin.Engine {
	controller := new(Controller)
	engine := gin.New()
	engine.GET("/word-cloud", controller.GetWordCloud)
	engine.GET("/word-cloud/:id", controller.GetWordCloudById)
	engine.GET("/word-cloud/:id/ws", controller.WSWordCloud)
	engine.POST("/word-cloud/:id/participant", controller.PostWordCloudParticipant)
	engine.POST("/word-cloud/:id/word", controller.PostWordCloudWord)
	engine.POST("/word-cloud", testutil.Authenticated, controller.PostWordCloud)
	engine.PATCH("/word-cloud/:id", testutil.Authenticated, controller.PatchWordCloud)
	engine.POST("/word-cloud/:id/duplicate", testutil.Authenticated, controller.PostWordCloudDuplicate)
	engine.DELETE("/word-cloud/:id", testutil.Authenticated, controller.DeleteWordCloud)
	engine.PATCH("/word-cloud/:id/words/:wordId", testutil.Authenticated, controller.PatchWord)
	engine.POST("/word-cloud/:id/words/merge", testutil.Authenticated, controller.PostMergeWords)
	return engine
}

func do(t *testing.T, method string, path string, body any, options ...testutil.Option) *httptest.ResponseRecorder {
	t.Helper()
	return testutil.Do(t, newEngine(), method, path, body, options...)
}

func resetLimiters() {
	participantLimiter = utils.NewRateLimiter(1, wordCooldown)
}
