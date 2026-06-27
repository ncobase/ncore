package jwt

import (
	"testing"
	"time"
)

func TestProvideTokenManagerEParsesDurations(t *testing.T) {
	manager, err := ProvideTokenManagerE(&Config{
		Secret:              "secret",
		AccessTokenExpiry:   "15m",
		RefreshTokenExpiry:  "2d",
		RegisterTokenExpiry: "1w",
	})
	if err != nil {
		t.Fatalf("provide token manager: %v", err)
	}

	if manager.accessTokenExpiry != 15*time.Minute {
		t.Fatalf("unexpected access expiry: %s", manager.accessTokenExpiry)
	}
	if manager.refreshTokenExpiry != 48*time.Hour {
		t.Fatalf("unexpected refresh expiry: %s", manager.refreshTokenExpiry)
	}
	if manager.registerTokenExpiry != 7*24*time.Hour {
		t.Fatalf("unexpected register expiry: %s", manager.registerTokenExpiry)
	}
}

func TestProvideTokenManagerERejectsInvalidDuration(t *testing.T) {
	_, err := ProvideTokenManagerE(&Config{
		Secret:            "secret",
		AccessTokenExpiry: "0s",
	})
	if err == nil {
		t.Fatal("expected invalid duration error")
	}
}
