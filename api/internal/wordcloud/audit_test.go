package wordcloud

import (
	"net/http"
	"testing"

	"dev/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestOwnerActionsAreAudited(t *testing.T) {
	testutil.Clean(t)
	owner := primitive.NewObjectID()

	created := do(t, http.MethodPost, "/word-cloud", map[string]any{"name": "Kick-off", "description": "What do you expect?"}, testutil.As(owner))
	require.Equal(t, http.StatusCreated, created.Code, created.Body.String())
	entries := testutil.Audit(t)
	require.Equal(t, []string{"wordcloud.created"}, testutil.AuditActions(t))
	id, err := primitive.ObjectIDFromHex(entries[0].Target.Id)
	require.NoError(t, err)
	session := storedSession(t, id)
	assert.Equal(t, owner, entries[0].UserId)
	assert.Equal(t, "wordcloud", entries[0].Target.Type)
	assert.Equal(t, "Kick-off", entries[0].Meta["name"])
	path := "/word-cloud/" + id.Hex()

	require.Equal(t, http.StatusOK, do(t, http.MethodPatch, path, map[string]any{"open": true}, testutil.As(owner)).Code)
	assert.Len(t, testutil.Audit(t), 1, "the session was already open")
	require.Equal(t, http.StatusOK, do(t, http.MethodPatch, path, map[string]any{"name": "Kick-off 2", "open": false}, testutil.As(owner)).Code)
	entries = testutil.Audit(t)
	require.Len(t, entries, 2)
	assert.Equal(t, "wordcloud.updated", entries[1].Action)
	assert.Equal(t, false, entries[1].Meta["open"])
	assert.ElementsMatch(t, []any{"name"}, entries[1].Meta["fields"])
	assert.Equal(t, id.Hex(), entries[1].Target.Id)

	duplicated := do(t, http.MethodPost, path+"/duplicate", nil, testutil.As(owner))
	require.Equal(t, http.StatusCreated, duplicated.Code, duplicated.Body.String())
	entries = testutil.Audit(t)
	require.Len(t, entries, 3)
	assert.Equal(t, "wordcloud.duplicated", entries[2].Action)
	assert.Equal(t, id.Hex(), entries[2].Meta["from"])
	assert.NotEqual(t, id.Hex(), entries[2].Target.Id)

	stranger := primitive.NewObjectID()
	require.Equal(t, http.StatusForbidden, do(t, http.MethodDelete, path, nil, testutil.As(stranger)).Code)
	entries = testutil.Audit(t)
	require.Len(t, entries, 4)
	assert.Equal(t, "access.denied", entries[3].Action)
	assert.Equal(t, stranger, entries[3].UserId)
	assert.Equal(t, id.Hex(), entries[3].Target.Id)

	require.Equal(t, http.StatusNoContent, do(t, http.MethodDelete, path, nil, testutil.As(owner)).Code)
	entries = testutil.Audit(t)
	require.Len(t, entries, 5)
	assert.Equal(t, "wordcloud.deleted", entries[4].Action)
	assert.Equal(t, session.UserId, entries[4].UserId)
	assert.Equal(t, "Kick-off 2", entries[4].Meta["name"])
}

func TestModerationIsAudited(t *testing.T) {
	testutil.Clean(t)
	owner := primitive.NewObjectID()
	session := createSessionFor(t, owner, withWords("svelte", "sveltekit"))
	path := "/word-cloud/" + session.Id.Hex() + "/words/"
	owned, err := OwnerView(ctx, session)
	require.NoError(t, err)
	wordId := owned.Words[0].Id.Hex()

	require.Equal(t, http.StatusOK, do(t, http.MethodPatch, path+wordId, map[string]any{"hidden": true}, testutil.As(owner)).Code)
	require.Equal(t, http.StatusNotFound, do(t, http.MethodPatch, path+primitive.NewObjectID().Hex(), map[string]any{"hidden": true}, testutil.As(owner)).Code)
	require.Equal(t, http.StatusOK, do(t, http.MethodPost, path+"merge", map[string]any{"from": []string{"sveltekit"}, "into": "svelte"}, testutil.As(owner)).Code)

	entries := testutil.Audit(t)
	require.Equal(t, []string{"wordcloud.word_hidden", "wordcloud.words_merged"}, testutil.AuditActions(t))
	assert.Equal(t, wordId, entries[0].Meta["wordId"])
	assert.Equal(t, true, entries[0].Meta["hidden"])
	assert.Equal(t, "svelte", entries[1].Meta["into"])
	assert.ElementsMatch(t, []any{"sveltekit"}, entries[1].Meta["from"])
	for _, entry := range entries {
		assert.Equal(t, owner, entry.UserId)
		assert.Equal(t, session.Id.Hex(), entry.Target.Id)
	}
}

func TestAudienceIsNotAudited(t *testing.T) {
	testutil.Clean(t)
	resetLimiters()
	session := createSessionFor(t, primitive.NewObjectID(), nil)

	token := joinSession(t, session.Id)
	require.Equal(t, http.StatusCreated, sendWord(t, session.Id, token, "svelte").Code)
	require.Equal(t, http.StatusOK, do(t, http.MethodGet, "/word-cloud/"+session.Id.Hex(), nil).Code)

	assert.Empty(t, testutil.Audit(t))
}
