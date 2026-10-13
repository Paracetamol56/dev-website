package contact

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Contact represents a contact entity.
type Contact struct {
	Id        primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	UserId    primitive.ObjectID `json:"userId" bson:"userId"`
	Name      string             `json:"name" bson:"name"`
	Email     string             `json:"email" bson:"email"`
	Message   string             `json:"message" bson:"message"`
	MailId    string             `json:"mailId" bson:"mailId"`
	CreatedAt primitive.DateTime `json:"createdAt" bson:"createdAt"`
}

type ContactBody struct {
	UserId  primitive.ObjectID `json:"userId" binding:""`
	Name    string             `json:"name" binding:"required,min=2,max=100"`
	Email   string             `json:"email" binding:"required,email"`
	Message string             `json:"message" binding:"required,min=10,max=1000"`
}
