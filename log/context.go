package log

import (
	"context"
	"ncobase/common/helper"
)

var traceKey = helper.TraceIDKey

// getTraceID gets a trace ID from the context.
func getTraceID(ctx context.Context) string {
	return helper.GetTraceID(ctx)
}

// setTraceID sets a trace ID to the context.
func setTraceID(ctx context.Context, traceID string) context.Context {
	return helper.SetTraceID(ctx, traceID)
}

// EnsureTraceID ensures that a trace ID exists in the context.
func EnsureTraceID(ctx context.Context) (context.Context, string) {
	return helper.EnsureTraceID(ctx)
}
