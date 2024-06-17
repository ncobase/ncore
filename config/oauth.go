package config

import "github.com/spf13/viper"

// OAuth oauth config struct
type OAuth struct {
	Github   Github
	Facebook Facebook
	Google   Google
}

func getOAuthConfig(v *viper.Viper) OAuth {
	return OAuth{
		Github:   getGithubConfig(v),
		Facebook: getFacebookConfig(v),
		Google:   getGoogleConfig(v),
	}
}

// Github github config struct
type Github struct {
	ID     string
	Secret string
}

func getGithubConfig(v *viper.Viper) Github {
	return Github{
		ID:     v.GetString("oauth.github.id"),
		Secret: v.GetString("oauth.github.secret"),
	}
}

// Facebook facebook config struct
type Facebook struct {
	ID     string
	Secret string
}

func getFacebookConfig(v *viper.Viper) Facebook {
	return Facebook{
		ID:     v.GetString("oauth.facebook.id"),
		Secret: v.GetString("oauth.facebook.secret"),
	}
}

// Google google config struct
type Google struct {
	ID     string
	Secret string
}

func getGoogleConfig(v *viper.Viper) Google {
	return Google{
		ID:     v.GetString("oauth.google.id"),
		Secret: v.GetString("oauth.google.secret"),
	}
}
