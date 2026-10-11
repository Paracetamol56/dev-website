package user

import (
	"testing"

	"dev/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendDeletionEmail(t *testing.T) {
	internet := testutil.FakeInternet(t)
	testutil.FakeBrevo(internet)

	require.NoError(t, SendDeletionEmail(ctx, &User{Name: "Ada", Email: "ada@example.test"}))

	emails := testutil.SentEmails(t, internet)
	require.Len(t, emails, 1)
	assert.Equal(t, "ada@example.test", emails[0].To[0].Email)
	assert.Equal(t, "Ada", emails[0].To[0].Name)
	assert.Equal(t, "Your account has been deleted", emails[0].Subject)

	testutil.FakeInternet(t)
	assert.Error(t, SendDeletionEmail(ctx, &User{Email: "ada@example.test"}))
}

func TestScheduleAccountDeletion(t *testing.T) {
	assert.NotPanics(t, ScheduleAccountDeletion)
}
