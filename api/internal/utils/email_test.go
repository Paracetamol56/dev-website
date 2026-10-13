package utils

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"dev/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const brevoHost = "api.brevo.com"

func TestSendEmail(t *testing.T) {
	internet := testutil.FakeInternet(t)
	internet.JSON(brevoHost, http.StatusCreated, `{"messageId":"<id-1@brevo>"}`)

	messageId, err := SendEmail(context.Background(), Email{
		To:          EmailAddress{Name: "Ada", Email: "ada@example.test"},
		ReplyTo:     &EmailAddress{Email: "reply@example.test"},
		Subject:     "Hello",
		HTMLContent: "<p>Hi</p>",
		TextContent: "Hi",
	})
	require.NoError(t, err)
	assert.Equal(t, "<id-1@brevo>", messageId)

	requests := internet.Requests(brevoHost)
	require.Len(t, requests, 1)
	assert.Equal(t, http.MethodPost, requests[0].Method)
	assert.Equal(t, "/v3/smtp/email", requests[0].Path)
	assert.Equal(t, "test-brevo-key", requests[0].Header.Get("api-key"))
	assert.JSONEq(t, `{
		"sender": {"name": "Matheo Galuba", "email": "admin@example.test"},
		"to": [{"name": "Ada", "email": "ada@example.test"}],
		"replyTo": {"email": "reply@example.test"},
		"subject": "Hello",
		"htmlContent": "<p>Hi</p>",
		"textContent": "Hi"
	}`, requests[0].Body)
}

func TestSendEmailOmitsEmptyParts(t *testing.T) {
	internet := testutil.FakeInternet(t)
	internet.JSON(brevoHost, http.StatusCreated, `{}`)

	messageId, err := SendEmail(context.Background(), Email{To: EmailAddress{Email: "ada@example.test"}, Subject: "Hello"})
	require.NoError(t, err)
	assert.Empty(t, messageId)

	var payload map[string]any
	require.NoError(t, json.Unmarshal([]byte(internet.Requests(brevoHost)[0].Body), &payload))
	assert.NotContains(t, payload, "replyTo")
	assert.NotContains(t, payload, "htmlContent")
	assert.NotContains(t, payload, "textContent")
}

func TestSendEmailFailures(t *testing.T) {
	internet := testutil.FakeInternet(t)
	internet.JSON(brevoHost, http.StatusUnauthorized, `{"message":"Key not found"}`)

	_, err := SendEmail(context.Background(), Email{To: EmailAddress{Email: "ada@example.test"}})
	assert.ErrorContains(t, err, "status 401")
	assert.ErrorContains(t, err, "Key not found")

	testutil.FakeInternet(t)
	_, err = SendEmail(context.Background(), Email{To: EmailAddress{Email: "ada@example.test"}})
	assert.ErrorContains(t, err, "unexpected request", "network failure")

	_, err = brevoRequest(context.Background(), "/x", make(chan int))
	assert.Error(t, err, "payload that cannot be encoded")
	_, err = brevoRequest(nil, "/x", nil)
	assert.Error(t, err, "nil context")
}

func TestAddEmailContact(t *testing.T) {
	internet := testutil.FakeInternet(t)
	internet.JSON(brevoHost, http.StatusNoContent, ``)

	require.NoError(t, AddEmailContact(context.Background(), "ada@example.test"))
	requests := internet.Requests(brevoHost)
	require.Len(t, requests, 1)
	assert.Equal(t, "/v3/contacts", requests[0].Path)
	assert.JSONEq(t, `{"email": "ada@example.test", "listIds": [7], "updateEnabled": true}`, requests[0].Body)

	internet.JSON(brevoHost, http.StatusBadRequest, `{}`)
	assert.ErrorContains(t, AddEmailContact(context.Background(), "ada@example.test"), "status 400")

	t.Setenv("BREVO_CONTACT_LIST_ID", "")
	assert.ErrorContains(t, AddEmailContact(context.Background(), "ada@example.test"), "invalid BREVO_CONTACT_LIST_ID")
	assert.Len(t, internet.Requests(brevoHost), 2, "no request without a list")
}
