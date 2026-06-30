package config

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// ExtensionBuiltInMode represents the build tag value for built-in extension mode.
const ExtensionBuiltInMode = "c2hlbgo"

// Extension contains extension system infrastructure settings.
type Extension struct {
	Mode      string   `json:"mode" yaml:"mode"`
	Path      string   `json:"path" yaml:"path"`
	Includes  []string `json:"includes" yaml:"includes"`
	Excludes  []string `json:"excludes" yaml:"excludes"`
	HotReload bool     `json:"hot_reload" yaml:"hot_reload"`

	MaxPlugins   int            `json:"max_plugins" yaml:"max_plugins"`
	PluginConfig map[string]any `json:"plugin_config" yaml:"plugin_config"`

	Security    *ExtensionSecurity    `json:"security" yaml:"security"`
	Performance *ExtensionPerformance `json:"performance" yaml:"performance"`
	Metrics     *ExtensionMetrics     `json:"metrics" yaml:"metrics"`
}

// ExtensionSecurity contains extension sandbox and trust settings.
type ExtensionSecurity struct {
	EnableSandbox     bool     `json:"enable_sandbox" yaml:"enable_sandbox"`
	AllowedPaths      []string `json:"allowed_paths" yaml:"allowed_paths"`
	BlockedExtensions []string `json:"blocked_extensions" yaml:"blocked_extensions"`
	TrustedSources    []string `json:"trusted_sources" yaml:"trusted_sources"`
	RequireSignature  bool     `json:"require_signature" yaml:"require_signature"`
	AllowUnsafe       bool     `json:"allow_unsafe" yaml:"allow_unsafe"`
}

// ExtensionPerformance contains extension runtime limits.
type ExtensionPerformance struct {
	MaxMemoryMB            int    `json:"max_memory_mb" yaml:"max_memory_mb"`
	MaxCPUPercent          int    `json:"max_cpu_percent" yaml:"max_cpu_percent"`
	GarbageCollectInterval string `json:"gc_interval" yaml:"gc_interval"`
	MaxConcurrentLoads     int    `json:"max_concurrent_loads" yaml:"max_concurrent_loads"`
}

// ExtensionMetrics contains extension metrics collection settings.
type ExtensionMetrics struct {
	Enabled       bool                     `json:"enabled" yaml:"enabled"`
	FlushInterval string                   `json:"flush_interval" yaml:"flush_interval"`
	BatchSize     int                      `json:"batch_size" yaml:"batch_size"`
	Retention     string                   `json:"retention" yaml:"retention"`
	Storage       *ExtensionMetricsStorage `json:"storage" yaml:"storage"`
}

// ExtensionMetricsStorage contains extension metrics persistence settings.
type ExtensionMetricsStorage struct {
	Type      string            `json:"type" yaml:"type"`
	KeyPrefix string            `json:"key_prefix" yaml:"key_prefix"`
	Options   map[string]string `json:"options" yaml:"options"`
}

// IsBuiltInMode reports whether built-in extension loading is enabled.
func (c *Extension) IsBuiltInMode() bool {
	return c != nil && c.Mode == ExtensionBuiltInMode
}

// GetRetentionDuration returns the configured metrics retention duration.
func (m *ExtensionMetrics) GetRetentionDuration() (time.Duration, error) {
	if m == nil || m.Retention == "" {
		return 168 * time.Hour, nil
	}
	return parseExtensionDuration(m.Retention)
}

// Validate validates extension infrastructure settings.
func (c *Extension) Validate() error {
	if c == nil {
		return fmt.Errorf("extension config is nil")
	}
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

// Validate validates extension metrics settings.
func (m *ExtensionMetrics) Validate() error {
	if m == nil || !m.Enabled {
		return nil
	}
	if m.FlushInterval != "" {
		if _, err := time.ParseDuration(m.FlushInterval); err != nil {
			return fmt.Errorf("invalid flush_interval: %v", err)
		}
	}
	if m.Retention != "" {
		if _, err := parseExtensionDuration(m.Retention); err != nil {
			return fmt.Errorf("invalid retention: %v", err)
		}
	}
	if m.BatchSize <= 0 {
		return fmt.Errorf("batch_size must be greater than 0")
	}
	return nil
}

// GetExtensionConfig returns extension configuration and panics on invalid values.
func GetExtensionConfig(v *viper.Viper) *Extension {
	config, err := GetExtensionConfigE(v)
	if err != nil {
		panic(err)
	}
	return config
}

// GetExtensionConfigE returns extension configuration and validation errors.
func GetExtensionConfigE(v *viper.Viper) (*Extension, error) {
	env := getStringWithDefault(v, "environment", "production")
	isDev := env == "development" || env == "dev"

	mode := getStringWithDefault(v, "extension.mode", "file")
	isBuiltIn := mode == ExtensionBuiltInMode

	config := &Extension{
		Mode:      mode,
		Path:      getStringWithDefault(v, "extension.path", getDefaultExtensionPath(isDev)),
		Includes:  v.GetStringSlice("extension.includes"),
		Excludes:  v.GetStringSlice("extension.excludes"),
		HotReload: isBuiltIn || getBoolWithDefault(v, "extension.hot_reload", false),

		MaxPlugins:   getIntWithDefault(v, "extension.max_plugins", 20),
		PluginConfig: v.GetStringMap("extension.plugin_config"),

		Security:    getExtensionSecurityConfig(v, isDev),
		Performance: getExtensionPerformanceConfig(v, isDev),
		Metrics:     getExtensionMetricsConfig(v, isDev),
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid extension config: %w", err)
	}
	return config, nil
}

func getExtensionConfig(v *viper.Viper) *Extension {
	return GetExtensionConfig(v)
}

func getDefaultExtensionPath(isDev bool) string {
	if isDev {
		return "./plugins"
	}
	return "/opt/ncore/plugins"
}

func getExtensionSecurityConfig(v *viper.Viper, isDev bool) *ExtensionSecurity {
	if !v.IsSet("extension.security") {
		return &ExtensionSecurity{
			EnableSandbox:     false,
			AllowedPaths:      []string{},
			BlockedExtensions: []string{".exe", ".bat", ".cmd"},
			TrustedSources:    []string{},
			RequireSignature:  false,
			AllowUnsafe:       isDev,
		}
	}

	return &ExtensionSecurity{
		EnableSandbox:     getBoolWithDefault(v, "extension.security.enable_sandbox", false),
		AllowedPaths:      v.GetStringSlice("extension.security.allowed_paths"),
		BlockedExtensions: getStringSliceWithDefault(v, "extension.security.blocked_extensions", []string{".exe", ".bat", ".cmd"}),
		TrustedSources:    v.GetStringSlice("extension.security.trusted_sources"),
		RequireSignature:  getBoolWithDefault(v, "extension.security.require_signature", false),
		AllowUnsafe:       getBoolWithDefault(v, "extension.security.allow_unsafe", isDev),
	}
}

func getExtensionPerformanceConfig(v *viper.Viper, isDev bool) *ExtensionPerformance {
	defaultMaxMem, defaultMaxCPU, defaultMaxLoads := 256, 30, 3
	defaultGC := "10m"
	if isDev {
		defaultMaxMem, defaultMaxCPU, defaultMaxLoads = 512, 50, 5
		defaultGC = "5m"
	}

	return &ExtensionPerformance{
		MaxMemoryMB:            getIntWithDefault(v, "extension.performance.max_memory_mb", defaultMaxMem),
		MaxCPUPercent:          getIntWithDefault(v, "extension.performance.max_cpu_percent", defaultMaxCPU),
		GarbageCollectInterval: getStringWithDefault(v, "extension.performance.gc_interval", defaultGC),
		MaxConcurrentLoads:     getIntWithDefault(v, "extension.performance.max_concurrent_loads", defaultMaxLoads),
	}
}

func getExtensionMetricsConfig(v *viper.Viper, isDev bool) *ExtensionMetrics {
	defaultBatch, defaultRetention, defaultFlush := 100, "7d", "60s"
	if isDev {
		defaultBatch, defaultRetention, defaultFlush = 50, "24h", "30s"
	}

	storage := &ExtensionMetricsStorage{
		Type:      "auto",
		KeyPrefix: "ncore_ext",
		Options:   make(map[string]string),
	}

	if v.IsSet("extension.metrics.storage") {
		storage.Type = getStringWithDefault(v, "extension.metrics.storage.type", "auto")
		storage.KeyPrefix = getStringWithDefault(v, "extension.metrics.storage.key_prefix", "ncore_ext")
		storage.Options = v.GetStringMapString("extension.metrics.storage.options")
	}

	return &ExtensionMetrics{
		Enabled:       getBoolWithDefault(v, "extension.metrics.enabled", false),
		FlushInterval: getStringWithDefault(v, "extension.metrics.flush_interval", defaultFlush),
		BatchSize:     getIntWithDefault(v, "extension.metrics.batch_size", defaultBatch),
		Retention:     getStringWithDefault(v, "extension.metrics.retention", defaultRetention),
		Storage:       storage,
	}
}

func parseExtensionDuration(s string) (time.Duration, error) {
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
