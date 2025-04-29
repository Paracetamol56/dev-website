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

type CodeCarbonEmissions struct {
	Id                primitive.ObjectID `json:"id" bson:"_id"`
	TimeStamp         time.Time          `json:"timestamp" bson:"timestamp"`
	Duration          int64              `json:"duration" bson:"duration"`
	EmissionsSum      float64            `json:"emissions_sum" bson:"emissions_sum"`
	EmissionsRate     float64            `json:"emissions_rate" bson:"emissions_rate"`
	CpuPower          float64            `json:"cpu_power" bson:"cpu_power"`
	GpuPower          float64            `json:"gpu_power" bson:"gpu_power"`
	RamPower          float64            `json:"ram_power" bson:"ram_power"`
	CpuEnergy         float64            `json:"cpu_energy" bson:"cpu_energy"`
	GpuEnergy         float64            `json:"gpu_energy" bson:"gpu_energy"`
	RamEnergy         float64            `json:"ram_energy" bson:"ram_energy"`
	EnergyConsumption float64            `json:"energy_consumption" bson:"energy_consumption"`
	CreatedAt         time.Time          `json:"created_at" bson:"created_at"`
}

type CodeCarbonRun struct {
	Id                primitive.ObjectID    `json:"id" bson:"_id"`
	Os                string                `json:"os" bson:"os"`
	PythonVersion     string                `json:"python_version" bson:"python_version"`
	CodeCarbonVersion string                `json:"codecarbon_version" bson:"codecarbon_version"`
	CpuCount          int32                 `json:"cpu_count" bson:"cpu_count"`
	CpuModel          string                `json:"cpu_model" bson:"cpu_model"`
	GpuCount          int32                 `json:"cpu_count" bson:"cpu_count"`
	GpuModel          string                `json:"gpu_model" bson:"gpu_model"`
	Location          []float64             `json:"location" bson:"location"`
	Region            string                `json:"region" bson:"region"`
	Provider          string                `json:"provider" bson:"provider"`
	RamTotalSize      float64               `json:"ram_total_size" bson:"ram_total_size"`
	TrackingMode      string                `json:"tracking_mode" bson:"tracking_mode"`
	Emissions         []CodeCarbonEmissions `json:"emissions" bson:"emissions"`
	CreatedAt         time.Time             `json:"created_at" bson:"created_at"`
}

type CodeCarbonExperiment struct {
	Id             primitive.ObjectID `json:"id" bson:"_id"`
	Name           string             `json:"name" bson:"name"`
	Description    string             `json:"description" bson:"description"`
	CountryIsoCode string             `json:"country_iso_code" bson:"country_iso_code"`
	Region         string             `json:"region" bson:"region"`
	OnCloud        bool               `json:"on_cloud" bson:"on_cloud"`
	CloudProvider  string             `json:"cloud_provider" bson:"cloud_provider"`
	CloudRegion    string             `json:"cloud_region" bson:"cloud_region"`
	Runs           []CodeCarbonRun    `json:"runs" bson:"runs"`
	CreatedAt      time.Time          `json:"created_at" bson:"created_at"`
}

type CodeCarbonToken struct {
	Id         primitive.ObjectID `json:"id" bson:"_id"`
	Name       string             `json:"name" bson:"name"`
	Token      string             `json:"token" bson:"token"`
	CreatedAt  time.Time          `json:"created_at" bson:"created_at"`
	LastUsedAt *time.Time         `json:"lastUsed_at" bson:"lastUsed_at,omitempty"`
	RevokedAt  *time.Time         `json:"revoked_at,omitempty" bson:"revoked_at,omitempty"`
}

type CodeCarbonProject struct {
	Id          primitive.ObjectID     `json:"id" bson:"_id"`
	Name        string                 `json:"name" bson:"name"`
	Description string                 `json:"description" bson:"description"`
	Experiments []CodeCarbonExperiment `json:"experiments" bson:"experiments"`
	Tokens      []CodeCarbonToken      `json:"tokens" bson:"tokens"`
	CreatedAt   time.Time              `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" bson:"updated_at"`
	ClosedAt    *time.Time             `json:"closed_at,omitempty" bson:"closed_at,omitempty"`
}

func GetCodeCarbonTokensByProject(c *gin.Context, userId primitive.ObjectID, projectId primitive.ObjectID) ([]*CodeCarbonToken, error) {
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
			bson.D{{"$match", bson.D{{"revokedAt", bson.M{"$exists": false}}}}},
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

func CreateCodeCarbonToken(c *gin.Context, userId primitive.ObjectID, projectId primitive.ObjectID, token *CodeCarbonToken) (*mongo.UpdateResult, error) {
	db := db.GetDB()
	collection := db.Collection("users")
	result, err := collection.UpdateOne(
		c,
		bson.M{"_id": userId, "codecarbon_projects._id": projectId},
		bson.D{
			{"$push", bson.M{"codecarbon_projects.$.tokens": token}},
			{"$set", bson.M{"codecarbon_projects.$.updated_at": time.Now()}},
		},
	)
	return result, err
}

func UseToken(c *gin.Context, userId primitive.ObjectID, projectId primitive.ObjectID, tokenId primitive.ObjectID) (*mongo.UpdateResult, error) {
	db := db.GetDB()
	collection := db.Collection("users")
	result, err := collection.UpdateOne(
		c,
		bson.M{"_id": userId},
		bson.D{
			{"$set", bson.M{"codecarbon_projects.$[projectId].tokens.$[tokenId].lastUsedAt": time.Now()}},
			{"$set", bson.M{"codecarbon_projects.$[projectId].updated_at": time.Now()}},
		},
		options.Update().SetArrayFilters(options.ArrayFilters{
			Filters: []interface{}{
				bson.D{{"projectId._id", projectId}},
				bson.D{{"token._id", tokenId}, {"token.revokedAt", bson.M{"$exists": false}}},
			},
		}),
	)
	return result, err
}

func RevokeToken(c *gin.Context, userId primitive.ObjectID, projectId primitive.ObjectID, tokenId primitive.ObjectID) (*mongo.UpdateResult, error) {
	db := db.GetDB()
	collection := db.Collection("users")
	result, err := collection.UpdateOne(
		c,
		bson.M{"_id": userId},
		bson.D{
			{"$set", bson.M{"codecarbon_projects.$[project].tokens.$[token].revokedAt": time.Now()}},
			{"$set", bson.M{"codecarbon_projects.$[project].updated_at": time.Now()}},
		},
		options.Update().SetArrayFilters(options.ArrayFilters{
			Filters: []interface{}{
				bson.D{{"project._id", projectId}},
				bson.D{{"token._id", tokenId}, {"token.revokedAt", bson.M{"$exists": false}}},
			},
		}),
	)
	return result, err
}

func GetCodeCarbonExperimentsByProject(c *gin.Context, userId primitive.ObjectID, projectId primitive.ObjectID) ([]*CodeCarbonExperiment, error) {
	db := db.GetDB()
	collection := db.Collection("users")
	cursor, err := collection.Aggregate(
		c,
		bson.A{
			bson.D{{"$match", bson.D{{"_id", userId}}}},
			bson.D{{"$unwind", bson.D{{"path", "$codecarbon_projects"}}}},
			bson.D{{"$match", bson.D{{"codecarbon_projects._id", projectId}}}},
			bson.D{{"$unwind", bson.D{{"path", "$codecarbon_projects.experiments"}}}},
			bson.D{{"$replaceRoot", bson.D{{"newRoot", "$codecarbon_projects.experiments"}}}},
		},
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(c)
	codeCarbonExperiments := make([]*CodeCarbonExperiment, 0)
	if err := cursor.All(c, &codeCarbonExperiments); err != nil {
		return nil, err
	}
	return codeCarbonExperiments, nil
}

func GetCodeCarbonExperimentsByProjectAndId(c *gin.Context, userId primitive.ObjectID, projectId primitive.ObjectID, experimentId primitive.ObjectID) (*CodeCarbonExperiment, error) {
	db := db.GetDB()
	collection := db.Collection("users")
	cursor, err := collection.Aggregate(
		c,
		bson.A{
			bson.D{{"$match", bson.D{{"_id", userId}}}},
			bson.D{{"$unwind", bson.D{{"path", "$codecarbon_projects"}}}},
			bson.D{{"$match", bson.D{{"codecarbon_projects._id", projectId}}}},
			bson.D{{"$unwind", bson.D{{"path", "$codecarbon_projects.experiments"}}}},
			bson.D{{"$match", bson.D{{"codecarbon_projects.experiments._id", experimentId}}}},
			bson.D{{"$limit", 1}},
			bson.D{{"$replaceRoot", bson.D{{"newRoot", "$codecarbon_projects.experiments"}}}},
		},
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(c)
	var experiment *CodeCarbonExperiment
	if !cursor.Next(c) {
		return nil, mongo.ErrNoDocuments
	}
	if err := cursor.Decode(&experiment); err != nil {
		return nil, err
	}
	return experiment, nil
}

func GetCodeCarbonExperimentsById(c *gin.Context, userId primitive.ObjectID, experimentId primitive.ObjectID) (*CodeCarbonExperiment, error) {
	db := db.GetDB()
	collection := db.Collection("users")
	cursor, err := collection.Aggregate(
		c,
		bson.A{
			bson.D{{"$match", bson.D{{"_id", userId}}}},
			bson.D{{"$unwind", bson.D{{"path", "$codecarbon_projects"}}}},
			bson.D{{"$unwind", bson.D{{"path", "$codecarbon_projects.experiments"}}}},
			bson.D{{"$match", bson.D{{"codecarbon_projects.experiments._id", experimentId}}}},
			bson.D{{"$limit", 1}},
			bson.D{{"$replaceRoot", bson.D{{"newRoot", "$codecarbon_projects.experiments"}}}},
		},
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(c)
	var experiment *CodeCarbonExperiment
	if !cursor.Next(c) {
		return nil, mongo.ErrNoDocuments
	}
	if err := cursor.Decode(&experiment); err != nil {
		return nil, err
	}
	return experiment, nil
}

func CreateCodeCarbonExperiment(c *gin.Context, userId primitive.ObjectID, projectId primitive.ObjectID, experiment *CodeCarbonExperiment) (*mongo.UpdateResult, error) {
	db := db.GetDB()
	collection := db.Collection("users")
	result, err := collection.UpdateOne(
		c,
		bson.M{"_id": userId, "codecarbon_projects._id": projectId},
		bson.D{
			{"$push", bson.M{"codecarbon_projects.$.experiments": experiment}},
			{"$set", bson.M{"codecarbon_projects.$.updated_at": time.Now()}},
		},
	)
	return result, err
}

func UpdateCodeCarbonExperiment(c *gin.Context, userId primitive.ObjectID, projectId primitive.ObjectID, experiment *CodeCarbonExperiment) (*mongo.UpdateResult, error) {
	db := db.GetDB()
	collection := db.Collection("users")
	result, err := collection.UpdateOne(
		c,
		bson.M{"_id": userId, "codecarbon_projects._id": projectId},
		bson.D{
			{"$set", bson.M{"codecarbon_projects.$[project].experiments.$[experiment]": experiment}},
			{"$set", bson.M{"codecarbon_projects.$[project].updated_at": time.Now()}},
		},
		options.Update().SetArrayFilters(options.ArrayFilters{
			Filters: []interface{}{
				bson.D{{"project._id", projectId}},
				bson.D{{"experiment._id", experiment.Id}},
			},
		}),
	)
	return result, err
}

func DeleteCodeCarbonExperiment(c *gin.Context, userId primitive.ObjectID, projectId primitive.ObjectID, experimentId primitive.ObjectID) (*mongo.UpdateResult, error) {
	db := db.GetDB()
	collection := db.Collection("users")
	result, err := collection.UpdateOne(
		c,
		bson.M{"_id": userId},
		bson.D{
			{"$pull", bson.M{"codecarbon_projects.$[project].experiments": bson.M{"_id": experimentId}}},
			{"$set", bson.M{"codecarbon_projects.$[project].updated_at": time.Now()}},
		},
		options.Update().SetArrayFilters(options.ArrayFilters{
			Filters: []interface{}{
				bson.D{{"project._id", projectId}},
			},
		}),
	)
	return result, err
}
