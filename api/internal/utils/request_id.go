package utils

import "context"

const RequestIdKey = "x-request-id"

// RequestId returns the id given to the request behind the context, if any
func RequestId(ctx context.Context) string {
	id, _ := ctx.Value(RequestIdKey).(string)
	return id
}
