package config

import (
	"testing"

	"github.com/spf13/viper"
)

func TestGetConfigEReturnsValidationError(t *testing.T) {
	v := viper.New()
	v.Set("extension.max_plugins", -1)

	if _, err := GetConfigE(v); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestGetConfigEAppliesDefaults(t *testing.T) {
	v := viper.New()

	cfg, err := GetConfigE(v)
	if err != nil {
		t.Fatalf("get config: %v", err)
	}
	if cfg.Mode != "file" {
		t.Fatalf("unexpected mode: %s", cfg.Mode)
	}
	if cfg.MaxPlugins != 20 {
		t.Fatalf("unexpected max plugins: %d", cfg.MaxPlugins)
	}
}
