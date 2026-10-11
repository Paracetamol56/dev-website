package controllers

import (
	"crypto/rand"
	"dev/internal/middlewares"
	"dev/internal/models"
	"dev/internal/utils"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type WordCloudController struct{}

const (
	codeLength = 5
	// No 0/O or 1/I/L, which are easily confused
	codeAlphabet = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"
	// Per participant rather than per IP: audiences often share one network
	wordCooldown = time.Second
	// Participants are offered more than they need, so their own words can be left out
	recentWords       = 8
	socketAuthTimeout = 5 * time.Second
)

var (
	codeRegex          = regexp.MustCompile("^[a-zA-Z0-9]+$")
	participantLimiter = utils.NewRateLimiter(1, wordCooldown)
)

func GenerateCode() string {
	code := make([]byte, codeLength)
	for i := range code {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(codeAlphabet))))
		if err != nil {
			panic(err) // crypto/rand does not fail on supported platforms
		}
		code[i] = codeAlphabet[n.Int64()]
	}
	return string(code)
}

func newUniqueCode(c *gin.Context) (string, error) {
	for range 10 {
		code := GenerateCode()
		inUse, err := models.IsCodeInUse(c, code)
		if err != nil {
			return "", err
		}
		if !inUse {
			return code, nil
		}
	}
	return "", errors.New("could not generate a unique code")
}

func ValidateCode(code string) error {
	if len(code) != codeLength {
		return fmt.Errorf("code must be %d characters long", codeLength)
	}
	if !codeRegex.MatchString(code) {
		return fmt.Errorf("code must contain only letters and numbers")
	}
	return nil
}

func requestUserId(c *gin.Context) (primitive.ObjectID, bool) {
	scheme, token, found := strings.Cut(c.GetHeader("Authorization"), " ")
	if !found || scheme != "Bearer" {
		return primitive.NilObjectID, false
	}
	return userIdFromAccessToken(token)
}

func userIdFromAccessToken(token string) (primitive.ObjectID, bool) {
	secret := os.Getenv("ACCESS_TOKEN_SECRET")
	if authorized, err := utils.IsAuthorized(token, secret); err != nil || !authorized {
		return primitive.NilObjectID, false
	}
	userId, err := utils.ExtractID(token, secret)
	return userId, err == nil
}

func isOpen(wordCloud *models.WordCloud) bool {
	return wordCloud.ClosedAt == nil
}

func publicSession(wordCloud *models.WordCloud) gin.H {
	session := gin.H{
		"id":          wordCloud.Id,
		"name":        wordCloud.Name,
		"description": wordCloud.Description,
		"code":        wordCloud.Code,
		"open":        isOpen(wordCloud),
		"settings": gin.H{
			"showCloudToAudience": wordCloud.Settings.ShowCloudToAudience,
			"allowUpvotes":        wordCloud.Settings.AllowUpvotes,
			"maxWords":            wordCloud.Settings.WordLimit(),
		},
	}
	texts := visibleTexts(wordCloud)
	if wordCloud.Settings.ShowCloudToAudience {
		session["words"] = texts
	}
	if wordCloud.Settings.AllowUpvotes {
		session["recent"] = recentDistinct(texts, recentWords)
	}
	return session
}

// recentDistinct returns up to count texts, most recent first, without repeats.
func recentDistinct(texts []string, count int) []string {
	recent := []string{}
	for i := len(texts) - 1; i >= 0 && len(recent) < count; i-- {
		if !slices.Contains(recent, texts[i]) {
			recent = append(recent, texts[i])
		}
	}
	return recent
}

func visibleTexts(wordCloud *models.WordCloud) []string {
	texts := []string{}
	for _, word := range wordCloud.Words {
		if !word.Hidden {
			texts = append(texts, word.Text)
		}
	}
	return texts
}

// broadcastSession sends the session to everyone, and the moderated words to owners.
func broadcastSession(wordCloud *models.WordCloud) {
	broadcast(wordCloud.Id,
		gin.H{"type": "session", "session": publicSession(wordCloud), "words": withSentiment(wordCloud, wordCloud.Words), "settings": wordCloud.Settings},
		gin.H{"type": "session", "session": publicSession(wordCloud)})
}

func GetWordCloudByCode(c *gin.Context, code string) {
	if err := ValidateCode(code); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	wordCloud, err := models.GetWordCloudByCode(c, code)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.JSON(http.StatusNotFound, gin.H{"error": "word cloud not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, publicSession(wordCloud))
}

func GetWordCloudByUser(c *gin.Context, userIdString string) {
	status := c.DefaultQuery("status", "open")

	queryUserId, err := primitive.ObjectIDFromHex(userIdString)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	tokenUserId, authorized := requestUserId(c)
	if !authorized || queryUserId != tokenUserId {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	wordClouds, err := models.GetWordCloudByUser(c, tokenUserId, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := []gin.H{}
	for _, wordCloud := range wordClouds {
		response = append(response, gin.H{
			"id":          wordCloud.Id,
			"name":        wordCloud.Name,
			"description": wordCloud.Description,
			"submissions": len(wordCloud.Words),
			"code":        wordCloud.Code,
			"open":        isOpen(wordCloud),
			"createdAt":   wordCloud.CreatedAt,
			"closedAt":    wordCloud.ClosedAt,
		})
	}
	c.JSON(http.StatusOK, response)
}

func (controller *WordCloudController) GetWordCloud(c *gin.Context) {
	if code := c.Query("code"); code != "" {
		GetWordCloudByCode(c, code)
		return
	}
	if user := c.Query("user"); user != "" {
		GetWordCloudByUser(c, user)
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": "missing query parameter"})
}

func sessionFromParam(c *gin.Context) (*models.WordCloud, bool) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return nil, false
	}
	wordCloud, err := models.GetWordCloudById(c, id)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.JSON(http.StatusNotFound, gin.H{"error": "word cloud not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return nil, false
	}
	return wordCloud, true
}

func (controller *WordCloudController) GetWordCloudById(c *gin.Context) {
	wordCloud, ok := sessionFromParam(c)
	if !ok {
		return
	}

	if userId, authorized := requestUserId(c); authorized && userId == wordCloud.UserId {
		if err := models.EnsureWordIds(c, wordCloud); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		withSentiment(wordCloud, wordCloud.Words)
		c.JSON(http.StatusOK, struct {
			*models.WordCloud
			Open bool `json:"open"`
		}{wordCloud, isOpen(wordCloud)})
		return
	}

	c.JSON(http.StatusOK, publicSession(wordCloud))
}

func (controller *WordCloudController) PostWordCloudParticipant(c *gin.Context) {
	wordCloud, ok := sessionFromParam(c)
	if !ok {
		return
	}
	if !isOpen(wordCloud) {
		c.JSON(http.StatusConflict, gin.H{"error": "word cloud closed"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"token": utils.SignParticipant(wordCloud.Id.Hex(), uuid.NewV4())})
}

type PostWordCloudWordBody struct {
	Text   string             `json:"text" binding:"required,min=1,max=100"`
	Token  string             `json:"token" binding:"required"`
	Client *models.ClientInfo `json:"client"`
}

func (controller *WordCloudController) PostWordCloudWord(c *gin.Context) {
	wordCloud, ok := sessionFromParam(c)
	if !ok {
		return
	}

	var body PostWordCloudWordBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	participant, err := utils.VerifyParticipant(wordCloud.Id.Hex(), body.Token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	if !isOpen(wordCloud) {
		c.JSON(http.StatusConflict, gin.H{"error": "word cloud closed"})
		return
	}

	text := utils.CleanText(body.Text)
	if text == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "word is empty"})
		return
	}
	submitted := 0
	for _, word := range wordCloud.Words {
		if word.UUID == participant {
			submitted++
		}
	}
	if limit := wordCloud.Settings.WordLimit(); limit > 0 && submitted >= limit {
		c.JSON(http.StatusForbidden, gin.H{"error": "word limit reached"})
		return
	}
	if allowed, retryAfter := participantLimiter.Allow(wordCloud.Id.Hex() + ":" + participant.String()); !allowed {
		middlewares.AbortTooManyRequests(c, retryAfter)
		return
	}

	word := models.Word{
		Id:        primitive.NewObjectID(),
		Text:      text,
		Hidden:    wordCloud.Settings.FilterProfanity && utils.IsProfane(text),
		UUID:      participant,
		UserAgent: c.Request.UserAgent(),
		CreatedAt: primitive.NewDateTimeFromTime(time.Now()),
		Client:    body.Client,
		Network:   utils.Fingerprint("network", c.ClientIP()),
	}
	if body.Client != nil {
		details, _ := json.Marshal(body.Client)
		word.Device = utils.Fingerprint("device", string(details)+c.Request.UserAgent())
	}
	if err := models.AddWordToWordCloud(c, wordCloud.Id, &word); err != nil {
		if errors.Is(err, models.ErrWordNotAdded) {
			c.JSON(http.StatusConflict, gin.H{"error": "word already submitted"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	var publicEvent any
	if (wordCloud.Settings.ShowCloudToAudience || wordCloud.Settings.AllowUpvotes) && !word.Hidden {
		publicEvent = gin.H{"type": "word", "word": gin.H{"text": word.Text}}
	}
	go broadcast(wordCloud.Id, gin.H{"type": "word", "word": withSentiment(wordCloud, []models.Word{word})[0]}, publicEvent)

	c.JSON(http.StatusCreated, gin.H{"text": text})
}

type PostWordCloudBody struct {
	Name        string           `json:"name" binding:"required,min=3,max=100"`
	Description string           `json:"description" binding:"omitempty,min=10,max=1000"`
	Settings    *models.Settings `json:"settings"`
}

func (controller *WordCloudController) PostWordCloud(c *gin.Context) {
	var postWordCloud PostWordCloudBody
	if err := c.ShouldBindJSON(&postWordCloud); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	settings := models.Settings{FilterProfanity: true, AllowUpvotes: true, Sentiment: true}
	if postWordCloud.Settings != nil {
		settings = *postWordCloud.Settings
	}
	createSession(c, c.MustGet("x-user-id").(primitive.ObjectID), postWordCloud.Name, postWordCloud.Description, settings)
}

func createSession(c *gin.Context, userId primitive.ObjectID, name string, description string, settings models.Settings) {
	code, err := newUniqueCode(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	now := primitive.NewDateTimeFromTime(time.Now())
	result, err := models.CreateWordCloud(c, &models.WordCloud{
		UserId:      userId,
		Name:        name,
		Description: description,
		Code:        code,
		Words:       []models.Word{},
		Settings:    settings,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": result.InsertedID})
}

func ownedSession(c *gin.Context) (*models.WordCloud, bool) {
	wordCloud, ok := sessionFromParam(c)
	if !ok {
		return nil, false
	}
	if wordCloud.UserId != c.MustGet("x-user-id").(primitive.ObjectID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return nil, false
	}
	return wordCloud, true
}

func (controller *WordCloudController) PostWordCloudDuplicate(c *gin.Context) {
	wordCloud, ok := ownedSession(c)
	if !ok {
		return
	}
	createSession(c, wordCloud.UserId, wordCloud.Name, wordCloud.Description, wordCloud.Settings)
}

type PatchWordCloudBody struct {
	Name        *string          `json:"name" binding:"omitempty,min=3,max=100"`
	Description *string          `json:"description" binding:"omitempty,max=1000"`
	Open        *bool            `json:"open"`
	Settings    *models.Settings `json:"settings"`
}

func (controller *WordCloudController) PatchWordCloud(c *gin.Context) {
	wordCloud, ok := ownedSession(c)
	if !ok {
		return
	}
	var body PatchWordCloudBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if body.Description != nil && *body.Description != "" && len(*body.Description) < 10 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "description must be empty or at least 10 characters long"})
		return
	}

	set, unset := bson.M{}, bson.M{}
	if body.Name != nil {
		wordCloud.Name = *body.Name
		set["name"] = wordCloud.Name
	}
	if body.Description != nil {
		wordCloud.Description = *body.Description
		set["description"] = wordCloud.Description
	}
	if body.Settings != nil {
		wordCloud.Settings = *body.Settings
		set["settings"] = wordCloud.Settings
	}
	if body.Open != nil && *body.Open != isOpen(wordCloud) {
		if *body.Open {
			inUse, err := models.IsCodeInUse(c, wordCloud.Code)
			if err == nil && inUse {
				wordCloud.Code, err = newUniqueCode(c)
			}
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			set["code"] = wordCloud.Code
			unset["closedAt"] = ""
			wordCloud.ClosedAt = nil
		} else {
			now := primitive.NewDateTimeFromTime(time.Now())
			set["closedAt"] = now
			wordCloud.ClosedAt = &now
		}
	}

	if err := models.UpdateWordCloud(c, wordCloud.Id, set, unset); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	go broadcastSession(wordCloud)
	c.JSON(http.StatusOK, publicSession(wordCloud))
}

type socketClient struct {
	conn  *websocket.Conn
	owner bool
	// gorilla/websocket allows one concurrent writer per connection
	write sync.Mutex
}

func (client *socketClient) send(event any) error {
	client.write.Lock()
	defer client.write.Unlock()
	return client.conn.WriteJSON(event)
}

var (
	socketClients = make(map[primitive.ObjectID][]*socketClient)
	socketMutex   = sync.Mutex{}
	upgrader      = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

// broadcast sends ownerEvent to owners and publicEvent to other clients; nil events are skipped.
func broadcast(sessionId primitive.ObjectID, ownerEvent any, publicEvent any) {
	socketMutex.Lock()
	clients := append([]*socketClient(nil), socketClients[sessionId]...)
	socketMutex.Unlock()

	for _, client := range clients {
		event := publicEvent
		if client.owner {
			event = ownerEvent
		}
		if event == nil {
			continue
		}
		if err := client.send(event); err != nil {
			log.Println("Error writing to word cloud socket:", err)
			client.conn.Close()
		}
	}
}

type socketAuthMessage struct {
	Type  string `json:"type"`
	Token string `json:"token"`
}

// WSWordCloud expects {"type": "auth", "token": accessToken} first; only the owner's token
// unlocks owner events.
func (controller *WordCloudController) WSWordCloud(c *gin.Context) {
	wordCloud, ok := sessionFromParam(c)
	if !ok {
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	var auth socketAuthMessage
	conn.SetReadDeadline(time.Now().Add(socketAuthTimeout))
	if err := conn.ReadJSON(&auth); err != nil || auth.Type != "auth" {
		return
	}
	conn.SetReadDeadline(time.Time{})

	client := &socketClient{conn: conn}
	if userId, authorized := userIdFromAccessToken(auth.Token); authorized && userId == wordCloud.UserId {
		client.owner = true
	}

	socketMutex.Lock()
	socketClients[wordCloud.Id] = append(socketClients[wordCloud.Id], client)
	socketMutex.Unlock()

	defer func() {
		socketMutex.Lock()
		clients := socketClients[wordCloud.Id]
		for i, registered := range clients {
			if registered == client {
				socketClients[wordCloud.Id] = append(clients[:i], clients[i+1:]...)
				break
			}
		}
		if len(socketClients[wordCloud.Id]) == 0 {
			delete(socketClients, wordCloud.Id)
		}
		socketMutex.Unlock()
	}()

	client.send(gin.H{"type": "ready", "owner": client.owner})

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}

func (controller *WordCloudController) DeleteWordCloud(c *gin.Context) {
	wordCloud, ok := ownedSession(c)
	if !ok {
		return
	}
	if err := models.DeleteWordCloud(c, wordCloud.Id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	event := gin.H{"type": "session", "session": nil}
	go broadcast(wordCloud.Id, event, event)
	c.Status(http.StatusNoContent)
}

type PatchWordBody struct {
	Hidden bool `json:"hidden"`
}

func (controller *WordCloudController) PatchWord(c *gin.Context) {
	wordCloud, ok := ownedSession(c)
	if !ok {
		return
	}
	wordId, err := primitive.ObjectIDFromHex(c.Param("wordId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid word id"})
		return
	}
	var body PatchWordBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	found, err := models.SetWordHidden(c, wordCloud.Id, wordId, body.Hidden)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "word not found"})
		return
	}
	respondWithWords(c, wordCloud.Id)
}

type PostMergeBody struct {
	From []string `json:"from" binding:"required,min=1,max=500,dive,min=1,max=100"`
	Into string   `json:"into" binding:"required,min=1,max=100"`
}

func (controller *WordCloudController) PostMergeWords(c *gin.Context) {
	wordCloud, ok := ownedSession(c)
	if !ok {
		return
	}
	var body PostMergeBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	into := utils.CleanText(body.Into)
	if into == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "word is empty"})
		return
	}
	if err := models.MergeWords(c, wordCloud.Id, body.From, into); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	respondWithWords(c, wordCloud.Id)
}

// respondWithWords returns the session's words after a moderation change and tells every viewer.
func respondWithWords(c *gin.Context, sessionId primitive.ObjectID) {
	wordCloud, err := models.GetWordCloudById(c, sessionId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	go broadcastSession(wordCloud)
	c.JSON(http.StatusOK, gin.H{"words": withSentiment(wordCloud, wordCloud.Words)})
}

func withSentiment(wordCloud *models.WordCloud, words []models.Word) []models.Word {
	for i := range words {
		if !wordCloud.Settings.Sentiment {
			words[i].Sentiment = nil
		} else if score, ok := utils.Sentiment(words[i].Text); ok {
			words[i].Sentiment = &score
		} else {
			words[i].Sentiment = nil
		}
	}
	return words
}
