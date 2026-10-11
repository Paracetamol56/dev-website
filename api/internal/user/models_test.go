package user

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestPasskeys(t *testing.T) {
	credential := &webauthn.Credential{ID: []byte{0xfb, 0xff, 0x01}, PublicKey: []byte("key")}
	assert.Equal(t, "-_8B", PasskeyId(credential), "URL-safe, unpadded")

	passkey, err := NewPasskey("Laptop", credential)
	require.NoError(t, err)
	assert.Equal(t, "-_8B", passkey.Id)
	assert.Equal(t, "Laptop", passkey.Name)
	assert.WithinDuration(t, time.Now(), passkey.CreatedAt, time.Second)
	assert.Nil(t, passkey.LastUsed)

	encoded, err := json.Marshal(passkey)
	require.NoError(t, err)
	assert.NotContains(t, string(encoded), "credential", "the credential is never sent to clients")

	user := &User{
		Id:       primitive.NewObjectID(),
		Email:    "ada@example.test",
		Passkeys: []Passkey{*passkey, {Id: "corrupt", Credential: []byte("{")}, {Id: "other", Credential: []byte(`{"id":"AQID"}`)}},
	}
	assert.Equal(t, user.Id[:], user.WebAuthnID())
	assert.Equal(t, "ada@example.test", user.WebAuthnName())
	assert.Equal(t, "ada@example.test", user.WebAuthnDisplayName())
	user.Name = "Ada"
	assert.Equal(t, "Ada", user.WebAuthnDisplayName())

	credentials := user.WebAuthnCredentials()
	require.Len(t, credentials, 2, "unreadable credentials are skipped")
	assert.Equal(t, credential.ID, credentials[0].ID)
	assert.Equal(t, []byte{1, 2, 3}, credentials[1].ID)
	assert.Empty(t, (&User{}).WebAuthnCredentials())

	assert.Same(t, &user.Passkeys[2], user.GetPasskey("other"))
	assert.Nil(t, user.GetPasskey("unknown"))
	assert.False(t, user.RemovePasskey("unknown"))
	assert.True(t, user.RemovePasskey("corrupt"))
	assert.Len(t, user.Passkeys, 2)
	assert.Nil(t, user.GetPasskey("corrupt"))
}

func TestIdentities(t *testing.T) {
	user := User{}
	assert.Nil(t, user.GetIdentity("github"))
	assert.False(t, user.RemoveIdentity("github"))

	user.SetIdentity(Identity{Provider: "email", Email: "ada@example.test"})
	user.SetIdentity(Identity{Provider: "github", ProviderId: "42", Username: "ada"})
	require.Len(t, user.Identities, 2)
	github := user.GetIdentity("github")
	require.NotNil(t, github)
	assert.WithinDuration(t, time.Now(), github.LinkedAt, time.Second)
	assert.Equal(t, github.LinkedAt, github.LastLogin)

	linkedAt := time.Now().Add(-48 * time.Hour)
	github.LinkedAt = linkedAt
	user.SetIdentity(Identity{Provider: "github", ProviderId: "42", Username: "countess"})
	require.Len(t, user.Identities, 2, "an identity is replaced, not duplicated")
	github = user.GetIdentity("github")
	assert.Equal(t, "countess", github.Username)
	assert.Equal(t, linkedAt, github.LinkedAt, "the link date survives a new login")
	assert.WithinDuration(t, time.Now(), github.LastLogin, time.Second)

	assert.True(t, user.RemoveIdentity("email"))
	assert.Nil(t, user.GetIdentity("email"))
	assert.Len(t, user.Identities, 1)
}

func normalized(identity Identity) Identity {
	identity.LinkedAt = identity.LinkedAt.UTC()
	identity.LastLogin = identity.LastLogin.UTC()
	return identity
}
