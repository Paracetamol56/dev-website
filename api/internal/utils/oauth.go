package utils

import (
	"context"
	"dev/internal/models"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

var oauthClient = &http.Client{Timeout: 10 * time.Second}

type OAuthProvider struct {
	Name         string
	AuthorizeURL string
	Scope        string

	tokenURL        string
	clientIdEnv     string
	clientSecretEnv string
	fetchIdentity   func(ctx context.Context, accessToken string) (*models.Identity, error)
}

var oauthProviders = []*OAuthProvider{
	{
		Name:            models.IdentityProviderGithub,
		AuthorizeURL:    "https://github.com/login/oauth/authorize",
		Scope:           "read:user user:email",
		tokenURL:        "https://github.com/login/oauth/access_token",
		clientIdEnv:     "GITHUB_CLIENT_ID",
		clientSecretEnv: "GITHUB_CLIENT_SECRET",
		fetchIdentity:   fetchGithubIdentity,
	},
	{
		Name:            models.IdentityProviderGoogle,
		AuthorizeURL:    "https://accounts.google.com/o/oauth2/v2/auth",
		Scope:           "openid email profile",
		tokenURL:        "https://oauth2.googleapis.com/token",
		clientIdEnv:     "GOOGLE_CLIENT_ID",
		clientSecretEnv: "GOOGLE_CLIENT_SECRET",
		fetchIdentity:   fetchGoogleIdentity,
	},
}

func (provider *OAuthProvider) ClientId() string {
	return os.Getenv(provider.clientIdEnv)
}

func (provider *OAuthProvider) enabled() bool {
	return provider.ClientId() != "" && os.Getenv(provider.clientSecretEnv) != ""
}

func GetOAuthProviders() []*OAuthProvider {
	enabled := []*OAuthProvider{}
	for _, provider := range oauthProviders {
		if provider.enabled() {
			enabled = append(enabled, provider)
		}
	}
	return enabled
}

func GetOAuthProvider(name string) *OAuthProvider {
	for _, provider := range GetOAuthProviders() {
		if provider.Name == name {
			return provider
		}
	}
	return nil
}

func (provider *OAuthProvider) Authenticate(ctx context.Context, code string, redirectURI string) (*models.Identity, error) {
	data := url.Values{}
	data.Set("client_id", provider.ClientId())
	data.Set("client_secret", os.Getenv(provider.clientSecretEnv))
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")
	if redirectURI != "" {
		data.Set("redirect_uri", redirectURI)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, provider.tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	response, err := oauthClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var token struct {
		AccessToken      string `json:"access_token"`
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
	}
	if err := json.NewDecoder(response.Body).Decode(&token); err != nil {
		return nil, err
	}
	if token.AccessToken == "" {
		return nil, fmt.Errorf("%s refused the authorization code: %s %s", provider.Name, token.Error, token.ErrorDescription)
	}

	identity, err := provider.fetchIdentity(ctx, token.AccessToken)
	if err != nil {
		return nil, err
	}
	identity.Provider = provider.Name
	identity.AccessToken = token.AccessToken
	return identity, nil
}

func getOAuthResource(ctx context.Context, url string, accessToken string, v any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	response, err := oauthClient.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s: status %d", url, response.StatusCode)
	}
	return json.NewDecoder(response.Body).Decode(v)
}

func fetchGithubIdentity(ctx context.Context, accessToken string) (*models.Identity, error) {
	var profile struct {
		Id        int    `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		AvatarUrl string `json:"avatar_url"`
		HtmlUrl   string `json:"html_url"`
	}
	if err := getOAuthResource(ctx, "https://api.github.com/user", accessToken, &profile); err != nil {
		return nil, err
	}

	// The email of the public profile is optional and unverified
	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := getOAuthResource(ctx, "https://api.github.com/user/emails", accessToken, &emails); err != nil {
		return nil, err
	}
	for _, email := range emails {
		if email.Primary && email.Verified {
			return &models.Identity{
				ProviderId: strconv.Itoa(profile.Id),
				Email:      email.Email,
				Username:   profile.Login,
				Name:       profile.Name,
				AvatarUrl:  profile.AvatarUrl,
				ProfileUrl: profile.HtmlUrl,
			}, nil
		}
	}
	return nil, errors.New("this GitHub account has no verified primary email")
}

func fetchGoogleIdentity(ctx context.Context, accessToken string) (*models.Identity, error) {
	var profile struct {
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}
	if err := getOAuthResource(ctx, "https://openidconnect.googleapis.com/v1/userinfo", accessToken, &profile); err != nil {
		return nil, err
	}
	if profile.Email == "" || !profile.EmailVerified {
		return nil, errors.New("this Google account has no verified email")
	}

	return &models.Identity{
		ProviderId: profile.Sub,
		Email:      profile.Email,
		Name:       profile.Name,
		AvatarUrl:  profile.Picture,
	}, nil
}
