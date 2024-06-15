package config

// OAuth oauth config struct
type OAuth struct {
	Github   Github
	Facebook Facebook
	Google   Google
}

func getOAuthConfig() OAuth {
	return OAuth{
		Github:   getGithubConfig(),
		Facebook: getFacebookConfig(),
		Google:   getGoogleConfig(),
	}
}

// Github github config struct
type Github struct {
	ID     string
	Secret string
}

func getGithubConfig() Github {
	return Github{
		ID:     c.GetString("oauth.github.id"),
		Secret: c.GetString("oauth.github.secret"),
	}
}

// Facebook facebook config struct
type Facebook struct {
	ID     string
	Secret string
}

func getFacebookConfig() Facebook {
	return Facebook{
		ID:     c.GetString("oauth.facebook.id"),
		Secret: c.GetString("oauth.facebook.secret"),
	}
}

// Google google config struct
type Google struct {
	ID     string
	Secret string
}

func getGoogleConfig() Google {
	return Google{
		ID:     c.GetString("oauth.google.id"),
		Secret: c.GetString("oauth.google.secret"),
	}
}
