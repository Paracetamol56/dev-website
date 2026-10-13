package contact

import (
	"dev/internal/db"
	"dev/internal/testutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestPostContact(t *testing.T) {
	testutil.Clean(t)
	internet := testutil.FakeInternet(t)
	testutil.FakeBrevo(internet)
	userId := primitive.NewObjectID()
	body := map[string]any{"userId": userId, "name": "Ada Lovelace", "email": "ada@example.test", "message": "I found a bug in the engine."}

	response := do(t, http.MethodPost, "/contact", body)
	assert.Equal(t, http.StatusCreated, response.Code)
	assert.JSONEq(t, `{"message": "Message sent"}`, response.Body.String())

	emails := testutil.SentEmails(t, internet)
	require.Len(t, emails, 1)
	assert.Equal(t, testutil.AdminEmail, emails[0].To[0].Email, "the message goes to the admin")
	assert.Equal(t, "ada@example.test", emails[0].ReplyTo.Email)
	assert.Equal(t, "Ada Lovelace", emails[0].ReplyTo.Name)
	assert.Contains(t, emails[0].TextContent, "I found a bug in the engine.")
	assert.Contains(t, emails[0].TextContent, userId.Hex())
	assert.Empty(t, emails[0].HTMLContent, "user input is never sent as HTML")

	var stored Contact
	require.NoError(t, db.GetDB().Collection("contacts").FindOne(ctx, bson.M{}).Decode(&stored))
	assert.Equal(t, userId, stored.UserId)
	assert.Equal(t, "Ada Lovelace", stored.Name)
	assert.Equal(t, "I found a bug in the engine.", stored.Message)
	assert.Equal(t, "<message@brevo>", stored.MailId)

	anonymous := map[string]any{"name": "Ada", "email": "ada@example.test", "message": "No account needed."}
	assert.Equal(t, http.StatusCreated, do(t, http.MethodPost, "/contact", anonymous).Code)
}

func TestPostContactRejections(t *testing.T) {
	testutil.Clean(t)
	internet := testutil.FakeInternet(t)
	testutil.FakeBrevo(internet)
	valid := func() map[string]any {
		return map[string]any{"name": "Ada", "email": "ada@example.test", "message": "Long enough message"}
	}

	for name, change := range map[string]func(map[string]any){
		"no name":           func(body map[string]any) { delete(body, "name") },
		"short name":        func(body map[string]any) { body["name"] = "A" },
		"long name":         func(body map[string]any) { body["name"] = strings.Repeat("a", 101) },
		"invalid email":     func(body map[string]any) { body["email"] = "ada@" },
		"no message":        func(body map[string]any) { delete(body, "message") },
		"short message":     func(body map[string]any) { body["message"] = "Too short" },
		"long message":      func(body map[string]any) { body["message"] = strings.Repeat("a", 1001) },
		"user id not an id": func(body map[string]any) { body["userId"] = "me" },
	} {
		body := valid()
		change(body)
		assert.Equal(t, http.StatusBadRequest, do(t, http.MethodPost, "/contact", body).Code, name)
	}
	assert.Equal(t, http.StatusBadRequest, do(t, http.MethodPost, "/contact", "not json").Code)
	assert.Empty(t, testutil.SentEmails(t, internet))
	assert.Zero(t, testutil.Count(t, "contacts"))

	internet.JSON(testutil.BrevoHost, http.StatusInternalServerError, `{}`)
	failed := do(t, http.MethodPost, "/contact", valid())
	assert.Equal(t, http.StatusInternalServerError, failed.Code)
	assert.Equal(t, "Error while sending email", testutil.ErrorOf(t, failed))
	assert.Zero(t, testutil.Count(t, "contacts"), "a message that was not sent is not stored")

	testutil.FakeBrevo(internet)
	testutil.BreakDB(t)
	assert.Equal(t, http.StatusInternalServerError, do(t, http.MethodPost, "/contact", valid()).Code)
}
