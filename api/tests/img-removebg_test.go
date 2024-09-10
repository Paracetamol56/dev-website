package controllers_test

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/Paracetamol56/dev-website/api/controllers"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func TestPostImgRemoveBg(t *testing.T) {
	godotenv.Load()
	if os.Getenv("REMBG_ENDPOINT") == "" {
		// Skip the test if REMBG_ENDPOINT is not set
		t.Skip("REMBG_ENDPOINT is not set")
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	imgRemoveBg := new(controllers.ImgRemoveBgController)

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	file, err := os.Open("test_assets/test.png")
	if err != nil {
		t.Error(err)
	}
	defer file.Close()

	part, err := writer.CreateFormFile("file", filepath.Base(file.Name()))
	if err != nil {
		t.Error(err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		t.Error(err)
	}

	err = writer.Close()
	if err != nil {
		t.Error(err)
	}

	c.Request = httptest.NewRequest("POST", "/img-removebg", body)
	c.Request.Header.Set("Content-Type", writer.FormDataContentType())

	imgRemoveBg.PostImgRemoveBg(c)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "image/png", w.Header().Get("Content-Type"))
	assert.NotEqual(t, 0, w.Body.Len())
}
