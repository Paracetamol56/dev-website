package models

import (
	"testing"
	"time"

	"dev/internal/testutil"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func createWordCloud(t *testing.T, wordCloud WordCloud) *WordCloud {
	t.Helper()
	if wordCloud.Words == nil {
		wordCloud.Words = []Word{}
	}
	result, err := CreateWordCloud(ginContext(), &wordCloud)
	require.NoError(t, err)
	return getWordCloud(t, result.InsertedID.(primitive.ObjectID))
}

func getWordCloud(t *testing.T, id primitive.ObjectID) *WordCloud {
	t.Helper()
	wordCloud, err := GetWordCloudById(ginContext(), id)
	require.NoError(t, err)
	return wordCloud
}

func texts(wordCloud *WordCloud) []string {
	result := []string{}
	for _, word := range wordCloud.Words {
		result = append(result, word.Text)
	}
	return result
}

func closedAt() *primitive.DateTime {
	now := primitive.NewDateTimeFromTime(time.Now())
	return &now
}

func TestWordLimit(t *testing.T) {
	limit := func(n int) *int { return &n }
	assert.Equal(t, DefaultMaxWords, Settings{}.WordLimit())
	assert.Equal(t, 0, Settings{MaxWords: limit(0)}.WordLimit())
	assert.Equal(t, 5, Settings{MaxWords: limit(5)}.WordLimit())
}

func TestFindWordClouds(t *testing.T) {
	testutil.Clean(t)
	userId, otherUser := primitive.NewObjectID(), primitive.NewObjectID()
	open := createWordCloud(t, WordCloud{UserId: userId, Name: "Open", Code: "ABCDE", Settings: Settings{AllowUpvotes: true}})
	closed := createWordCloud(t, WordCloud{UserId: userId, Name: "Closed", Code: "FGHJK", ClosedAt: closedAt()})
	createWordCloud(t, WordCloud{UserId: otherUser, Name: "Other", Code: "MNPQR"})

	byCode, err := GetWordCloudByCode(ginContext(), "abcde")
	require.NoError(t, err, "codes are matched without case")
	assert.Equal(t, open, byCode)
	_, err = GetWordCloudByCode(ginContext(), "FGHJK")
	assert.ErrorIs(t, err, mongo.ErrNoDocuments, "a closed session cannot be joined")
	_, err = GetWordCloudByCode(ginContext(), "ZZZZZ")
	assert.ErrorIs(t, err, mongo.ErrNoDocuments)

	for code, expected := range map[string]bool{"ABCDE": true, "abcde": true, "FGHJK": false, "ZZZZZ": false} {
		inUse, err := IsCodeInUse(ginContext(), code)
		require.NoError(t, err)
		assert.Equal(t, expected, inUse, code)
	}

	assert.Equal(t, closed, getWordCloud(t, closed.Id))
	_, err = GetWordCloudById(ginContext(), primitive.NewObjectID())
	assert.ErrorIs(t, err, mongo.ErrNoDocuments)

	opened, err := GetWordCloudByUser(ginContext(), userId, "open")
	require.NoError(t, err)
	assert.Equal(t, []*WordCloud{open}, opened)
	closedOnes, err := GetWordCloudByUser(ginContext(), userId, "closed")
	require.NoError(t, err)
	assert.Equal(t, []*WordCloud{closed}, closedOnes)
	none, err := GetWordCloudByUser(ginContext(), primitive.NewObjectID(), "open")
	require.NoError(t, err)
	assert.Empty(t, none)
	_, err = GetWordCloudByUser(ginContext(), userId, "archived")
	assert.EqualError(t, err, "invalid status")
}

func TestAddWordToWordCloud(t *testing.T) {
	testutil.Clean(t)
	session := createWordCloud(t, WordCloud{Name: "Session", Code: "ABCDE"})
	closed := createWordCloud(t, WordCloud{Name: "Closed", Code: "FGHJK", ClosedAt: closedAt()})
	ada, grace := uuid.NewV4(), uuid.NewV4()
	word := func(text string, participant uuid.UUID) *Word {
		return &Word{Id: primitive.NewObjectID(), Text: text, UUID: participant, CreatedAt: primitive.NewDateTimeFromTime(time.Now()),
			Client: &ClientInfo{Timezone: "Europe/Paris", Languages: []string{"fr"}}, Network: "hash"}
	}

	require.NoError(t, AddWordToWordCloud(ginContext(), session.Id, word("trust", ada)))
	require.NoError(t, AddWordToWordCloud(ginContext(), session.Id, word("trust", grace)), "the same word from someone else is an upvote")
	require.NoError(t, AddWordToWordCloud(ginContext(), session.Id, word("fun", ada)))
	assert.ErrorIs(t, AddWordToWordCloud(ginContext(), session.Id, word("trust", ada)), ErrWordNotAdded, "a participant sends a word once")
	assert.ErrorIs(t, AddWordToWordCloud(ginContext(), closed.Id, word("late", ada)), ErrWordNotAdded)
	assert.ErrorIs(t, AddWordToWordCloud(ginContext(), primitive.NewObjectID(), word("lost", ada)), ErrWordNotAdded)

	stored := getWordCloud(t, session.Id)
	assert.Equal(t, []string{"trust", "trust", "fun"}, texts(stored))
	assert.Equal(t, ada, stored.Words[0].UUID)
	assert.Equal(t, "Europe/Paris", stored.Words[0].Client.Timezone)
	assert.Empty(t, getWordCloud(t, closed.Id).Words)
}

func TestUpdateAndDeleteWordCloud(t *testing.T) {
	testutil.Clean(t)
	created := primitive.NewDateTimeFromTime(time.Now().Add(-time.Hour))
	session := createWordCloud(t, WordCloud{Name: "Session", Code: "ABCDE", ClosedAt: closedAt(), CreatedAt: created, UpdatedAt: created})

	require.NoError(t, UpdateWordCloud(ginContext(), session.Id, bson.M{"name": "Renamed", "code": "FGHJK"}, bson.M{"closedAt": ""}))
	updated := getWordCloud(t, session.Id)
	assert.Equal(t, "Renamed", updated.Name)
	assert.Equal(t, "FGHJK", updated.Code)
	assert.Nil(t, updated.ClosedAt)
	assert.Equal(t, created, updated.CreatedAt)
	assert.Greater(t, updated.UpdatedAt, created, "updates are dated")

	require.NoError(t, UpdateWordCloud(ginContext(), session.Id, nil, nil))
	assert.Equal(t, "Renamed", getWordCloud(t, session.Id).Name)

	require.NoError(t, DeleteWordCloud(ginContext(), session.Id))
	_, err := GetWordCloudById(ginContext(), session.Id)
	assert.ErrorIs(t, err, mongo.ErrNoDocuments)
	require.NoError(t, DeleteWordCloud(ginContext(), session.Id), "deleting twice is not an error")
}

func TestEnsureWordIds(t *testing.T) {
	testutil.Clean(t)
	existing := primitive.NewObjectID()
	session := createWordCloud(t, WordCloud{Name: "Old session", Code: "ABCDE", Words: []Word{{Text: "old"}, {Id: existing, Text: "new"}}})

	require.NoError(t, EnsureWordIds(ginContext(), session))
	assert.False(t, session.Words[0].Id.IsZero())
	assert.Equal(t, existing, session.Words[1].Id)
	assert.Equal(t, session.Words, getWordCloud(t, session.Id).Words, "the ids are stored")

	stored := session.Words[0].Id
	require.NoError(t, EnsureWordIds(ginContext(), session))
	assert.Equal(t, stored, session.Words[0].Id, "nothing changes once every word has an id")

	stale := createWordCloud(t, WordCloud{Name: "Busy session", Code: "FGHJK", Words: []Word{{Text: "first"}}})
	require.NoError(t, AddWordToWordCloud(ginContext(), stale.Id, &Word{Id: primitive.NewObjectID(), Text: "second", UUID: uuid.NewV4()}))
	require.NoError(t, EnsureWordIds(ginContext(), stale))
	assert.Equal(t, []string{"first", "second"}, texts(getWordCloud(t, stale.Id)), "a word added meanwhile is not overwritten")
}

func TestSetWordHidden(t *testing.T) {
	testutil.Clean(t)
	first, second := primitive.NewObjectID(), primitive.NewObjectID()
	session := createWordCloud(t, WordCloud{Name: "Session", Code: "ABCDE", Words: []Word{{Id: first, Text: "one"}, {Id: second, Text: "two"}}})

	found, err := SetWordHidden(ginContext(), session.Id, second, true)
	require.NoError(t, err)
	assert.True(t, found)
	words := getWordCloud(t, session.Id).Words
	assert.False(t, words[0].Hidden)
	assert.True(t, words[1].Hidden)

	found, err = SetWordHidden(ginContext(), session.Id, second, false)
	require.NoError(t, err)
	assert.True(t, found)
	assert.False(t, getWordCloud(t, session.Id).Words[1].Hidden)

	found, err = SetWordHidden(ginContext(), session.Id, primitive.NewObjectID(), true)
	require.NoError(t, err)
	assert.False(t, found)
	found, err = SetWordHidden(ginContext(), primitive.NewObjectID(), first, true)
	require.NoError(t, err)
	assert.False(t, found)
}

func TestMergeWords(t *testing.T) {
	testutil.Clean(t)
	session := createWordCloud(t, WordCloud{Name: "Session", Code: "ABCDE", Words: []Word{
		{Id: primitive.NewObjectID(), Text: "colour"},
		{Id: primitive.NewObjectID(), Text: "color", Hidden: true},
		{Id: primitive.NewObjectID(), Text: "couleur"},
		{Id: primitive.NewObjectID(), Text: "shape"},
	}})

	require.NoError(t, MergeWords(ginContext(), session.Id, []string{"colour", "couleur"}, "color"))
	merged := getWordCloud(t, session.Id).Words
	assert.Equal(t, []string{"color", "color", "color", "shape"}, texts(&WordCloud{Words: merged}))
	assert.Equal(t, "colour", merged[0].OriginalText)
	assert.Empty(t, merged[1].OriginalText, "a word that already had the text is untouched")
	assert.True(t, merged[1].Hidden)
	assert.Equal(t, "couleur", merged[2].OriginalText)
	assert.Equal(t, session.Words[0].Id, merged[0].Id)

	require.NoError(t, MergeWords(ginContext(), session.Id, []string{"color"}, "hue"))
	merged = getWordCloud(t, session.Id).Words
	assert.Equal(t, []string{"hue", "hue", "hue", "shape"}, texts(&WordCloud{Words: merged}))
	assert.Equal(t, "colour", merged[0].OriginalText, "the submitted text survives several merges")
	assert.Equal(t, "color", merged[1].OriginalText)

	require.NoError(t, MergeWords(ginContext(), primitive.NewObjectID(), []string{"a"}, "b"))
}

func TestWordCloudDatabaseFailures(t *testing.T) {
	session := createWordCloud(t, WordCloud{Name: "Session", Code: "QRSTU", Words: []Word{{Text: "no id"}}})
	testutil.BreakDB(t)
	id := primitive.NewObjectID()

	_, err := GetWordCloudByCode(ginContext(), "ABCDE")
	assert.Error(t, err)
	_, err = IsCodeInUse(ginContext(), "ABCDE")
	assert.Error(t, err)
	_, err = GetWordCloudById(ginContext(), id)
	assert.Error(t, err)
	_, err = GetWordCloudByUser(ginContext(), id, "open")
	assert.Error(t, err)
	_, err = CreateWordCloud(ginContext(), &WordCloud{})
	assert.Error(t, err)
	err = AddWordToWordCloud(ginContext(), id, &Word{})
	assert.Error(t, err)
	assert.NotErrorIs(t, err, ErrWordNotAdded)
	assert.Error(t, UpdateWordCloud(ginContext(), id, bson.M{"name": "x"}, nil))
	assert.Error(t, DeleteWordCloud(ginContext(), id))
	assert.Error(t, EnsureWordIds(ginContext(), session))
	_, err = SetWordHidden(ginContext(), id, id, true)
	assert.Error(t, err)
	assert.Error(t, MergeWords(ginContext(), id, []string{"a"}, "b"))
}
