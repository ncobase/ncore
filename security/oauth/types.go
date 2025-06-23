package oauth

import (
	"context"
	"time"
)

// Provider represents OAuth provider type
type Provider string

const (
	ProviderGoogle    Provider = "google"
	ProviderGitHub    Provider = "github"
	ProviderFacebook  Provider = "facebook"
	ProviderMicrosoft Provider = "microsoft"
	ProviderApple     Provider = "apple"
	ProviderTwitter   Provider = "twitter"
	ProviderLinkedIn  Provider = "linkedin"
	ProviderTikTok    Provider = "tiktok"
	ProviderWeChat    Provider = "wechat"
	ProviderAlipay    Provider = "alipay"
	ProviderBaidu     Provider = "baidu"
	ProviderWeibo     Provider = "weibo"
	ProviderQQ        Provider = "qq"
)

// Profile represents user profile from OAuth provider
type Profile struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Avatar   string `json:"avatar"`
	Username string `json:"username"`
	Provider string `json:"provider"`
	Verified bool   `json:"verified"`
	Locale   string `json:"locale,omitempty"`
}

// TokenResponse represents OAuth token response
type TokenResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
	Scope        string    `json:"scope"`
	IDToken      string    `json:"id_token,omitempty"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// StateData represents OAuth state information
type StateData struct {
	Provider  string    `json:"provider"`
	NextURL   string    `json:"next_url,omitempty"`
	UserID    string    `json:"user_id,omitempty"`
	Action    string    `json:"action,omitempty"` // "login", "register", "link"
	Timestamp int64     `json:"timestamp"`
	Nonce     string    `json:"nonce"`
	PKCE      *PKCEData `json:"pkce,omitempty"`
}

// PKCEData represents PKCE challenge data
type PKCEData struct {
	CodeVerifier  string `json:"code_verifier"`
	CodeChallenge string `json:"code_challenge"`
	Method        string `json:"method"` // "S256" or "plain"
}

// ClientInterface defines OAuth client interface
type ClientInterface interface {
	GetAuthURL(provider Provider, state string, additionalParams map[string]string) (string, error)
	ExchangeCodeForToken(ctx context.Context, provider Provider, code string, codeVerifier ...string) (*TokenResponse, error)
	RefreshAccessToken(ctx context.Context, provider Provider, refreshToken string) (*TokenResponse, error)
	GetUserProfile(ctx context.Context, provider Provider, accessToken string) (*Profile, error)
	ValidateToken(ctx context.Context, provider Provider, token string) (*TokenInfo, error)
	RevokeToken(ctx context.Context, provider Provider, token string) error
}

// TokenInfo represents token validation information
type TokenInfo struct {
	Valid     bool      `json:"valid"`
	ExpiresAt time.Time `json:"expires_at"`
	Scope     string    `json:"scope"`
	ClientID  string    `json:"client_id"`
}

// ProviderInfo represents provider capability information
type ProviderInfo struct {
	Name               string   `json:"name"`
	DisplayName        string   `json:"display_name"`
	Icon               string   `json:"icon"`
	SupportedScopes    []string `json:"supported_scopes"`
	RequiredScopes     []string `json:"required_scopes"`
	SupportsPKCE       bool     `json:"supports_pkce"`
	SupportsRefresh    bool     `json:"supports_refresh"`
	SupportsRevocation bool     `json:"supports_revocation"`
}
