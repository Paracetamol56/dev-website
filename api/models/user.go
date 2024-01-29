package models

import (
	"time"

	"github.com/Paracetamol56/dev-website/api/db"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type GitHubUser struct {
	Login             string `json:"login" bson:"login"`
	Id                int    `json:"id" bson:"id"`
	NodeId            string `json:"node_id" bson:"node_id"`
	AvatarUrl         string `json:"avatar_url" bson:"avatar_url"`
	GravatarId        string `json:"gravatar_id" bson:"gravatar_id"`
	Url               string `json:"url" bson:"url"`
	HtmlUrl           string `json:"html_url" bson:"html_url"`
	FollowersUrl      string `json:"followers_url" bson:"followers_url"`
	FollowingUrl      string `json:"following_url" bson:"following_url"`
	GistsUrl          string `json:"gists_url" bson:"gists_url"`
	StarredUrl        string `json:"starred_url" bson:"starred_url"`
	SubscriptionsUrl  string `json:"subscriptions_url" bson:"subscriptions_url"`
	OrganizationsUrl  string `json:"organizations_url" bson:"organizations_url"`
	ReposUrl          string `json:"repos_url" bson:"repos_url"`
	EventsUrl         string `json:"events_url" bson:"events_url"`
	ReceivedEventsUrl string `json:"received_events_url" bson:"received_events_url"`
	Type              string `json:"type" bson:"type"`
	SiteAdmin         bool   `json:"site_admin" bson:"site_admin"`
	Name              string `json:"name" bson:"name"`
	Company           string `json:"company" bson:"company"`
	Blog              string `json:"blog" bson:"blog"`
	Location          string `json:"location" bson:"location"`
	Email             string `json:"email" bson:"email"`
	Hireable          bool   `json:"hireable" bson:"hireable"`
	Bio               string `json:"bio" bson:"bio"`
	TwitterUsername   string `json:"twitter_username" bson:"twitter_username"`
	PublicRepos       int    `json:"public_repos" bson:"public_repos"`
	PublicGists       int    `json:"public_gists" bson:"public_gists"`
	Followers         int    `json:"followers" bson:"followers"`
	Following         int    `json:"following" bson:"following"`
	CreatedAt         string `json:"created_at" bson:"created_at"`
	UpdatedAt         string `json:"updated_at" bson:"updated_at"`
}

type User struct {
	Id             primitive.ObjectID `json:"id" bson:"_id"`
	Name           string             `json:"name" bson:"name"`
	ProfilePicture string             `json:"profilePicture" bson:"profilePicture"`
}

type FullUser struct {
	Id                primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name              string             `json:"name" bson:"name"`
	Email             string             `json:"email" bson:"email"`
	CreatedAt         time.Time          `json:"createdAt" bson:"createdAt,omitempty"`
	LastLogin         time.Time          `json:"lastLogin" bson:"lastLogin,omitempty"`
	LastRefresh       time.Time          `json:"lastRefresh" bson:"lastRefresh,omitempty"`
	Flavour           string             `json:"flavour" bson:"flavour"`
	ProfilePicture    string             `json:"profilePicture,omitempty" bson:"profilePicture,omitempty"`
	GitHubAccessToken string             `json:"githubAccessToken,omitempty" bson:"githubAccessToken,omitempty"`
	Github            *GitHubUser        `json:"github,omitempty" bson:"github,omitempty"`
}

func CreateUser(c *gin.Context, user *FullUser) (*mongo.InsertOneResult, error) {
	db := db.GetDB()
	collection := db.Collection("users")
	result, err := collection.InsertOne(c, user)
	return result, err
}

func GetUserById(c *gin.Context, id primitive.ObjectID) (*User, error) {
	db := db.GetDB()
	collection := db.Collection("users")
	var user User
	if err := collection.FindOne(c, bson.M{"_id": id, "deletedAt": bson.M{"$exists": false}}).Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

func GetFullUserById(c *gin.Context, id primitive.ObjectID) (*FullUser, error) {
	db := db.GetDB()
	collection := db.Collection("users")
	var user FullUser
	if err := collection.FindOne(c, bson.M{"_id": id, "deletedAt": bson.M{"$exists": false}}).Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

func GetFullUserByEmail(c *gin.Context, email string) (*FullUser, error) {
	db := db.GetDB()
	collection := db.Collection("users")
	var user FullUser
	if err := collection.FindOne(c, bson.M{"email": email, "deletedAt": bson.M{"$exists": false}}).Decode(&user); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func UpdateUser(c *gin.Context, id primitive.ObjectID, user *FullUser) (*mongo.UpdateResult, error) {
	db := db.GetDB()
	collection := db.Collection("users")
	result, err := collection.UpdateOne(c, bson.M{"_id": id}, bson.M{"$set": bson.M{
		"name":              user.Name,
		"email":             user.Email,
		"flavour":           user.Flavour,
		"profilePicture":    user.ProfilePicture,
		"lastLogin":         user.LastLogin,
		"lastRefresh":       user.LastRefresh,
		"githubAccessToken": user.GitHubAccessToken,
		"github":            user.Github,
	}})
	return result, err
}

func UpdateGithubUser(c *gin.Context, id primitive.ObjectID, githubUser *GitHubUser) (*mongo.UpdateResult, error) {
	db := db.GetDB()
	collection := db.Collection("users")
	result, err := collection.UpdateOne(c, bson.M{"_id": id}, bson.M{"$set": bson.M{
		"github": githubUser,
	}})
	return result, err
}

func DeleteUser(c *gin.Context, id primitive.ObjectID) (*mongo.UpdateResult, error) {
	db := db.GetDB()
	collection := db.Collection("users")
	result, err := collection.UpdateOne(c, bson.M{"_id": id}, bson.M{"$set": bson.M{
		"deletedAt": bson.M{"$exists": true},
	},
	})
	return result, err
}
