package config

import "github.com/spf13/viper"

// Plugin plugin config struct
type Plugin struct {
	Mode      string
	Path      string
	Includes  []string
	Excludes  []string
	HotReload bool
}

// getPluginConfig returns the plugin config
func getPluginConfig(v *viper.Viper) *Plugin {
	return &Plugin{
		Mode:      v.GetString("plugin.mode"),
		Path:      v.GetString("plugin.path"),
		Includes:  v.GetStringSlice("plugin.includes"),
		Excludes:  v.GetStringSlice("plugin.excludes"),
		HotReload: v.GetBool("plugin.hot_reload"),
	}
}
