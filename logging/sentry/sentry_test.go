package sentry

import "testing"

func TestNewSentrySkipsNilOptions(t *testing.T) {
	if err := NewSentry(nil); err != nil {
		t.Fatalf("NewSentry(nil) returned unexpected error: %v", err)
	}
}

func TestNewSentryRejectsInvalidSampleRate(t *testing.T) {
	if err := NewSentry(&SentryOptions{SampleRate: 1.1}); err == nil {
		t.Fatal("expected invalid sample rate error")
	}
}
