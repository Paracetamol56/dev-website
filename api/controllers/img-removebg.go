package controllers

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type ImgRemoveBgController struct {
}

func (imgRemoveBg *ImgRemoveBgController) PostImgRemoveBg(c *gin.Context) {
	// Post a request to the REMBG_ENDPOINT with the file data
	resp, err := http.Post(os.Getenv("REMBG_ENDPOINT"), c.Request.Header.Get("Content-Type"), c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	// Return the response from the REMBG_ENDPOINT
	c.DataFromReader(resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
}
