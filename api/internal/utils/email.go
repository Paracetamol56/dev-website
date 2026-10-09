package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

const brevoURL = "https://api.brevo.com/v3"

var brevoClient = &http.Client{Timeout: 10 * time.Second}

type EmailAddress struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email"`
}

type Email struct {
	To          EmailAddress
	ReplyTo     *EmailAddress
	Subject     string
	HTMLContent string
	TextContent string
}

func brevoRequest(ctx context.Context, path string, payload any) ([]byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, brevoURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("api-key", os.Getenv("BREVO_API_KEY"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	response, err := brevoClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	responseBody, _ := io.ReadAll(response.Body)
	if response.StatusCode < 200 || response.StatusCode > 299 {
		return nil, fmt.Errorf("brevo API error: status %d, body: %s", response.StatusCode, responseBody)
	}
	return responseBody, nil
}

func SendEmail(ctx context.Context, email Email) (string, error) {
	payload := map[string]any{
		"sender":  EmailAddress{Name: "Matheo Galuba", Email: os.Getenv("ADMIN_EMAIL")},
		"to":      []EmailAddress{email.To},
		"subject": email.Subject,
	}
	if email.ReplyTo != nil {
		payload["replyTo"] = email.ReplyTo
	}
	if email.HTMLContent != "" {
		payload["htmlContent"] = email.HTMLContent
	}
	if email.TextContent != "" {
		payload["textContent"] = email.TextContent
	}

	responseBody, err := brevoRequest(ctx, "/smtp/email", payload)
	if err != nil {
		log.Printf("Failed to send email %q to %s: %v", email.Subject, email.To.Email, err)
		return "", err
	}

	var result struct {
		MessageId string `json:"messageId"`
	}
	json.Unmarshal(responseBody, &result)
	log.Printf("Email %q sent to %s (%s)", email.Subject, email.To.Email, result.MessageId)
	return result.MessageId, nil
}

func AddEmailContact(ctx context.Context, address string) error {
	listId, err := strconv.Atoi(os.Getenv("BREVO_CONTACT_LIST_ID"))
	if err != nil {
		return fmt.Errorf("invalid BREVO_CONTACT_LIST_ID: %w", err)
	}

	_, err = brevoRequest(ctx, "/contacts", map[string]any{
		"email":         address,
		"listIds":       []int{listId},
		"updateEnabled": true,
	})
	return err
}
