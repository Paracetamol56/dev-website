package slides

import (
	"dev/internal/testutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func insertUser(t *testing.T, name string) primitive.ObjectID {
	t.Helper()
	id := primitive.NewObjectID()
	testutil.InsertRaw(t, "users", bson.M{"_id": id, "name": name, "email": strings.ToLower(name) + "@example.test"})
	return id
}

func createLink(t *testing.T, owner primitive.ObjectID, deck primitive.ObjectID, role string) string {
	t.Helper()
	response := do(t, http.MethodPut, "/slides/"+deck.Hex()+"/links/"+role, nil, testutil.As(owner))
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	links := testutil.Decode[Links](t, response)
	token := links.Viewer
	if role == RoleEditor {
		token = links.Editor
	}
	require.NotNil(t, token)
	return *token
}

func join(t *testing.T, userId primitive.ObjectID, token string) JoinedDeck {
	t.Helper()
	response := do(t, http.MethodPost, "/slides/join", map[string]any{"token": token}, testutil.As(userId))
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	return testutil.Decode[JoinedDeck](t, response)
}

func deckView(t *testing.T, userId primitive.ObjectID, id primitive.ObjectID) (int, map[string]any) {
	t.Helper()
	response := do(t, http.MethodGet, "/slides/"+id.Hex(), nil, testutil.As(userId))
	if response.Code != http.StatusOK {
		return response.Code, nil
	}
	return response.Code, testutil.Decode[map[string]any](t, response)
}

func TestShareLinks(t *testing.T) {
	testutil.Clean(t)
	owner := insertUser(t, "Owner")
	ada := insertUser(t, "Ada")
	deck := postDeck(t, owner, map[string]any{"title": "Shared"})

	code, _ := deckView(t, ada, deck.Id)
	assert.Equal(t, http.StatusNotFound, code, "a deck is private until shared")
	assert.Equal(t, http.StatusNotFound, do(t, http.MethodPut, "/slides/"+deck.Id.Hex()+"/links/editor", nil, testutil.As(ada)).Code)
	assert.Equal(t, http.StatusBadRequest, do(t, http.MethodPut, "/slides/"+deck.Id.Hex()+"/links/owner", nil, testutil.As(owner)).Code)

	viewer := createLink(t, owner, deck.Id, RoleViewer)
	assert.Len(t, viewer, 43)
	joined := join(t, ada, viewer)
	assert.Equal(t, JoinedDeck{Id: deck.Id, Role: RoleViewer}, joined)

	code, view := deckView(t, ada, deck.Id)
	require.Equal(t, http.StatusOK, code)
	assert.Equal(t, "viewer", view["role"])
	assert.Equal(t, deck.Content, view["content"])
	assert.NotContains(t, view, "links", "only the owner sees the links")
	assert.NotContains(t, view, "members")

	_, view = deckView(t, owner, deck.Id)
	assert.Equal(t, "owner", view["role"])
	assert.Equal(t, map[string]any{"viewer": viewer}, view["links"])
	members := view["members"].([]any)
	require.Len(t, members, 1)
	assert.Equal(t, ada.Hex(), members[0].(map[string]any)["user"])
	assert.Equal(t, "Ada", members[0].(map[string]any)["name"])
	assert.Equal(t, "viewer", members[0].(map[string]any)["role"])

	renewed := createLink(t, owner, deck.Id, RoleViewer)
	assert.NotEqual(t, viewer, renewed)
	assert.Equal(t, http.StatusNotFound, do(t, http.MethodPost, "/slides/join", map[string]any{"token": viewer}, testutil.As(insertUser(t, "Late"))).Code, "a replaced link stops working")

	assert.Equal(t, http.StatusNoContent, do(t, http.MethodDelete, "/slides/"+deck.Id.Hex()+"/links/viewer", nil, testutil.As(owner)).Code)
	assert.Equal(t, http.StatusNotFound, do(t, http.MethodPost, "/slides/join", map[string]any{"token": renewed}, testutil.As(insertUser(t, "Later"))).Code)
	code, _ = deckView(t, ada, deck.Id)
	assert.Equal(t, http.StatusOK, code, "members keep their access when a link is turned off")

	assert.Equal(t, []string{"slides.created", "slides.link_created", "slides.joined", "slides.link_created", "slides.link_revoked"}, testutil.AuditActions(t))
}

func TestJoinDeckRoles(t *testing.T) {
	testutil.Clean(t)
	owner := insertUser(t, "Owner")
	ada := insertUser(t, "Ada")
	deck := postDeck(t, owner, map[string]any{"title": "Shared"})
	editor := createLink(t, owner, deck.Id, RoleEditor)
	viewer := createLink(t, owner, deck.Id, RoleViewer)

	assert.Equal(t, RoleOwner, join(t, owner, editor).Role, "the owner stays the owner")
	assert.Equal(t, RoleViewer, join(t, ada, viewer).Role)
	assert.Equal(t, RoleEditor, join(t, ada, editor).Role, "an editor link upgrades a viewer")
	assert.Equal(t, RoleEditor, join(t, ada, viewer).Role, "a viewer link never downgrades")

	stored := storedDeck(t, deck.Id)
	require.Len(t, stored.Members, 1, "joining twice keeps one membership")
	assert.Equal(t, RoleEditor, stored.Members[0].Role)

	for name, body := range map[string]any{"no token": map[string]any{}, "unknown token": map[string]any{"token": "nope"}} {
		response := do(t, http.MethodPost, "/slides/join", body, testutil.As(ada))
		assert.Contains(t, []int{http.StatusBadRequest, http.StatusNotFound}, response.Code, name)
	}
}

func TestSharedDecksList(t *testing.T) {
	testutil.Clean(t)
	owner := insertUser(t, "Owner")
	ada := insertUser(t, "Ada")
	shared := postDeck(t, owner, map[string]any{"title": "Shared"})
	postDeck(t, owner, map[string]any{"title": "Private"})
	postDeck(t, ada, map[string]any{"title": "Mine"})
	join(t, ada, createLink(t, owner, shared.Id, RoleEditor))

	page := listDecks(t, ada, "")
	assert.Equal(t, []string{"Mine", "Shared"}, titles(page), "owned and shared decks are listed together, last updated first")
	assert.EqualValues(t, 2, page.Total)
	assert.Equal(t, RoleOwner, page.Items[0].Role)
	assert.Empty(t, page.Items[0].OwnerName)
	assert.Equal(t, RoleEditor, page.Items[1].Role)
	assert.Equal(t, "Owner", page.Items[1].OwnerName)

	assert.Equal(t, []string{"Shared"}, titles(listDecks(t, ada, "search=sha")))
	assert.Equal(t, []string{"Private", "Shared"}, titles(listDecks(t, owner, "")), "members do not see the other decks of the owner, and the owner does not see theirs")
}

func TestMemberRights(t *testing.T) {
	testutil.Clean(t)
	owner := insertUser(t, "Owner")
	ada := insertUser(t, "Ada")
	bob := insertUser(t, "Bob")
	deck := postDeck(t, owner, map[string]any{"title": "Shared"})
	join(t, ada, createLink(t, owner, deck.Id, RoleEditor))
	join(t, bob, createLink(t, owner, deck.Id, RoleViewer))
	path := "/slides/" + deck.Id.Hex()

	assert.Equal(t, http.StatusForbidden, do(t, http.MethodPost, path+"/duplicate", nil, testutil.As(bob)).Code, "viewers cannot duplicate")
	response := do(t, http.MethodPost, path+"/duplicate", nil, testutil.As(ada))
	require.Equal(t, http.StatusCreated, response.Code)
	copied := storedDeck(t, testutil.Decode[Deck](t, response).Id)
	assert.Equal(t, ada, copied.UserId, "editors get their own copy")
	assert.Empty(t, copied.Members)
	assert.Nil(t, copied.Links.Editor)

	for _, request := range []struct{ method, path string }{
		{http.MethodDelete, path},
		{http.MethodPut, path + "/links/viewer"},
		{http.MethodDelete, path + "/links/viewer"},
		{http.MethodPatch, path + "/members/" + bob.Hex()},
		{http.MethodDelete, path + "/members/" + bob.Hex()},
	} {
		response := do(t, request.method, request.path, map[string]any{"role": "editor"}, testutil.As(ada))
		assert.Equal(t, http.StatusForbidden, response.Code, "%s %s", request.method, request.path)
	}

	assert.Equal(t, http.StatusBadRequest, do(t, http.MethodPatch, path+"/members/"+bob.Hex(), map[string]any{"role": "owner"}, testutil.As(owner)).Code)
	assert.Equal(t, http.StatusNotFound, do(t, http.MethodPatch, path+"/members/"+owner.Hex(), map[string]any{"role": "editor"}, testutil.As(owner)).Code)
	assert.Equal(t, http.StatusNoContent, do(t, http.MethodPatch, path+"/members/"+bob.Hex(), map[string]any{"role": "editor"}, testutil.As(owner)).Code)
	_, view := deckView(t, bob, deck.Id)
	assert.Equal(t, "editor", view["role"])

	assert.Equal(t, http.StatusNoContent, do(t, http.MethodDelete, path+"/members/"+ada.Hex(), nil, testutil.As(ada)).Code, "a member can leave")
	assert.Equal(t, http.StatusNoContent, do(t, http.MethodDelete, path+"/members/"+bob.Hex(), nil, testutil.As(owner)).Code)
	code, _ := deckView(t, bob, deck.Id)
	assert.Equal(t, http.StatusNotFound, code)
	assert.Equal(t, http.StatusNotFound, do(t, http.MethodDelete, path+"/members/"+bob.Hex(), nil, testutil.As(owner)).Code)
	assert.Empty(t, storedDeck(t, deck.Id).Members)
}

func TestEmailLink(t *testing.T) {
	testutil.Clean(t)
	internet := testutil.FakeInternet(t)
	testutil.FakeBrevo(internet)
	owner := insertUser(t, "Owner <script>")
	deck := postDeck(t, owner, map[string]any{"title": "Talk & co"})
	path := "/slides/" + deck.Id.Hex() + "/links/editor/email"

	response := do(t, http.MethodPost, path, map[string]any{"email": "ada@example.org"}, testutil.As(owner), testutil.FromApp)
	require.Equal(t, http.StatusNoContent, response.Code, response.Body.String())
	stored := storedDeck(t, deck.Id)
	require.NotNil(t, stored.Links.Editor, "sending turns the link on")

	emails := testutil.SentEmails(t, internet)
	require.Len(t, emails, 1)
	assert.Equal(t, "ada@example.org", emails[0].To[0].Email)
	assert.Equal(t, "Owner <script> shared a slide deck with you", emails[0].Subject)
	assert.Contains(t, emails[0].HTMLContent, testutil.Origin+"/tools/slides/join/"+*stored.Links.Editor)
	assert.Contains(t, emails[0].HTMLContent, "Talk &amp; co")
	assert.Contains(t, emails[0].HTMLContent, "Owner &lt;script&gt;")
	assert.Contains(t, emails[0].HTMLContent, "to edit")

	do(t, http.MethodPost, "/slides/"+deck.Id.Hex()+"/links/viewer/email", map[string]any{"email": "bob@example.org"}, testutil.As(owner), testutil.Header("Origin", "https://evil.example"))
	emails = testutil.SentEmails(t, internet)
	require.Len(t, emails, 2)
	assert.Contains(t, emails[1].HTMLContent, defaultSite+"/tools/slides/join/", "unknown origins get the public site")
	assert.Equal(t, *stored.Links.Editor, *storedDeck(t, deck.Id).Links.Editor, "an existing link is reused")

	assert.Equal(t, http.StatusBadRequest, do(t, http.MethodPost, path, map[string]any{"email": "nope"}, testutil.As(owner)).Code)
	assert.Equal(t, http.StatusNotFound, do(t, http.MethodPost, path, map[string]any{"email": "a@b.org"}, testutil.As(primitive.NewObjectID())).Code)

	entries := testutil.Audit(t)
	emailed := entries[2]
	assert.Equal(t, "slides.link_emailed", emailed.Action)
	assert.Equal(t, "editor", emailed.Meta["role"])
	assert.Equal(t, "example.org", emailed.Meta["domain"])
}

func TestEmailLinkLimit(t *testing.T) {
	testutil.Clean(t)
	internet := testutil.FakeInternet(t)
	testutil.FakeBrevo(internet)
	owner := insertUser(t, "Owner")
	deck := postDeck(t, owner, map[string]any{"title": "Talk"})
	path := "/slides/" + deck.Id.Hex() + "/links/viewer/email"
	for i := 0; i < 10; i++ {
		require.Equal(t, http.StatusNoContent, do(t, http.MethodPost, path, map[string]any{"email": "a@example.org"}, testutil.As(owner)).Code)
	}
	assert.Equal(t, http.StatusTooManyRequests, do(t, http.MethodPost, path, map[string]any{"email": "a@example.org"}, testutil.As(owner)).Code)
	assert.Len(t, testutil.SentEmails(t, internet), 10)
}
