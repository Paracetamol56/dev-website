package models

import (
	"context"
	"dev/internal/db"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	IdentityProviderEmail  = "email"
	IdentityProviderGithub = "github"
	IdentityProviderGoogle = "google"
)

type Identity struct {
	Provider    string    `json:"provider" bson:"provider" enums:"email,github,google"`
	ProviderId  string    `json:"providerId" bson:"providerId"`
	Email       string    `json:"email" bson:"email"`
	Username    string    `json:"username,omitempty" bson:"username,omitempty"`
	Name        string    `json:"name,omitempty" bson:"name,omitempty"`
	AvatarUrl   string    `json:"avatarUrl,omitempty" bson:"avatarUrl,omitempty"`
	ProfileUrl  string    `json:"profileUrl,omitempty" bson:"profileUrl,omitempty"`
	AccessToken string    `json:"-" bson:"accessToken,omitempty"`
	LinkedAt    time.Time `json:"linkedAt" bson:"linkedAt"`
	LastLogin   time.Time `json:"lastLogin" bson:"lastLogin"`
}

// UserLight represents a lightweight version of a user.
type UserLight struct {
	Id             primitive.ObjectID `json:"id" bson:"_id"`
	Name           string             `json:"name" bson:"name"`
	ProfilePicture string             `json:"profilePicture" bson:"profilePicture"`
}

// User represents a user entity.
type User struct {
	Id             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name           string             `json:"name" bson:"name"`
	Email          string             `json:"email" bson:"email"`
	CreatedAt      time.Time          `json:"createdAt" bson:"createdAt,omitempty"`
	LastLogin      time.Time          `json:"lastLogin" bson:"lastLogin,omitempty"`
	LastRefresh    time.Time          `json:"lastRefresh" bson:"lastRefresh,omitempty"`
	Flavour        string             `json:"flavour" bson:"flavour"`
	ProfilePicture string             `json:"profilePicture,omitempty" bson:"profilePicture,omitempty"`
	Identities     []Identity         `json:"identities" bson:"identities"`
	Passkeys       []Passkey          `json:"passkeys" bson:"passkeys"`
}

// Users created before createdAt was stored fall back to the timestamp embedded in their id
func (user *User) fillCreatedAt() {
	if user.CreatedAt.IsZero() {
		user.CreatedAt = user.Id.Timestamp()
	}
}

func (user *User) GetIdentity(provider string) *Identity {
	for i := range user.Identities {
		if user.Identities[i].Provider == provider {
			return &user.Identities[i]
		}
	}
	return nil
}

func (user *User) SetIdentity(identity Identity) {
	identity.LastLogin = time.Now()
	identity.LinkedAt = identity.LastLogin
	if existing := user.GetIdentity(identity.Provider); existing != nil {
		identity.LinkedAt = existing.LinkedAt
		*existing = identity
		return
	}
	user.Identities = append(user.Identities, identity)
}

func (user *User) RemoveIdentity(provider string) bool {
	for i := range user.Identities {
		if user.Identities[i].Provider == provider {
			user.Identities = append(user.Identities[:i], user.Identities[i+1:]...)
			return true
		}
	}
	return false
}

// CreateUser creates a new user in the database.
// It takes a gin.Context object and a pointer to a User struct as parameters.
// It returns the result of the insertion operation (*mongo.InsertOneResult) and any error encountered.
func CreateUser(c *gin.Context, user *User) (*mongo.InsertOneResult, error) {
	db := db.GetDB()
	collection := db.Collection("users")
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}
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
	user.fillCreatedAt()
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
	caseInsensitive := options.FindOne().SetCollation(&options.Collation{Locale: "en", Strength: 2})
	if err := collection.FindOne(c, bson.M{"email": email, "deletedAt": bson.M{"$exists": false}}, caseInsensitive).Decode(&user); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	user.fillCreatedAt()
	return &user, nil
}

// UpdateUser updates a user in the database with the specified ID.
// It takes a gin.Context, an ID of type primitive.ObjectID, and a user object as parameters.
// It returns a pointer to mongo.UpdateResult and an error.
// The function updates the user's name, email, flavour, profile picture, last login, last refresh
// and identities in the database.
func UpdateUser(c *gin.Context, id primitive.ObjectID, user *User) (*mongo.UpdateResult, error) {
	db := db.GetDB()
	collection := db.Collection("users")
	result, err := collection.UpdateOne(c, bson.M{"_id": id}, bson.M{"$set": bson.M{
		"name":           user.Name,
		"email":          user.Email,
		"flavour":        user.Flavour,
		"profilePicture": user.ProfilePicture,
		"lastLogin":      user.LastLogin,
		"lastRefresh":    user.LastRefresh,
		"identities":     user.Identities,
		"passkeys":       user.Passkeys,
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
	result, err := collection.UpdateOne(c, bson.M{"_id": id}, bson.M{
		"$set": bson.M{
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

// DetachIdentity unlinks a provider account from every user but the given one, so that an
// account whose email changed on the provider side only stays linked to its current owner.
func DetachIdentity(ctx context.Context, provider string, providerId string, exceptUserId primitive.ObjectID) error {
	_, err := db.GetDB().Collection("users").UpdateMany(ctx,
		bson.M{"_id": bson.M{"$ne": exceptUserId}},
		bson.M{"$pull": bson.M{"identities": bson.M{"provider": provider, "providerId": providerId}}},
	)
	return err
}

func MigrateUserIdentities(ctx context.Context) (int, error) {
	collection := db.GetDB().Collection("users")

	cursor, err := collection.Find(ctx, bson.M{"github": bson.M{"$exists": true}})
	if err != nil {
		return 0, err
	}
	var legacyUsers []struct {
		Id          primitive.ObjectID `bson:"_id"`
		Email       string             `bson:"email"`
		LastLogin   time.Time          `bson:"lastLogin"`
		AccessToken string             `bson:"githubAccessToken"`
		Identities  []Identity         `bson:"identities"`
		Github      *struct {
			Id        int    `bson:"id"`
			Login     string `bson:"login"`
			Name      string `bson:"name"`
			AvatarUrl string `bson:"avatar_url"`
			HtmlUrl   string `bson:"html_url"`
		} `bson:"github"`
	}
	if err = cursor.All(ctx, &legacyUsers); err != nil {
		return 0, err
	}

	for _, legacy := range legacyUsers {
		user := User{Identities: legacy.Identities}
		if legacy.Github != nil && user.GetIdentity(IdentityProviderGithub) == nil {
			user.Identities = append(user.Identities, Identity{
				Provider:    IdentityProviderGithub,
				ProviderId:  strconv.Itoa(legacy.Github.Id),
				Email:       legacy.Email,
				Username:    legacy.Github.Login,
				Name:        legacy.Github.Name,
				AvatarUrl:   legacy.Github.AvatarUrl,
				ProfileUrl:  legacy.Github.HtmlUrl,
				AccessToken: legacy.AccessToken,
				LinkedAt:    legacy.LastLogin,
				LastLogin:   legacy.LastLogin,
			})
		}
		if user.Identities == nil {
			user.Identities = []Identity{}
		}
		if _, err := collection.UpdateOne(ctx, bson.M{"_id": legacy.Id}, bson.M{
			"$set":   bson.M{"identities": user.Identities},
			"$unset": bson.M{"github": "", "githubAccessToken": ""},
		}); err != nil {
			return 0, err
		}
	}
	return len(legacyUsers), nil
}
