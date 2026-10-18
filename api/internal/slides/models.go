package slides

import "go.mongodb.org/mongo-driver/bson/primitive"

const DefaultLimit = 12

const defaultContent = `<section>
	<h1>Title</h1>
	<p>Your first slide</p>
</section>
`

type Deck struct {
	Id        primitive.ObjectID `json:"id" bson:"_id,omitempty" swaggertype:"string"`
	UserId    primitive.ObjectID `json:"user" bson:"user" swaggertype:"string"`
	Title     string             `json:"title" bson:"title"`
	Content   string             `json:"content" bson:"content"`
	Outline   []OutlineEntry     `json:"outline" bson:"outline"`
	Settings  Settings           `json:"settings" bson:"settings"`
	Revision  int64              `json:"revision" bson:"revision"`
	Links     Links              `json:"-" bson:"links"`
	Members   []Member           `json:"-" bson:"members,omitempty"`
	CreatedAt primitive.DateTime `json:"createdAt" bson:"createdAt" swaggertype:"string" format:"date-time"`
	UpdatedAt primitive.DateTime `json:"updatedAt" bson:"updatedAt" swaggertype:"string" format:"date-time"`
}

const (
	RoleOwner  = "owner"
	RoleEditor = "editor"
	RoleViewer = "viewer"
)

// Links holds the secret token of each share link; a missing one is turned off
type Links struct {
	Editor *string `json:"editor,omitempty" bson:"editor,omitempty"`
	Viewer *string `json:"viewer,omitempty" bson:"viewer,omitempty"`
}

// Member is someone who joined a deck through one of its links
type Member struct {
	UserId   primitive.ObjectID `json:"user" bson:"user" swaggertype:"string"`
	Role     string             `json:"role" bson:"role" enums:"editor,viewer"`
	JoinedAt primitive.DateTime `json:"joinedAt" bson:"joinedAt" swaggertype:"string" format:"date-time"`
}

type MemberView struct {
	Member
	Name string `json:"name"`
}

// DeckView is a deck as its reader sees it: only the owner gets the links and the members
type DeckView struct {
	*Deck
	Role    string       `json:"role" enums:"owner,editor,viewer"`
	Links   *Links       `json:"links,omitempty"`
	Members []MemberView `json:"members,omitempty"`
}

// DeckSummary is a deck without its content, for lists
type DeckSummary struct {
	Id        primitive.ObjectID `json:"id" bson:"_id" swaggertype:"string"`
	UserId    primitive.ObjectID `json:"-" bson:"user"`
	Members   []Member           `json:"-" bson:"members,omitempty"`
	Role      string             `json:"role" bson:"-" enums:"owner,editor,viewer"`
	OwnerName string             `json:"ownerName,omitempty" bson:"-"`
	Title     string             `json:"title" bson:"title"`
	Outline   []OutlineEntry     `json:"outline" bson:"outline"`
	CreatedAt primitive.DateTime `json:"createdAt" bson:"createdAt" swaggertype:"string" format:"date-time"`
	UpdatedAt primitive.DateTime `json:"updatedAt" bson:"updatedAt" swaggertype:"string" format:"date-time"`
}

type DeckPage struct {
	Items []DeckSummary `json:"items"`
	Total int64         `json:"total"`
	Page  int           `json:"page"`
	Limit int           `json:"limit"`
}

type DecksQuery struct {
	Page   int    `form:"page" binding:"omitempty,min=1"`
	Limit  int    `form:"limit" binding:"omitempty,min=1,max=100"`
	Search string `form:"search" binding:"max=100"`
}

type JoinBody struct {
	Token string `json:"token" binding:"required,max=100"`
}

type JoinedDeck struct {
	Id   primitive.ObjectID `json:"id" swaggertype:"string"`
	Role string             `json:"role" enums:"owner,editor,viewer"`
}

type MemberRoleBody struct {
	Role string `json:"role" binding:"required,oneof=editor viewer"`
}

type EmailLinkBody struct {
	Email string `json:"email" binding:"required,email,max=254"`
}

type PostDeckBody struct {
	Title   string  `json:"title" binding:"required,min=1,max=100"`
	Content *string `json:"content" binding:"omitempty,max=500000"`
}
