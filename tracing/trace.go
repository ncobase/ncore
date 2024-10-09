package tracing

import (
	"context"
	"ncobase/common/uuid"
)

const TraceIDKey = "trace_id"

// GetTraceID gets a trace ID from the context.
func GetTraceID(ctx context.Context) string {
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok {
		return traceID
	}
	return ""
}

// SetTraceID sets a trace ID to the context.
func SetTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}

// EnsureTraceID ensures that a trace ID exists in the context.
func EnsureTraceID(ctx context.Context) (context.Context, string) {
	if traceID := GetTraceID(ctx); traceID != "" {
		return ctx, traceID
	}
	traceID := uuid.NewString()
	return SetTraceID(ctx, traceID), traceID
}
