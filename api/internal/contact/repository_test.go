package contact

import (
	"dev/internal/db"
	"dev/internal/testutil"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestCreateContact(t *testing.T) {
	testutil.Clean(t)
	contact := Contact{
		UserId: primitive.NewObjectID(), Name: "Ada", Email: "ada@example.test", Message: "Hello there",
		MailId: "<id@brevo>", CreatedAt: primitive.NewDateTimeFromTime(time.Now()),
	}

	id, err := CreateContact(testutil.Context(), contact)
	require.NoError(t, err)

	var stored Contact
	require.NoError(t, db.GetDB().Collection("contacts").FindOne(ctx, bson.M{"_id": id}).Decode(&stored))
	contact.Id = *id
	assert.Equal(t, contact, stored)

	testutil.BreakDB(t)
	_, err = CreateContact(testutil.Context(), contact)
	assert.Error(t, err)
}
