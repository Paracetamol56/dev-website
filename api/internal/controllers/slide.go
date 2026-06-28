package controllers

import (
	"dev/internal/models"
	"dev/internal/repositories"
	"dev/internal/utils"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SlideController struct {
	slideRepo *repositories.SlideRepository
}

func NewSlideController() *SlideController {
	return &SlideController{
		slideRepo: repositories.NewSlideDeckRepository(),
	}
}

// GetSlidesByUser godoc
func (controller *SlideController) GetSlidesByUser(c *gin.Context) {
	userId := c.MustGet("x-user-id").(primitive.ObjectID)
	var pagination utils.Pagination
	if err := c.ShouldBindQuery(&pagination); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	codeCarbonProjects, err := controller.slideRepo.GetByUser(c, userId, false, &pagination)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get codecarbon projects"})
		return
	}

	c.JSON(http.StatusOK, codeCarbonProjects.ToPaginatedArrayResult(&pagination))
}

// GetSlideById godoc
//
// @Summary		Get slide by id
// @Description	Get slide by id
// @Tags			slides
// @Produce		json
// @Param			id	path		string	true	"Slide ID"
// @Success	200	{object}	models.Slide
// @Failure	400
// @Failure	404
// @Security		Bearer
// @Router	/slides/{id} [get]
func (controller *SlideController) GetSlideById(c *gin.Context) {
	userId := c.MustGet("x-user-id").(primitive.ObjectID)

	slideId, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid id"})
		return
	}

	slide, err := controller.slideRepo.GetByIdForUser(c, slideId, userId)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Slide not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get slide"})
		return
	}

	c.JSON(http.StatusOK, slide)
}

type PostSlideBody struct {
	Title   string `json:"title" binding:"required,max=100"`
	Content string `json:"content" binding:"required"`
}

// PostSlide godoc
//
// @Summary		Create slide
// @Description	Create a new slide
// @Tags			slides
// @Accept			json
// @Produce		json
// @Param			slide body		models.CreateSlideRequest	true	"Slide data"
// @Success	201	{object}	models.Slide
// @Failure	400
// @Security		Bearer
// @Router	/slides [post]
func (controller *SlideController) PostSlide(c *gin.Context) {
	var body PostSlideBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userId := c.MustGet("x-user-id").(primitive.ObjectID)

	slide := models.SlideDeck{
		Title:      body.Title,
		Content:    body.Content,
		UserId:     userId,
		Status:     models.StatusDraft,
		Assets:     make([]models.Asset, 0),
		Operations: make([]models.SlideUpdate, 0),
	}

	_, err := controller.slideRepo.Create(c, &slide)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create slide"})
		return
	}

	c.JSON(http.StatusCreated, slide)
}

type SlideClient struct {
	Conn        *websocket.Conn
	ConnectedAt time.Time
}

var (
	slideClients   = make(map[primitive.ObjectID][]*SlideClient)
	slideConnMutex = sync.Mutex{} // Mutex to protect concurrent access to the map
	slideUpgrader  = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func (controller *SlideController) GetWebSocket(c *gin.Context) {
	slideId, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid slide ID"})
		return
	}

	conn, err := slideUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	client := controller.registerSlideClient(slideId, conn)
	defer controller.unregisterSlideClient(slideId, client)

	slideRepo := controller.slideRepo
	controller.handleSlideMessages(c, conn, slideId, slideRepo)
}

func (controller *SlideController) registerSlideClient(slideId primitive.ObjectID, conn *websocket.Conn) *SlideClient {
	client := &SlideClient{
		Conn:        conn,
		ConnectedAt: time.Now(),
	}
	slideConnMutex.Lock()
	slideClients[slideId] = append(slideClients[slideId], client)
	slideConnMutex.Unlock()
	return client
}

func (controller *SlideController) unregisterSlideClient(slideId primitive.ObjectID, client *SlideClient) {
	slideConnMutex.Lock()
	clients := slideClients[slideId]
	for i, c := range clients {
		if c.Conn == client.Conn {
			slideClients[slideId] = append(clients[:i], clients[i+1:]...)
			break
		}
	}
	slideConnMutex.Unlock()
	client.Conn.Close()
}

func (controller *SlideController) handleSlideMessages(c *gin.Context, conn *websocket.Conn, slideId primitive.ObjectID, slideRepo *repositories.SlideRepository) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}

		if err := controller.processIncomingMessage(c, message, slideId, conn); err != nil {
			log.Printf("Failed to process incoming message: %v", err)
		}
	}
}

func (controller *SlideController) processIncomingMessage(c *gin.Context, message []byte, slideId primitive.ObjectID, senderConn *websocket.Conn) error {
	var incoming struct {
		Type    string               `json:"type"`
		Version int64                `json:"version"`
		Updates []models.SlideUpdate `json:"updates"`
	}

	if err := json.Unmarshal(message, &incoming); err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	slide, err := controller.slideRepo.GetById(c, slideId)
	if err != nil {
		return fmt.Errorf("fetch error: %w", err)
	}

	switch incoming.Type {
	case "getDocument":
		return controller.handleGetDocument(senderConn, slide)
	case "pullUpdates":
		return controller.handlePullUpdates(senderConn, slide, incoming.Version)
	case "pushUpdates":
		return controller.handlePushUpdates(c, senderConn, slideId, slide, incoming.Version, incoming.Updates)
	default:
		return fmt.Errorf("unknown message type: %s", incoming.Type)
	}
}

func (controller *SlideController) handleGetDocument(conn *websocket.Conn, slide *models.SlideDeck) error {
	response := map[string]interface{}{
		"type":    "getDocument",
		"version": slide.Revision,
		"doc":     slide.Content,
	}
	msg, _ := json.Marshal(response)
	return conn.WriteMessage(websocket.TextMessage, msg)
}

func (controller *SlideController) handlePullUpdates(conn *websocket.Conn, slide *models.SlideDeck, version int64) error {
	if version < slide.Revision {
		response := map[string]interface{}{
			"type":    "pullUpdates",
			"version": slide.Revision,
			"doc":     slide.Content,
		}
		msg, _ := json.Marshal(response)
		return conn.WriteMessage(websocket.TextMessage, msg)
	}
	return nil
}

func (controller *SlideController) handlePushUpdates(c *gin.Context, senderConn *websocket.Conn, slideId primitive.ObjectID, slide *models.SlideDeck, version int64, updates []models.SlideUpdate) error {
	if version != slide.Revision {
		return fmt.Errorf("version mismatch: client %d vs server %d", version, slide.Revision)
	}

	updatedContent, err := controller.applyUpdates(slide.Content, updates)
	if err != nil {
		return fmt.Errorf("apply updates error: %w", err)
	}

	slide.Content = updatedContent
	slide.Revision++
	slide.UpdatedAt = time.Now()

	/*if result, err := controller.slideRepo.UpdateContent(c, slide.Id, slide.Content, slide.Revision); err != nil || result.MatchedCount != 1 {
		return fmt.Errorf("update error: %w", err)
	}*/

	var outgoing struct {
		Type    string               `json:"type"`
		Version int64                `json:"version"`
		Updates []models.SlideUpdate `json:"updates"`
	}
	outgoing.Type = "pushUpdates"
	outgoing.Version = slide.Revision
	outgoing.Updates = updates
	broadcastMsg, _ := json.Marshal(outgoing)
	fmt.Println("Broadcasting updates to other clients:", string(broadcastMsg))

	slideConnMutex.Lock()
	defer slideConnMutex.Unlock()
	for _, otherClient := range slideClients[slideId] {
		if err := otherClient.Conn.WriteMessage(websocket.TextMessage, broadcastMsg); err != nil {
			log.Printf("Error broadcasting to client: %v", err)
		}
	}

	return nil
}

func (controller *SlideController) applyUpdates(text string, updates []models.SlideUpdate) (string, error) {
	for _, update := range updates {
		var parsed []interface{}
		if err := json.Unmarshal(update.Changes, &parsed); err != nil {
			return "", fmt.Errorf("failed to parse changes: %w", err)
		}

		// Structure:
		// [
		// 	28, // index
		// 	[
		// 		0, // operation type (0 for insert, >= 1 for delete)
		// 		"hello" // text to insert or replace
		// 		"" // possibly on multiple lines
		// 	},
		// 	32 // I have no idea what this is
		// ]

		index := int(parsed[0].(float64))
		if index < 0 || index > len(text) {
			return "", fmt.Errorf("index out of bounds for insert operation")
		}

		operation := parsed[1].([]interface{})
		if len(operation) < 1 {
			return "", fmt.Errorf("invalid operation format")
		}

		removeNumber := int(operation[0].(float64))

		// Delete operation
		if removeNumber > 0 {
			text = text[:index] + text[index+removeNumber:]
		}

		// Insert operation
		if len(operation) > 1 {
			textToInsert := operation[1].(string)
			if len(operation) > 2 {
				// Handle multiple lines if provided
				for i := 2; i < len(operation); i++ {
					textToInsert = textToInsert + "\n" + operation[i].(string)
				}
			}
			text = text[:index] + textToInsert + text[index:]
		}
	}

	return text, nil
}
