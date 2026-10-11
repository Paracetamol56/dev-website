package models

import (
	"context"
	"net/http/httptest"
	"testing"

	"dev/internal/db"
	"dev/internal/testutil"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	testutil.Main(m)
}

var ctx = context.Background()

func ginContext() *gin.Context {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	return c
}

func insertRaw(t *testing.T, collection string, documents ...any) {
	t.Helper()
	_, err := db.GetDB().Collection(collection).InsertMany(ctx, documents)
	require.NoError(t, err)
}
