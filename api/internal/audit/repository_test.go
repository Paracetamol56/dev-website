package audit

import (
	"dev/internal/db"
	"dev/internal/testutil"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestEnsureIndexes(t *testing.T) {
	testutil.Clean(t)
	require.NoError(t, EnsureIndexes(ctx))
	require.NoError(t, EnsureIndexes(ctx))

	cursor, err := db.GetDB().Collection(collectionName).Indexes().List(ctx)
	require.NoError(t, err)
	var indexes []bson.M
	require.NoError(t, cursor.All(ctx, &indexes))

	expirations := map[string]any{}
	for _, index := range indexes {
		expirations[index["name"].(string)] = index["expireAfterSeconds"]
	}
	assert.Len(t, indexes, 4)
	assert.EqualValues(t, Retention.Seconds(), expirations["createdAt_1"])

	testutil.BreakDB(t)
	assert.Error(t, EnsureIndexes(testutil.Context()))
}

func TestCreateEntry(t *testing.T) {
	testutil.Clean(t)
	userId := primitive.NewObjectID()
	entry := Entry{
		Action:    WordCloudDeleted,
		CreatedAt: time.Now().UTC().Truncate(time.Millisecond),
		UserId:    &userId,
		Ip:        "203.0.113.7",
		UserAgent: "test-agent",
		RequestId: "request-1",
		Target:    &Target{Type: TargetWordCloud, Id: primitive.NewObjectID().Hex()},
		Meta:      map[string]any{"name": "Kick-off"},
	}

	require.NoError(t, CreateEntry(ctx, &entry))
	assert.False(t, entry.Id.IsZero())

	var stored Entry
	require.NoError(t, db.GetDB().Collection(collectionName).FindOne(ctx, bson.M{"_id": entry.Id}).Decode(&stored))
	stored.CreatedAt = stored.CreatedAt.UTC()
	assert.Equal(t, entry, stored)

	testutil.BreakDB(t)
	assert.Error(t, CreateEntry(testutil.Context(), &entry))
}

func TestCreateEntryWithoutUser(t *testing.T) {
	testutil.Clean(t)
	entry := Entry{Action: AuthLoginFailed, CreatedAt: time.Now()}
	require.NoError(t, CreateEntry(ctx, &entry))

	var raw bson.M
	require.NoError(t, db.GetDB().Collection(collectionName).FindOne(ctx, bson.M{"_id": entry.Id}).Decode(&raw))
	assert.NotContains(t, raw, "userId")
	assert.NotContains(t, raw, "target")
	assert.NotContains(t, raw, "meta")
}

func TestGetEntriesByUserId(t *testing.T) {
	testutil.Clean(t)
	userId, otherId := primitive.NewObjectID(), primitive.NewObjectID()
	now := time.Now()
	for _, entry := range []Entry{
		{Action: AuthLogin, UserId: &userId, CreatedAt: now.Add(-time.Hour)},
		{Action: UserProfileUpdated, UserId: &userId, CreatedAt: now},
		{Action: AuthLogin, UserId: &otherId, CreatedAt: now},
		{Action: AuthLoginFailed, CreatedAt: now},
	} {
		require.NoError(t, CreateEntry(ctx, &entry))
	}

	entries, err := GetEntriesByUserId(ctx, userId)
	require.NoError(t, err)
	require.Len(t, entries, 2)
	assert.Equal(t, UserProfileUpdated, entries[0].Action)
	assert.Equal(t, AuthLogin, entries[1].Action)

	entries, err = GetEntriesByUserId(ctx, primitive.NewObjectID())
	require.NoError(t, err)
	assert.Empty(t, entries)

	testutil.BreakDB(t)
	_, err = GetEntriesByUserId(testutil.Context(), userId)
	assert.Error(t, err)
}
