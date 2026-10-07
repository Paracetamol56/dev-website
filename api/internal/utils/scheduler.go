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

// ScheduleIconRefresh seeds the icons collection if it is empty, then refreshes it every Sunday at 00:00 UTC.
func ScheduleIconRefresh() {
	go SeedIcons(context.Background())

	s := gocron.NewScheduler(time.UTC)
	s.Every(1).Week().Sunday().At("00:00").Do(func() {
		RefreshIcons(context.Background())
	})
	s.StartAsync()
}
