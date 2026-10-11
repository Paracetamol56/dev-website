//go:build e2e

// Package e2e drives the whole API over HTTP, the way the web app does: real router,
// middlewares, tokens and database. Only the services of other companies are faked.
package e2e

import (
	"bytes"
	"encoding/json"
	"html"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	_ "dev/docs"
	"dev/internal/server"
	"dev/internal/testutil"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	testutil.Main(m)
}

const brevoHost = "api.brevo.com"

type app struct {
	t        *testing.T
	server   *httptest.Server
	internet *testutil.Internet
	run      int64
}

// start serves a fresh API (rate limits included) on a local port
func start(t *testing.T) *app {
	t.Helper()
	testutil.Clean(t)
	internet := testutil.FakeInternet(t)
	internet.JSON(brevoHost, http.StatusCreated, `{"messageId":"<message@brevo>"}`)
	apiServer := httptest.NewServer(server.InitRouter())
	t.Cleanup(apiServer.Close)
	return &app{t: t, server: apiServer, internet: internet, run: time.Now().UnixNano()}
}

// browser is one visitor of the site, logged in once it holds an access token
type browser struct {
	*app
	accessToken  string
	refreshToken string
	userId       string
}

func (app *app) visitor() *browser {
	return &browser{app: app}
}

type response struct {
	t      *testing.T
	Status int
	Header http.Header
	Body   []byte
}

func (browser *browser) call(method string, path string, body any) *response {
	browser.t.Helper()
	var reader io.Reader
	if text, raw := body.(string); raw {
		reader = strings.NewReader(text)
	} else if body != nil {
		encoded, err := json.Marshal(body)
		require.NoError(browser.t, err)
		reader = bytes.NewReader(encoded)
	}
	request, err := http.NewRequest(method, browser.server.URL+"/api"+path, reader)
	require.NoError(browser.t, err)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", testutil.Origin)
	request.Header.Set("Referer", testutil.Origin+"/settings")
	if browser.accessToken != "" {
		request.Header.Set("Authorization", "Bearer "+browser.accessToken)
	}

	result, err := http.DefaultClient.Do(request)
	require.NoError(browser.t, err)
	defer result.Body.Close()
	content, err := io.ReadAll(result.Body)
	require.NoError(browser.t, err)
	return &response{t: browser.t, Status: result.StatusCode, Header: result.Header, Body: content}
}

func (browser *browser) get(path string) *response { return browser.call(http.MethodGet, path, nil) }
func (browser *browser) post(path string, body any) *response {
	return browser.call(http.MethodPost, path, body)
}
func (browser *browser) patch(path string, body any) *response {
	return browser.call(http.MethodPatch, path, body)
}
func (browser *browser) delete(path string) *response {
	return browser.call(http.MethodDelete, path, nil)
}

// expect fails the test when the status is not the expected one
func (response *response) expect(status int) *response {
	response.t.Helper()
	require.Equal(response.t, status, response.Status, string(response.Body))
	return response
}

func (response *response) into(target any) {
	response.t.Helper()
	require.NoError(response.t, json.Unmarshal(response.Body, target), string(response.Body))
}

func (response *response) object() map[string]any {
	response.t.Helper()
	var value map[string]any
	response.into(&value)
	return value
}

func (response *response) list() []map[string]any {
	response.t.Helper()
	var value []map[string]any
	response.into(&value)
	return value
}

func (response *response) error() string {
	response.t.Helper()
	message, _ := response.object()["error"].(string)
	return message
}

type email struct {
	To []struct {
		Email string `json:"email"`
	} `json:"to"`
	ReplyTo *struct {
		Email string `json:"email"`
	} `json:"replyTo"`
	Subject     string `json:"subject"`
	HTMLContent string `json:"htmlContent"`
	TextContent string `json:"textContent"`
}

// inbox returns the emails sent to an address, oldest first
func (app *app) inbox(address string) []email {
	app.t.Helper()
	received := []email{}
	for _, request := range app.internet.Requests(brevoHost) {
		if request.Path != "/v3/smtp/email" {
			continue
		}
		var message email
		require.NoError(app.t, json.Unmarshal([]byte(request.Body), &message))
		if len(message.To) == 1 && strings.EqualFold(message.To[0].Email, address) {
			received = append(received, message)
		}
	}
	return received
}

var linkPattern = regexp.MustCompile(`<a href="([^"]+)">Verify your email address</a>`)

// login goes through the magic link flow: ask for a link, open the email, follow the link
func (browser *browser) login(address string) {
	browser.t.Helper()
	browser.post("/auth/login", map[string]string{"email": address}).expect(http.StatusNoContent)

	emails := browser.inbox(address)
	require.NotEmpty(browser.t, emails, "no email received")
	match := linkPattern.FindStringSubmatch(emails[len(emails)-1].HTMLContent)
	require.NotNil(browser.t, match, "no link in the email")
	link, err := url.Parse(html.UnescapeString(match[1]))
	require.NoError(browser.t, err)
	require.Equal(browser.t, testutil.Origin, link.Scheme+"://"+link.Host, "the link must lead back to the app")

	browser.keepSession(browser.post("/auth/verify", map[string]string{"token": link.Query().Get("token")}).expect(http.StatusOK))
}

func (browser *browser) keepSession(response *response) {
	browser.t.Helper()
	var session struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
		User         struct {
			Id string `json:"id"`
		} `json:"user"`
	}
	response.into(&session)
	require.NotEmpty(browser.t, session.AccessToken)
	require.NotEmpty(browser.t, session.RefreshToken)
	browser.accessToken, browser.refreshToken, browser.userId = session.AccessToken, session.RefreshToken, session.User.Id
}

// address returns an email of its own for each run of a test: login emails are limited per address
func (app *app) address(name string) string {
	return name + "." + strconv.FormatInt(app.run, 36) + "@example.test"
}

type socket struct {
	t    *testing.T
	conn *websocket.Conn
}

func (browser *browser) socket(path string) *socket {
	browser.t.Helper()
	conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(browser.server.URL, "http")+"/api"+path, nil)
	require.NoError(browser.t, err)
	browser.t.Cleanup(func() { conn.Close() })
	opened := &socket{t: browser.t, conn: conn}
	require.NoError(browser.t, conn.WriteJSON(map[string]string{"type": "auth", "token": browser.accessToken}))
	return opened
}

func (socket *socket) next() map[string]any {
	socket.t.Helper()
	var event map[string]any
	socket.conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	require.NoError(socket.t, socket.conn.ReadJSON(&event))
	return event
}
