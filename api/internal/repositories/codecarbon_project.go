package repositories

import (
	"context"
	"dev/internal/db"
	"dev/internal/models"
	"dev/internal/utils"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type CodeCarbonProjectRepository struct {
	collection *mongo.Collection
}

func NewCodeCarbonProjectRepository() *CodeCarbonProjectRepository {
	db := db.GetDB()
	collection := db.Collection("users")

	return &CodeCarbonProjectRepository{collection}
}

func (r *CodeCarbonProjectRepository) GetCodeCarbonProjectsByUser(
	c context.Context,
	userId primitive.ObjectID,
	pagination *utils.Pagination,
) (*utils.PaginatedFacetResult[models.CodeCarbonProject], error) {
	fmt.Println("GetCodeCarbonProjectsByUser")
	fmt.Println(userId)
	fmt.Println(pagination)
	fmt.Println(pagination.GetFacetStage())
	cursor, err := r.collection.Aggregate(
		c,
		bson.A{
			bson.D{{"$match", bson.D{{"_id", userId}}}},
			bson.D{{"$unwind", bson.D{{"path", "$codecarbon_projects"}}}},
			bson.D{{"$replaceRoot", bson.D{{"newRoot", "$codecarbon_projects"}}}},
			bson.D{{"$sort", bson.D{{"updated_at", -1}}}},
			pagination.GetFacetStage(),
		},
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(c)
	var result utils.PaginatedFacetResult[models.CodeCarbonProject]
	if !cursor.Next(c) {
		return nil, mongo.ErrNoDocuments
	}
	if err := cursor.Decode(&result); err != nil {
		return nil, err
	}
	fmt.Println(result)
	return &result, nil
}

func (r *CodeCarbonProjectRepository) GetCodeCarbonProjectById(c context.Context, userId primitive.ObjectID, projectId primitive.ObjectID) (*models.CodeCarbonProject, error) {
	cursor, err := r.collection.Aggregate(
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
	var project *models.CodeCarbonProject
	if !cursor.Next(c) {
		return nil, mongo.ErrNoDocuments
	}
	if err := cursor.Decode(&project); err != nil {
		return nil, err
	}
	return project, nil
}

func (r *CodeCarbonProjectRepository) CreateCodeCarbonProject(c context.Context, userId primitive.ObjectID, codeCarbonProject *models.CodeCarbonProject) (*mongo.UpdateResult, error) {
	result, err := r.collection.UpdateOne(
		c,
		bson.M{"_id": userId},
		bson.D{
			{"$push", bson.M{"codecarbon_projects": codeCarbonProject}},
		},
	)
	return result, err
}

func (r *CodeCarbonProjectRepository) UpdateCodeCarbonProject(c context.Context, userId primitive.ObjectID, codeCarbonProject *models.CodeCarbonProject) (*mongo.UpdateResult, error) {
	result, err := r.collection.UpdateOne(
		c,
		bson.M{"_id": userId, "codecarbon_projects._id": codeCarbonProject.Id},
		bson.D{
			{"$set", bson.M{"codecarbon_projects.$": codeCarbonProject}},
		},
	)

	return result, err
}

func (r *CodeCarbonProjectRepository) DeleteCodeCarbonProject(c context.Context, userId primitive.ObjectID, projectId primitive.ObjectID) (*mongo.UpdateResult, error) {
	result, err := r.collection.UpdateOne(
		c,
		bson.M{"_id": userId},
		bson.D{
			{"$pull", bson.M{"codecarbon_projects": bson.M{"_id": projectId}}},
		},
	)
	return result, err
}
