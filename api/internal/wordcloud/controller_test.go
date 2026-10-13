package wordcloud

import (
	"dev/internal/testutil"
	"dev/internal/utils"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func createSessionFor(t *testing.T, owner primitive.ObjectID, change func(*WordCloud)) *WordCloud {
	t.Helper()
	now := primitive.NewDateTimeFromTime(time.Now())
	wordCloud := &WordCloud{
		UserId: owner, Name: "Retrospective", Description: "What went well?", Code: GenerateCode(), Words: []Word{},
		Settings: Settings{FilterProfanity: true, AllowUpvotes: true, Sentiment: true}, CreatedAt: now, UpdatedAt: now,
	}
	if change != nil {
		change(wordCloud)
	}
	result, err := CreateWordCloud(testutil.Context(), wordCloud)
	require.NoError(t, err)
	return storedSession(t, result.InsertedID.(primitive.ObjectID))
}

func storedSession(t *testing.T, id primitive.ObjectID) *WordCloud {
	t.Helper()
	wordCloud, err := GetWordCloudById(testutil.Context(), id)
	require.NoError(t, err)
	return wordCloud
}

func closed(wordCloud *WordCloud) {
	now := primitive.NewDateTimeFromTime(time.Now())
	wordCloud.ClosedAt = &now
}

func withWords(texts ...string) func(*WordCloud) {
	return func(wordCloud *WordCloud) {
		for _, text := range texts {
			wordCloud.Words = append(wordCloud.Words, Word{Id: primitive.NewObjectID(), Text: text, UUID: uuid.NewV4()})
		}
	}
}

func joinSession(t *testing.T, id primitive.ObjectID) string {
	t.Helper()
	response := do(t, http.MethodPost, "/word-cloud/"+id.Hex()+"/participant", nil)
	require.Equal(t, http.StatusCreated, response.Code, response.Body.String())
	return testutil.Decode[map[string]string](t, response)["token"]
}

func sendWord(t *testing.T, id primitive.ObjectID, token string, text string, options ...testutil.Option) *httptest.ResponseRecorder {
	t.Helper()
	resetLimiters()
	return do(t, http.MethodPost, "/word-cloud/"+id.Hex()+"/word", map[string]any{"text": text, "token": token}, options...)
}

func wordTexts(wordCloud *WordCloud) []string {
	texts := []string{}
	for _, word := range wordCloud.Words {
		texts = append(texts, word.Text)
	}
	return texts
}

type publicView struct {
	Id          primitive.ObjectID `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Code        string             `json:"code"`
	Open        bool               `json:"open"`
	Settings    map[string]any     `json:"settings"`
	Words       *[]string          `json:"words"`
	Recent      *[]string          `json:"recent"`
}

func TestGetWordCloudByCode(t *testing.T) {
	testutil.Clean(t)
	owner := primitive.NewObjectID()
	session := createSessionFor(t, owner, func(wordCloud *WordCloud) {
		wordCloud.Code = "ABCDE"
		wordCloud.Settings = Settings{ShowCloudToAudience: true, AllowUpvotes: true, FilterProfanity: true}
		withWords("trust", "fun", "trust")(wordCloud)
		wordCloud.Words[1].Hidden = true
	})

	response := do(t, http.MethodGet, "/word-cloud?code=abcde", nil)
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	view := testutil.Decode[publicView](t, response)
	assert.Equal(t, session.Id, view.Id)
	assert.Equal(t, "Retrospective", view.Name)
	assert.Equal(t, "ABCDE", view.Code)
	assert.True(t, view.Open)
	assert.Equal(t, map[string]any{"showCloudToAudience": true, "allowUpvotes": true, "maxWords": 20.0}, view.Settings)
	assert.Equal(t, []string{"trust", "trust"}, *view.Words, "hidden words are not shown")
	assert.Equal(t, []string{"trust"}, *view.Recent)
	for _, private := range []string{"filterProfanity", "uuid", "user", owner.Hex()} {
		assert.NotContains(t, response.Body.String(), private)
	}

	for code, status := range map[string]int{"ABCD": http.StatusBadRequest, "AB-DE": http.StatusBadRequest, "ZZZZZ": http.StatusNotFound} {
		assert.Equal(t, status, do(t, http.MethodGet, "/word-cloud?code="+code, nil).Code, code)
	}
	missing := do(t, http.MethodGet, "/word-cloud", nil)
	assert.Equal(t, http.StatusBadRequest, missing.Code)
	assert.Equal(t, "missing query parameter", testutil.ErrorOf(t, missing))

	testutil.BreakDB(t)
	assert.Equal(t, http.StatusInternalServerError, do(t, http.MethodGet, "/word-cloud?code=ABCDE", nil).Code)
}

func TestPublicSessionHidesWhatTheHostDisabled(t *testing.T) {
	testutil.Clean(t)
	noLimit := 0
	session := createSessionFor(t, primitive.NewObjectID(), func(wordCloud *WordCloud) {
		wordCloud.Settings = Settings{MaxWords: &noLimit}
		withWords("trust")(wordCloud)
	})

	view := testutil.Decode[publicView](t, do(t, http.MethodGet, "/word-cloud/"+session.Id.Hex(), nil))
	assert.Nil(t, view.Words)
	assert.Nil(t, view.Recent)
	assert.Equal(t, 0.0, view.Settings["maxWords"])
}

func TestGetWordCloudByUser(t *testing.T) {
	testutil.Clean(t)
	owner, stranger := primitive.NewObjectID(), primitive.NewObjectID()
	open := createSessionFor(t, owner, withWords("one", "two"))
	createSessionFor(t, owner, closed)
	createSessionFor(t, stranger, nil)
	path := "/word-cloud?user=" + owner.Hex()

	response := do(t, http.MethodGet, path, nil, testutil.Bearer(t, owner))
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	sessions := testutil.Decode[[]map[string]any](t, response)
	require.Len(t, sessions, 1, "open sessions by default")
	assert.Equal(t, open.Id.Hex(), sessions[0]["id"])
	assert.Equal(t, 2.0, sessions[0]["submissions"])
	assert.Equal(t, true, sessions[0]["open"])
	assert.Nil(t, sessions[0]["closedAt"])
	assert.NotContains(t, sessions[0], "words")

	closedOnes := testutil.Decode[[]map[string]any](t, do(t, http.MethodGet, path+"&status=closed", nil, testutil.Bearer(t, owner)))
	require.Len(t, closedOnes, 1)
	assert.Equal(t, false, closedOnes[0]["open"])
	assert.NotNil(t, closedOnes[0]["closedAt"])
	newcomer := primitive.NewObjectID()
	assert.JSONEq(t, `[]`, do(t, http.MethodGet, "/word-cloud?user="+newcomer.Hex(), nil, testutil.Bearer(t, newcomer)).Body.String())

	assert.Equal(t, http.StatusBadRequest, do(t, http.MethodGet, "/word-cloud?user=me", nil, testutil.Bearer(t, owner)).Code)
	assert.Equal(t, http.StatusUnauthorized, do(t, http.MethodGet, path, nil).Code)
	assert.Equal(t, http.StatusUnauthorized, do(t, http.MethodGet, path, nil, testutil.Bearer(t, stranger)).Code, "sessions of someone else")
	for _, authorization := range []string{"Bearer garbage", "Basic " + owner.Hex(), "Bearer"} {
		assert.Equal(t, http.StatusUnauthorized, do(t, http.MethodGet, path, nil, testutil.Header("Authorization", authorization)).Code, authorization)
	}
	invalid := do(t, http.MethodGet, path+"&status=archived", nil, testutil.Bearer(t, owner))
	assert.Equal(t, http.StatusBadRequest, invalid.Code)
	assert.Equal(t, "invalid status", testutil.ErrorOf(t, invalid))
}

func TestGetWordCloudById(t *testing.T) {
	testutil.Clean(t)
	owner := primitive.NewObjectID()
	session := createSessionFor(t, owner, func(wordCloud *WordCloud) {
		withWords("good", "teamwork")(wordCloud)
		wordCloud.Words[1].Id = primitive.NilObjectID
		wordCloud.Words[1].Network = "network-hash"
		wordCloud.Words[1].Client = &ClientInfo{Timezone: "Europe/Paris"}
	})
	path := "/word-cloud/" + session.Id.Hex()

	public := do(t, http.MethodGet, path, nil)
	require.Equal(t, http.StatusOK, public.Code)
	assert.Nil(t, testutil.Decode[publicView](t, public).Words, "the cloud is not shown to the audience by default")
	assert.NotContains(t, public.Body.String(), "network-hash")
	assert.Equal(t, public.Body.String(), do(t, http.MethodGet, path, nil, testutil.Bearer(t, primitive.NewObjectID())).Body.String(),
		"another user is a participant like any other")

	response := do(t, http.MethodGet, path, nil, testutil.Bearer(t, owner))
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	owned := testutil.Decode[struct {
		WordCloud
		Open bool `json:"open"`
	}](t, response)
	assert.True(t, owned.Open)
	assert.Equal(t, owner, owned.UserId)
	require.Len(t, owned.Words, 2)
	assert.Equal(t, 3, *owned.Words[0].Sentiment)
	assert.Nil(t, owned.Words[1].Sentiment, "a word that is not rated")
	assert.Equal(t, "network-hash", owned.Words[1].Network)
	assert.NotContains(t, response.Body.String(), "Europe/Paris", "raw client details stay on the server")
	assert.False(t, owned.Words[1].Id.IsZero(), "old words get an id when the host opens the session")
	assert.Equal(t, owned.Words[1].Id, storedSession(t, session.Id).Words[1].Id)

	assert.Equal(t, http.StatusBadRequest, do(t, http.MethodGet, "/word-cloud/first", nil).Code)
	missing := do(t, http.MethodGet, "/word-cloud/"+primitive.NewObjectID().Hex(), nil)
	assert.Equal(t, http.StatusNotFound, missing.Code)
	assert.Equal(t, "word cloud not found", testutil.ErrorOf(t, missing))

	testutil.BreakDB(t)
	assert.Equal(t, http.StatusInternalServerError, do(t, http.MethodGet, path, nil).Code)
}

func TestPostWordCloudParticipant(t *testing.T) {
	testutil.Clean(t)
	session := createSessionFor(t, primitive.NewObjectID(), nil)

	first, second := joinSession(t, session.Id), joinSession(t, session.Id)
	assert.NotEqual(t, first, second)
	_, err := VerifyParticipant(session.Id.Hex(), first)
	assert.NoError(t, err)

	over := createSessionFor(t, primitive.NewObjectID(), closed)
	late := do(t, http.MethodPost, "/word-cloud/"+over.Id.Hex()+"/participant", nil)
	assert.Equal(t, http.StatusConflict, late.Code)
	assert.Equal(t, "word cloud closed", testutil.ErrorOf(t, late))
	assert.Equal(t, http.StatusNotFound, do(t, http.MethodPost, "/word-cloud/"+primitive.NewObjectID().Hex()+"/participant", nil).Code)
	assert.Equal(t, http.StatusBadRequest, do(t, http.MethodPost, "/word-cloud/first/participant", nil).Code)
}

func TestPostWordCloudWord(t *testing.T) {
	testutil.Clean(t)
	session := createSessionFor(t, primitive.NewObjectID(), nil)
	token := joinSession(t, session.Id)
	participant, err := VerifyParticipant(session.Id.Hex(), token)
	require.NoError(t, err)

	resetLimiters()
	response := do(t, http.MethodPost, "/word-cloud/"+session.Id.Hex()+"/word", map[string]any{
		"text": "  Très   Bien! ", "token": token,
		"client": map[string]any{"timezone": "Europe/Paris", "languages": []string{"fr-FR"}, "platform": "Linux", "cores": 8},
	}, testutil.Header("User-Agent", "TestBrowser/1.0"), func(request *http.Request) { request.RemoteAddr = "203.0.113.7:4000" })
	require.Equal(t, http.StatusCreated, response.Code, response.Body.String())
	assert.JSONEq(t, `{"text": "tres bien"}`, response.Body.String(), "the cleaned word is returned")

	words := storedSession(t, session.Id).Words
	require.Len(t, words, 1)
	assert.Equal(t, "tres bien", words[0].Text)
	assert.False(t, words[0].Id.IsZero())
	assert.False(t, words[0].Hidden)
	assert.Equal(t, participant, words[0].UUID)
	assert.Equal(t, "TestBrowser/1.0", words[0].UserAgent)
	assert.Equal(t, "Europe/Paris", words[0].Client.Timezone)
	assert.Equal(t, Fingerprint("network", "203.0.113.7"), words[0].Network, "the address is only stored hashed")
	assert.Len(t, words[0].Device, 16)
	assert.WithinDuration(t, time.Now(), words[0].CreatedAt.Time(), 5*time.Second)

	assert.Equal(t, http.StatusCreated, sendWord(t, session.Id, token, "fun").Code)
	assert.Empty(t, storedSession(t, session.Id).Words[1].Device, "no device hash without client details")

	duplicate := sendWord(t, session.Id, token, "FUN!")
	assert.Equal(t, http.StatusConflict, duplicate.Code)
	assert.Equal(t, "word already submitted", testutil.ErrorOf(t, duplicate))
	assert.Equal(t, http.StatusCreated, sendWord(t, session.Id, joinSession(t, session.Id), "fun").Code, "someone else can send it too")
	assert.Equal(t, []string{"tres bien", "fun", "fun"}, wordTexts(storedSession(t, session.Id)))
}

func TestPostWordCloudWordRejections(t *testing.T) {
	testutil.Clean(t)
	session := createSessionFor(t, primitive.NewObjectID(), nil)
	other := createSessionFor(t, primitive.NewObjectID(), nil)
	token := joinSession(t, session.Id)
	path := "/word-cloud/" + session.Id.Hex() + "/word"

	for name, body := range map[string]any{
		"not json":       "not json",
		"no text":        map[string]any{"token": token},
		"no token":       map[string]any{"text": "word"},
		"long text":      map[string]any{"token": token, "text": strings.Repeat("a", 101)},
		"client details": map[string]any{"token": token, "text": "word", "client": map[string]any{"timezone": strings.Repeat("a", 65)}},
	} {
		assert.Equal(t, http.StatusBadRequest, do(t, http.MethodPost, path, body).Code, name)
	}
	empty := sendWord(t, session.Id, token, "!?")
	assert.Equal(t, http.StatusBadRequest, empty.Code)
	assert.Equal(t, "word is empty", testutil.ErrorOf(t, empty))

	assert.Equal(t, http.StatusUnauthorized, sendWord(t, session.Id, "forged", "word").Code)
	assert.Equal(t, http.StatusUnauthorized, sendWord(t, session.Id, joinSession(t, other.Id), "word").Code, "token of another session")
	assert.Equal(t, http.StatusNotFound, sendWord(t, primitive.NewObjectID(), token, "word").Code)
	assert.Equal(t, http.StatusBadRequest, do(t, http.MethodPost, "/word-cloud/first/word", nil).Code)

	over := createSessionFor(t, primitive.NewObjectID(), nil)
	lateToken := joinSession(t, over.Id)
	require.NoError(t, UpdateWordCloud(testutil.Context(), over.Id, map[string]any{"closedAt": time.Now()}, nil))
	late := sendWord(t, over.Id, lateToken, "word")
	assert.Equal(t, http.StatusConflict, late.Code)
	assert.Equal(t, "word cloud closed", testutil.ErrorOf(t, late))

	assert.Empty(t, storedSession(t, session.Id).Words)
}

func TestPostWordCloudWordLimits(t *testing.T) {
	testutil.Clean(t)
	two := 2
	limited := createSessionFor(t, primitive.NewObjectID(), func(wordCloud *WordCloud) { wordCloud.Settings.MaxWords = &two })
	token := joinSession(t, limited.Id)

	assert.Equal(t, http.StatusCreated, sendWord(t, limited.Id, token, "one").Code)
	assert.Equal(t, http.StatusCreated, sendWord(t, limited.Id, token, "two").Code)
	refused := sendWord(t, limited.Id, token, "three")
	assert.Equal(t, http.StatusForbidden, refused.Code)
	assert.Equal(t, "word limit reached", testutil.ErrorOf(t, refused))
	assert.Equal(t, http.StatusCreated, sendWord(t, limited.Id, joinSession(t, limited.Id), "three").Code, "the limit is per participant")

	zero := 0
	unlimited := createSessionFor(t, primitive.NewObjectID(), func(wordCloud *WordCloud) { wordCloud.Settings.MaxWords = &zero })
	token = joinSession(t, unlimited.Id)
	for i := range DefaultMaxWords + 1 {
		assert.Equal(t, http.StatusCreated, sendWord(t, unlimited.Id, token, "word"+strings.Repeat("a", i)).Code)
	}

	resetLimiters()
	path := "/word-cloud/" + unlimited.Id.Hex() + "/word"
	assert.Equal(t, http.StatusCreated, do(t, http.MethodPost, path, map[string]any{"text": "quick", "token": token}).Code)
	tooFast := do(t, http.MethodPost, path, map[string]any{"text": "quicker", "token": token})
	assert.Equal(t, http.StatusTooManyRequests, tooFast.Code, "one word per second and per participant")
	assert.Equal(t, "1", tooFast.Header().Get("Retry-After"))
	assert.Equal(t, http.StatusCreated, do(t, http.MethodPost, path, map[string]any{"text": "quicker", "token": joinSession(t, unlimited.Id)}).Code)
}

func TestPostWordCloudWordProfanity(t *testing.T) {
	testutil.Clean(t)
	filtered := createSessionFor(t, primitive.NewObjectID(), nil)
	unfiltered := createSessionFor(t, primitive.NewObjectID(), func(wordCloud *WordCloud) { wordCloud.Settings.FilterProfanity = false })

	for _, session := range []*WordCloud{filtered, unfiltered} {
		token := joinSession(t, session.Id)
		assert.Equal(t, http.StatusCreated, sendWord(t, session.Id, token, "Merde").Code, "the participant is not told")
		assert.Equal(t, http.StatusCreated, sendWord(t, session.Id, token, "classic").Code)
	}

	words := storedSession(t, filtered.Id).Words
	assert.True(t, words[0].Hidden, "profanity waits for the host")
	assert.False(t, words[1].Hidden)
	assert.False(t, storedSession(t, unfiltered.Id).Words[0].Hidden)
}

func TestPostWordCloud(t *testing.T) {
	testutil.Clean(t)
	owner := primitive.NewObjectID()

	response := do(t, http.MethodPost, "/word-cloud", map[string]any{"name": "Retrospective", "description": "What went well?"}, testutil.As(owner))
	require.Equal(t, http.StatusCreated, response.Code, response.Body.String())
	session := storedSession(t, testutil.Decode[map[string]primitive.ObjectID](t, response)["id"])
	assert.Equal(t, owner, session.UserId)
	assert.Equal(t, "Retrospective", session.Name)
	assert.Equal(t, "What went well?", session.Description)
	assert.NoError(t, ValidateCode(session.Code))
	assert.Empty(t, session.Words)
	assert.Nil(t, session.ClosedAt)
	assert.Equal(t, Settings{FilterProfanity: true, AllowUpvotes: true, Sentiment: true}, session.Settings, "defaults")
	assert.WithinDuration(t, time.Now(), session.CreatedAt.Time(), 5*time.Second)

	response = do(t, http.MethodPost, "/word-cloud", map[string]any{"name": "Quiet", "settings": map[string]any{"showCloudToAudience": true, "maxWords": 3}}, testutil.As(owner))
	require.Equal(t, http.StatusCreated, response.Code)
	custom := storedSession(t, testutil.Decode[map[string]primitive.ObjectID](t, response)["id"])
	assert.True(t, custom.Settings.ShowCloudToAudience)
	assert.False(t, custom.Settings.FilterProfanity, "given settings replace the defaults")
	assert.Equal(t, 3, custom.Settings.WordLimit())
	assert.NotEqual(t, session.Code, custom.Code)

	for name, body := range map[string]any{
		"not json":          "not json",
		"no name":           map[string]any{"description": "Long enough text"},
		"short name":        map[string]any{"name": "ab"},
		"long name":         map[string]any{"name": strings.Repeat("a", 101)},
		"short description": map[string]any{"name": "Valid", "description": "short"},
		"too many words":    map[string]any{"name": "Valid", "settings": map[string]any{"maxWords": 101}},
		"negative words":    map[string]any{"name": "Valid", "settings": map[string]any{"maxWords": -1}},
	} {
		assert.Equal(t, http.StatusBadRequest, do(t, http.MethodPost, "/word-cloud", body, testutil.As(owner)).Code, name)
	}
	assert.Equal(t, http.StatusUnauthorized, do(t, http.MethodPost, "/word-cloud", map[string]any{"name": "Anonymous"}).Code)
	assert.Equal(t, 2, testutil.Count(t, "word_cloud_sessions"))

	testutil.BreakDB(t)
	assert.Equal(t, http.StatusInternalServerError, do(t, http.MethodPost, "/word-cloud", map[string]any{"name": "Valid"}, testutil.As(owner)).Code)
}

func TestPostWordCloudDuplicate(t *testing.T) {
	testutil.Clean(t)
	owner := primitive.NewObjectID()
	five := 5
	original := createSessionFor(t, owner, func(wordCloud *WordCloud) {
		wordCloud.Settings = Settings{ShowCloudToAudience: true, MaxWords: &five}
		withWords("kept", "here")(wordCloud)
		closed(wordCloud)
	})
	path := "/word-cloud/" + original.Id.Hex() + "/duplicate"

	response := do(t, http.MethodPost, path, nil, testutil.As(owner))
	require.Equal(t, http.StatusCreated, response.Code, response.Body.String())
	copied := storedSession(t, testutil.Decode[map[string]primitive.ObjectID](t, response)["id"])
	assert.NotEqual(t, original.Id, copied.Id)
	assert.Equal(t, original.Name, copied.Name)
	assert.Equal(t, original.Description, copied.Description)
	assert.Equal(t, original.Settings, copied.Settings)
	assert.Equal(t, owner, copied.UserId)
	assert.Empty(t, copied.Words, "the copy starts empty")
	assert.Nil(t, copied.ClosedAt, "and open")
	assert.NotEqual(t, original.Code, copied.Code)
	assert.Len(t, storedSession(t, original.Id).Words, 2)

	forbidden := do(t, http.MethodPost, path, nil, testutil.As(primitive.NewObjectID()))
	assert.Equal(t, http.StatusForbidden, forbidden.Code)
	assert.Equal(t, "forbidden", testutil.ErrorOf(t, forbidden))
	assert.Equal(t, http.StatusNotFound, do(t, http.MethodPost, "/word-cloud/"+primitive.NewObjectID().Hex()+"/duplicate", nil, testutil.As(owner)).Code)
	assert.Equal(t, http.StatusBadRequest, do(t, http.MethodPost, "/word-cloud/first/duplicate", nil, testutil.As(owner)).Code)
}

func TestPatchWordCloud(t *testing.T) {
	testutil.Clean(t)
	owner := primitive.NewObjectID()
	session := createSessionFor(t, owner, nil)
	path := "/word-cloud/" + session.Id.Hex()

	response := do(t, http.MethodPatch, path, map[string]any{
		"name": "Renamed", "description": "", "settings": map[string]any{"showCloudToAudience": true, "maxWords": 0},
	}, testutil.As(owner))
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	view := testutil.Decode[publicView](t, response)
	assert.Equal(t, "Renamed", view.Name)
	assert.NotNil(t, view.Words)
	stored := storedSession(t, session.Id)
	assert.Equal(t, "Renamed", stored.Name)
	assert.Empty(t, stored.Description)
	assert.True(t, stored.Settings.ShowCloudToAudience)
	assert.False(t, stored.Settings.AllowUpvotes, "settings are replaced as a whole")
	assert.Equal(t, 0, stored.Settings.WordLimit())
	assert.Equal(t, session.Code, stored.Code)

	assert.Equal(t, http.StatusOK, do(t, http.MethodPatch, path, map[string]any{}, testutil.As(owner)).Code)
	assert.Equal(t, "Renamed", storedSession(t, session.Id).Name, "an empty patch changes nothing")

	for name, body := range map[string]any{
		"not json":          "not json",
		"short name":        map[string]any{"name": "ab"},
		"short description": map[string]any{"description": "short"},
		"long description":  map[string]any{"description": strings.Repeat("a", 1001)},
		"too many words":    map[string]any{"settings": map[string]any{"maxWords": 101}},
	} {
		assert.Equal(t, http.StatusBadRequest, do(t, http.MethodPatch, path, body, testutil.As(owner)).Code, name)
	}
	assert.Equal(t, http.StatusForbidden, do(t, http.MethodPatch, path, map[string]any{"name": "Stolen"}, testutil.As(primitive.NewObjectID())).Code)
	assert.Equal(t, http.StatusNotFound, do(t, http.MethodPatch, "/word-cloud/"+primitive.NewObjectID().Hex(), map[string]any{}, testutil.As(owner)).Code)
	assert.Equal(t, "Renamed", storedSession(t, session.Id).Name)
}

func TestPatchWordCloudClosesAndReopens(t *testing.T) {
	testutil.Clean(t)
	owner := primitive.NewObjectID()
	session := createSessionFor(t, owner, nil)
	path := "/word-cloud/" + session.Id.Hex()

	response := do(t, http.MethodPatch, path, map[string]any{"open": false}, testutil.As(owner))
	require.Equal(t, http.StatusOK, response.Code)
	assert.False(t, testutil.Decode[publicView](t, response).Open)
	closedAt := storedSession(t, session.Id).ClosedAt
	require.NotNil(t, closedAt)
	assert.Equal(t, http.StatusNotFound, do(t, http.MethodGet, "/word-cloud?code="+session.Code, nil).Code, "a closed session cannot be joined")

	assert.Equal(t, http.StatusOK, do(t, http.MethodPatch, path, map[string]any{"open": false}, testutil.As(owner)).Code)
	assert.Equal(t, closedAt, storedSession(t, session.Id).ClosedAt, "closing twice keeps the first date")

	response = do(t, http.MethodPatch, path, map[string]any{"open": true}, testutil.As(owner))
	require.Equal(t, http.StatusOK, response.Code)
	reopened := testutil.Decode[publicView](t, response)
	assert.True(t, reopened.Open)
	assert.Equal(t, session.Code, reopened.Code, "the code is kept when it is still free")
	assert.Nil(t, storedSession(t, session.Id).ClosedAt)

	require.Equal(t, http.StatusOK, do(t, http.MethodPatch, path, map[string]any{"open": false}, testutil.As(owner)).Code)
	createSessionFor(t, primitive.NewObjectID(), func(wordCloud *WordCloud) { wordCloud.Code = session.Code })
	response = do(t, http.MethodPatch, path, map[string]any{"open": true}, testutil.As(owner))
	require.Equal(t, http.StatusOK, response.Code)
	renewed := testutil.Decode[publicView](t, response).Code
	assert.NotEqual(t, session.Code, renewed, "a code taken meanwhile is replaced")
	assert.NoError(t, ValidateCode(renewed))
	assert.Equal(t, renewed, storedSession(t, session.Id).Code)
}

func TestDeleteWordCloud(t *testing.T) {
	testutil.Clean(t)
	owner := primitive.NewObjectID()
	session := createSessionFor(t, owner, nil)
	path := "/word-cloud/" + session.Id.Hex()

	assert.Equal(t, http.StatusForbidden, do(t, http.MethodDelete, path, nil, testutil.As(primitive.NewObjectID())).Code)
	assert.Equal(t, 1, testutil.Count(t, "word_cloud_sessions"))

	deleted := do(t, http.MethodDelete, path, nil, testutil.As(owner))
	assert.Equal(t, http.StatusNoContent, deleted.Code)
	assert.Zero(t, testutil.Count(t, "word_cloud_sessions"))
	assert.Equal(t, http.StatusNotFound, do(t, http.MethodDelete, path, nil, testutil.As(owner)).Code)
}

func TestPatchWord(t *testing.T) {
	testutil.Clean(t)
	owner := primitive.NewObjectID()
	session := createSessionFor(t, owner, withWords("good", "bad"))
	path := "/word-cloud/" + session.Id.Hex() + "/words/"
	target := session.Words[1].Id.Hex()

	response := do(t, http.MethodPatch, path+target, map[string]any{"hidden": true}, testutil.As(owner))
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	words := testutil.Decode[map[string][]Word](t, response)["words"]
	require.Len(t, words, 2)
	assert.False(t, words[0].Hidden)
	assert.True(t, words[1].Hidden)
	assert.Equal(t, -3, *words[1].Sentiment)
	assert.True(t, storedSession(t, session.Id).Words[1].Hidden)

	require.Equal(t, http.StatusOK, do(t, http.MethodPatch, path+target, map[string]any{"hidden": false}, testutil.As(owner)).Code)
	assert.False(t, storedSession(t, session.Id).Words[1].Hidden)

	assert.Equal(t, http.StatusBadRequest, do(t, http.MethodPatch, path+"second", map[string]any{"hidden": true}, testutil.As(owner)).Code)
	assert.Equal(t, http.StatusBadRequest, do(t, http.MethodPatch, path+target, "not json", testutil.As(owner)).Code)
	missing := do(t, http.MethodPatch, path+primitive.NewObjectID().Hex(), map[string]any{"hidden": true}, testutil.As(owner))
	assert.Equal(t, http.StatusNotFound, missing.Code)
	assert.Equal(t, "word not found", testutil.ErrorOf(t, missing))
	assert.Equal(t, http.StatusForbidden, do(t, http.MethodPatch, path+target, map[string]any{"hidden": true}, testutil.As(primitive.NewObjectID())).Code)
	assert.False(t, storedSession(t, session.Id).Words[1].Hidden)
}

func TestPostMergeWords(t *testing.T) {
	testutil.Clean(t)
	owner := primitive.NewObjectID()
	session := createSessionFor(t, owner, withWords("colour", "couleur", "shape"))
	path := "/word-cloud/" + session.Id.Hex() + "/words/merge"

	response := do(t, http.MethodPost, path, map[string]any{"from": []string{"colour", "couleur"}, "into": " Color! "}, testutil.As(owner))
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	words := testutil.Decode[map[string][]Word](t, response)["words"]
	assert.Equal(t, []string{"color", "color", "shape"}, wordTexts(&WordCloud{Words: words}), "the target is cleaned like a submission")
	assert.Equal(t, "colour", words[0].OriginalText)
	assert.Equal(t, []string{"color", "color", "shape"}, wordTexts(storedSession(t, session.Id)))

	many := make([]string, 501)
	for i := range many {
		many[i] = "word"
	}
	for name, body := range map[string]any{
		"not json":    "not json",
		"no source":   map[string]any{"into": "color"},
		"empty list":  map[string]any{"from": []string{}, "into": "color"},
		"empty word":  map[string]any{"from": []string{""}, "into": "color"},
		"too many":    map[string]any{"from": many, "into": "color"},
		"no target":   map[string]any{"from": []string{"shape"}},
		"long target": map[string]any{"from": []string{"shape"}, "into": strings.Repeat("a", 101)},
	} {
		assert.Equal(t, http.StatusBadRequest, do(t, http.MethodPost, path, body, testutil.As(owner)).Code, name)
	}
	symbols := do(t, http.MethodPost, path, map[string]any{"from": []string{"shape"}, "into": "?!"}, testutil.As(owner))
	assert.Equal(t, http.StatusBadRequest, symbols.Code)
	assert.Equal(t, "word is empty", testutil.ErrorOf(t, symbols))
	assert.Equal(t, http.StatusForbidden, do(t, http.MethodPost, path, map[string]any{"from": []string{"shape"}, "into": "form"}, testutil.As(primitive.NewObjectID())).Code)
	assert.Equal(t, []string{"color", "color", "shape"}, wordTexts(storedSession(t, session.Id)))
}

func TestOwnerRoutesFailWithoutDatabase(t *testing.T) {
	owner := primitive.NewObjectID()
	testutil.BreakDB(t)
	id := primitive.NewObjectID().Hex()
	for _, route := range []string{
		"PATCH /word-cloud/" + id, "DELETE /word-cloud/" + id, "POST /word-cloud/" + id + "/duplicate",
		"PATCH /word-cloud/" + id + "/words/" + id, "POST /word-cloud/" + id + "/words/merge",
		"POST /word-cloud/" + id + "/participant", "POST /word-cloud/" + id + "/word",
	} {
		method, path, _ := strings.Cut(route, " ")
		assert.Equal(t, http.StatusInternalServerError, do(t, method, path, map[string]any{}, testutil.As(owner)).Code, route)
	}
	assert.Equal(t, http.StatusInternalServerError, do(t, http.MethodGet, "/word-cloud?user="+owner.Hex(), nil, testutil.Bearer(t, owner)).Code)
}

type socket struct {
	t    *testing.T
	conn *websocket.Conn
}

func openSocket(t *testing.T, server *httptest.Server, id primitive.ObjectID) *socket {
	t.Helper()
	url := "ws" + strings.TrimPrefix(server.URL, "http") + "/word-cloud/" + id.Hex() + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })
	return &socket{t: t, conn: conn}
}

func (socket *socket) authenticate(token string) map[string]any {
	socket.t.Helper()
	require.NoError(socket.t, socket.conn.WriteJSON(gin.H{"type": "auth", "token": token}))
	return socket.next()
}

func (socket *socket) next() map[string]any {
	socket.t.Helper()
	var event map[string]any
	socket.conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	require.NoError(socket.t, socket.conn.ReadJSON(&event))
	return event
}

func accessToken(t *testing.T, userId primitive.ObjectID) string {
	t.Helper()
	token, err := utils.SignAccessToken(userId.Hex(), 1)
	require.NoError(t, err)
	return token
}

func TestWordCloudSocket(t *testing.T) {
	testutil.Clean(t)
	server := httptest.NewServer(newEngine())
	defer server.Close()
	owner := primitive.NewObjectID()
	session := createSessionFor(t, owner, nil)

	host := openSocket(t, server, session.Id)
	assert.Equal(t, map[string]any{"type": "ready", "owner": true}, host.authenticate(accessToken(t, owner)))
	audience := openSocket(t, server, session.Id)
	assert.Equal(t, map[string]any{"type": "ready", "owner": false}, audience.authenticate(""))
	impostor := openSocket(t, server, session.Id)
	assert.Equal(t, map[string]any{"type": "ready", "owner": false}, impostor.authenticate(accessToken(t, primitive.NewObjectID())),
		"only the owner's token unlocks owner events")

	token := joinSession(t, session.Id)
	require.Equal(t, http.StatusCreated, sendWord(t, session.Id, token, "Good").Code)
	event := host.next()
	assert.Equal(t, "word", event["type"])
	word := event["word"].(map[string]any)
	assert.Equal(t, "good", word["text"])
	assert.Equal(t, 3.0, word["sentiment"])
	assert.NotEmpty(t, word["uuid"], "the host can tell participants apart")
	assert.Equal(t, map[string]any{"type": "word", "word": map[string]any{"text": "good"}}, audience.next(), "the audience only gets the text")
	impostor.next()

	require.Equal(t, http.StatusCreated, sendWord(t, session.Id, token, "merde").Code)
	assert.Equal(t, true, host.next()["word"].(map[string]any)["hidden"])

	require.Equal(t, http.StatusOK, do(t, http.MethodPatch, "/word-cloud/"+session.Id.Hex(), map[string]any{"name": "Renamed"}, testutil.As(owner)).Code)
	event = host.next()
	assert.Equal(t, "session", event["type"])
	assert.Equal(t, "Renamed", event["session"].(map[string]any)["name"])
	assert.Len(t, event["words"], 2)
	assert.Contains(t, event, "settings")
	event = audience.next()
	assert.Equal(t, "session", event["type"], "the hidden word was not sent to the audience")
	assert.Equal(t, "Renamed", event["session"].(map[string]any)["name"])
	assert.NotContains(t, event, "words")
	assert.NotContains(t, event, "settings")
	impostor.next()

	wordId := storedSession(t, session.Id).Words[0].Id.Hex()
	require.Equal(t, http.StatusOK, do(t, http.MethodPatch, "/word-cloud/"+session.Id.Hex()+"/words/"+wordId, map[string]any{"hidden": true}, testutil.As(owner)).Code)
	assert.Equal(t, true, host.next()["words"].([]any)[0].(map[string]any)["hidden"])
	assert.Equal(t, []any{}, audience.next()["session"].(map[string]any)["recent"], "a hidden word leaves the suggestions")
	impostor.next()

	impostor.conn.Close()
	assert.Eventually(t, func() bool {
		socketMutex.Lock()
		defer socketMutex.Unlock()
		return len(socketClients[session.Id]) == 2
	}, 3*time.Second, 10*time.Millisecond, "a closed socket is forgotten")

	require.Equal(t, http.StatusNoContent, do(t, http.MethodDelete, "/word-cloud/"+session.Id.Hex(), nil, testutil.As(owner)).Code)
	assert.Equal(t, map[string]any{"type": "session", "session": nil}, host.next())
	assert.Equal(t, map[string]any{"type": "session", "session": nil}, audience.next())

	host.conn.Close()
	audience.conn.Close()
	assert.Eventually(t, func() bool {
		socketMutex.Lock()
		defer socketMutex.Unlock()
		_, registered := socketClients[session.Id]
		return !registered
	}, 3*time.Second, 10*time.Millisecond)
}

func TestWordCloudSocketWithoutPublicEvents(t *testing.T) {
	testutil.Clean(t)
	server := httptest.NewServer(newEngine())
	defer server.Close()
	owner := primitive.NewObjectID()
	session := createSessionFor(t, owner, func(wordCloud *WordCloud) { wordCloud.Settings = Settings{} })

	host := openSocket(t, server, session.Id)
	host.authenticate(accessToken(t, owner))
	audience := openSocket(t, server, session.Id)
	audience.authenticate("")

	require.Equal(t, http.StatusCreated, sendWord(t, session.Id, joinSession(t, session.Id), "good").Code)
	word := host.next()["word"].(map[string]any)
	assert.NotContains(t, word, "sentiment", "sentiment turned off")

	require.Equal(t, http.StatusOK, do(t, http.MethodPatch, "/word-cloud/"+session.Id.Hex(), map[string]any{"open": false}, testutil.As(owner)).Code)
	event := audience.next()
	assert.Equal(t, "session", event["type"], "words are not sent to the audience when neither the cloud nor upvotes are on")
	assert.Equal(t, false, event["session"].(map[string]any)["open"])
}

func TestWordCloudSocketRejections(t *testing.T) {
	testutil.Clean(t)
	server := httptest.NewServer(newEngine())
	defer server.Close()
	session := createSessionFor(t, primitive.NewObjectID(), nil)
	base := "ws" + strings.TrimPrefix(server.URL, "http") + "/word-cloud/"

	_, response, err := websocket.DefaultDialer.Dial(base+primitive.NewObjectID().Hex()+"/ws", nil)
	require.Error(t, err)
	assert.Equal(t, http.StatusNotFound, response.StatusCode)
	_, response, err = websocket.DefaultDialer.Dial(base+"first/ws", nil)
	require.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
	assert.Equal(t, http.StatusBadRequest, do(t, http.MethodGet, "/word-cloud/"+session.Id.Hex()+"/ws", nil).Code, "not a WebSocket handshake")

	for _, first := range []any{gin.H{"type": "hello"}, "not an object"} {
		impatient := openSocket(t, server, session.Id)
		require.NoError(t, impatient.conn.WriteJSON(first))
		impatient.conn.SetReadDeadline(time.Now().Add(3 * time.Second))
		_, _, err = impatient.conn.ReadMessage()
		assert.Error(t, err, "the connection is closed when the first message is not an authentication")
	}
	socketMutex.Lock()
	assert.Empty(t, socketClients[session.Id])
	socketMutex.Unlock()
}
