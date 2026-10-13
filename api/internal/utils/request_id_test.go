package utils

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequestId(t *testing.T) {
	assert.Empty(t, RequestId(context.Background()))
	assert.Empty(t, RequestId(context.WithValue(context.Background(), RequestIdKey, 42)))
	assert.Equal(t, "abc", RequestId(context.WithValue(context.Background(), RequestIdKey, "abc")))
}
