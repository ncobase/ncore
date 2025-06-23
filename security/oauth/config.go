package oauth

import (
	"strings"

	"github.com/spf13/viper"
)

// Config represents OAuth configuration
type Config struct {
	Providers    map[string]*ProviderConfig `json:"providers" yaml:"providers"`
	DefaultScope []string                   `json:"default_scope" yaml:"default_scope"`
	EnablePKCE   bool                       `json:"enable_pkce" yaml:"enable_pkce"`
	StateSecret  string                     `json:"state_secret" yaml:"state_secret"`
}

// ProviderConfig represents OAuth provider configuration
type ProviderConfig struct {
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

// GetConfig loads OAuth configuration from viper
func GetConfig(v *viper.Viper) *Config {
	config := &Config{
		Providers:    make(map[string]*ProviderConfig),
		DefaultScope: v.GetStringSlice("oauth.default_scope"),
		EnablePKCE:   v.GetBool("oauth.enable_pkce"),
		StateSecret:  v.GetString("oauth.state_secret"),
	}

	// Load predefined providers
	providers := GetSupportedProviders()
	for _, provider := range providers {
		if v.IsSet("oauth." + provider) {
			config.Providers[provider] = getProviderConfig(v, provider)
		}
	}

	// Load custom providers
	if v.IsSet("oauth.custom") {
		customProviders := v.GetStringMap("oauth.custom")
		for name := range customProviders {
			config.Providers[name] = getProviderConfig(v, "custom."+name)
		}
	}

	return config
}

// getProviderConfig loads provider-specific configuration
func getProviderConfig(v *viper.Viper, provider string) *ProviderConfig {
	prefix := "oauth." + provider

	// Handle custom provider prefix
	if strings.HasPrefix(provider, "custom.") {
		prefix = "oauth." + provider
	}

	pc := &ProviderConfig{
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

	// Set defaults for known providers
	providerName := strings.TrimPrefix(provider, "custom.")
	setProviderDefaults(providerName, pc)

	return pc
}

// setProviderDefaults sets default URLs and scopes for known providers
func setProviderDefaults(provider string, config *ProviderConfig) {
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

// GetProviderInfo returns provider capability information
func GetProviderInfo(provider string) *ProviderInfo {
	infos := map[string]*ProviderInfo{
		"google": {
			Name:               "google",
			DisplayName:        "Google",
			Icon:               "google",
			SupportedScopes:    []string{"openid", "email", "profile"},
			RequiredScopes:     []string{"email"},
			SupportsPKCE:       true,
			SupportsRefresh:    true,
			SupportsRevocation: true,
		},
		"github": {
			Name:               "github",
			DisplayName:        "GitHub",
			Icon:               "github",
			SupportedScopes:    []string{"user", "user:email", "repo", "read:org"},
			RequiredScopes:     []string{"user:email"},
			SupportsPKCE:       true,
			SupportsRefresh:    false,
			SupportsRevocation: false,
		},
		"facebook": {
			Name:               "facebook",
			DisplayName:        "Facebook",
			Icon:               "facebook",
			SupportedScopes:    []string{"email", "public_profile"},
			RequiredScopes:     []string{"email"},
			SupportsPKCE:       true,
			SupportsRefresh:    false,
			SupportsRevocation: false,
		},
		"microsoft": {
			Name:               "microsoft",
			DisplayName:        "Microsoft",
			Icon:               "microsoft",
			SupportedScopes:    []string{"openid", "email", "profile"},
			RequiredScopes:     []string{"email"},
			SupportsPKCE:       true,
			SupportsRefresh:    true,
			SupportsRevocation: true,
		},
		"apple": {
			Name:               "apple",
			DisplayName:        "Apple",
			Icon:               "apple",
			SupportedScopes:    []string{"name", "email"},
			RequiredScopes:     []string{"email"},
			SupportsPKCE:       true,
			SupportsRefresh:    true,
			SupportsRevocation: true,
		},
		"twitter": {
			Name:               "twitter",
			DisplayName:        "Twitter",
			Icon:               "twitter",
			SupportedScopes:    []string{"tweet.read", "users.read"},
			RequiredScopes:     []string{"users.read"},
			SupportsPKCE:       true,
			SupportsRefresh:    true,
			SupportsRevocation: true,
		},
		"linkedin": {
			Name:               "linkedin",
			DisplayName:        "LinkedIn",
			Icon:               "linkedin",
			SupportedScopes:    []string{"r_liteprofile", "r_emailaddress"},
			RequiredScopes:     []string{"r_emailaddress"},
			SupportsPKCE:       false,
			SupportsRefresh:    true,
			SupportsRevocation: false,
		},
		"tiktok": {
			Name:               "tiktok",
			DisplayName:        "TikTok",
			Icon:               "tiktok",
			SupportedScopes:    []string{"user.info.basic", "user.info.profile", "video.list"},
			RequiredScopes:     []string{"user.info.basic"},
			SupportsPKCE:       true,
			SupportsRefresh:    true,
			SupportsRevocation: false,
		},
		"wechat": {
			Name:               "wechat",
			DisplayName:        "微信",
			Icon:               "wechat",
			SupportedScopes:    []string{"snsapi_login", "snsapi_userinfo"},
			RequiredScopes:     []string{"snsapi_login"},
			SupportsPKCE:       false,
			SupportsRefresh:    true,
			SupportsRevocation: false,
		},
		"alipay": {
			Name:               "alipay",
			DisplayName:        "支付宝",
			Icon:               "alipay",
			SupportedScopes:    []string{"auth_user", "auth_base"},
			RequiredScopes:     []string{"auth_user"},
			SupportsPKCE:       false,
			SupportsRefresh:    false,
			SupportsRevocation: false,
		},
		"baidu": {
			Name:               "baidu",
			DisplayName:        "百度",
			Icon:               "baidu",
			SupportedScopes:    []string{"basic", "netdisk"},
			RequiredScopes:     []string{"basic"},
			SupportsPKCE:       false,
			SupportsRefresh:    true,
			SupportsRevocation: false,
		},
		"weibo": {
			Name:               "weibo",
			DisplayName:        "微博",
			Icon:               "weibo",
			SupportedScopes:    []string{"email", "direct_messages_read"},
			RequiredScopes:     []string{"email"},
			SupportsPKCE:       false,
			SupportsRefresh:    false,
			SupportsRevocation: false,
		},
		"qq": {
			Name:               "qq",
			DisplayName:        "QQ",
			Icon:               "qq",
			SupportedScopes:    []string{"get_user_info", "list_album"},
			RequiredScopes:     []string{"get_user_info"},
			SupportsPKCE:       false,
			SupportsRefresh:    false,
			SupportsRevocation: false,
		},
	}

	if info, exists := infos[provider]; exists {
		return info
	}

	return &ProviderInfo{
		Name:        provider,
		DisplayName: strings.Title(provider),
	}
}

// GetSupportedProviders returns list of supported providers
func GetSupportedProviders() []string {
	return []string{
		string(ProviderGoogle),
		string(ProviderGitHub),
		string(ProviderFacebook),
		string(ProviderMicrosoft),
		string(ProviderApple),
		string(ProviderTwitter),
		string(ProviderLinkedIn),
		string(ProviderTikTok),
		string(ProviderWeChat),
		string(ProviderAlipay),
		string(ProviderBaidu),
		string(ProviderWeibo),
		string(ProviderQQ),
	}
}

// ValidateProvider checks if provider is supported
func ValidateProvider(provider string) bool {
	supportedProviders := GetSupportedProviders()
	for _, p := range supportedProviders {
		if p == provider {
			return true
		}
	}
	return false
}
