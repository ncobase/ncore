package storage

import (
	"context"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/ncobase/ncore/config"
	"github.com/ncobase/ncore/ctxutil"
	"github.com/ncobase/ncore/oss"
)

type mockStorage struct{}

func (m *mockStorage) Get(_ string) (*os.File, error) { return nil, errors.New("not implemented") }
func (m *mockStorage) GetStream(_ string) (io.ReadCloser, error) {
	return nil, errors.New("not implemented")
}
func (m *mockStorage) Put(_ string, _ io.Reader) (*oss.Object, error) {
	return nil, errors.New("not implemented")
}
func (m *mockStorage) Delete(_ string) error                { return nil }
func (m *mockStorage) List(_ string) ([]*oss.Object, error) { return nil, nil }
func (m *mockStorage) GetURL(_ string) (string, error)      { return "", nil }
func (m *mockStorage) GetEndpoint() string                  { return "mock-endpoint" }
func (m *mockStorage) Exists(_ string) (bool, error)        { return false, nil }
func (m *mockStorage) Stat(_ string) (*oss.Object, error)   { return nil, errors.New("not implemented") }

func TestStorageHelpers(t *testing.T) {
	cfg := &config.Config{
		Storage: &config.Storage{
			Provider: "filesystem",
			Bucket:   t.TempDir(),
		},
	}

	mock := &mockStorage{}
	ctx := Set(ctxutil.SetConfig(context.Background(), cfg), mock)
	gotStorage, gotCfg := Get(ctx)
	if gotStorage == nil {
		t.Fatal("expected Get to return mock storage")
	}
	if gotStorage != mock {
		t.Fatalf("expected same storage instance, got %#v", gotStorage)
	}
	if gotCfg == nil || gotCfg.Provider != "filesystem" {
		t.Fatalf("unexpected storage config: %#v", gotCfg)
	}

	ctxAuto := ctxutil.SetConfig(context.Background(), &config.Config{
		Storage: &config.Storage{Provider: "filesystem", Bucket: t.TempDir()},
	})
	autoStorage, autoCfg := Get(ctxAuto)
	if autoStorage == nil {
		t.Fatal("expected auto storage creation to succeed")
	}
	if autoCfg == nil || autoCfg.Provider != "filesystem" {
		t.Fatalf("unexpected auto storage config: %#v", autoCfg)
	}

	ctxInvalid := ctxutil.SetConfig(context.Background(), &config.Config{Storage: nil})
	invalidStorage, invalidCfg := Get(ctxInvalid)
	if invalidStorage != nil || invalidCfg != nil {
		t.Fatalf("expected invalid storage config to return nil values, got storage=%#v cfg=%#v", invalidStorage, invalidCfg)
	}

	ctxMissingConfig := ctxutil.SetConfig(context.Background(), nil)
	missingStorage, missingCfg := Get(ctxMissingConfig)
	if missingStorage != nil || missingCfg != nil {
		t.Fatalf("expected missing config to return nil values, got storage=%#v cfg=%#v", missingStorage, missingCfg)
	}
}

func TestToOSSConfig(t *testing.T) {
	cfg := ToOSSConfig(&config.Storage{
		Provider: "s3",
		ID:       "id",
		Secret:   "secret",
		Region:   "us-east-1",
		Bucket:   "bucket",
		Endpoint: "https://example.com",
	})
	if cfg == nil {
		t.Fatal("expected converted config")
	}
	if cfg.Provider != "s3" || cfg.Bucket != "bucket" || cfg.Endpoint != "https://example.com" {
		t.Fatalf("unexpected converted config: %#v", cfg)
	}
}
