package config

import (
	"fmt"
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

	// Basic limits and timeouts
	MaxPlugins        int            `json:"max_plugins" yaml:"max_plugins"`
	LoadTimeout       string         `json:"load_timeout" yaml:"load_timeout"`
	InitTimeout       string         `json:"init_timeout" yaml:"init_timeout"`
	DependencyTimeout string         `json:"dependency_timeout" yaml:"dependency_timeout"`
	PluginConfig      map[string]any `json:"plugin_config" yaml:"plugin_config"`

	// Feature toggles
	Security    *SecurityConfig    `json:"security" yaml:"security"`
	Performance *PerformanceConfig `json:"performance" yaml:"performance"`
	Monitoring  *MonitoringConfig  `json:"monitoring" yaml:"monitoring"`
}

// SecurityConfig defines security-related extension settings
type SecurityConfig struct {
	EnableSandbox     bool     `json:"enable_sandbox" yaml:"enable_sandbox"`
	AllowedPaths      []string `json:"allowed_paths" yaml:"allowed_paths"`
	BlockedExtensions []string `json:"blocked_extensions" yaml:"blocked_extensions"`
	TrustedSources    []string `json:"trusted_sources" yaml:"trusted_sources"`
	RequireSignature  bool     `json:"require_signature" yaml:"require_signature"`
	AllowUnsafe       bool     `json:"allow_unsafe" yaml:"allow_unsafe"` // For development
}

// PerformanceConfig defines performance-related extension settings
type PerformanceConfig struct {
	MaxMemoryMB            int    `json:"max_memory_mb" yaml:"max_memory_mb"`
	MaxCPUPercent          int    `json:"max_cpu_percent" yaml:"max_cpu_percent"`
	EnableMetrics          bool   `json:"enable_metrics" yaml:"enable_metrics"`
	MetricsInterval        string `json:"metrics_interval" yaml:"metrics_interval"`
	GarbageCollectInterval string `json:"gc_interval" yaml:"gc_interval"`
	MaxConcurrentLoads     int    `json:"max_concurrent_loads" yaml:"max_concurrent_loads"`
}

// MonitoringConfig defines monitoring-related settings
type MonitoringConfig struct {
	EnableHealthCheck     bool   `json:"enable_health_check" yaml:"enable_health_check"`
	HealthCheckInterval   string `json:"health_check_interval" yaml:"health_check_interval"`
	EnableDetailedMetrics bool   `json:"enable_detailed_metrics" yaml:"enable_detailed_metrics"`
	MetricsRetention      string `json:"metrics_retention" yaml:"metrics_retention"`
}

// IsBuiltInMode checks if running in built-in mode
func (c *Config) IsBuiltInMode() bool {
	return c.Mode == "c2hlbgo" // built-in mode
}

// ShouldEnableHotReload checks if hot reload is enabled and supported
func (c *Config) ShouldEnableHotReload() bool {
	return c.HotReload && !c.IsBuiltInMode()
}

// IsDevelopmentMode checks if running in development mode
func (c *Config) IsDevelopmentMode() bool {
	return c.Mode == "development" || c.Mode == "dev"
}

// IsProductionMode checks if running in production mode
func (c *Config) IsProductionMode() bool {
	return c.Mode == "production" || c.Mode == "prod"
}

// Validate validates the configuration with simplified logic
func (c *Config) Validate() error {
	if c.MaxPlugins <= 0 {
		return fmt.Errorf("max_plugins must be greater than 0, got: %d", c.MaxPlugins)
	}

	// Validate timeout values if provided
	timeouts := map[string]string{
		"load_timeout":       c.LoadTimeout,
		"init_timeout":       c.InitTimeout,
		"dependency_timeout": c.DependencyTimeout,
	}

	for name, timeout := range timeouts {
		if timeout != "" {
			if _, err := time.ParseDuration(timeout); err != nil {
				return fmt.Errorf("invalid %s: %v", name, err)
			}
		}
	}

	return nil
}

// GetDefaultSecurityConfig returns default security configuration
func GetDefaultSecurityConfig(isDevelopment bool) *SecurityConfig {
	if isDevelopment {
		return &SecurityConfig{
			EnableSandbox:     false,
			AllowedPaths:      []string{},
			BlockedExtensions: []string{".exe", ".bat", ".cmd"},
			TrustedSources:    []string{}, // Allow any source in development
			RequireSignature:  false,
			AllowUnsafe:       true,
		}
	}

	return &SecurityConfig{
		EnableSandbox:     true,
		AllowedPaths:      []string{"/opt/plugins", "/usr/local/plugins"},
		BlockedExtensions: []string{".exe", ".bat", ".cmd", ".ps1", ".sh"},
		TrustedSources:    []string{}, // Empty means require explicit configuration
		RequireSignature:  true,
		AllowUnsafe:       false,
	}
}

// GetDefaultPerformanceConfig returns default performance configuration
func GetDefaultPerformanceConfig(isDevelopment bool) *PerformanceConfig {
	if isDevelopment {
		return &PerformanceConfig{
			MaxMemoryMB:            1024,
			MaxCPUPercent:          90,
			EnableMetrics:          true,
			MetricsInterval:        "10s",
			GarbageCollectInterval: "2m",
			MaxConcurrentLoads:     10,
		}
	}

	return &PerformanceConfig{
		MaxMemoryMB:            512,
		MaxCPUPercent:          70,
		EnableMetrics:          true,
		MetricsInterval:        "30s",
		GarbageCollectInterval: "5m",
		MaxConcurrentLoads:     5,
	}
}

// GetDefaultMonitoringConfig returns default monitoring configuration
func GetDefaultMonitoringConfig() *MonitoringConfig {
	return &MonitoringConfig{
		EnableHealthCheck:     true,
		HealthCheckInterval:   "60s",
		EnableDetailedMetrics: false,
		MetricsRetention:      "24h",
	}
}

// GetConfig returns extension config from viper
func GetConfig(v *viper.Viper) *Config {
	// Determine environment
	env := v.GetString("environment")
	if env == "" {
		env = v.GetString("extension.environment")
	}
	if env == "" {
		env = "development" // Default to development
	}

	isDevelopment := env == "development" || env == "dev"

	config := &Config{
		Mode:      getStringWithDefault(v, "extension.mode", "file"),
		Path:      getStringWithDefault(v, "extension.path", getDefaultPluginPath(isDevelopment)),
		Includes:  v.GetStringSlice("extension.includes"),
		Excludes:  v.GetStringSlice("extension.excludes"),
		HotReload: getBoolWithDefault(v, "extension.hot_reload", isDevelopment),

		MaxPlugins:        getIntWithDefault(v, "extension.max_plugins", 50),
		LoadTimeout:       getStringWithDefault(v, "extension.load_timeout", "30s"),
		InitTimeout:       getStringWithDefault(v, "extension.init_timeout", "60s"),
		DependencyTimeout: getStringWithDefault(v, "extension.dependency_timeout", "15s"),
		PluginConfig:      v.GetStringMap("extension.plugin_config"),

		Security:    getSecurityConfig(v, isDevelopment),
		Performance: getPerformanceConfig(v, isDevelopment),
		Monitoring:  getMonitoringConfig(v),
	}

	if err := config.Validate(); err != nil {
		panic(fmt.Sprintf("invalid extension config: %v", err))
	}

	return config
}

// getSecurityConfig returns security configuration
func getSecurityConfig(v *viper.Viper, isDevelopment bool) *SecurityConfig {
	defaultConfig := GetDefaultSecurityConfig(isDevelopment)

	if !v.IsSet("extension.security") {
		return defaultConfig
	}

	return &SecurityConfig{
		EnableSandbox:     getBoolWithDefault(v, "extension.security.enable_sandbox", defaultConfig.EnableSandbox),
		AllowedPaths:      getStringSliceWithDefault(v, "extension.security.allowed_paths", defaultConfig.AllowedPaths),
		BlockedExtensions: getStringSliceWithDefault(v, "extension.security.blocked_extensions", defaultConfig.BlockedExtensions),
		TrustedSources:    getStringSliceWithDefault(v, "extension.security.trusted_sources", defaultConfig.TrustedSources),
		RequireSignature:  getBoolWithDefault(v, "extension.security.require_signature", defaultConfig.RequireSignature),
		AllowUnsafe:       getBoolWithDefault(v, "extension.security.allow_unsafe", defaultConfig.AllowUnsafe),
	}
}

// getPerformanceConfig returns performance configuration
func getPerformanceConfig(v *viper.Viper, isDevelopment bool) *PerformanceConfig {
	defaultConfig := GetDefaultPerformanceConfig(isDevelopment)

	return &PerformanceConfig{
		MaxMemoryMB:            getIntWithDefault(v, "extension.performance.max_memory_mb", defaultConfig.MaxMemoryMB),
		MaxCPUPercent:          getIntWithDefault(v, "extension.performance.max_cpu_percent", defaultConfig.MaxCPUPercent),
		EnableMetrics:          getBoolWithDefault(v, "extension.performance.enable_metrics", defaultConfig.EnableMetrics),
		MetricsInterval:        getStringWithDefault(v, "extension.performance.metrics_interval", defaultConfig.MetricsInterval),
		GarbageCollectInterval: getStringWithDefault(v, "extension.performance.gc_interval", defaultConfig.GarbageCollectInterval),
		MaxConcurrentLoads:     getIntWithDefault(v, "extension.performance.max_concurrent_loads", defaultConfig.MaxConcurrentLoads),
	}
}

// getMonitoringConfig returns monitoring configuration
func getMonitoringConfig(v *viper.Viper) *MonitoringConfig {
	defaultConfig := GetDefaultMonitoringConfig()

	if !v.IsSet("extension.monitoring") {
		return defaultConfig
	}

	return &MonitoringConfig{
		EnableHealthCheck:     getBoolWithDefault(v, "extension.monitoring.enable_health_check", defaultConfig.EnableHealthCheck),
		HealthCheckInterval:   getStringWithDefault(v, "extension.monitoring.health_check_interval", defaultConfig.HealthCheckInterval),
		EnableDetailedMetrics: getBoolWithDefault(v, "extension.monitoring.enable_detailed_metrics", defaultConfig.EnableDetailedMetrics),
		MetricsRetention:      getStringWithDefault(v, "extension.monitoring.metrics_retention", defaultConfig.MetricsRetention),
	}
}

// getDefaultPluginPath returns default plugin path
func getDefaultPluginPath(isDevelopment bool) string {
	if isDevelopment {
		return "./plugins"
	}
	return "/opt/ncore/plugins"
}

// getStringWithDefault returns string value with default
func getStringWithDefault(v *viper.Viper, key, defaultValue string) string {
	if v.IsSet(key) {
		return v.GetString(key)
	}
	return defaultValue
}

// getIntWithDefault returns int value with default
func getIntWithDefault(v *viper.Viper, key string, defaultValue int) int {
	if v.IsSet(key) {
		return v.GetInt(key)
	}
	return defaultValue
}

// getBoolWithDefault returns bool value with default
func getBoolWithDefault(v *viper.Viper, key string, defaultValue bool) bool {
	if v.IsSet(key) {
		return v.GetBool(key)
	}
	return defaultValue
}

// getStringSliceWithDefault returns string slice value with default
func getStringSliceWithDefault(v *viper.Viper, key string, defaultValue []string) []string {
	if v.IsSet(key) {
		return v.GetStringSlice(key)
	}
	return defaultValue
}
