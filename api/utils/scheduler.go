package utils

import (
	"context"
	"time"

	"github.com/Paracetamol56/dev-website/api/models"
	"github.com/go-co-op/gocron"
)

func ScheduleAccountDeletion() {
	s := gocron.NewScheduler(time.UTC)
	s.Every(1).Day().At("00:00").Do(func() {
		models.DeleteOldUsers(context.Background())
	})
	s.StartAsync()
}
