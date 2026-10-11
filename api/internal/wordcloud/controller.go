package wordcloud

import (
	"dev/internal/middlewares"
	"dev/internal/utils"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Controller struct{}

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

func sessionId(c *gin.Context) (primitive.ObjectID, bool) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return primitive.NilObjectID, false
	}
	return id, true
}

func sessionFromParam(c *gin.Context) (*WordCloud, bool) {
	id, ok := sessionId(c)
	if !ok {
		return nil, false
	}
	wordCloud, err := GetSession(c, id)
	if err != nil {
		middlewares.RespondError(c, err)
		return nil, false
	}
	return wordCloud, true
}

func ownedSession(c *gin.Context) (*WordCloud, bool) {
	id, ok := sessionId(c)
	if !ok {
		return nil, false
	}
	wordCloud, err := GetOwnedSession(c, id, c.MustGet("x-user-id").(primitive.ObjectID))
	if err != nil {
		middlewares.RespondError(c, err)
		return nil, false
	}
	return wordCloud, true
}

func (controller *Controller) GetWordCloud(c *gin.Context) {
	if code := c.Query("code"); code != "" {
		wordCloud, err := GetByCode(c, code)
		if err != nil {
			middlewares.RespondError(c, err)
			return
		}
		c.JSON(http.StatusOK, PublicView(wordCloud))
		return
	}

	if user := c.Query("user"); user != "" {
		userId, err := primitive.ObjectIDFromHex(user)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
			return
		}
		requester, authenticated := requestUserId(c)
		wordClouds, err := ListSessions(c, requester, authenticated, userId, c.DefaultQuery("status", "open"))
		if err != nil {
			middlewares.RespondError(c, err)
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
		return
	}

	c.JSON(http.StatusBadRequest, gin.H{"error": "missing query parameter"})
}

func (controller *Controller) GetWordCloudById(c *gin.Context) {
	wordCloud, ok := sessionFromParam(c)
	if !ok {
		return
	}

	if userId, authenticated := requestUserId(c); IsOwner(wordCloud, userId, authenticated) {
		owned, err := OwnerView(c, wordCloud)
		if err != nil {
			middlewares.RespondError(c, err)
			return
		}
		c.JSON(http.StatusOK, struct {
			*WordCloud
			Open bool `json:"open"`
		}{owned, isOpen(owned)})
		return
	}

	c.JSON(http.StatusOK, PublicView(wordCloud))
}

func (controller *Controller) PostWordCloudParticipant(c *gin.Context) {
	wordCloud, ok := sessionFromParam(c)
	if !ok {
		return
	}
	token, err := Join(wordCloud)
	if err != nil {
		middlewares.RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"token": token})
}

type PostWordCloudWordBody struct {
	Text   string      `json:"text" binding:"required,min=1,max=100"`
	Token  string      `json:"token" binding:"required"`
	Client *ClientInfo `json:"client"`
}

func (controller *Controller) PostWordCloudWord(c *gin.Context) {
	wordCloud, ok := sessionFromParam(c)
	if !ok {
		return
	}
	var body PostWordCloudWordBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	text, err := SubmitWord(c, wordCloud, Submission{
		Text:      body.Text,
		Token:     body.Token,
		Client:    body.Client,
		UserAgent: c.Request.UserAgent(),
		IP:        c.ClientIP(),
	})
	if err != nil {
		middlewares.RespondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"text": text})
}

func (controller *Controller) PostWordCloud(c *gin.Context) {
	var body PostWordCloudBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := CreateSession(c, c.MustGet("x-user-id").(primitive.ObjectID), body)
	if err != nil {
		middlewares.RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (controller *Controller) PostWordCloudDuplicate(c *gin.Context) {
	wordCloud, ok := ownedSession(c)
	if !ok {
		return
	}
	id, err := DuplicateSession(c, wordCloud)
	if err != nil {
		middlewares.RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (controller *Controller) PatchWordCloud(c *gin.Context) {
	wordCloud, ok := ownedSession(c)
	if !ok {
		return
	}
	var body PatchWordCloudBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	changed, err := ChangeSession(c, wordCloud, body)
	if err != nil {
		middlewares.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, PublicView(changed))
}

type socketAuthMessage struct {
	Type  string `json:"type"`
	Token string `json:"token"`
}

// WSWordCloud expects {"type": "auth", "token": accessToken} first; only the owner's token
// unlocks owner events.
func (controller *Controller) WSWordCloud(c *gin.Context) {
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

	userId, authenticated := userIdFromAccessToken(auth.Token)
	client := subscribe(wordCloud.Id, conn, IsOwner(wordCloud, userId, authenticated))
	defer unsubscribe(wordCloud.Id, client)

	client.send(gin.H{"type": "ready", "owner": client.owner})

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}

func (controller *Controller) DeleteWordCloud(c *gin.Context) {
	wordCloud, ok := ownedSession(c)
	if !ok {
		return
	}
	if err := RemoveSession(c, wordCloud); err != nil {
		middlewares.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

type PatchWordBody struct {
	Hidden bool `json:"hidden"`
}

func (controller *Controller) PatchWord(c *gin.Context) {
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

	words, err := HideWord(c, wordCloud, wordId, body.Hidden)
	if err != nil {
		middlewares.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"words": words})
}

type PostMergeBody struct {
	From []string `json:"from" binding:"required,min=1,max=500,dive,min=1,max=100"`
	Into string   `json:"into" binding:"required,min=1,max=100"`
}

func (controller *Controller) PostMergeWords(c *gin.Context) {
	wordCloud, ok := ownedSession(c)
	if !ok {
		return
	}
	var body PostMergeBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	words, err := MergeInto(c, wordCloud, body.From, body.Into)
	if err != nil {
		middlewares.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"words": words})
}
