package slides

import (
	"context"
	"dev/internal/testutil"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func postDeck(t *testing.T, userId primitive.ObjectID, body map[string]any) Deck {
	t.Helper()
	response := do(t, http.MethodPost, "/slides", body, testutil.As(userId))
	require.Equal(t, http.StatusCreated, response.Code, response.Body.String())
	return testutil.Decode[Deck](t, response)
}

func listDecks(t *testing.T, userId primitive.ObjectID, query string) DeckPage {
	t.Helper()
	response := do(t, http.MethodGet, "/slides?"+query, nil, testutil.As(userId))
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	return testutil.Decode[DeckPage](t, response)
}

func titles(page DeckPage) []string {
	names := []string{}
	for _, deck := range page.Items {
		names = append(names, deck.Title)
	}
	return names
}

func TestPostDeck(t *testing.T) {
	testutil.Clean(t)
	userId := primitive.NewObjectID()

	deck := postDeck(t, userId, map[string]any{"title": "Talk"})
	assert.False(t, deck.Id.IsZero())
	assert.Equal(t, userId, deck.UserId)
	assert.Equal(t, "Talk", deck.Title)
	assert.Equal(t, defaultContent, deck.Content)
	assert.Equal(t, []OutlineEntry{{Title: "Title", From: 0, To: len(defaultContent) - 1}}, deck.Outline)
	assert.EqualValues(t, 0, deck.Revision)
	assert.WithinDuration(t, time.Now(), deck.CreatedAt.Time(), 5*time.Second)
	assert.Equal(t, deck.CreatedAt, deck.UpdatedAt)

	empty := postDeck(t, userId, map[string]any{"title": "Empty", "content": ""})
	assert.Equal(t, "", empty.Content)
	assert.Empty(t, empty.Outline)

	stored, err := GetDeckByUserIdAndId(context.Background(), userId, deck.Id)
	require.NoError(t, err)
	assert.Equal(t, defaultContent, stored.Content)

	entries := testutil.Audit(t)
	require.Len(t, entries, 2)
	assert.Equal(t, "slides.created", entries[0].Action)
	assert.Equal(t, userId, entries[0].UserId)
	assert.Equal(t, "slides", entries[0].Target.Type)
	assert.Equal(t, deck.Id.Hex(), entries[0].Target.Id)
	assert.Equal(t, "Talk", entries[0].Meta["title"])
}

func TestPostDeckInvalid(t *testing.T) {
	testutil.Clean(t)
	for name, body := range map[string]any{
		"no title":         map[string]any{},
		"empty title":      map[string]any{"title": ""},
		"long title":       map[string]any{"title": strings.Repeat("a", 101)},
		"too much content": map[string]any{"title": "Big", "content": strings.Repeat("a", 500001)},
		"not json":         "{",
	} {
		t.Run(name, func(t *testing.T) {
			response := do(t, http.MethodPost, "/slides", body, testutil.As(primitive.NewObjectID()))
			assert.Equal(t, http.StatusBadRequest, response.Code, response.Body.String())
		})
	}
	assert.Zero(t, testutil.Count(t, "slides"))
}

func TestGetDeck(t *testing.T) {
	testutil.Clean(t)
	owner := primitive.NewObjectID()
	deck := postDeck(t, owner, map[string]any{"title": "Mine", "content": "<section>Hi</section>"})

	response := do(t, http.MethodGet, "/slides/"+deck.Id.Hex(), nil, testutil.As(owner))
	require.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, "<section>Hi</section>", testutil.Decode[Deck](t, response).Content)

	response = do(t, http.MethodGet, "/slides/"+deck.Id.Hex(), nil, testutil.As(primitive.NewObjectID()))
	assert.Equal(t, http.StatusNotFound, response.Code, "another user cannot tell the deck exists")

	assert.Equal(t, http.StatusNotFound, do(t, http.MethodGet, "/slides/"+primitive.NewObjectID().Hex(), nil, testutil.As(owner)).Code)
	assert.Equal(t, http.StatusBadRequest, do(t, http.MethodGet, "/slides/nope", nil, testutil.As(owner)).Code)
}

func TestGetDecks(t *testing.T) {
	testutil.Clean(t)
	userId := primitive.NewObjectID()
	for i := 1; i <= 5; i++ {
		postDeck(t, userId, map[string]any{"title": "Deck " + strconv.Itoa(i)})
		time.Sleep(2 * time.Millisecond)
	}
	postDeck(t, primitive.NewObjectID(), map[string]any{"title": "Someone else's"})

	page := listDecks(t, userId, "")
	assert.EqualValues(t, 5, page.Total)
	assert.Equal(t, 1, page.Page)
	assert.Equal(t, DefaultLimit, page.Limit)
	assert.Equal(t, []string{"Deck 5", "Deck 4", "Deck 3", "Deck 2", "Deck 1"}, titles(page), "last updated first")
	assert.Len(t, page.Items[0].Outline, 1)

	page = listDecks(t, userId, "page=2&limit=2")
	assert.EqualValues(t, 5, page.Total)
	assert.Equal(t, []string{"Deck 3", "Deck 2"}, titles(page))

	assert.Equal(t, []string{"Deck 3"}, titles(listDecks(t, userId, "search=k+3")))
	assert.Empty(t, listDecks(t, userId, "search=.*").Items, "the search is not a pattern")
	assert.Empty(t, listDecks(t, primitive.NewObjectID(), "").Items)

	response := do(t, http.MethodGet, "/slides", nil, testutil.As(userId))
	assert.NotContains(t, response.Body.String(), "content", "lists leave the content out")

	for _, query := range []string{"page=-1", "limit=101", "limit=-1", "page=x"} {
		assert.Equal(t, http.StatusBadRequest, do(t, http.MethodGet, "/slides?"+query, nil, testutil.As(userId)).Code, query)
	}
}

func TestPostDeckDuplicate(t *testing.T) {
	testutil.Clean(t)
	userId := primitive.NewObjectID()
	deck := postDeck(t, userId, map[string]any{"title": "Original", "content": "<section>A</section>"})

	response := do(t, http.MethodPost, "/slides/"+deck.Id.Hex()+"/duplicate", nil, testutil.As(userId))
	require.Equal(t, http.StatusCreated, response.Code, response.Body.String())
	copied := testutil.Decode[Deck](t, response)
	assert.NotEqual(t, deck.Id, copied.Id)
	assert.Equal(t, "Original (copy)", copied.Title)
	assert.Equal(t, deck.Content, copied.Content)
	assert.EqualValues(t, 0, copied.Revision)

	long := postDeck(t, userId, map[string]any{"title": strings.Repeat("a", 100)})
	response = do(t, http.MethodPost, "/slides/"+long.Id.Hex()+"/duplicate", nil, testutil.As(userId))
	require.Equal(t, http.StatusCreated, response.Code)
	assert.Equal(t, long.Title, testutil.Decode[Deck](t, response).Title, "the title stays within its limit")

	assert.Equal(t, http.StatusNotFound, do(t, http.MethodPost, "/slides/"+deck.Id.Hex()+"/duplicate", nil, testutil.As(primitive.NewObjectID())).Code)
	assert.Equal(t, 4, testutil.Count(t, "slides"))

	entries := testutil.Audit(t)
	assert.Equal(t, "slides.duplicated", entries[1].Action)
	assert.Equal(t, copied.Id.Hex(), entries[1].Target.Id)
	assert.Equal(t, deck.Id.Hex(), entries[1].Meta["from"])
}

func TestDeleteDeck(t *testing.T) {
	testutil.Clean(t)
	userId := primitive.NewObjectID()
	deck := postDeck(t, userId, map[string]any{"title": "Gone"})
	path := "/slides/" + deck.Id.Hex()

	assert.Equal(t, http.StatusNotFound, do(t, http.MethodDelete, path, nil, testutil.As(primitive.NewObjectID())).Code)
	assert.Equal(t, 1, testutil.Count(t, "slides"))

	assert.Equal(t, http.StatusNoContent, do(t, http.MethodDelete, path, nil, testutil.As(userId)).Code)
	assert.Zero(t, testutil.Count(t, "slides"))
	assert.Equal(t, http.StatusNotFound, do(t, http.MethodDelete, path, nil, testutil.As(userId)).Code)

	assert.Equal(t, []string{"slides.created", "slides.deleted"}, testutil.AuditActions(t))
}

func TestDatabaseFailure(t *testing.T) {
	testutil.Clean(t)
	userId := primitive.NewObjectID()
	deck := postDeck(t, userId, map[string]any{"title": "Deck"})
	testutil.BreakDB(t)

	path := "/slides/" + deck.Id.Hex()
	for _, request := range []struct {
		method, path string
		body         any
	}{
		{http.MethodGet, "/slides", nil},
		{http.MethodPost, "/slides", map[string]any{"title": "New"}},
		{http.MethodGet, path, nil},
		{http.MethodPost, path + "/duplicate", nil},
		{http.MethodDelete, path, nil},
	} {
		response := do(t, request.method, request.path, request.body, testutil.As(userId))
		assert.Equal(t, http.StatusInternalServerError, response.Code, request.method+" "+request.path)
	}
}

func TestDeckSettingsDefaults(t *testing.T) {
	testutil.Clean(t)
	userId := primitive.NewObjectID()
	deck := postDeck(t, userId, map[string]any{"title": "Styled"})
	assert.Equal(t, DefaultSettings(), deck.Settings)

	custom := DefaultSettings()
	custom.Theme = "moon"
	require.NoError(t, SetDeckFields(context.Background(), deck.Id, bson.M{"settings": custom}))
	response := do(t, http.MethodPost, "/slides/"+deck.Id.Hex()+"/duplicate", nil, testutil.As(userId))
	require.Equal(t, http.StatusCreated, response.Code)
	assert.Equal(t, custom, testutil.Decode[Deck](t, response).Settings, "a copy keeps the settings")
}

func TestBackfillSettings(t *testing.T) {
	testutil.Clean(t)
	old := primitive.NewObjectID()
	owner := primitive.NewObjectID()
	testutil.InsertRaw(t, "slides", bson.M{"_id": old, "user": owner, "title": "Old", "content": ""})
	styled := postDeck(t, owner, map[string]any{"title": "New"})
	custom := DefaultSettings()
	custom.Theme = "white"
	require.NoError(t, SetDeckFields(context.Background(), styled.Id, bson.M{"settings": custom}))

	filled, err := BackfillSettings(context.Background())
	require.NoError(t, err)
	assert.EqualValues(t, 1, filled)

	deck, err := GetDeckByUserIdAndId(context.Background(), owner, old)
	require.NoError(t, err)
	assert.Equal(t, DefaultSettings(), deck.Settings)
	deck, err = GetDeckByUserIdAndId(context.Background(), owner, styled.Id)
	require.NoError(t, err)
	assert.Equal(t, "white", deck.Settings.Theme)
}
