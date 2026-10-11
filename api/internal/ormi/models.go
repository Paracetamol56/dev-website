package ormi

import (
	"bytes"
	"encoding/json"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	TodoStateTodo       = "TODO"
	TodoStateInProgress = "IN_PROGRESS"
	TodoStateStandby    = "STANDBY"
	TodoStateDone       = "DONE"
	TodoStateCancelled  = "CANCELLED"
)

type OrmiEvent struct {
	PreviousState string             `json:"previousState" bson:"previousState" example:"TODO"`
	NewState      string             `json:"newState" bson:"newState" example:"DONE"`
	UpdatedAt     primitive.DateTime `json:"updatedAt" bson:"updatedAt" swaggertype:"string" format:"date-time"`
}

type Todo struct {
	Id          primitive.ObjectID `json:"id" bson:"_id,omitempty" swaggertype:"string"`
	User        primitive.ObjectID `json:"user,omitempty" bson:"user" swaggertype:"string"`
	Title       string             `json:"title" bson:"title"`
	Description string             `json:"description,omitempty" bson:"description"`
	DueDate     primitive.DateTime `json:"dueDate,omitempty" bson:"dueDate,omitempty" swaggertype:"string" format:"date-time"`
	Labels      []string           `json:"labels,omitempty" bson:"labels"`
	History     []OrmiEvent        `json:"history,omitempty" bson:"history"`
	State       string             `json:"state" bson:"state" enums:"TODO,IN_PROGRESS,STANDBY,DONE,CANCELLED"`
	Position    int                `json:"position" bson:"position"`
	GitURL      string             `json:"gitURL,omitempty" bson:"gitURL,omitempty"`
	GitIssue    uint32             `json:"gitIssue,omitempty" bson:"gitIssue,omitempty"`
	CreatedAt   primitive.DateTime `json:"createdAt" bson:"createdAt,omitempty" swaggertype:"string" format:"date-time"`
}

type TodoDayCount struct {
	Date  string `json:"date" bson:"_id" example:"2026-10-09"`
	Count int    `json:"count" bson:"count"`
}

type TodoStats struct {
	Days          []TodoDayCount `json:"days"`
	FirstYear     int            `json:"firstYear" example:"2024"`
	CurrentStreak int            `json:"currentStreak"`
	LongestStreak int            `json:"longestStreak"`
	Todo          int            `json:"todo"`
	InProgress    int            `json:"inProgress"`
	Standby       int            `json:"standby"`
	Overdue       int            `json:"overdue"`
}

type TodoFilter struct {
	States []string   `form:"state" binding:"omitempty,dive,oneof=TODO IN_PROGRESS STANDBY DONE CANCELLED"`
	Labels []string   `form:"label" binding:"omitempty,max=20,dive,max=50"`
	From   *time.Time `form:"from" time_format:"2006-01-02T15:04:05Z07:00"`
	To     *time.Time `form:"to" time_format:"2006-01-02T15:04:05Z07:00"`
	Sort   string     `form:"sort,default=position" binding:"oneof=position due created updated"`
	Order  string     `form:"order" binding:"omitempty,oneof=asc desc"`
	Limit  int64      `form:"limit,default=50" binding:"min=1,max=200"`
	Offset int64      `form:"offset,default=0" binding:"min=0"`
}

type TodoPage struct {
	Total int    `json:"total" example:"42"`
	Items []Todo `json:"items"`
}

type TodoBody struct {
	Title       string             `json:"title" binding:"required,min=2,max=100"`
	Description string             `json:"description" binding:"omitempty,max=1000"`
	DueDate     primitive.DateTime `json:"dueDate" binding:"omitempty" swaggertype:"string" format:"date-time"`
	Labels      []string           `json:"labels" binding:"omitempty"`
	GitURL      string             `json:"gitURL" binding:"omitempty,url"`
	GitIssue    uint32             `json:"gitIssue" binding:"omitempty,numeric,min=1"`
}

type OptionalDate struct {
	Set   bool
	Value primitive.DateTime
}

func (d *OptionalDate) UnmarshalJSON(data []byte) error {
	d.Set = true
	if bytes.Equal(data, []byte("null")) {
		d.Value = 0
		return nil
	}
	return json.Unmarshal(data, &d.Value)
}

type PatchTodoBody struct {
	Title       *string      `json:"title" binding:"omitempty,min=2,max=100"`
	Description *string      `json:"description" binding:"omitempty,max=1000"`
	DueDate     OptionalDate `json:"dueDate" swaggertype:"string" format:"date-time"`
	Labels      *[]string    `json:"labels" binding:"omitempty"`
	GitURL      *string      `json:"gitURL" binding:"omitempty,url"`
	GitIssue    *uint32      `json:"gitIssue" binding:"omitempty,numeric,min=1"`
	State       *string      `json:"state" binding:"omitempty,oneof=TODO IN_PROGRESS STANDBY DONE CANCELLED"`
}
