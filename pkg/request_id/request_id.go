package request_id

import (
	"context"
)

const ReqIDKey string = "request_id"

func GetReqID(ctx context.Context) string {
	if reqID, ok := ctx.Value(ReqIDKey).(string); ok {
		return reqID
	}
	return ""
}

func SetReqID(ctx context.Context, reqID string) context.Context {
	return context.WithValue(ctx, ReqIDKey, reqID)
}
