package log

import (
	"context"
	"ncobase/common/tracing"
)

var traceKey = tracing.TraceIDKey

// getTraceID gets a trace ID from the context.
func getTraceID(ctx context.Context) string {
	return tracing.GetTraceID(ctx)
}

// setTraceID sets a trace ID to the context.
func setTraceID(ctx context.Context, traceID string) context.Context {
	return tracing.SetTraceID(ctx, traceID)
}

// EnsureTraceID ensures that a trace ID exists in the context.
func EnsureTraceID(ctx context.Context) (context.Context, string) {
	return tracing.EnsureTraceID(ctx)
}
