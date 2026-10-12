package audit

import (
	"context"
	"testing"

	"dev/internal/testutil"
)

func TestMain(m *testing.M) {
	testutil.Main(m)
}

var ctx = context.Background()
