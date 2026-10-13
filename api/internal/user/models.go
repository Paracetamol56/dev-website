package user

import (
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const MaxPasskeys = 10

type Passkey struct {
	Id         string     `json:"id" bson:"id"`
	Name       string     `json:"name" bson:"name"`
	CreatedAt  time.Time  `json:"createdAt" bson:"createdAt"`
	LastUsed   *time.Time `json:"lastUsed,omitempty" bson:"lastUsed,omitempty"`
	Credential []byte     `json:"-" bson:"credential"`
}

func PasskeyId(credential *webauthn.Credential) string {
	return base64.RawURLEncoding.EncodeToString(credential.ID)
}

func NewPasskey(name string, credential *webauthn.Credential) (*Passkey, error) {
	encoded, err := json.Marshal(credential)
	if err != nil {
		return nil, err
	}
	return &Passkey{
		Id:         PasskeyId(credential),
		Name:       name,
		CreatedAt:  time.Now(),
		Credential: encoded,
	}, nil
}

func (user *User) WebAuthnID() []byte {
	return user.Id[:]
}

func (user *User) WebAuthnName() string {
	return user.Email
}

func (user *User) WebAuthnDisplayName() string {
	if user.Name != "" {
		return user.Name
	}
	return user.Email
}

func (user *User) WebAuthnCredentials() []webauthn.Credential {
	credentials := []webauthn.Credential{}
	for _, passkey := range user.Passkeys {
		var credential webauthn.Credential
		if err := json.Unmarshal(passkey.Credential, &credential); err == nil {
			credentials = append(credentials, credential)
		}
	}
	return credentials
}

func (user *User) GetPasskey(id string) *Passkey {
	for i := range user.Passkeys {
		if user.Passkeys[i].Id == id {
			return &user.Passkeys[i]
		}
	}
	return nil
}

func (user *User) RemovePasskey(id string) bool {
	for i := range user.Passkeys {
		if user.Passkeys[i].Id == id {
			user.Passkeys = append(user.Passkeys[:i], user.Passkeys[i+1:]...)
			return true
		}
	}
	return false
}

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
	Role           string             `json:"role,omitempty" bson:"role,omitempty" enums:"admin"`
}

const RoleAdmin = "admin"

func (user *User) IsAdmin() bool {
	return user.Role == RoleAdmin
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

type PatchUserBody struct {
	Name           string `json:"name" binding:"omitempty,min=2,max=100"`
	Flavour        string `json:"flavour" binding:"omitempty,oneof=latte frappe macchiato mocha"`
	ProfilePicture string `json:"profilePicture" binding:"omitempty,url"`
}
