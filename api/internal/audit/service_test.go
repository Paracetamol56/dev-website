package audit

import (
	"dev/internal/testutil"
	"dev/internal/utils"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestRecord(t *testing.T) {
	testutil.Clean(t)
	userId := primitive.NewObjectID()

	Record(ctx, Entry{Action: AuthLogin, UserId: &userId, Meta: map[string]any{"method": "passkey"}})

	entries, err := GetEntriesByUserId(ctx, userId)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, AuthLogin, entries[0].Action)
	assert.WithinDuration(t, time.Now(), entries[0].CreatedAt, 5*time.Second)

	c := testutil.Context()
	c.Set(utils.RequestIdKey, "request-1")
	Record(c, Entry{Action: UserDataExported, UserId: &userId})
	Record(c, Entry{Action: UserDataExported, UserId: &userId, RequestId: "explicit"})
	entries, err = GetEntriesByUserId(ctx, userId)
	require.NoError(t, err)
	ids := []string{}
	for _, entry := range entries {
		ids = append(ids, entry.RequestId)
	}
	assert.ElementsMatch(t, []string{"", "request-1", "explicit"}, ids)

	testutil.BreakDB(t)
	assert.NotPanics(t, func() { Record(testutil.Context(), Entry{Action: AuthLogin}) })
}
