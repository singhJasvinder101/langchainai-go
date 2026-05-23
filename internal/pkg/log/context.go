package log

import "context"

type contextKey string

const requestIDKey contextKey = "request_id"

func ContextWithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

func RequestIDFromContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	requestID, ok := ctx.Value(requestIDKey).(string)
	return requestID, ok && requestID != ""
}
