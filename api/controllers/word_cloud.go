package controllers

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Paracetamol56/dev-website/api/models"
	"github.com/Paracetamol56/dev-website/api/utils"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WordCloudController struct{}

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

func cleanText(text string) string {
	res := strings.TrimSpace(text)
	res = strings.ToLower(res)
	re := regexp.MustCompile(`[^\w-]+`)
	res = re.ReplaceAllString(res, "")

	return res
}

func broadcastToAdmins(sessionId primitive.ObjectID, data models.Word) {
	connMutex.Lock()
	defer connMutex.Unlock()

	// Get connections for the specific session
	connections, exists := adminConnections[sessionId]
	if !exists {
		return // No admins connected to this session
	}

	// Broadcast message to all connections
	for _, admin := range connections {
		err := admin.Conn.WriteJSON(data)
		if err != nil {
			// Remove broken connections
			log.Println("Error writing to admin connection:", err)
			admin.Conn.Close()

			// Remove the connection from the list
			adminConnections[sessionId] = removeConnection(connections, admin)
		}
	}
}

func removeConnection(connections []*AdminConnection, connectionToRemove *AdminConnection) []*AdminConnection {
	for i, conn := range connections {
		if conn == connectionToRemove {
			return append(connections[:i], connections[i+1:]...)
		}
	}
	return connections
}

func GetWordCloudByCode(c *gin.Context, code string) {
	if err := ValidateCode(code); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	wordCloud, err := models.GetWordCloudByCode(c, code)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			c.JSON(http.StatusNotFound, gin.H{"error": "word cloud not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":          wordCloud.Id,
		"name":        wordCloud.Name,
		"description": wordCloud.Description,
		"code":        wordCloud.Code,
	})
}

func GetWordCloudByUser(c *gin.Context, userIdString string) {
	status := c.DefaultQuery("status", "open")

	queryUserId, err := primitive.ObjectIDFromHex(userIdString)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	// Extract the auth header and verify the token
	authHeader := c.GetHeader("Authorization")
	t := strings.Split(authHeader, " ")
	if len(t) != 2 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	authToken := t[1]
	// Check if the token is authorized and if the user ID matches the token
	authorized, _ := utils.IsAuthorized(authToken, os.Getenv("ACCESS_TOKEN_SECRET"))
	tokenUserId, _ := utils.ExtractID(authToken, os.Getenv("ACCESS_TOKEN_SECRET"))
	if !authorized || queryUserId != tokenUserId {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	wordClouds, err := models.GetWordCloudByUser(c, tokenUserId, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	var response []gin.H
	for _, wordCloud := range wordClouds {
		response = append(response, gin.H{
			"id":          wordCloud.Id,
			"name":        wordCloud.Name,
			"description": wordCloud.Description,
			"submitions":  len(wordCloud.Words),
			"code":        wordCloud.Code,
		})
	}

	c.JSON(http.StatusOK, response)
}

func (controller *WordCloudController) GetWordCloud(c *gin.Context) {
	code := c.Query("code")
	user := c.Query("user")

	if code != "" {
		GetWordCloudByCode(c, code)
		return
	} else if user != "" {
		GetWordCloudByUser(c, user)
		return
	}

	// Missing required query parameter
	c.JSON(http.StatusBadRequest, gin.H{"error": "missing query parameter"})
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
		if err.Error() == "mongo: no documents in result" {
			c.JSON(http.StatusNotFound, gin.H{"error": "word cloud not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		t := strings.Split(authHeader, " ")
		if len(t) != 2 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}
		authToken := t[1]
		authorized, err := utils.IsAuthorized(authToken, os.Getenv("ACCESS_TOKEN_SECRET"))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		if authorized {
			userId, err := utils.ExtractID(authToken, os.Getenv("ACCESS_TOKEN_SECRET"))
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
				return
			}
			if wordCloud.UserId == userId {
				c.JSON(http.StatusOK, wordCloud)
				return
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"id":          wordCloud.Id,
		"name":        wordCloud.Name,
		"description": wordCloud.Description,
		"code":        wordCloud.Code,
		"uuid":        uuid.NewV4(),
	})
}

type PostWordCloudWordBody struct {
	Text      string    `json:"text" binding:"required,min=1,max=100"`
	UUID      uuid.UUID `json:"uuid" binding:"required"`
	UserAgent string    `json:"userAgent" binding:"required"`
}

func (controllers *WordCloudController) PostWordCloudWord(c *gin.Context) {
	idString := c.Param("id")
	sessionId, err := primitive.ObjectIDFromHex(idString)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var body PostWordCloudWordBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	word := models.Word{
		Text:      cleanText(body.Text),
		UUID:      body.UUID,
		UserAgent: body.UserAgent,
		CreatedAt: primitive.NewDateTimeFromTime(time.Now()),
	}

	if _, err := models.AddWordToWordCloud(c, sessionId, &word); err != nil {
		if err.Error() == "word already submitted" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	go broadcastToAdmins(sessionId, word)

	c.JSON(http.StatusCreated, gin.H{})
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

type AdminConnection struct {
	Conn        *websocket.Conn
	ConnectedAt time.Time
	SessionID   primitive.ObjectID
}

var (
	adminConnections = make(map[primitive.ObjectID][]*AdminConnection)
	connMutex        = sync.Mutex{} // Mutex to protect concurrent access to the map
	upgrader         = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func (controller *WordCloudController) WSWordCloud(c *gin.Context) {
	idString := c.Param("id")
	sessionId, err := primitive.ObjectIDFromHex(idString)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	connection := &AdminConnection{
		Conn:        conn,
		ConnectedAt: time.Now(),
		SessionID:   sessionId,
	}

	connMutex.Lock()
	adminConnections[sessionId] = append(adminConnections[sessionId], connection)
	connMutex.Unlock()

	defer func() {
		// Clean up on disconnect
		connMutex.Lock()
		connections := adminConnections[sessionId]
		for i, conn := range connections {
			if conn == connection {
				adminConnections[sessionId] = append(connections[:i], connections[i+1:]...)
				conn.Conn.Close()
				break
			}
		}
		connMutex.Unlock()
	}()

	// Keep the connection alive
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break // Disconnect on error
		}
	}
}

func (controller *WordCloudController) DeleteWordCloud(c *gin.Context) {
	idString := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idString)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	userId := c.MustGet("x-user-id").(primitive.ObjectID)
	wordCloud, err := models.GetWordCloudById(c, id)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			c.JSON(http.StatusNotFound, gin.H{"error": "word cloud not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if wordCloud.UserId != userId {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if _, err := models.CloseWordCloud(c, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}
