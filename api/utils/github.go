package utils

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/Paracetamol56/dev-website/api/models"
)

func GetGithubAccessToken(code string) (string, error) {
	data := url.Values{}
	data.Set("client_id", os.Getenv("GITHUB_CLIENT_ID"))
	data.Set("client_secret", os.Getenv("GITHUB_CLIENT_SECRET"))
	data.Set("code", code)

	req, err := http.NewRequest("POST", "https://github.com/login/oauth/access_token", strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var body struct {
		AccessToken string `json:"access_token"`
	}
	if err := decodeBody(resp.Body, &body); err != nil {
		return "", err
	}

	return body.AccessToken, nil
}

func GetGithubUser(accessToken string) (*models.GitHubUser, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "token "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var githubUser models.GitHubUser
	if err := decodeBody(resp.Body, &githubUser); err != nil {
		return nil, err
	}

	return &githubUser, nil
}

func decodeBody(body io.ReadCloser, v interface{}) error {
	defer body.Close()
	return json.NewDecoder(body).Decode(v)
}
