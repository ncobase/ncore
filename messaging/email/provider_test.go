package email

import (
	"strings"
	"testing"
)

func TestProvideSenderAllowsUnconfiguredEmail(t *testing.T) {
	sender, err := ProvideSender(nil)
	if err != nil {
		t.Fatalf("nil config should not error: %v", err)
	}
	if sender != nil {
		t.Fatalf("expected nil sender for nil config, got %#v", sender)
	}

	sender, err = ProvideSender(&Email{})
	if err != nil {
		t.Fatalf("empty provider should not error: %v", err)
	}
	if sender != nil {
		t.Fatalf("expected nil sender for empty provider, got %#v", sender)
	}
}

func TestProvideSenderRejectsUnknownProvider(t *testing.T) {
	_, err := ProvideSender(&Email{Provider: "unknown"})
	if err == nil || !strings.Contains(err.Error(), "unsupported email provider") {
		t.Fatalf("expected unsupported provider error, got %v", err)
	}
}
