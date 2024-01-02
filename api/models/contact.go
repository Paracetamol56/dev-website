package models

import (
	"github.com/Paracetamol56/dev-website/api/db"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Contact struct {
	Id        primitive.ObjectID `json:"_id" bson:"_id"`
	UserId    primitive.ObjectID `json:"userId" bson:"userId"`
	Name      string             `json:"name" bson:"name"`
	Email     string             `json:"email" bson:"email"`
	Message   string             `json:"message" bson:"message"`
	MailId    string             `json:"mailId" bson:"mailId"`
	CreatedAt primitive.DateTime `json:"createdAt" bson:"createdAt"`
}

func CreateContact(c *gin.Context, contact Contact) (*primitive.ObjectID, error) {
	db := db.GetDB()
	collection := db.Collection("contacts")
	result, err := collection.InsertOne(c, bson.M{
		"userId":    contact.UserId,
		"name":      contact.Name,
		"email":     contact.Email,
		"message":   contact.Message,
		"mailId":    contact.MailId,
		"createdAt": contact.CreatedAt,
	})
	if err != nil {
		return nil, err
	}
	insertedID := result.InsertedID.(primitive.ObjectID)
	return &insertedID, nil
}
