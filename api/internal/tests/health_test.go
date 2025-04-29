package controllers_test

import (
	"dev/internal/controllers"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetHealth(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	health := new(controllers.HealthController)

	health.GetHealth(c)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, `{"status":"ok"}`, w.Body.String())
}
