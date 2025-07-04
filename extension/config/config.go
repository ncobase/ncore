package config

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config extension config struct
type Config struct {
	Mode      string   `json:"mode" yaml:"mode"`
	Path      string   `json:"path" yaml:"path"`
	Includes  []string `json:"includes" yaml:"includes"`
	Excludes  []string `json:"excludes" yaml:"excludes"`
	HotReload bool     `json:"hot_reload" yaml:"hot_reload"`

	MaxPlugins   int            `json:"max_plugins" yaml:"max_plugins"`
	PluginConfig map[string]any `json:"plugin_config" yaml:"plugin_config"`

	Security    *SecurityConfig    `json:"security" yaml:"security"`
	Performance *PerformanceConfig `json:"performance" yaml:"performance"`
	Metrics     *MetricsConfig     `json:"metrics" yaml:"metrics"`
}

// SecurityConfig security settings
type SecurityConfig struct {
	EnableSandbox     bool     `json:"enable_sandbox" yaml:"enable_sandbox"`
	AllowedPaths      []string `json:"allowed_paths" yaml:"allowed_paths"`
	BlockedExtensions []string `json:"blocked_extensions" yaml:"blocked_extensions"`
	TrustedSources    []string `json:"trusted_sources" yaml:"trusted_sources"`
	RequireSignature  bool     `json:"require_signature" yaml:"require_signature"`
	AllowUnsafe       bool     `json:"allow_unsafe" yaml:"allow_unsafe"`
}

// PerformanceConfig performance settings
type PerformanceConfig struct {
	MaxMemoryMB            int    `json:"max_memory_mb" yaml:"max_memory_mb"`
	MaxCPUPercent          int    `json:"max_cpu_percent" yaml:"max_cpu_percent"`
	GarbageCollectInterval string `json:"gc_interval" yaml:"gc_interval"`
	MaxConcurrentLoads     int    `json:"max_concurrent_loads" yaml:"max_concurrent_loads"`
}

// MetricsConfig metrics configuration
type MetricsConfig struct {
	Enabled       bool           `json:"enabled" yaml:"enabled"`
	FlushInterval string         `json:"flush_interval" yaml:"flush_interval"`
	BatchSize     int            `json:"batch_size" yaml:"batch_size"`
	Retention     string         `json:"retention" yaml:"retention"`
	Storage       *StorageConfig `json:"storage" yaml:"storage"`
}

// StorageConfig metrics storage configuration
type StorageConfig struct {
	Type      string            `json:"type" yaml:"type"`
	KeyPrefix string            `json:"key_prefix" yaml:"key_prefix"`
	Options   map[string]string `json:"options" yaml:"options"`
}

// IsBuiltInMode checks if running in built-in mode
func (c *Config) IsBuiltInMode() bool {
	return c.Mode == "c2hlbgo"
}

// GetRetentionDuration returns the retention duration
func (m *MetricsConfig) GetRetentionDuration() (time.Duration, error) {
	if m.Retention == "" {
		return 168 * time.Hour, nil
	}
	return parseDuration(m.Retention)
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.MaxPlugins <= 0 {
		return fmt.Errorf("max_plugins must be greater than 0")
	}

	if c.Performance != nil && c.Performance.GarbageCollectInterval != "" {
		if _, err := time.ParseDuration(c.Performance.GarbageCollectInterval); err != nil {
			return fmt.Errorf("invalid gc_interval: %v", err)
		}
	}

	if c.Metrics != nil {
		if err := c.Metrics.Validate(); err != nil {
			return fmt.Errorf("metrics config error: %v", err)
		}
	}

	return nil
}

// Validate validates the metrics configuration
func (m *MetricsConfig) Validate() error {
	if !m.Enabled {
		return nil
	}

	if m.FlushInterval != "" {
		if _, err := time.ParseDuration(m.FlushInterval); err != nil {
			return fmt.Errorf("invalid flush_interval: %v", err)
		}
	}

	if m.Retention != "" {
		if _, err := parseDuration(m.Retention); err != nil {
			return fmt.Errorf("invalid retention: %v", err)
		}
	}

	if m.BatchSize <= 0 {
		return fmt.Errorf("batch_size must be greater than 0")
	}

	return nil
}

// parseDuration parses duration with support for days (d) and weeks (w)
func parseDuration(s string) (time.Duration, error) {
	if strings.HasSuffix(s, "d") {
		days, err := strconv.Atoi(strings.TrimSuffix(s, "d"))
		if err != nil {
			return 0, fmt.Errorf("invalid days format: %s", s)
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}

	if strings.HasSuffix(s, "w") {
		weeks, err := strconv.Atoi(strings.TrimSuffix(s, "w"))
		if err != nil {
			return 0, fmt.Errorf("invalid weeks format: %s", s)
		}
		return time.Duration(weeks) * 7 * 24 * time.Hour, nil
	}

	return time.ParseDuration(s)
}

// GetConfig returns extension config from viper
func GetConfig(v *viper.Viper) *Config {
	env := getStringWithDefault(v, "environment", "production")
	isDev := env == "development" || env == "dev"

	mode := getStringWithDefault(v, "extension.mode", "file")
	isBuiltIn := mode == "c2hlbgo"

	config := &Config{
		Mode:      mode,
		Path:      getStringWithDefault(v, "extension.path", getDefaultPath(isDev)),
		Includes:  v.GetStringSlice("extension.includes"),
		Excludes:  v.GetStringSlice("extension.excludes"),
		HotReload: isBuiltIn || getBoolWithDefault(v, "extension.hot_reload", false),

		MaxPlugins:   getIntWithDefault(v, "extension.max_plugins", 20),
		PluginConfig: v.GetStringMap("extension.plugin_config"),

		Security:    getSecurityConfig(v, isDev),
		Performance: getPerformanceConfig(v, isDev),
		Metrics:     getMetricsConfig(v, isDev),
	}

	if err := config.Validate(); err != nil {
		panic(fmt.Sprintf("invalid extension config: %v", err))
	}

	return config
}

func getDefaultPath(isDev bool) string {
	if isDev {
		return "./plugins"
	}
	return "/opt/ncore/plugins"
}

func getSecurityConfig(v *viper.Viper, isDev bool) *SecurityConfig {
	if !v.IsSet("extension.security") {
		return &SecurityConfig{
			EnableSandbox:     false,
			AllowedPaths:      []string{},
			BlockedExtensions: []string{".exe", ".bat", ".cmd"},
			TrustedSources:    []string{},
			RequireSignature:  false,
			AllowUnsafe:       isDev,
		}
	}

	return &SecurityConfig{
		EnableSandbox:     getBoolWithDefault(v, "extension.security.enable_sandbox", false),
		AllowedPaths:      v.GetStringSlice("extension.security.allowed_paths"),
		BlockedExtensions: getStringSliceWithDefault(v, "extension.security.blocked_extensions", []string{".exe", ".bat", ".cmd"}),
		TrustedSources:    v.GetStringSlice("extension.security.trusted_sources"),
		RequireSignature:  getBoolWithDefault(v, "extension.security.require_signature", false),
		AllowUnsafe:       getBoolWithDefault(v, "extension.security.allow_unsafe", isDev),
	}
}

func getPerformanceConfig(v *viper.Viper, isDev bool) *PerformanceConfig {
	defaultMaxMem, defaultMaxCPU, defaultMaxLoads := 256, 30, 3
	defaultGC := "10m"
	if isDev {
		defaultMaxMem, defaultMaxCPU, defaultMaxLoads = 512, 50, 5
		defaultGC = "5m"
	}

	return &PerformanceConfig{
		MaxMemoryMB:            getIntWithDefault(v, "extension.performance.max_memory_mb", defaultMaxMem),
		MaxCPUPercent:          getIntWithDefault(v, "extension.performance.max_cpu_percent", defaultMaxCPU),
		GarbageCollectInterval: getStringWithDefault(v, "extension.performance.gc_interval", defaultGC),
		MaxConcurrentLoads:     getIntWithDefault(v, "extension.performance.max_concurrent_loads", defaultMaxLoads),
	}
}

func getMetricsConfig(v *viper.Viper, isDev bool) *MetricsConfig {
	defaultBatch, defaultRetention, defaultFlush := 100, "7d", "60s"
	if isDev {
		defaultBatch, defaultRetention, defaultFlush = 50, "24h", "30s"
	}

	storage := &StorageConfig{
		Type:      "auto",
		KeyPrefix: "ncore_ext",
		Options:   make(map[string]string),
	}

	if v.IsSet("extension.metrics.storage") {
		storage.Type = getStringWithDefault(v, "extension.metrics.storage.type", "auto")
		storage.KeyPrefix = getStringWithDefault(v, "extension.metrics.storage.key_prefix", "ncore_ext")
		storage.Options = v.GetStringMapString("extension.metrics.storage.options")
	}

	return &MetricsConfig{
		Enabled:       getBoolWithDefault(v, "extension.metrics.enabled", false),
		FlushInterval: getStringWithDefault(v, "extension.metrics.flush_interval", defaultFlush),
		BatchSize:     getIntWithDefault(v, "extension.metrics.batch_size", defaultBatch),
		Retention:     getStringWithDefault(v, "extension.metrics.retention", defaultRetention),
		Storage:       storage,
	}
}

func getStringWithDefault(v *viper.Viper, key, defaultValue string) string {
	if v.IsSet(key) {
		return v.GetString(key)
	}
	return defaultValue
}

func getIntWithDefault(v *viper.Viper, key string, defaultValue int) int {
	if v.IsSet(key) {
		return v.GetInt(key)
	}
	return defaultValue
}

func getBoolWithDefault(v *viper.Viper, key string, defaultValue bool) bool {
	if v.IsSet(key) {
		return v.GetBool(key)
	}
	return defaultValue
}

func getStringSliceWithDefault(v *viper.Viper, key string, defaultValue []string) []string {
	if v.IsSet(key) {
		return v.GetStringSlice(key)
	}
	return defaultValue
}
