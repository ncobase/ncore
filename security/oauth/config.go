package oauth

import (
	"strings"

	nconfig "github.com/ncobase/ncore/config"
	"github.com/spf13/viper"
)

// Config contains OAuth infrastructure settings.
type Config = nconfig.OAuth

// ProviderConfig contains provider-specific OAuth settings.
type ProviderConfig = nconfig.OAuthProvider

// GetConfig loads OAuth configuration from viper.
func GetConfig(v *viper.Viper) *Config {
	return nconfig.GetOAuthConfig(v)
}

// GetProviderInfo returns provider capability information.
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
			DisplayName:        "WeChat",
			Icon:               "wechat",
			SupportedScopes:    []string{"snsapi_login", "snsapi_userinfo"},
			RequiredScopes:     []string{"snsapi_login"},
			SupportsPKCE:       false,
			SupportsRefresh:    true,
			SupportsRevocation: false,
		},
		"alipay": {
			Name:               "alipay",
			DisplayName:        "Alipay",
			Icon:               "alipay",
			SupportedScopes:    []string{"auth_user", "auth_base"},
			RequiredScopes:     []string{"auth_user"},
			SupportsPKCE:       false,
			SupportsRefresh:    false,
			SupportsRevocation: false,
		},
		"baidu": {
			Name:               "baidu",
			DisplayName:        "Baidu",
			Icon:               "baidu",
			SupportedScopes:    []string{"basic", "netdisk"},
			RequiredScopes:     []string{"basic"},
			SupportsPKCE:       false,
			SupportsRefresh:    true,
			SupportsRevocation: false,
		},
		"weibo": {
			Name:               "weibo",
			DisplayName:        "Weibo",
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

// GetSupportedProviders returns all built-in OAuth provider names.
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

// ValidateProvider checks whether provider is a built-in OAuth provider.
func ValidateProvider(provider string) bool {
	for _, supported := range GetSupportedProviders() {
		if supported == provider {
			return true
		}
	}
	return false
}
