package models

import (
	"dev/internal/db"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CodeCarbonRun struct {
	Id primitive.ObjectID `json:"id" bson:"_id"`
}

type CodeCarbonToken struct {
	Id         primitive.ObjectID `json:"id" bson:"_id"`
	Name       string             `json:"name" bson:"name"`
	Token      string             `json:"token" bson:"token"`
	CreatedAt  time.Time          `json:"createdAt" bson:"createdAt"`
	LastUsedAt *time.Time         `json:"lastUsedAt" bson:"lastUsedAt,omitempty"`
	RevokedAt  *time.Time         `json:"revokedAt,omitempty" bson:"revokedAt,omitempty"`
}

type CodeCarbonProject struct {
	Id          primitive.ObjectID `json:"id" bson:"_id"`
	Name        string             `json:"name" bson:"name"`
	Description string             `json:"description" bson:"description"`
	Runs        []CodeCarbonRun    `json:"runs" bson:"runs"`
	Tokens      []CodeCarbonToken  `json:"tokens" bson:"tokens"`
	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt   time.Time          `json:"updatedAt" bson:"updatedAt"`
	ClosedAt    *time.Time         `json:"closedAt,omitempty" bson:"closedAt,omitempty"`
}

func GetCodeCarbonProjectsByUser(c *gin.Context, userId primitive.ObjectID) ([]*CodeCarbonProject, error) {
	db := db.GetDB()
	collection := db.Collection("users")
	cursor, err := collection.Aggregate(
		c,
		bson.A{
			bson.D{{"$match", bson.D{{"_id", userId}}}},
			bson.D{{"$unwind", bson.D{{"path", "$codecarbon_projects"}}}},
			bson.D{{"$replaceRoot", bson.D{{"newRoot", "$codecarbon_projects"}}}},
		},
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(c)
	var codeCarbonProjects []*CodeCarbonProject
	if err := cursor.All(c, &codeCarbonProjects); err != nil {
		return nil, err
	}
	return codeCarbonProjects, nil
}

func GetCodeCarbonProjectById(c *gin.Context, userId primitive.ObjectID, projectId primitive.ObjectID) (*CodeCarbonProject, error) {
	db := db.GetDB()
	collection := db.Collection("users")
	cursor, err := collection.Aggregate(
		c,
		bson.A{
			bson.D{{"$match", bson.D{{"_id", userId}}}},
			bson.D{{"$unwind", bson.D{{"path", "$codecarbon_projects"}}}},
			bson.D{{"$match", bson.D{{"codecarbon_projects._id", projectId}}}},
			bson.D{{"$limit", 1}},
			bson.D{{"$replaceRoot", bson.D{{"newRoot", "$codecarbon_projects"}}}},
		},
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(c)
	var project *CodeCarbonProject
	if !cursor.Next(c) {
		return nil, mongo.ErrNoDocuments
	}
	if err := cursor.Decode(&project); err != nil {
		return nil, err
	}
	return project, nil
}

func CreateCodeCarbonProject(c *gin.Context, userId primitive.ObjectID, codeCarbonProject *CodeCarbonProject) (*mongo.UpdateResult, error) {
	db := db.GetDB()
	collection := db.Collection("users")
	result, err := collection.UpdateOne(
		c,
		bson.M{"_id": userId},
		bson.D{
			{"$push", bson.M{"codecarbon_projects": codeCarbonProject}},
		},
	)
	return result, err
}

func UpdateCodeCarbonProject(c *gin.Context, userId primitive.ObjectID, codeCarbonProject *CodeCarbonProject) (*mongo.UpdateResult, error) {
	db := db.GetDB()
	collection := db.Collection("users")

	result, err := collection.UpdateOne(
		c,
		bson.M{"_id": userId, "codecarbon_projects._id": codeCarbonProject.Id},
		bson.D{
			{"$set", bson.M{"codecarbon_projects.$": codeCarbonProject}},
		},
	)

	return result, err
}

func DeleteCodeCarbonProject(c *gin.Context, userId primitive.ObjectID, projectId primitive.ObjectID) (*mongo.UpdateResult, error) {
	db := db.GetDB()
	collection := db.Collection("users")
	result, err := collection.UpdateOne(
		c,
		bson.M{"_id": userId},
		bson.D{
			{"$pull", bson.M{"codecarbon_projects": bson.M{"_id": projectId}}},
		},
	)
	return result, err
}

func GetCodeCarbonProjectTokensByProject(c *gin.Context, userId primitive.ObjectID, projectId primitive.ObjectID) ([]*CodeCarbonToken, error) {
	db := db.GetDB()
	collection := db.Collection("users")
	cursor, err := collection.Aggregate(
		c,
		bson.A{
			bson.D{{"$match", bson.D{{"_id", userId}}}},
			bson.D{{"$unwind", bson.D{{"path", "$codecarbon_projects"}}}},
			bson.D{{"$match", bson.D{{"codecarbon_projects._id", projectId}}}},
			bson.D{{"$unwind", bson.D{{"path", "$codecarbon_projects.tokens"}}}},
			bson.D{{"$replaceRoot", bson.D{{"newRoot", "$codecarbon_projects.tokens"}}}},
			bson.D{{"$match", bson.D{{"revokedAt", bson.D{{"$exists", false}}}}}},
		},
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(c)
	codeCarbonTokens := make([]*CodeCarbonToken, 0)
	if err := cursor.All(c, &codeCarbonTokens); err != nil {
		return nil, err
	}
	return codeCarbonTokens, nil
}

func CreateCodeCarbonProjectToken(c *gin.Context, userId primitive.ObjectID, projectId primitive.ObjectID, token *CodeCarbonToken) (*mongo.UpdateResult, error) {
	db := db.GetDB()
	collection := db.Collection("users")
	result, err := collection.UpdateOne(
		c,
		bson.M{"_id": userId, "codecarbon_projects._id": projectId},
		bson.D{
			{"$push", bson.M{"codecarbon_projects.$.tokens": token}},
		},
	)
	return result, err
}

func UpdateCodeCarbonProjectToken(c *gin.Context, userId primitive.ObjectID, projectId primitive.ObjectID, token *CodeCarbonToken) (*mongo.UpdateResult, error) {
	db := db.GetDB()
	collection := db.Collection("users")
	result, err := collection.UpdateOne(
		c,
		bson.M{"_id": userId, "codecarbon_projects._id": projectId, "codecarbon_projects.tokens._id": token.Id},
		bson.D{
			{"$set", bson.M{"codecarbon_projects.tokens.$": token}},
		},
	)
	return result, err
}

func UseProjectToken(c *gin.Context, userId primitive.ObjectID, projectId primitive.ObjectID, tokenId primitive.ObjectID) (*mongo.UpdateResult, error) {
	db := db.GetDB()
	collection := db.Collection("users")
	result, err := collection.UpdateOne(
		c,
		bson.M{"_id": userId, "codecarbon_projects.$[projectId].tokens.$[tokenId].revokedAt": bson.M{"$exists": false}},
		bson.D{
			{"$set", bson.M{"codecarbon_projects.$[projectId].tokens.$[tokenId].lastUsedAt": time.Now()}},
		},
		options.Update().SetArrayFilters(options.ArrayFilters{
			Filters: []interface{}{
				bson.D{{"projectId._id", projectId}},
				bson.D{{"tokenId._id", tokenId}},
			},
		}),
	)
	return result, err
}

func RevokeProjectToken(c *gin.Context, userId primitive.ObjectID, projectId primitive.ObjectID, tokenId primitive.ObjectID) (*mongo.UpdateResult, error) {
	db := db.GetDB()
	collection := db.Collection("users")
	result, err := collection.UpdateOne(
		c,
		bson.M{"_id": userId, "codecarbon_projects.$[projectId].tokens.$[tokenId].revokedAt": bson.M{"$exists": false}},
		bson.D{
			{"$set", bson.M{"codecarbon_projects.$[projectId].tokens.$[tokenId].revokedAt": time.Now()}},
		},
		options.Update().SetArrayFilters(options.ArrayFilters{
			Filters: []interface{}{
				bson.D{{"projectId._id", projectId}},
				bson.D{{"tokenId._id", tokenId}},
			},
		}),
	)
	return result, err
}
