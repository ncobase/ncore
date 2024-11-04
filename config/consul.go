package config

import "github.com/spf13/viper"

// Consul config struct
type Consul struct {
	Address   string `json:"address"`
	Scheme    string `json:"scheme"`
	Discovery struct {
		DefaultTags   []string          `json:"default_tags"`
		DefaultMeta   map[string]string `json:"default_meta"`
		HealthCheck   bool              `json:"health_check"`
		CheckInterval string            `json:"check_interval"`
		Timeout       string            `json:"timeout"`
	} `json:"discovery"`
}

// getConsulConfig get consul config
func getConsulConfig(v *viper.Viper) *Consul {
	// Get consul config
	consul := &Consul{
		Address: v.GetString("consul.address"),
		Scheme:  v.GetString("consul.scheme"),
	}

	// Get consul discovery config
	consul.Discovery.DefaultTags = v.GetStringSlice("consul.discovery.default_tags")
	consul.Discovery.DefaultMeta = v.GetStringMapString("consul.discovery.default_meta")
	consul.Discovery.HealthCheck = v.GetBool("consul.discovery.health_check")
	consul.Discovery.CheckInterval = v.GetString("consul.discovery.check_interval")
	consul.Discovery.Timeout = v.GetString("consul.discovery.timeout")

	// Set default values if not set
	if consul.Discovery.CheckInterval == "" {
		consul.Discovery.CheckInterval = "10s"
	}
	if consul.Discovery.Timeout == "" {
		consul.Discovery.Timeout = "5s"
	}

	return consul
}
