package audit

import (
	"context"
	"dev/internal/utils"
	"log"
	"time"
)

func Record(ctx context.Context, entry Entry) {
	entry.CreatedAt = time.Now()
	if entry.RequestId == "" {
		entry.RequestId = utils.RequestId(ctx)
	}
	if err := CreateEntry(ctx, &entry); err != nil {
		log.Printf("Failed to record audit entry %s: %v", entry.Action, err)
	}
}
