package config

import "github.com/spf13/viper"

// Feature feature config struct
type Feature struct {
	Mode      string
	Path      string
	Includes  []string
	Excludes  []string
	HotReload bool
}

// getFeatureConfig returns the feature config
func getFeatureConfig(v *viper.Viper) *Feature {
	return &Feature{
		Mode:      v.GetString("feature.mode"),
		Path:      v.GetString("feature.path"),
		Includes:  v.GetStringSlice("feature.includes"),
		Excludes:  v.GetStringSlice("feature.excludes"),
		HotReload: v.GetBool("feature.hot_reload"),
	}
}
