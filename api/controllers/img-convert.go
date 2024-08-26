package controllers

import (
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"mime/multipart"
	"strconv"
	"strings"

	"github.com/gen2brain/avif"
	"github.com/gen2brain/webp"
	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"

	"github.com/gin-gonic/gin"
)

type ImgConvertController struct {
}

func (controller *ImgConvertController) PostImgConvert(c *gin.Context) {
	destFormat := strings.ToLower(strings.TrimSpace(c.PostForm("format")))

	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(400, gin.H{"error": "No image provided"})
		return
	}

	if file.Size > 20*1024*1024 {
		c.JSON(400, gin.H{"error": "File size too large"})
		return
	}

	// Convert the file to image
	src, err := file.Open()
	if err != nil {
		c.JSON(400, gin.H{"error": "Unable to open image"})
		return
	}
	defer src.Close()

	img, err := DecodeImage(src, file.Filename[strings.LastIndex(file.Filename, ".")+1:])
	if err != nil {
		c.JSON(400, gin.H{"error": "Unable to decode image"})
		return
	}

	c.Writer.Header().Set("Content-Type", "image/"+destFormat)
	c.Writer.Header().Set("Content-Disposition", "attachment; filename=image."+destFormat)

	EncodeImage(c, img, destFormat)
}

func DecodeImage(src multipart.File, format string) (image.Image, error) {
	switch format {
	case "bmp":
		return bmp.Decode(src)
	case "tiff":
		return tiff.Decode(src)
	case "avif":
		return avif.Decode(src)
	}
	img, _, err := image.Decode(src)
	return img, err
}

func EncodeImage(c *gin.Context, img image.Image, destFormat string) {
	switch destFormat {
	case "jpg", "jpeg":
		destQuality, err := strconv.Atoi(c.DefaultPostForm("quality", "100"))
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid quality"})
			return
		}
		o := jpeg.Options{Quality: destQuality}
		if err := jpeg.Encode(c.Writer, img, &o); err != nil {
			c.JSON(500, gin.H{"error": "Failed to convert image"})
			return
		}
	case "png":
		if err := png.Encode(c.Writer, img); err != nil {
			c.JSON(500, gin.H{"error": "Failed to convert image"})
			return
		}
	case "gif":
		if err := gif.Encode(c.Writer, img, nil); err != nil {
			c.JSON(500, gin.H{"error": "Failed to convert image"})
			return
		}
	case "bmp":
		if err := bmp.Encode(c.Writer, img); err != nil {
			c.JSON(500, gin.H{"error": "Failed to convert image"})
			return
		}
	case "tiff":
		if err := tiff.Encode(c.Writer, img, &tiff.Options{Compression: tiff.Uncompressed}); err != nil {
			c.JSON(500, gin.H{"error": "Failed to convert image"})
			return
		}
	case "webp":
		destQuality, err := strconv.Atoi(c.DefaultPostForm("quality", "100"))
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid quality"})
			return
		}
		o := webp.Options{Lossless: false, Quality: destQuality}
		if err := webp.Encode(c.Writer, img, o); err != nil {
			c.JSON(500, gin.H{"error": "Failed to convert image"})
			return
		}
	case "avif":
		destQuality, err := strconv.Atoi(c.DefaultPostForm("quality", "100"))
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid quality"})
			return
		}
		o := avif.Options{Quality: destQuality}
		if err := avif.Encode(c.Writer, img, o); err != nil {
			println(err.Error())
			c.JSON(500, gin.H{"error": "Failed to convert image"})
			return
		}
	default:
		c.JSON(400, gin.H{"error": "Invalid format"})
		return
	}

	c.Writer.Header().Set("Content-Length", strconv.Itoa(c.Writer.Size()))
	c.Status(200)
}
