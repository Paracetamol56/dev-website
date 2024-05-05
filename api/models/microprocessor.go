package models

import (
	"github.com/Paracetamol56/dev-website/api/db"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Microprocessor struct {
	Id          string             `json:"id" bson:"_id"`
	Name        string             `json:"name" bson:"name"`
	Type        string             `json:"type" bson:"type"`
	Release     primitive.DateTime `json:"release" bson:"release"`
	GateSize    int                `json:"gateSize" bson:"gate_size"`
	TDP         int                `json:"tdp" bson:"tdp"`
	DieSize     int                `json:"dieSize" bson:"die_size"`
	Transistors int                `json:"transistors" bson:"transistors"`
	Frequency   int                `json:"frequency" bson:"frequency"`
	Vendor      string             `json:"vendor" bson:"vendor"`
}

func GetAllMicroprocessors(c *gin.Context) ([]Microprocessor, error) {
	db := db.GetDB()
	collection := db.Collection("transistors_per_microprocessor")

	cursor, err := collection.Find(c, bson.M{}, nil)
	if err != nil {
		return nil, err
	}

	var value []Microprocessor
	if err = cursor.All(c, &value); err != nil {
		return nil, err
	}

	return value, nil
}

func GetMicroprocessorById(c *gin.Context, id primitive.ObjectID) (*Microprocessor, error) {
	db := db.GetDB()
	collection := db.Collection("transistors_per_microprocessor")

	var value Microprocessor
	if err := collection.FindOne(c, bson.M{"_id": id}).Decode(&value); err != nil {
		return nil, err
	}

	return &value, nil
}
