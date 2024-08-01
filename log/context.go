package log

import (
	"context"
	"ncobase/common/uuid"
)

// getValue retrieves a value from the context.
func getValue(ctx context.Context, key string) any {
	return ctx.Value(key)
}

// setValue sets a value to the context.
func setValue(ctx context.Context, key string, val any) context.Context {
	return context.WithValue(ctx, key, val)
}

// setTraceID sets a trace ID to the context.
func setTraceID(ctx context.Context, traceID string) context.Context {
	return setValue(ctx, TraceIDKey, traceID)
}

// getTraceID gets a trace ID from the context.
func getTraceID(ctx context.Context) string {
	if traceID, ok := getValue(ctx, TraceIDKey).(string); ok {
		return traceID
	}
	return ""
}

// getOrCreateTraceID gets a trace ID from the context or creates a new one.
func getOrCreateTraceID(ctx context.Context) string {
	if traceID, ok := getValue(ctx, TraceIDKey).(string); ok {
		return traceID
	}
	return uuid.NewString()
}

// EnsureTraceID ensures that a trace ID exists in the context.
func EnsureTraceID(ctx context.Context) (context.Context, string) {
	traceID := getOrCreateTraceID(ctx)
	ctx = setValue(ctx, TraceIDKey, traceID)
	return ctx, traceID
}

// GetTraceID gets a trace ID from the context.
func GetTraceID(ctx context.Context) string {
	return getTraceID(ctx)
}

// SetTraceID sets a trace ID to the context.
func SetTraceID(ctx context.Context, traceID string) context.Context {
	return setTraceID(ctx, traceID)
}
