package request_id

import (
	"context"
)

type ctxKeyRequestID string

const ReqIDKey ctxKeyRequestID = "request_id"

const LogFieldRequestID = string(ReqIDKey)

func GetReqID(ctx context.Context) string {
	if reqID, ok := ctx.Value(ReqIDKey).(ctxKeyRequestID); ok {
		return string(reqID)
	}
	return ""
}

func SetReqID(ctx context.Context, reqID string) context.Context {
	return context.WithValue(ctx, ReqIDKey, ctxKeyRequestID(reqID))
}
