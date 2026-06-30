package logger

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync/atomic"
	"time"
)

const traceKey = "trace_id"

var traceFallbackCounter uint64

// getTraceID gets a trace ID from the context.
func getTraceID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if traceID, ok := ctx.Value(traceKey).(string); ok {
		return traceID
	}
	return ""
}

// EnsureTraceID ensures that a trace ID exists in the context.
func EnsureTraceID(ctx context.Context) (context.Context, string) {
	if ctx == nil {
		ctx = context.Background()
	}
	if traceID := getTraceID(ctx); traceID != "" {
		return ctx, traceID
	}
	traceID := newTraceID()
	return context.WithValue(ctx, traceKey, traceID), traceID
}

func newTraceID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err == nil {
		return hex.EncodeToString(b[:])
	}
	return fmt.Sprintf("%x%x", time.Now().UnixNano(), atomic.AddUint64(&traceFallbackCounter, 1))
}
