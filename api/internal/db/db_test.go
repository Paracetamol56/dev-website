package db_test

import (
	"context"
	"testing"

	"dev/internal/db"
	"dev/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

func TestInit(t *testing.T) {
	t.Setenv("DB_URI", testutil.URI())

	t.Setenv("DB_NAME", "")
	require.NoError(t, db.Init())
	assert.Equal(t, "dev", db.GetDB().Name())
	db.GetDB().Client().Disconnect(context.Background())

	t.Setenv("DB_NAME", "dev_test_init")
	require.NoError(t, db.Init())
	assert.Equal(t, "dev_test_init", db.GetDB().Name())
	defer db.GetDB().Client().Disconnect(context.Background())
	defer db.GetDB().Drop(context.Background())

	_, err := db.GetDB().Collection("probe").InsertOne(context.Background(), bson.M{"ok": true})
	require.NoError(t, err)
	count, err := db.GetDB().Collection("probe").CountDocuments(context.Background(), bson.M{})
	require.NoError(t, err)
	assert.EqualValues(t, 1, count)
}

func TestUse(t *testing.T) {
	t.Setenv("DB_URI", testutil.URI())
	t.Setenv("DB_NAME", "dev_test_init")
	require.NoError(t, db.Init())
	defer db.GetDB().Client().Disconnect(context.Background())

	other := db.GetDB().Client().Database("dev_test_other")
	db.Use(other)
	assert.Same(t, other, db.GetDB())
}
