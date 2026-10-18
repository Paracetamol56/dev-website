package slides

import (
	"dev/internal/middlewares"
	"dev/internal/utils"
	"net/http"
	"slices"

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
//	@Description	List a page of the slide decks the authenticated user owns or is a member of, without their content, last updated first
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
//	@Description	Get a slide deck the authenticated user owns or is a member of, with its content and their role. The owner also gets the share links and the members.
//	@Tags			slides
//	@Produce		json
//	@Param			id	path		string	true	"Deck ID"
//	@Success		200	{object}	DeckView
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
//	@Description	Create a copy, owned by the authenticated user, of a deck they own or edit
//	@Tags			slides
//	@Produce		json
//	@Param			id	path		string	true	"Deck ID"
//	@Success		201	{object}	Deck
//	@Failure		400	{object}	map[string]string
//	@Failure		401	{object}	map[string]string
//	@Failure		403	{object}	map[string]string
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
//	@Failure		403	{object}	map[string]string
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

func memberId(c *gin.Context) (primitive.ObjectID, bool) {
	id, err := primitive.ObjectIDFromHex(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return primitive.NilObjectID, false
	}
	return id, true
}

// site is where share links point: the site the request comes from when it is a known one
func site(c *gin.Context) string {
	origin := c.GetHeader("Origin")
	if slices.Contains(utils.AllowedOrigins, origin) {
		return origin
	}
	return defaultSite
}

// PutDeckLink godoc
//
//	@Summary		Create a share link
//	@Description	Turn on the share link giving a role in a deck, or replace its token so the previous link stops working. Only the owner can do this.
//	@Tags			slides
//	@Produce		json
//	@Param			id		path		string	true	"Deck ID"
//	@Param			role	path		string	true	"Role given by the link"	Enums(editor, viewer)
//	@Success		200		{object}	Links
//	@Failure		400		{object}	map[string]string
//	@Failure		401		{object}	map[string]string
//	@Failure		403		{object}	map[string]string
//	@Failure		404		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Security		Bearer
//	@Router			/slides/{id}/links/{role} [put]
func (controller *Controller) PutDeckLink(c *gin.Context) {
	id, ok := deckId(c)
	if !ok {
		return
	}
	links, err := CreateLink(c, requester(c), id, c.Param("role"))
	if err != nil {
		middlewares.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, links)
}

// DeleteDeckLink godoc
//
//	@Summary		Turn off a share link
//	@Description	Turn off the share link giving a role in a deck. The members who joined through it keep their access. Only the owner can do this.
//	@Tags			slides
//	@Param			id		path	string	true	"Deck ID"
//	@Param			role	path	string	true	"Role given by the link"	Enums(editor, viewer)
//	@Success		204
//	@Failure		400	{object}	map[string]string
//	@Failure		401	{object}	map[string]string
//	@Failure		403	{object}	map[string]string
//	@Failure		404	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Security		Bearer
//	@Router			/slides/{id}/links/{role} [delete]
func (controller *Controller) DeleteDeckLink(c *gin.Context) {
	id, ok := deckId(c)
	if !ok {
		return
	}
	if err := RevokeLink(c, requester(c), id, c.Param("role")); err != nil {
		middlewares.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// PostDeckLinkEmail godoc
//
//	@Summary		Email a share link
//	@Description	Send the share link giving a role in a deck to an email address, turning the link on if needed. Only the owner can do this, a few times per hour.
//	@Tags			slides
//	@Accept			json
//	@Param			id		path	string			true	"Deck ID"
//	@Param			role	path	string			true	"Role given by the link"	Enums(editor, viewer)
//	@Param			body	body	EmailLinkBody	true	"Recipient"
//	@Success		204
//	@Failure		400	{object}	map[string]string
//	@Failure		401	{object}	map[string]string
//	@Failure		403	{object}	map[string]string
//	@Failure		404	{object}	map[string]string
//	@Failure		429	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Security		Bearer
//	@Router			/slides/{id}/links/{role}/email [post]
func (controller *Controller) PostDeckLinkEmail(c *gin.Context) {
	id, ok := deckId(c)
	if !ok {
		return
	}
	var body EmailLinkBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := EmailLink(c, requester(c), id, c.Param("role"), body.Email, site(c)); err != nil {
		middlewares.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// PostDeckJoin godoc
//
//	@Summary		Join a slide deck
//	@Description	Become a member of the deck of a share link, with the role of the link. An editor link upgrades a viewer; nothing changes for the owner.
//	@Tags			slides
//	@Accept			json
//	@Produce		json
//	@Param			body	body		JoinBody	true	"Token of the link"
//	@Success		200		{object}	JoinedDeck
//	@Failure		400		{object}	map[string]string
//	@Failure		401		{object}	map[string]string
//	@Failure		404		{object}	map[string]string
//	@Failure		500		{object}	map[string]string
//	@Security		Bearer
//	@Router			/slides/join [post]
func (controller *Controller) PostDeckJoin(c *gin.Context) {
	var body JoinBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	joined, err := JoinDeck(c, requester(c), body.Token)
	if err != nil {
		middlewares.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, joined)
}

// PatchDeckMember godoc
//
//	@Summary		Change the role of a member
//	@Description	Make a member of a deck an editor or a viewer. Only the owner can do this.
//	@Tags			slides
//	@Accept			json
//	@Param			id		path	string			true	"Deck ID"
//	@Param			userId	path	string			true	"User ID of the member"
//	@Param			body	body	MemberRoleBody	true	"New role"
//	@Success		204
//	@Failure		400	{object}	map[string]string
//	@Failure		401	{object}	map[string]string
//	@Failure		403	{object}	map[string]string
//	@Failure		404	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Security		Bearer
//	@Router			/slides/{id}/members/{userId} [patch]
func (controller *Controller) PatchDeckMember(c *gin.Context) {
	id, ok := deckId(c)
	if !ok {
		return
	}
	member, ok := memberId(c)
	if !ok {
		return
	}
	var body MemberRoleBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := ChangeMemberRole(c, requester(c), id, member, body.Role); err != nil {
		middlewares.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// DeleteDeckMember godoc
//
//	@Summary		Remove a member
//	@Description	Take a member out of a deck: the owner can remove anyone, and a member can leave. A removed member is disconnected from the live deck.
//	@Tags			slides
//	@Param			id		path	string	true	"Deck ID"
//	@Param			userId	path	string	true	"User ID of the member"
//	@Success		204
//	@Failure		400	{object}	map[string]string
//	@Failure		401	{object}	map[string]string
//	@Failure		403	{object}	map[string]string
//	@Failure		404	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Security		Bearer
//	@Router			/slides/{id}/members/{userId} [delete]
func (controller *Controller) DeleteDeckMember(c *gin.Context) {
	id, ok := deckId(c)
	if !ok {
		return
	}
	member, ok := memberId(c)
	if !ok {
		return
	}
	if err := RemoveMember(c, requester(c), id, member); err != nil {
		middlewares.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
