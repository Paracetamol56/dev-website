package models

import (
	"github.com/Paracetamol56/dev-website/api/db"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type HipparcosHR struct {
	HIP  int     `json:"hip" bson:"HIP"`
	Amag float64 `json:"Amag" bson:"Amag"`
	BV   float64 `json:"BV" bson:"BV"`
}

func GetAllHipparcosHR(c *gin.Context) ([]HipparcosHR, error) {
	db := db.GetDB()
	collection := db.Collection("hipparcos")

	matchStage := bson.D{{Key: "$match", Value: bson.D{
		{Key: "B-V", Value: bson.D{{Key: "$gte", Value: -0.4}, {Key: "$lte", Value: 2.2}}},
		{Key: "Plx", Value: bson.D{{Key: "$gte", Value: 0.1}}},
	}}}

	projectStage := bson.D{{Key: "$project", Value: bson.D{
		{Key: "_id", Value: 0},
		{Key: "HIP", Value: 1},
		{Key: "Amag", Value: bson.D{{Key: "$subtract", Value: bson.A{
			"$Vmag",
			bson.D{{Key: "$multiply", Value: bson.A{
				5,
				bson.D{{Key: "$log10", Value: bson.D{{Key: "$divide", Value: bson.A{
					bson.D{{Key: "$divide", Value: bson.A{1000, "$Plx"}}},
					10,
				}}}},
				}}},
			}}},
		}},
		{Key: "BV", Value: "$B-V"},
	}}}

	cursor, err := collection.Aggregate(c, mongo.Pipeline{matchStage, projectStage})
	if err != nil {
		return nil, err
	}

	var value []HipparcosHR
	if err = cursor.All(c, &value); err != nil {
		return nil, err
	}

	return value, nil
}

func GetHipparcosHRByHIP(c *gin.Context, hip int) ([]HipparcosHR, error) {
	db := db.GetDB()
	collection := db.Collection("hipparcos")

	matchStage := bson.D{{Key: "$match", Value: bson.D{
		{Key: "HIP", Value: hip},
		{Key: "B-V", Value: bson.D{{Key: "$gte", Value: -0.4}, {Key: "$lte", Value: 2.2}}},
		{Key: "Plx", Value: bson.D{{Key: "$gte", Value: 0.1}}},
	}}}

	projectStage := bson.D{{Key: "$project", Value: bson.D{
		{Key: "_id", Value: 0},
		{Key: "HIP", Value: 1},
		{Key: "Amag", Value: bson.D{{Key: "$subtract", Value: bson.A{
			"$Vmag",
			bson.D{{Key: "$multiply", Value: bson.A{
				5,
				bson.D{{Key: "$log10", Value: bson.D{{Key: "$divide", Value: bson.A{
					bson.D{{Key: "$divide", Value: bson.A{1000, "$Plx"}}},
					10,
				}}}},
				}}},
			}}},
		}},
		{Key: "BV", Value: "$B-V"},
	}}}

	cursor, err := collection.Aggregate(c, mongo.Pipeline{matchStage, projectStage})
	if err != nil {
		return nil, err
	}

	var value []HipparcosHR
	if err = cursor.All(c, &value); err != nil {
		return nil, err
	}

	return value, nil
}
