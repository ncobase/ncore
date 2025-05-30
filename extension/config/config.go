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

	// Advanced configuration
	MaxPlugins        int                `json:"max_plugins" yaml:"max_plugins"`
	LoadTimeout       string             `json:"load_timeout" yaml:"load_timeout"`
	InitTimeout       string             `json:"init_timeout" yaml:"init_timeout"`
	DependencyTimeout string             `json:"dependency_timeout" yaml:"dependency_timeout"`
	PluginConfig      map[string]any     `json:"plugin_config" yaml:"plugin_config"`
	Security          *SecurityConfig    `json:"security" yaml:"security"`
	Performance       *PerformanceConfig `json:"performance" yaml:"performance"`
	Monitoring        *MonitoringConfig  `json:"monitoring" yaml:"monitoring"`
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

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.MaxPlugins <= 0 {
		return fmt.Errorf("max_plugins must be greater than 0, got: %d", c.MaxPlugins)
	}

	// Validate timeout values
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

	// Validate security config
	if c.Security != nil {
		if err := c.Security.Validate(); err != nil {
			return fmt.Errorf("security config validation failed: %v", err)
		}
	}

	// Validate performance config
	if c.Performance != nil {
		if err := c.Performance.Validate(); err != nil {
			return fmt.Errorf("performance config validation failed: %v", err)
		}
	}

	return nil
}

// SecurityConfig defines security-related extension settings
type SecurityConfig struct {
	EnableSandbox     bool     `json:"enable_sandbox" yaml:"enable_sandbox"`
	AllowedPaths      []string `json:"allowed_paths" yaml:"allowed_paths"`
	BlockedExtensions []string `json:"blocked_extensions" yaml:"blocked_extensions"`
	TrustedSources    []string `json:"trusted_sources" yaml:"trusted_sources"`
	RequireSignature  bool     `json:"require_signature" yaml:"require_signature"`
	AllowUnsafe       bool     `json:"allow_unsafe" yaml:"allow_unsafe"`       // For development
	SkipValidation    bool     `json:"skip_validation" yaml:"skip_validation"` // For development
}

// Validate validates security configuration
func (s *SecurityConfig) Validate() error {
	if s.EnableSandbox && len(s.AllowedPaths) == 0 && !s.AllowUnsafe {
		return fmt.Errorf("sandbox enabled but no allowed paths specified and unsafe mode disabled")
	}

	if s.RequireSignature && s.SkipValidation {
		return fmt.Errorf("cannot require signature while skipping validation")
	}

	return nil
}

// PerformanceConfig defines performance-related extension settings
type PerformanceConfig struct {
	MaxMemoryMB            int    `json:"max_memory_mb" yaml:"max_memory_mb"`
	MaxCPUPercent          int    `json:"max_cpu_percent" yaml:"max_cpu_percent"`
	EnableMetrics          bool   `json:"enable_metrics" yaml:"enable_metrics"`
	MetricsInterval        string `json:"metrics_interval" yaml:"metrics_interval"`
	EnableProfiling        bool   `json:"enable_profiling" yaml:"enable_profiling"`
	GarbageCollectInterval string `json:"gc_interval" yaml:"gc_interval"`
	MaxConcurrentLoads     int    `json:"max_concurrent_loads" yaml:"max_concurrent_loads"`
	MemoryThresholdPercent int    `json:"memory_threshold_percent" yaml:"memory_threshold_percent"`
}

// Validate validates performance configuration
func (p *PerformanceConfig) Validate() error {
	if p.MaxMemoryMB <= 0 {
		return fmt.Errorf("max_memory_mb must be greater than 0, got: %d", p.MaxMemoryMB)
	}

	if p.MaxCPUPercent <= 0 || p.MaxCPUPercent > 100 {
		return fmt.Errorf("max_cpu_percent must be between 1-100, got: %d", p.MaxCPUPercent)
	}

	if p.MaxConcurrentLoads <= 0 {
		return fmt.Errorf("max_concurrent_loads must be greater than 0, got: %d", p.MaxConcurrentLoads)
	}

	if p.MemoryThresholdPercent <= 0 || p.MemoryThresholdPercent > 100 {
		return fmt.Errorf("memory_threshold_percent must be between 1-100, got: %d", p.MemoryThresholdPercent)
	}

	// Validate interval values
	intervals := map[string]string{
		"metrics_interval": p.MetricsInterval,
		"gc_interval":      p.GarbageCollectInterval,
	}

	for name, interval := range intervals {
		if interval != "" {
			if _, err := time.ParseDuration(interval); err != nil {
				return fmt.Errorf("invalid %s: %v", name, err)
			}
		}
	}

	return nil
}

// MonitoringConfig defines monitoring-related settings
type MonitoringConfig struct {
	EnableHealthCheck      bool   `json:"enable_health_check" yaml:"enable_health_check"`
	HealthCheckInterval    string `json:"health_check_interval" yaml:"health_check_interval"`
	EnableDetailedMetrics  bool   `json:"enable_detailed_metrics" yaml:"enable_detailed_metrics"`
	MetricsRetention       string `json:"metrics_retention" yaml:"metrics_retention"`
	AlertingEnabled        bool   `json:"alerting_enabled" yaml:"alerting_enabled"`
	SlowOperationThreshold string `json:"slow_operation_threshold" yaml:"slow_operation_threshold"`
}

// Validate validates monitoring configuration
func (m *MonitoringConfig) Validate() error {
	intervals := map[string]string{
		"health_check_interval":    m.HealthCheckInterval,
		"metrics_retention":        m.MetricsRetention,
		"slow_operation_threshold": m.SlowOperationThreshold,
	}

	for name, interval := range intervals {
		if interval != "" {
			if _, err := time.ParseDuration(interval); err != nil {
				return fmt.Errorf("invalid %s: %v", name, err)
			}
		}
	}

	return nil
}

// GetDefaultSecurityConfig returns environment-specific default security configuration
func GetDefaultSecurityConfig(isDevelopment bool) *SecurityConfig {
	if isDevelopment {
		return &SecurityConfig{
			EnableSandbox:     false, // Disabled for development ease
			AllowedPaths:      []string{},
			BlockedExtensions: []string{".exe", ".bat", ".cmd", ".ps1"},
			TrustedSources:    []string{},
			RequireSignature:  false, // Disabled for development
			AllowUnsafe:       true,  // Allow for development
			SkipValidation:    true,  // Skip for development
		}
	}

	return &SecurityConfig{
		EnableSandbox:     true,
		AllowedPaths:      []string{"/opt/plugins", "/usr/local/plugins"},
		BlockedExtensions: []string{".exe", ".bat", ".cmd", ".ps1", ".sh", ".py"},
		TrustedSources:    []string{},
		RequireSignature:  true,
		AllowUnsafe:       false,
		SkipValidation:    false,
	}
}

// GetDefaultPerformanceConfig returns environment-specific default performance configuration
func GetDefaultPerformanceConfig(isDevelopment bool) *PerformanceConfig {
	if isDevelopment {
		return &PerformanceConfig{
			MaxMemoryMB:            1024, // Higher for development
			MaxCPUPercent:          90,   // Higher for development
			EnableMetrics:          true,
			MetricsInterval:        "10s", // More frequent for development
			EnableProfiling:        true,  // Enabled for development
			GarbageCollectInterval: "2m",  // More frequent for development
			MaxConcurrentLoads:     10,    // Higher for development
			MemoryThresholdPercent: 85,    // Higher threshold for development
		}
	}

	return &PerformanceConfig{
		MaxMemoryMB:            512,
		MaxCPUPercent:          70, // More conservative for production
		EnableMetrics:          true,
		MetricsInterval:        "30s",
		EnableProfiling:        false, // Disabled for production
		GarbageCollectInterval: "5m",
		MaxConcurrentLoads:     5,  // Conservative for production
		MemoryThresholdPercent: 75, // Conservative for production
	}
}

// GetDefaultMonitoringConfig returns default monitoring configuration
func GetDefaultMonitoringConfig(isDevelopment bool) *MonitoringConfig {
	if isDevelopment {
		return &MonitoringConfig{
			EnableHealthCheck:      true,
			HealthCheckInterval:    "30s",
			EnableDetailedMetrics:  true,
			MetricsRetention:       "1h",  // Shorter retention for development
			AlertingEnabled:        false, // Disabled for development
			SlowOperationThreshold: "1s",  // Lower threshold for development
		}
	}

	return &MonitoringConfig{
		EnableHealthCheck:      true,
		HealthCheckInterval:    "60s", // Less frequent for production
		EnableDetailedMetrics:  false, // Conservative for production
		MetricsRetention:       "24h", // Longer retention for production
		AlertingEnabled:        true,  // Enabled for production
		SlowOperationThreshold: "5s",  // Higher threshold for production
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
		HotReload: getBoolWithDefault(v, "extension.hot_reload", isDevelopment), // Enable for dev

		// Improved defaults
		MaxPlugins:        getIntWithDefault(v, "extension.max_plugins", getDefaultMaxPlugins(isDevelopment)),
		LoadTimeout:       getStringWithDefault(v, "extension.load_timeout", getDefaultLoadTimeout(isDevelopment)),
		InitTimeout:       getStringWithDefault(v, "extension.init_timeout", getDefaultInitTimeout(isDevelopment)),
		DependencyTimeout: getStringWithDefault(v, "extension.dependency_timeout", "15s"),
		PluginConfig:      v.GetStringMap("extension.plugin_config"),

		Security:    getSecurityConfig(v, isDevelopment),
		Performance: getPerformanceConfig(v, isDevelopment),
		Monitoring:  getMonitoringConfig(v, isDevelopment),
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
		SkipValidation:    getBoolWithDefault(v, "extension.security.skip_validation", defaultConfig.SkipValidation),
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
		EnableProfiling:        getBoolWithDefault(v, "extension.performance.enable_profiling", defaultConfig.EnableProfiling),
		GarbageCollectInterval: getStringWithDefault(v, "extension.performance.gc_interval", defaultConfig.GarbageCollectInterval),
		MaxConcurrentLoads:     getIntWithDefault(v, "extension.performance.max_concurrent_loads", defaultConfig.MaxConcurrentLoads),
		MemoryThresholdPercent: getIntWithDefault(v, "extension.performance.memory_threshold_percent", defaultConfig.MemoryThresholdPercent),
	}
}

// getMonitoringConfig returns monitoring configuration
func getMonitoringConfig(v *viper.Viper, isDevelopment bool) *MonitoringConfig {
	defaultConfig := GetDefaultMonitoringConfig(isDevelopment)

	if !v.IsSet("extension.monitoring") {
		return defaultConfig
	}

	return &MonitoringConfig{
		EnableHealthCheck:      getBoolWithDefault(v, "extension.monitoring.enable_health_check", defaultConfig.EnableHealthCheck),
		HealthCheckInterval:    getStringWithDefault(v, "extension.monitoring.health_check_interval", defaultConfig.HealthCheckInterval),
		EnableDetailedMetrics:  getBoolWithDefault(v, "extension.monitoring.enable_detailed_metrics", defaultConfig.EnableDetailedMetrics),
		MetricsRetention:       getStringWithDefault(v, "extension.monitoring.metrics_retention", defaultConfig.MetricsRetention),
		AlertingEnabled:        getBoolWithDefault(v, "extension.monitoring.alerting_enabled", defaultConfig.AlertingEnabled),
		SlowOperationThreshold: getStringWithDefault(v, "extension.monitoring.slow_operation_threshold", defaultConfig.SlowOperationThreshold),
	}
}

// Environment-specific default value helpers

func getDefaultPluginPath(isDevelopment bool) string {
	if isDevelopment {
		return "./plugins"
	}
	return "/opt/ncore/plugins"
}

func getDefaultMaxPlugins(isDevelopment bool) int {
	if isDevelopment {
		return 50 // Lower for development
	}
	return 100 // Higher for production
}

func getDefaultLoadTimeout(isDevelopment bool) string {
	if isDevelopment {
		return "60s" // Longer for development (debugging)
	}
	return "30s" // Shorter for production
}

func getDefaultInitTimeout(isDevelopment bool) string {
	if isDevelopment {
		return "120s" // Longer for development (debugging)
	}
	return "60s" // Shorter for production
}

// Helper functions for default values

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
