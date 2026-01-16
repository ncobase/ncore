package logger

import (
	"context"

	"github.com/ncobase/ncore/ctxutil"
)

var traceKey = ctxutil.TraceIDKey

// getTraceID gets a trace ID from the context.
func getTraceID(ctx context.Context) string {
	return ctxutil.GetTraceID(ctx)
}

// EnsureTraceID ensures that a trace ID exists in the context.
func EnsureTraceID(ctx context.Context) (context.Context, string) {
	return ctxutil.EnsureTraceID(ctx)
}
