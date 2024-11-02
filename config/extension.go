package config

import "github.com/spf13/viper"

// Extension extension config struct
type Extension struct {
	Mode      string
	Path      string
	Includes  []string
	Excludes  []string
	HotReload bool
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
