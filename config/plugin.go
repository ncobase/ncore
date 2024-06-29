package config

import "github.com/spf13/viper"

// Plugin plugin config struct
type Plugin struct {
	Path     string
	Includes []string
	Excludes []string
}

// getPluginConfig returns the plugin config
func getPluginConfig(v *viper.Viper) Plugin {
	return Plugin{
		Path:     v.GetString("plugin.path"),
		Includes: v.GetStringSlice("plugin.includes"),
		Excludes: v.GetStringSlice("plugin.excludes"),
	}
}
