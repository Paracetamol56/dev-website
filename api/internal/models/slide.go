package models

import (
	"encoding/json"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SlideDeck struct {
	Id      primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Title   string             `json:"title" bson:"title" validate:"required,min=1,max=255"`
	Content string             `json:"content" bson:"content"`
	UserId  primitive.ObjectID `json:"user_id" bson:"user_id" validate:"required"`

	// Document versioning and OT support
	Revision   int64         `json:"revision" bson:"revision"`
	Operations []SlideUpdate `json:"operations,omitempty" bson:"operations,omitempty"`

	// Asset management
	Assets []Asset `json:"assets,omitempty" bson:"assets,omitempty"`

	// Status management
	Status SlideStatus `json:"status" bson:"status"`

	CreatedAt  time.Time  `json:"created_at" bson:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at" bson:"updated_at"`
	ArchivedAt *time.Time `json:"archived_at,omitempty" bson:"archived_at,omitempty"`
}

// Operation represents an operational transform operation
type SlideUpdate struct {
	ClientID string          `json:"clientID" bson:"client_id"`
	Changes  json.RawMessage `json:"changes" bson:"changes"`
	Effects  json.RawMessage `json:"effects,omitempty" bson:"effects,omitempty"`
}

// Asset represents uploaded files (images, etc.)
type Asset struct {
	Id         primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Filename   string             `json:"filename" bson:"filename"`
	Path       string             `json:"path" bson:"path"`
	MimeType   string             `json:"mime_type" bson:"mime_type"`
	Size       int64              `json:"size" bson:"size"`
	URL        string             `json:"url" bson:"url"`
	UploadedAt time.Time          `json:"uploaded_at" bson:"uploaded_at"`
}

// Enums
type SlideStatus string

const (
	StatusDraft     SlideStatus = "draft"
	StatusPublished SlideStatus = "published"
	StatusPrivate   SlideStatus = "private"
)
