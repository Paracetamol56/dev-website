package controllers_test

import (
	"net/http/httptest"

	"github.com/Paracetamol56/dev-website/api/controllers"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"testing"
)

func TestGetHealth(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	health := new(controllers.HealthController)

	health.GetHealth(c)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, `{"status":"ok"}`, w.Body.String())
}
