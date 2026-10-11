package wordcloud

import (
	"errors"

	uuid "github.com/satori/go.uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var ErrWordNotAdded = errors.New("word not added")

// Word represents a word in a word cloud.
type Word struct {
	Id   primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Text string             `json:"text" bson:"text"`
	// The text as submitted, kept when words are merged
	OriginalText string             `json:"originalText,omitempty" bson:"originalText,omitempty"`
	Hidden       bool               `json:"hidden" bson:"hidden"`
	UUID         uuid.UUID          `json:"uuid" bson:"uuid"`
	UserAgent    string             `json:"userAgent" bson:"userAgent"`
	CreatedAt    primitive.DateTime `json:"createdAt" bson:"createdAt"`
	// Kept for abuse investigations; only the hashes below are ever sent to the host
	Client  *ClientInfo `json:"-" bson:"client,omitempty"`
	Device  string      `json:"device,omitempty" bson:"device,omitempty"`
	Network string      `json:"network,omitempty" bson:"network,omitempty"`
	// Computed on read, so merged words and lexicon updates are always reflected
	Sentiment *int `json:"sentiment,omitempty" bson:"-"`
}

type ClientInfo struct {
	Timezone       string   `json:"timezone" bson:"timezone" binding:"max=64"`
	Languages      []string `json:"languages" bson:"languages" binding:"max=10,dive,max=35"`
	Platform       string   `json:"platform" bson:"platform" binding:"max=64"`
	Mobile         bool     `json:"mobile" bson:"mobile"`
	Screen         string   `json:"screen" bson:"screen" binding:"max=32"`
	AvailableArea  string   `json:"availableArea" bson:"availableArea" binding:"max=32"`
	PixelRatio     float64  `json:"pixelRatio" bson:"pixelRatio" binding:"min=0,max=16"`
	ColorDepth     int      `json:"colorDepth" bson:"colorDepth" binding:"min=0,max=64"`
	Cores          int      `json:"cores" bson:"cores" binding:"min=0,max=1024"`
	Memory         float64  `json:"memory" bson:"memory" binding:"min=0,max=1024"`
	TouchPoints    int      `json:"touchPoints" bson:"touchPoints" binding:"min=0,max=64"`
	ColorScheme    string   `json:"colorScheme" bson:"colorScheme" binding:"max=16"`
	ReducedMotion  bool     `json:"reducedMotion" bson:"reducedMotion"`
	CookiesEnabled bool     `json:"cookiesEnabled" bson:"cookiesEnabled"`
}

type Settings struct {
	ShowCloudToAudience  bool `json:"showCloudToAudience" bson:"showCloudToAudience"`
	FilterProfanity      bool `json:"filterProfanity" bson:"filterProfanity"`
	AllowUpvotes         bool `json:"allowUpvotes" bson:"allowUpvotes"`
	Sentiment            bool `json:"sentiment" bson:"sentiment"`
	SentimentInPresenter bool `json:"sentimentInPresenter" bson:"sentimentInPresenter"`
	// Unset for the default limit, 0 for no limit
	MaxWords *int `json:"maxWords" bson:"maxWords,omitempty" binding:"omitempty,min=0,max=100"`
}

const DefaultMaxWords = 20

// WordLimit returns the words allowed per participant, 0 meaning no limit.
func (settings Settings) WordLimit() int {
	if settings.MaxWords == nil {
		return DefaultMaxWords
	}
	return *settings.MaxWords
}

// WordCloud represents a word cloud object.
type WordCloud struct {
	Id          primitive.ObjectID  `json:"id" bson:"_id,omitempty"`
	UserId      primitive.ObjectID  `json:"user" bson:"user"`
	Name        string              `json:"name" bson:"name"`
	Description string              `json:"description" bson:"description"`
	Code        string              `json:"code" bson:"code"`
	Words       []Word              `json:"words" bson:"words"`
	Settings    Settings            `json:"settings" bson:"settings"`
	CreatedAt   primitive.DateTime  `json:"createdAt" bson:"createdAt"`
	UpdatedAt   primitive.DateTime  `json:"updatedAt" bson:"updatedAt"`
	ClosedAt    *primitive.DateTime `json:"closedAt,omitempty" bson:"closedAt,omitempty"`
}

type PostWordCloudBody struct {
	Name        string    `json:"name" binding:"required,min=3,max=100"`
	Description string    `json:"description" binding:"omitempty,min=10,max=1000"`
	Settings    *Settings `json:"settings"`
}

type PatchWordCloudBody struct {
	Name        *string   `json:"name" binding:"omitempty,min=3,max=100"`
	Description *string   `json:"description" binding:"omitempty,max=1000"`
	Open        *bool     `json:"open"`
	Settings    *Settings `json:"settings"`
}
