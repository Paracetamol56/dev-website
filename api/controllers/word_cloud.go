package controllers

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/Paracetamol56/dev-website/api/models"
	"github.com/Paracetamol56/dev-website/api/utils"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WordCloudController struct {
}

func GenerateCode() string {
	// Create a 5 characters long alphanumeric code
	code := ""
	for i := 0; i < 5; i++ {
		randInt := rand.Intn(62)
		if randInt < 10 {
			code += string(randInt + 48) // ASCII range for numbers: 48-57
		} else if randInt < 36 {
			code += string(randInt + 55) // ASCII range for uppercase letters: 65-90
		} else {
			code += string(randInt + 61) // ASCII range for lowercase letters: 97-122
		}
	}
	return code
}

func ValidateCode(code string) error {
	if len(code) != 5 {
		return fmt.Errorf("code must be 5 characters long")
	}
	codeRegex := regexp.MustCompile("^[a-zA-Z0-9]*$")
	if !codeRegex.MatchString(code) {
		return fmt.Errorf("code must contain only letters and numbers")
	}
	return nil
}

func GetWordCloudByCode(c *gin.Context, code string) (*models.WordCloud, error) {
	if err := ValidateCode(code); err != nil {
		return nil, err
	}

	wordCloud, err := models.GetWordCloudByCode(c, code)
	if err != nil {
		return nil, err
	}
	return wordCloud, nil
}

func GetWordCloudByUser(c *gin.Context) ([]*models.WordCloud, error) {
	// Extract the user ID from the query string
	userIdString := c.Query("user")
	queryUserId, err := primitive.ObjectIDFromHex(userIdString)
	if err != nil {
		return nil, fmt.Errorf("invalid user id")
	}

	// Extract the auth header and verify the token
	authHeader := c.GetHeader("Authorization")
	t := strings.Split(authHeader, " ")
	if len(t) != 2 {
		return nil, fmt.Errorf("unauthorized")
	}
	authToken := t[1]
	// Check if the token is authorized and if the user ID matches the token
	authorized, _ := utils.IsAuthorized(authToken, os.Getenv("ACCESS_TOKEN_SECRET"))
	tokenUserId, _ := utils.ExtractID(authToken, os.Getenv("ACCESS_TOKEN_SECRET"))
	if !authorized || queryUserId != tokenUserId {
		return nil, fmt.Errorf("unauthorized")
	}

	wordClouds, err := models.GetWordCloudByUser(c, tokenUserId)
	if err != nil {
		return nil, err
	}
	return wordClouds, nil
}

func (controller *WordCloudController) GetWordCloud(c *gin.Context) {
	code := c.Query("code")

	if code != "" {
		wordCloud, err := GetWordCloudByCode(c, code)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"_id":         wordCloud.Id,
			"name":        wordCloud.Name,
			"description": wordCloud.Description,
			"code":        wordCloud.Code,
			"open":        wordCloud.Open,
		})
	} else {
		wordClouds, err := GetWordCloudByUser(c)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		var response []gin.H
		for _, wordCloud := range wordClouds {
			response = append(response, gin.H{
				"id":          wordCloud.Id,
				"name":        wordCloud.Name,
				"description": wordCloud.Description,
				"submitions":  len(wordCloud.Words),
				"code":        wordCloud.Code,
				"open":        wordCloud.Open,
			})
		}
		c.JSON(http.StatusOK, response)
	}
}

func (controller *WordCloudController) GetWordCloudById(c *gin.Context) {
	idString := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idString)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	wordCloud, err := models.GetWordCloudById(c, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if wordCloud.UserId != c.MustGet("x-user-id").(primitive.ObjectID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	c.JSON(http.StatusOK, wordCloud)
}

type PostWordCloudBody struct {
	Name        string `json:"name" binding:"required,min=3,max=100"`
	Description string `json:"description" binding:"omitempty,min=10,max=1000"`
}

func (controller *WordCloudController) PostWordCloud(c *gin.Context) {
	var postWordCloud PostWordCloudBody
	if err := c.ShouldBindJSON(&postWordCloud); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userId := c.MustGet("x-user-id").(primitive.ObjectID)

	result, err := models.CreateWordCloud(c, &models.WordCloud{
		UserId:      userId,
		Name:        postWordCloud.Name,
		Description: postWordCloud.Description,
		Code:        GenerateCode(),
		Open:        true,
		Words:       []models.Word{},
		CreatedAt:   primitive.NewDateTimeFromTime(time.Now()),
		UpdatedAt:   primitive.NewDateTimeFromTime(time.Now()),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id": result.InsertedID,
	})
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type WSMessage struct {
	Word      string `json:"word"`
	IP        string `json:"ip"`
	UserAgent string `json:"userAgent"`
}

func (controller *WordCloudController) GetWebSocket(c *gin.Context) {
	sessionId := c.Param("id")
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer ws.Close()
	clientId := uuid.NewV4()
	ws.WriteJSON(gin.H{
		"sessionId": sessionId,
		"status":    "ready",
		"connId":    clientId.String(),
	})
	for {
		var message WSMessage
		err := ws.ReadJSON(&message)
		if err != nil {
			break
		}
		_, err := models.AddWordToWordCloud(c, sessionId, message.Word, message.IP, message.UserAgent)
	}
}
