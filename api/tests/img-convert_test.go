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
	"github.com/stretchr/testify/assert"
)

func TestConvertImage(t *testing.T) {
	t.Parallel()
	formats := []string{"jpg", "png", "gif", "bmp", "tiff", "webp", "avif"}

	for _, inputFormat := range formats {
		for _, outputFormat := range formats {
			t.Run(inputFormat+" to "+outputFormat, func(t *testing.T) {
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				imgConvert := new(controllers.ImgConvertController)

				body := new(bytes.Buffer)
				writer := multipart.NewWriter(body)

				file, err := os.Open("test_assets/test." + inputFormat)
				if err != nil {
					t.Error(err)
				}
				defer file.Close()

				part, err := writer.CreateFormFile("image", filepath.Base(file.Name()))
				if err != nil {
					t.Error(err)
				}

				_, err = io.Copy(part, file)
				if err != nil {
					t.Error(err)
				}

				part, err = writer.CreateFormField("format")
				if err != nil {
					t.Error(err)
				}
				part.Write([]byte(outputFormat))

				err = writer.Close()
				if err != nil {
					t.Error(err)
				}

				c.Request = httptest.NewRequest("POST", "/img-convert", body)
				c.Request.Header.Set("Content-Type", writer.FormDataContentType())

				imgConvert.PostImgConvert(c)

				if w.Code != 200 {
					t.Error(w.Body.String())
				}
				assert.Equal(t, 200, w.Code)
				assert.Equal(t, "image/"+outputFormat, w.Header().Get("Content-Type"))
				assert.Equal(t, "attachment; filename=image."+outputFormat, w.Header().Get("Content-Disposition"))
			})
		}
	}
}
