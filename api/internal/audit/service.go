package audit

import (
	"context"
	"dev/internal/utils"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const maxUserAgentLength = 256

func Record(ctx context.Context, entry Entry) {
	entry.CreatedAt = time.Now()
	if entry.RequestId == "" {
		entry.RequestId = utils.RequestId(ctx)
	}
	if entry.UserId == nil {
		if userId, ok := ctx.Value("x-user-id").(primitive.ObjectID); ok {
			entry.UserId = &userId
		}
	}
	if c, ok := ctx.(*gin.Context); ok && c.Request != nil {
		entry.Ip = c.ClientIP()
		entry.UserAgent = c.Request.UserAgent()
		if len(entry.UserAgent) > maxUserAgentLength {
			entry.UserAgent = entry.UserAgent[:maxUserAgentLength]
		}
	}
	if err := CreateEntry(ctx, &entry); err != nil {
		log.Printf("Failed to record audit entry %s: %v", entry.Action, err)
	}
}
