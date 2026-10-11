//go:build e2e

package e2e

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"dev/internal/db"
	"dev/internal/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

func TestHealthDocumentationAndCors(t *testing.T) {
	app := start(t)
	visitor := app.visitor()

	health := visitor.get("/health").expect(http.StatusOK)
	assert.Equal(t, map[string]any{"status": "ok"}, health.object())
	assert.NotEmpty(t, health.Header.Get("x-request-id"))
	assert.Equal(t, "http://localhost:5173", health.Header.Get("Access-Control-Allow-Origin"))

	var specification struct {
		Paths map[string]any `json:"paths"`
	}
	visitor.get("/doc/doc.json").expect(http.StatusOK).into(&specification)
	for _, path := range []string{"/health", "/auth/login", "/ormi", "/chips", "/icons/search", "/users/{id}"} {
		assert.Contains(t, specification.Paths, path, "documented")
	}

	missing, err := http.Get(app.server.URL + "/api/nothing-here")
	require.NoError(t, err)
	missing.Body.Close()
	assert.Equal(t, http.StatusNotFound, missing.StatusCode)
}

func TestCatalogsAreImportedThenServed(t *testing.T) {
	app := start(t)
	app.internet.Handle("unpkg.com", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/lucide-static@latest/tags.json" {
			io.WriteString(w, `{"zap": ["flash", "electricity"], "zap-off": ["flash"], "arrow-down": ["direction"]}`)
			return
		}
		io.WriteString(w, `[{"title": "Zapier", "slug": "zapier", "hex": "FF4F00"}, {"title": "GitHub", "slug": "github", "hex": "181717"}]`)
	})
	page, err := json.Marshal(map[string]any{"parse": map[string]any{"revid": 99, "text": `
		<h2 id="Microprocessors"></h2>
		<table class="wikitable">
			<tr><th>Processor</th><th>Transistor count</th><th>Year</th><th>Designer</th><th>Process (nm)</th><th>Area (mm2)</th></tr>
			<tr><td>Intel 4004</td><td>2,250</td><td>1971</td><td>Intel</td><td>10,000 nm</td><td>12 mm2</td></tr>
			<tr><td>Apple M1</td><td>16,000,000,000</td><td>2020</td><td>Apple</td><td>5 nm</td><td>119 mm2</td></tr>
		</table>
		<h2 id="GPUs"></h2>
		<table class="wikitable">
			<tr><th>Processor</th><th>Transistor count</th><th>Year</th><th>Designer(s)</th><th>Fab(s)</th></tr>
			<tr><td>GeForce 256</td><td>17,000,000</td><td>1999</td><td>Nvidia</td><td>TSMC</td></tr>
		</table>`}})
	require.NoError(t, err)
	app.internet.JSON("en.wikipedia.org", http.StatusOK, string(page))
	_, err = db.GetDB().Collection("hipparcos").InsertMany(context.Background(), []any{
		bson.M{"HIP": 32349, "Vmag": -1.44, "Plx": 379.21, "B-V": 0.009},
		bson.M{"HIP": 1, "Vmag": 9.1, "Plx": 3.54, "B-V": 0.482},
		bson.M{"HIP": 2, "Vmag": 9.27, "Plx": -1.2, "B-V": 0.999},
	})
	require.NoError(t, err)

	utils.SeedIcons(context.Background())
	utils.SeedMicroprocessors(context.Background())
	visitor := app.visitor()

	assert.Len(t, visitor.get("/icons").expect(http.StatusOK).list(), 5)
	assert.Len(t, visitor.get("/icons?source=simpleicons").expect(http.StatusOK).list(), 2)
	var found struct {
		Total int              `json:"total"`
		Items []map[string]any `json:"items"`
	}
	search := visitor.get("/icons/search?q=zap&limit=2").expect(http.StatusOK)
	search.into(&found)
	assert.Equal(t, 3, found.Total)
	require.Len(t, found.Items, 2)
	assert.Equal(t, "lucide:zap", found.Items[0]["id"], "the exact match comes first")
	assert.Equal(t, "public, max-age=3600", search.Header.Get("Cache-Control"))
	visitor.get("/icons/search?tag=flash&source=lucide").expect(http.StatusOK).into(&found)
	assert.Equal(t, 2, found.Total)
	visitor.get("/icons/search?limit=9999").expect(http.StatusBadRequest)

	chips := visitor.get("/chips").expect(http.StatusOK).list()
	assert.Len(t, chips, 3)
	gpus := visitor.get("/chips?type=GPU").expect(http.StatusOK).list()
	require.Len(t, gpus, 1)
	assert.Equal(t, "GeForce 256", gpus[0]["name"])
	assert.Equal(t, "NVIDIA", gpus[0]["vendor"])
	assert.Equal(t, "TSMC", gpus[0]["manufacturer"])
	assert.Equal(t, 17.0, gpus[0]["transistors"], "in millions")
	m1 := visitor.get("/chips/wikipedia:cpu-apple-m1-2020").expect(http.StatusOK).object()
	assert.Equal(t, 16000.0, m1["transistors"])
	assert.Equal(t, "2020-01-01T00:00:00Z", m1["release"])
	assert.Equal(t, map[string]any{"value": 5.0, "unit": "nm"}, m1["process"])
	assert.Equal(t, "https://en.wikipedia.org/w/index.php?title=Transistor_count&oldid=99", m1["sourceUrl"])
	visitor.get("/chips/unknown").expect(http.StatusNotFound)

	stars := visitor.get("/hipparcos").expect(http.StatusOK).list()
	assert.Len(t, stars, 2, "stars without a usable parallax are left out")
	sirius := visitor.get("/hipparcos/32349").expect(http.StatusOK).object()
	assert.InDelta(t, 1.45, sirius["Amag"], 0.01)
	visitor.get("/hipparcos/2").expect(http.StatusNotFound)
	visitor.get("/hipparcos/sirius").expect(http.StatusBadRequest)
}

func TestContactForm(t *testing.T) {
	app := start(t)
	visitor := app.visitor()
	message := map[string]any{"name": "Ada Lovelace", "email": "ada@example.test", "message": "Your chart of Moore's law is great."}

	visitor.post("/contact", map[string]any{"name": "Ada", "email": "not-an-email", "message": "Hello there, friend"}).expect(http.StatusBadRequest)
	visitor.post("/contact", message).expect(http.StatusCreated)

	received := app.inbox("admin@example.test")
	require.Len(t, received, 1)
	assert.Equal(t, "ada@example.test", received[0].ReplyTo.Email)
	assert.Contains(t, received[0].TextContent, "Your chart of Moore's law is great.")
	stored, err := db.GetDB().Collection("contacts").CountDocuments(context.Background(), bson.M{"email": "ada@example.test"})
	require.NoError(t, err)
	assert.EqualValues(t, 1, stored)

	for range 3 {
		visitor.post("/contact", message).expect(http.StatusCreated)
	}
	limited := visitor.post("/contact", message).expect(http.StatusTooManyRequests)
	assert.NotEmpty(t, limited.Header.Get("Retry-After"))
	assert.Len(t, app.inbox("admin@example.test"), 4, "five attempts an hour, valid or not")
}

func TestTokensAreChecked(t *testing.T) {
	app := start(t)
	ada := app.visitor()
	ada.login(app.address("ada"))
	ada.get("/ormi").expect(http.StatusOK)

	refreshAsAccess := app.visitor()
	refreshAsAccess.accessToken = ada.refreshToken
	refreshAsAccess.get("/ormi").expect(http.StatusUnauthorized)

	tampered := app.visitor()
	tampered.accessToken = ada.accessToken[:len(ada.accessToken)-2] + "xx"
	tampered.get("/ormi").expect(http.StatusUnauthorized)
	tampered.get("/users/" + ada.userId).expect(http.StatusUnauthorized)
	tampered.post("/word-cloud", map[string]any{"name": "Forged"}).expect(http.StatusUnauthorized)
	tampered.post("/auth/refresh", map[string]string{"token": ada.refreshToken + "x"}).expect(http.StatusBadRequest)
}
