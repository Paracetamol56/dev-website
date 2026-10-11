package health

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetHealth(t *testing.T) {
	response := do(t, http.MethodGet, "/health", nil)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, `{"status":"ok"}`, response.Body.String())
}
