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
	CreatedAt primitive.DateTime `json:"createdAt" bson:"createdAt" swaggertype:"string" format:"date-time"`
	UpdatedAt primitive.DateTime `json:"updatedAt" bson:"updatedAt" swaggertype:"string" format:"date-time"`
}

// DeckSummary is a deck without its content, for lists
type DeckSummary struct {
	Id        primitive.ObjectID `json:"id" bson:"_id" swaggertype:"string"`
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

type PostDeckBody struct {
	Title   string  `json:"title" binding:"required,min=1,max=100"`
	Content *string `json:"content" binding:"omitempty,max=500000"`
}
