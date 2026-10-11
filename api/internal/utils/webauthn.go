package utils

import (
	"errors"
	"net/url"
	"slices"

	"github.com/go-webauthn/webauthn/webauthn"
)

var AllowedOrigins = []string{"http://localhost:8000", "http://localhost:5173", "https://dev.matheo-galuba.com", "https://dev-uat.matheo-galuba.com"}

// A passkey is bound to the domain it was created on, hence one relying party per origin
func NewWebAuthn(origin string) (*webauthn.WebAuthn, error) {
	if !slices.Contains(AllowedOrigins, origin) {
		return nil, errors.New("origin not allowed")
	}
	parsed, err := url.Parse(origin)
	if err != nil {
		return nil, err
	}

	return webauthn.New(&webauthn.Config{
		RPID:          parsed.Hostname(),
		RPDisplayName: "Mathéo Galuba - Dev",
		RPOrigins:     []string{origin},
	})
}
