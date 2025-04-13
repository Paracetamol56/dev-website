package utils

import (
	"context"
	"dev/internal/models"
	"time"

	"github.com/go-co-op/gocron"
)

func ScheduleAccountDeletion() {
	s := gocron.NewScheduler(time.UTC)
	s.Every(1).Day().At("00:00").Do(func() {
		models.DeleteOldUsers(context.Background())
	})
	s.StartAsync()
}
