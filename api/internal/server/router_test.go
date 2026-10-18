package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"dev/internal/testutil"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	testutil.Main(m)
}

func request(router *gin.Engine, method string, path string, configure func(*http.Request)) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, strings.NewReader("{}"))
	if configure != nil {
		configure(req)
	}
	router.ServeHTTP(recorder, req)
	return recorder
}

func TestRoutes(t *testing.T) {
	registered := map[string]bool{}
	for _, route := range InitRouter().Routes() {
		registered[route.Method+" "+route.Path] = true
	}

	for _, route := range []string{
		"GET /api/doc/*any", "GET /api/health", "POST /api/contact",
		"POST /api/auth/login", "POST /api/auth/verify", "POST /api/auth/refresh", "GET /api/auth/providers", "POST /api/auth/:provider",
		"POST /api/auth/passkey/begin", "POST /api/auth/passkey/finish",
		"GET /api/hipparcos", "GET /api/hipparcos/:hip",
		"GET /api/word-cloud", "GET /api/word-cloud/:id", "GET /api/word-cloud/:id/ws", "POST /api/word-cloud/:id/participant",
		"POST /api/word-cloud/:id/word", "POST /api/word-cloud", "PATCH /api/word-cloud/:id", "POST /api/word-cloud/:id/duplicate",
		"DELETE /api/word-cloud/:id", "PATCH /api/word-cloud/:id/words/:wordId", "POST /api/word-cloud/:id/words/merge",
		"GET /api/icons", "GET /api/icons/search", "GET /api/chips", "GET /api/chips/:id",
		"GET /api/ormi", "GET /api/ormi/stats", "GET /api/ormi/stats/:year", "GET /api/ormi/labels", "PUT /api/ormi/order",
		"GET /api/ormi/:id", "POST /api/ormi", "PATCH /api/ormi/:id", "DELETE /api/ormi/:id",
		"GET /api/slides/:id/ws", "GET /api/slides", "POST /api/slides", "GET /api/slides/:id",
		"POST /api/slides/:id/duplicate", "DELETE /api/slides/:id", "POST /api/slides/join",
		"PUT /api/slides/:id/links/:role", "DELETE /api/slides/:id/links/:role", "POST /api/slides/:id/links/:role/email",
		"PATCH /api/slides/:id/members/:userId", "DELETE /api/slides/:id/members/:userId",
		"GET /api/users/:id", "PATCH /api/users/:id", "DELETE /api/users/:id", "GET /api/users/:id/export",
		"DELETE /api/users/:id/identities/:provider", "POST /api/users/:id/passkeys/begin", "POST /api/users/:id/passkeys/finish",
		"DELETE /api/users/:id/passkeys/:passkeyId",
		"GET /api/admin/users", "GET /api/admin/users/:id", "PATCH /api/admin/users/:id/role", "DELETE /api/admin/users/:id",
		"GET /api/admin/audit",
	} {
		assert.True(t, registered[route], route)
		delete(registered, route)
	}
	assert.Empty(t, registered, "routes missing from this list")
}

func TestHealthAndRequestId(t *testing.T) {
	recorder := request(InitRouter(), http.MethodGet, "/api/health", nil)
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.JSONEq(t, `{"status": "ok"}`, recorder.Body.String())
	assert.NotEmpty(t, recorder.Header().Get("x-request-id"))

	recorder = request(InitRouter(), http.MethodGet, "/not-an-api-route", nil)
	assert.NotEmpty(t, recorder.Header().Get("x-request-id"))
}

func TestProtectedRoutesRequireAToken(t *testing.T) {
	router := InitRouter()
	id := "65f000000000000000000000"
	for _, route := range []string{
		"POST /api/word-cloud", "PATCH /api/word-cloud/" + id, "POST /api/word-cloud/" + id + "/duplicate", "DELETE /api/word-cloud/" + id,
		"PATCH /api/word-cloud/" + id + "/words/" + id, "POST /api/word-cloud/" + id + "/words/merge",
		"GET /api/ormi", "GET /api/ormi/stats", "GET /api/ormi/stats/2026", "GET /api/ormi/labels", "PUT /api/ormi/order",
		"GET /api/ormi/" + id, "POST /api/ormi", "PATCH /api/ormi/" + id, "DELETE /api/ormi/" + id,
		"GET /api/slides", "POST /api/slides", "GET /api/slides/" + id,
		"POST /api/slides/" + id + "/duplicate", "DELETE /api/slides/" + id, "POST /api/slides/join",
		"PUT /api/slides/" + id + "/links/editor", "DELETE /api/slides/" + id + "/links/editor",
		"POST /api/slides/" + id + "/links/editor/email", "PATCH /api/slides/" + id + "/members/" + id,
		"DELETE /api/slides/" + id + "/members/" + id,
		"GET /api/users/" + id, "PATCH /api/users/" + id, "DELETE /api/users/" + id, "GET /api/users/" + id + "/export",
		"DELETE /api/users/" + id + "/identities/github", "POST /api/users/" + id + "/passkeys/begin",
		"POST /api/users/" + id + "/passkeys/finish", "DELETE /api/users/" + id + "/passkeys/key",
		"GET /api/admin/users", "GET /api/admin/users/" + id, "PATCH /api/admin/users/" + id + "/role",
		"DELETE /api/admin/users/" + id, "GET /api/admin/audit",
	} {
		method, path, _ := strings.Cut(route, " ")
		assert.Equal(t, http.StatusUnauthorized, request(router, method, path, nil).Code, route)
	}
}

func TestCors(t *testing.T) {
	router := InitRouter()
	preflight := func(origin string) *httptest.ResponseRecorder {
		return request(router, http.MethodOptions, "/api/ormi", func(req *http.Request) {
			req.Header.Set("Origin", origin)
			req.Header.Set("Access-Control-Request-Method", "PATCH")
			req.Header.Set("Access-Control-Request-Headers", "Authorization")
		})
	}

	allowed := preflight("https://dev.matheo-galuba.com")
	assert.Equal(t, http.StatusNoContent, allowed.Code)
	assert.Equal(t, "https://dev.matheo-galuba.com", allowed.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, allowed.Header().Get("Access-Control-Allow-Methods"), "PATCH")
	assert.Contains(t, allowed.Header().Get("Access-Control-Allow-Headers"), "Authorization")

	refused := preflight("https://attacker.example")
	assert.Equal(t, http.StatusForbidden, refused.Code)
	assert.Empty(t, refused.Header().Get("Access-Control-Allow-Origin"))
}

func TestSwaggerAndStaticFallback(t *testing.T) {
	router := InitRouter()

	doc := request(router, http.MethodGet, "/api/doc/index.html", nil)
	assert.Equal(t, http.StatusOK, doc.Code)
	assert.Contains(t, doc.Body.String(), "swagger")

	directory := t.TempDir()
	t.Chdir(directory)
	assert.Equal(t, http.StatusNotFound, request(router, http.MethodGet, "/page/boids", nil).Code, "no build directory")

	assert.NoError(t, os.Mkdir("build", 0o755))
	assert.NoError(t, os.WriteFile("build/index.html", []byte("<title>app</title>"), 0o644))
	assert.NoError(t, os.WriteFile("build/robots.txt", []byte("User-agent: *"), 0o644))
	router = InitRouter()

	asset := request(router, http.MethodGet, "/robots.txt", nil)
	assert.Equal(t, http.StatusOK, asset.Code)
	assert.Equal(t, "User-agent: *", asset.Body.String())
	fallback := request(router, http.MethodGet, "/page/boids", nil)
	assert.Equal(t, http.StatusOK, fallback.Code, "unknown paths are left to the single-page app")
	assert.Equal(t, "<title>app</title>", fallback.Body.String())
	barnesHut := request(router, http.MethodGet, "/page/barnes-hut", nil)
	assert.Equal(t, "require-corp", barnesHut.Header().Get("Cross-Origin-Embedder-Policy"))
}

func postContact(router *gin.Engine, remoteAddr string, forwardedFor string) *httptest.ResponseRecorder {
	return request(router, http.MethodPost, "/api/contact", func(req *http.Request) {
		req.RemoteAddr = remoteAddr
		if forwardedFor != "" {
			req.Header.Set("X-Forwarded-For", forwardedFor)
		}
	})
}

func TestRateLimitIgnoresSpoofedForwardedFor(t *testing.T) {
	router := InitRouter()

	for i := range 5 {
		assert.Equal(t, http.StatusBadRequest, postContact(router, "203.0.113.7:4000", "198.51.100."+strconv.Itoa(i)).Code)
	}
	limited := postContact(router, "203.0.113.7:4000", "198.51.100.99")
	assert.Equal(t, http.StatusTooManyRequests, limited.Code)
	assert.NotEmpty(t, limited.Header().Get("Retry-After"))
}

func TestRateLimitTrustsPrivateProxy(t *testing.T) {
	router := InitRouter()

	for range 5 {
		assert.Equal(t, http.StatusBadRequest, postContact(router, "172.18.0.2:4000", "198.51.100.1").Code)
	}
	assert.Equal(t, http.StatusTooManyRequests, postContact(router, "172.18.0.2:4000", "198.51.100.1").Code)
	assert.Equal(t, http.StatusBadRequest, postContact(router, "172.18.0.2:4000", "198.51.100.2").Code)
}

func TestAuthRoutesAreRateLimited(t *testing.T) {
	router := InitRouter()
	providers := func() int {
		return request(router, http.MethodGet, "/api/auth/providers", func(req *http.Request) { req.RemoteAddr = "203.0.113.9:4000" }).Code
	}

	for range 30 {
		assert.Equal(t, http.StatusOK, providers())
	}
	assert.Equal(t, http.StatusTooManyRequests, providers())
}
