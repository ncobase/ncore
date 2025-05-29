package config

import "github.com/spf13/viper"

// Extension extension config struct
type Extension struct {
	Mode      string   `json:"mode" yaml:"mode"`
	Path      string   `json:"path" yaml:"path"`
	Includes  []string `json:"includes" yaml:"includes"`
	Excludes  []string `json:"excludes" yaml:"excludes"`
	HotReload bool     `json:"hot_reload" yaml:"hot_reload"`
}

// getExtensionConfig returns the extension config
func getExtensionConfig(v *viper.Viper) *Extension {
	return &Extension{
		Mode:      v.GetString("extension.mode"),
		Path:      v.GetString("extension.path"),
		Includes:  v.GetStringSlice("extension.includes"),
		Excludes:  v.GetStringSlice("extension.excludes"),
		HotReload: v.GetBool("extension.hot_reload"),
	}
}
