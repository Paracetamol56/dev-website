package models

import (
	"context"
	"time"

	"github.com/Paracetamol56/dev-website/api/db"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// GitHubUser represents a user from GitHub.
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

// UserLight represents a lightweight version of a user.
type UserLight struct {
	Id             primitive.ObjectID `json:"id" bson:"_id"`
	Name           string             `json:"name" bson:"name"`
	ProfilePicture string             `json:"profilePicture" bson:"profilePicture"`
}

// User represents a user entity.
type User struct {
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

// CreateUser creates a new user in the database.
// It takes a gin.Context object and a pointer to a User struct as parameters.
// It returns the result of the insertion operation (*mongo.InsertOneResult) and any error encountered.
func CreateUser(c *gin.Context, user *User) (*mongo.InsertOneResult, error) {
	db := db.GetDB()
	collection := db.Collection("users")
	result, err := collection.InsertOne(c, user)
	return result, err
}

// GetUserById retrieves a user by their ID from the database.
// It takes a gin.Context object and the ID of the user as parameters.
// It returns a pointer to a UserLight struct and an error.
// The UserLight struct represents a simplified version of the user model.
// If the user is found, the function returns the user object.
// If the user is not found or an error occurs, it returns nil and the error.
func GetUserById(c *gin.Context, id primitive.ObjectID) (*UserLight, error) {
	db := db.GetDB()
	collection := db.Collection("users")
	var user UserLight
	if err := collection.FindOne(c, bson.M{"_id": id, "deletedAt": bson.M{"$exists": false}}).Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

// GetFullUserById retrieves a full user by their ID.
// It takes a gin.Context object and an ID of type primitive.ObjectID as parameters.
// It returns a pointer to a User struct and an error.
// The function queries the database to find a user with the specified ID and no "deletedAt" field.
// If the user is found, it is decoded into the user variable and returned.
// If an error occurs during the query or decoding, the function returns nil and the error.
func GetFullUserById(c *gin.Context, id primitive.ObjectID) (*User, error) {
	db := db.GetDB()
	collection := db.Collection("users")
	var user User
	if err := collection.FindOne(c, bson.M{"_id": id, "deletedAt": bson.M{"$exists": false}}).Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

// GetFullUserByEmail retrieves the full user information by email.
// It takes a gin.Context and an email string as parameters.
// It returns a pointer to a User struct and an error.
// If the user is not found, it returns nil, nil.
// If an error occurs during the retrieval, it returns nil and the error.
func GetFullUserByEmail(c *gin.Context, email string) (*User, error) {
	db := db.GetDB()
	collection := db.Collection("users")
	var user User
	if err := collection.FindOne(c, bson.M{"email": email, "deletedAt": bson.M{"$exists": false}}).Decode(&user); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// UpdateUser updates a user in the database with the specified ID.
// It takes a gin.Context, an ID of type primitive.ObjectID, and a user object as parameters.
// It returns a pointer to mongo.UpdateResult and an error.
// The function updates the user's name, email, flavour, profile picture, last login, last refresh,
// GitHub access token, and GitHub information in the database.
func UpdateUser(c *gin.Context, id primitive.ObjectID, user *User) (*mongo.UpdateResult, error) {
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

// UpdateGithubUser updates the GitHub user information in the database.
// It takes a gin.Context, an ObjectID representing the user ID, and a GitHubUser struct as parameters.
// It returns a pointer to mongo.UpdateResult and an error.
// The function updates the "github" field of the user document with the provided GitHubUser struct.
func DeleteUser(c *gin.Context, id primitive.ObjectID) (*mongo.UpdateResult, error) {
	db := db.GetDB()
	collection := db.Collection("users")
	result, err := collection.UpdateOne(c, bson.M{"_id": id}, bson.M{"$set": bson.M{
		"deletedAt": time.Now(),
	},
	})
	return result, err
}

// DeleteOldUsers deletes old users from the database.
// It takes a context as input and returns the delete result and an error, if any.
// The function queries the "users" collection in the database and deletes users whose "deletedAt" field is older than 30 days ago.
func DeleteOldUsers(c context.Context) (*mongo.DeleteResult, error) {
	db := db.GetDB()
	collection := db.Collection("users")
	result, err := collection.DeleteMany(c, bson.M{"deletedAt": bson.M{"$lt": time.Now().AddDate(0, 0, -30)}})
	return result, err
}
