package slides

import (
	"dev/internal/middlewares"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Controller struct{}

func deckId(c *gin.Context) (primitive.ObjectID, bool) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return primitive.NilObjectID, false
	}
	return id, true
}

func requester(c *gin.Context) primitive.ObjectID {
	return c.MustGet("x-user-id").(primitive.ObjectID)
}

// GetDecks godoc
//
//	@Summary		List slide decks
//	@Description	List a page of the slide decks of the authenticated user, without their content, last updated first
//	@Tags			slides
//	@Produce		json
//	@Param			search	query		string	false	"Only return decks whose title contains this text"
//	@Param			page	query		int		false	"Page number"				default(1)	minimum(1)
//	@Param			limit	query		int		false	"Number of decks per page"	default(12)	minimum(1)	maximum(100)
//	@Success		200		{object}	DeckPage
//	@Failure		400		{object}	map[string]string
//	@Failure		401		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Security		Bearer
//	@Router			/slides [get]
func (controller *Controller) GetDecks(c *gin.Context) {
	var query DecksQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	page, err := ListDecks(c, requester(c), query)
	if err != nil {
		middlewares.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, page)
}

// GetDeck godoc
//
//	@Summary		Get a slide deck
//	@Description	Get a slide deck of the authenticated user, with its content
//	@Tags			slides
//	@Produce		json
//	@Param			id	path		string	true	"Deck ID"
//	@Success		200	{object}	Deck
//	@Failure		400	{object}	map[string]string
//	@Failure		401	{object}	map[string]string
//	@Failure		404	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Security		Bearer
//	@Router			/slides/{id} [get]
func (controller *Controller) GetDeck(c *gin.Context) {
	id, ok := deckId(c)
	if !ok {
		return
	}
	deck, err := GetDeck(c, requester(c), id)
	if err != nil {
		middlewares.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, deck)
}

// PostDeck godoc
//
//	@Summary		Create a slide deck
//	@Description	Create a slide deck, with a first slide when no content is given
//	@Tags			slides
//	@Accept			json
//	@Produce		json
//	@Param			body	body		PostDeckBody	true	"Deck"
//	@Success		201		{object}	Deck
//	@Failure		400		{object}	map[string]string
//	@Failure		401		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Security		Bearer
//	@Router			/slides [post]
func (controller *Controller) PostDeck(c *gin.Context) {
	var body PostDeckBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	deck, err := AddDeck(c, requester(c), body)
	if err != nil {
		middlewares.RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, deck)
}

// PostDeckDuplicate godoc
//
//	@Summary		Duplicate a slide deck
//	@Description	Create a copy of a slide deck of the authenticated user
//	@Tags			slides
//	@Produce		json
//	@Param			id	path		string	true	"Deck ID"
//	@Success		201	{object}	Deck
//	@Failure		400	{object}	map[string]string
//	@Failure		401	{object}	map[string]string
//	@Failure		404	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Security		Bearer
//	@Router			/slides/{id}/duplicate [post]
func (controller *Controller) PostDeckDuplicate(c *gin.Context) {
	id, ok := deckId(c)
	if !ok {
		return
	}
	deck, err := DuplicateDeck(c, requester(c), id)
	if err != nil {
		middlewares.RespondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, deck)
}

// DeleteDeck godoc
//
//	@Summary		Delete a slide deck
//	@Description	Delete a slide deck of the authenticated user for good
//	@Tags			slides
//	@Param			id	path	string	true	"Deck ID"
//	@Success		204
//	@Failure		400	{object}	map[string]string
//	@Failure		401	{object}	map[string]string
//	@Failure		404	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Security		Bearer
//	@Router			/slides/{id} [delete]
func (controller *Controller) DeleteDeck(c *gin.Context) {
	id, ok := deckId(c)
	if !ok {
		return
	}
	if err := RemoveDeck(c, requester(c), id); err != nil {
		middlewares.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
