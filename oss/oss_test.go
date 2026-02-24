package oss

import "testing"

func TestConfigValidateAliases(t *testing.T) {
	tests := []struct {
		name         string
		cfg          Config
		wantProvider string
		wantRegion   string
		wantErr      bool
	}{
		{
			name: "local alias",
			cfg: Config{
				Provider: "local",
				Bucket:   "",
			},
			wantProvider: "filesystem",
			wantErr:      false,
		},
		{
			name: "provider normalization with spaces and uppercase",
			cfg: Config{
				Provider: "  COS  ",
				ID:       "ak",
				Secret:   "sk",
				Bucket:   "example-1250000000",
				Region:   " ap-guangzhou ",
			},
			wantProvider: "tencent",
			wantRegion:   "ap-guangzhou",
			wantErr:      false,
		},
		{
			name: "oss alias",
			cfg: Config{
				Provider: "oss",
				ID:       "ak",
				Secret:   "sk",
				Bucket:   "bucket",
			},
			wantProvider: "aliyun",
			wantRegion:   "cn-hangzhou",
			wantErr:      false,
		},
		{
			name: "cos alias accepts full bucket with appid suffix",
			cfg: Config{
				Provider: "cos",
				ID:       "ak",
				Secret:   "sk",
				Bucket:   "examplebucket-1250000000",
				Region:   "ap-guangzhou",
			},
			wantProvider: "tencent",
			wantRegion:   "ap-guangzhou",
			wantErr:      false,
		},
		{
			name: "cos alias appid mismatch should fail",
			cfg: Config{
				Provider: "cos",
				ID:       "ak",
				Secret:   "sk",
				Bucket:   "examplebucket-1250000000",
				AppID:    "1250000001",
				Region:   "ap-guangzhou",
			},
			wantErr: true,
		},
		{
			name: "r2 alias defaults region auto",
			cfg: Config{
				Provider: "r2",
				ID:       "ak",
				Secret:   "sk",
				Bucket:   "bucket",
				Endpoint: "example.r2.cloudflarestorage.com",
			},
			wantProvider: "s3",
			wantRegion:   "auto",
			wantErr:      false,
		},
		{
			name: "b2 alias requires region",
			cfg: Config{
				Provider: "b2",
				ID:       "ak",
				Secret:   "sk",
				Bucket:   "bucket",
				Endpoint: "https://s3.us-west-004.backblazeb2.com",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.cfg
			err := cfg.Validate()
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantProvider != "" && cfg.Provider != tt.wantProvider {
				t.Fatalf("provider mismatch: got %q want %q", cfg.Provider, tt.wantProvider)
			}
			if tt.wantRegion != "" && cfg.Region != tt.wantRegion {
				t.Fatalf("region mismatch: got %q want %q", cfg.Region, tt.wantRegion)
			}
			if tt.name == "r2 alias defaults region auto" && cfg.Endpoint != "https://example.r2.cloudflarestorage.com" {
				t.Fatalf("endpoint mismatch: got %q", cfg.Endpoint)
			}
			if tt.name == "cos alias accepts full bucket with appid suffix" {
				if cfg.Bucket != "examplebucket" {
					t.Fatalf("bucket mismatch: got %q", cfg.Bucket)
				}
				if cfg.AppID != "1250000000" {
					t.Fatalf("app_id mismatch: got %q", cfg.AppID)
				}
			}
		})
	}
}

func TestNewStorageNilConfig(t *testing.T) {
	_, err := NewStorage(nil)
	if err == nil {
		t.Fatal("expected error for nil config, got nil")
	}
}
