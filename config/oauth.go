package config

import (
	"strings"

	"github.com/spf13/viper"
)

// OAuth contains OAuth infrastructure settings.
type OAuth struct {
	Providers    map[string]*OAuthProvider `json:"providers" yaml:"providers"`
	DefaultScope []string                  `json:"default_scope" yaml:"default_scope"`
	EnablePKCE   bool                      `json:"enable_pkce" yaml:"enable_pkce"`
	StateSecret  string                    `json:"state_secret" yaml:"state_secret"`
}

// OAuthProvider contains provider-specific OAuth settings.
type OAuthProvider struct {
	ClientID     string            `json:"client_id" yaml:"client_id"`
	ClientSecret string            `json:"client_secret" yaml:"client_secret"`
	RedirectURL  string            `json:"redirect_url" yaml:"redirect_url"`
	Scopes       []string          `json:"scopes" yaml:"scopes"`
	AuthURL      string            `json:"auth_url" yaml:"auth_url"`
	TokenURL     string            `json:"token_url" yaml:"token_url"`
	UserInfoURL  string            `json:"user_info_url" yaml:"user_info_url"`
	RevokeURL    string            `json:"revoke_url" yaml:"revoke_url"`
	Enabled      bool              `json:"enabled" yaml:"enabled"`
	ExtraParams  map[string]string `json:"extra_params" yaml:"extra_params"`
}

// GetOAuthConfig loads OAuth configuration from viper.
func GetOAuthConfig(v *viper.Viper) *OAuth {
	config := &OAuth{
		Providers:    make(map[string]*OAuthProvider),
		DefaultScope: v.GetStringSlice("oauth.default_scope"),
		EnablePKCE:   v.GetBool("oauth.enable_pkce"),
		StateSecret:  v.GetString("oauth.state_secret"),
	}

	for _, provider := range supportedOAuthProviders() {
		if v.IsSet("oauth." + provider) {
			config.Providers[provider] = getOAuthProviderConfig(v, provider)
		}
	}

	if v.IsSet("oauth.custom") {
		customProviders := v.GetStringMap("oauth.custom")
		for name := range customProviders {
			config.Providers[name] = getOAuthProviderConfig(v, "custom."+name)
		}
	}

	return config
}

func getOAuthConfig(v *viper.Viper) *OAuth {
	return GetOAuthConfig(v)
}

func getOAuthProviderConfig(v *viper.Viper, provider string) *OAuthProvider {
	prefix := "oauth." + provider
	config := &OAuthProvider{
		ClientID:     v.GetString(prefix + ".client_id"),
		ClientSecret: v.GetString(prefix + ".client_secret"),
		RedirectURL:  v.GetString(prefix + ".redirect_url"),
		Scopes:       v.GetStringSlice(prefix + ".scopes"),
		AuthURL:      v.GetString(prefix + ".auth_url"),
		TokenURL:     v.GetString(prefix + ".token_url"),
		UserInfoURL:  v.GetString(prefix + ".user_info_url"),
		RevokeURL:    v.GetString(prefix + ".revoke_url"),
		Enabled:      v.GetBool(prefix + ".enabled"),
		ExtraParams:  v.GetStringMapString(prefix + ".extra_params"),
	}

	setOAuthProviderDefaults(strings.TrimPrefix(provider, "custom."), config)
	return config
}

func supportedOAuthProviders() []string {
	return []string{
		"google",
		"github",
		"facebook",
		"microsoft",
		"apple",
		"twitter",
		"linkedin",
		"tiktok",
		"wechat",
		"alipay",
		"baidu",
		"weibo",
		"qq",
	}
}

func setOAuthProviderDefaults(provider string, config *OAuthProvider) {
	switch provider {
	case "google":
		if config.AuthURL == "" {
			config.AuthURL = "https://accounts.google.com/o/oauth2/v2/auth"
		}
		if config.TokenURL == "" {
			config.TokenURL = "https://oauth2.googleapis.com/token"
		}
		if config.UserInfoURL == "" {
			config.UserInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"
		}
		if config.RevokeURL == "" {
			config.RevokeURL = "https://oauth2.googleapis.com/revoke"
		}
		if len(config.Scopes) == 0 {
			config.Scopes = []string{"openid", "email", "profile"}
		}
	case "github":
		if config.AuthURL == "" {
			config.AuthURL = "https://github.com/login/oauth/authorize"
		}
		if config.TokenURL == "" {
			config.TokenURL = "https://github.com/login/oauth/access_token"
		}
		if config.UserInfoURL == "" {
			config.UserInfoURL = "https://api.github.com/user"
		}
		if len(config.Scopes) == 0 {
			config.Scopes = []string{"user:email"}
		}
	case "facebook":
		if config.AuthURL == "" {
			config.AuthURL = "https://www.facebook.com/v18.0/dialog/oauth"
		}
		if config.TokenURL == "" {
			config.TokenURL = "https://graph.facebook.com/v18.0/oauth/access_token"
		}
		if config.UserInfoURL == "" {
			config.UserInfoURL = "https://graph.facebook.com/v18.0/me"
		}
		if len(config.Scopes) == 0 {
			config.Scopes = []string{"email", "public_profile"}
		}
	case "microsoft":
		if config.AuthURL == "" {
			config.AuthURL = "https://login.microsoftonline.com/common/oauth2/v2.0/authorize"
		}
		if config.TokenURL == "" {
			config.TokenURL = "https://login.microsoftonline.com/common/oauth2/v2.0/token"
		}
		if config.UserInfoURL == "" {
			config.UserInfoURL = "https://graph.microsoft.com/v1.0/me"
		}
		if config.RevokeURL == "" {
			config.RevokeURL = "https://login.microsoftonline.com/common/oauth2/v2.0/logout"
		}
		if len(config.Scopes) == 0 {
			config.Scopes = []string{"openid", "email", "profile"}
		}
	case "apple":
		if config.AuthURL == "" {
			config.AuthURL = "https://appleid.apple.com/auth/authorize"
		}
		if config.TokenURL == "" {
			config.TokenURL = "https://appleid.apple.com/auth/token"
		}
		if config.RevokeURL == "" {
			config.RevokeURL = "https://appleid.apple.com/auth/revoke"
		}
		if len(config.Scopes) == 0 {
			config.Scopes = []string{"name", "email"}
		}
	case "twitter":
		if config.AuthURL == "" {
			config.AuthURL = "https://twitter.com/i/oauth2/authorize"
		}
		if config.TokenURL == "" {
			config.TokenURL = "https://api.twitter.com/2/oauth2/token"
		}
		if config.UserInfoURL == "" {
			config.UserInfoURL = "https://api.twitter.com/2/users/me"
		}
		if config.RevokeURL == "" {
			config.RevokeURL = "https://api.twitter.com/2/oauth2/revoke"
		}
		if len(config.Scopes) == 0 {
			config.Scopes = []string{"tweet.read", "users.read"}
		}
	case "linkedin":
		if config.AuthURL == "" {
			config.AuthURL = "https://www.linkedin.com/oauth/v2/authorization"
		}
		if config.TokenURL == "" {
			config.TokenURL = "https://www.linkedin.com/oauth/v2/accessToken"
		}
		if config.UserInfoURL == "" {
			config.UserInfoURL = "https://api.linkedin.com/v2/people/~"
		}
		if len(config.Scopes) == 0 {
			config.Scopes = []string{"r_liteprofile", "r_emailaddress"}
		}
	case "tiktok":
		if config.AuthURL == "" {
			config.AuthURL = "https://www.tiktok.com/auth/authorize/"
		}
		if config.TokenURL == "" {
			config.TokenURL = "https://open-api.tiktok.com/oauth/access_token/"
		}
		if config.UserInfoURL == "" {
			config.UserInfoURL = "https://open-api.tiktok.com/oauth/userinfo/"
		}
		if len(config.Scopes) == 0 {
			config.Scopes = []string{"user.info.basic"}
		}
	case "wechat":
		if config.AuthURL == "" {
			config.AuthURL = "https://open.weixin.qq.com/connect/qrconnect"
		}
		if config.TokenURL == "" {
			config.TokenURL = "https://api.weixin.qq.com/sns/oauth2/access_token"
		}
		if config.UserInfoURL == "" {
			config.UserInfoURL = "https://api.weixin.qq.com/sns/userinfo"
		}
		if len(config.Scopes) == 0 {
			config.Scopes = []string{"snsapi_login"}
		}
	case "alipay":
		if config.AuthURL == "" {
			config.AuthURL = "https://openauth.alipay.com/oauth2/publicAppAuthorize.htm"
		}
		if config.TokenURL == "" {
			config.TokenURL = "https://openapi.alipay.com/gateway.do"
		}
		if config.UserInfoURL == "" {
			config.UserInfoURL = "https://openapi.alipay.com/gateway.do"
		}
		if len(config.Scopes) == 0 {
			config.Scopes = []string{"auth_user"}
		}
	case "baidu":
		if config.AuthURL == "" {
			config.AuthURL = "https://openapi.baidu.com/oauth/2.0/authorize"
		}
		if config.TokenURL == "" {
			config.TokenURL = "https://openapi.baidu.com/oauth/2.0/token"
		}
		if config.UserInfoURL == "" {
			config.UserInfoURL = "https://openapi.baidu.com/rest/2.0/passport/users/getInfo"
		}
		if len(config.Scopes) == 0 {
			config.Scopes = []string{"basic"}
		}
	case "weibo":
		if config.AuthURL == "" {
			config.AuthURL = "https://api.weibo.com/oauth2/authorize"
		}
		if config.TokenURL == "" {
			config.TokenURL = "https://api.weibo.com/oauth2/access_token"
		}
		if config.UserInfoURL == "" {
			config.UserInfoURL = "https://api.weibo.com/2/users/show.json"
		}
		if len(config.Scopes) == 0 {
			config.Scopes = []string{"email"}
		}
	case "qq":
		if config.AuthURL == "" {
			config.AuthURL = "https://graph.qq.com/oauth2.0/authorize"
		}
		if config.TokenURL == "" {
			config.TokenURL = "https://graph.qq.com/oauth2.0/token"
		}
		if config.UserInfoURL == "" {
			config.UserInfoURL = "https://graph.qq.com/user/get_user_info"
		}
		if len(config.Scopes) == 0 {
			config.Scopes = []string{"get_user_info"}
		}
	}
}
