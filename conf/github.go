package conf

// GithubConfig github config
type GithubConfig struct {
	ID     string
	Secret string
}

func getGithubConfig() *GithubConfig {
	return &GithubConfig{
		ID:     c.GetString("github.id"),
		Secret: c.GetString("github.secret"),
	}
}
