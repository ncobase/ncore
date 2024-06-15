package oauth

import (
	"encoding/json"
	"errors"
	"fmt"

	"golang.org/x/oauth2"
)

// ProviderConfig holds the configuration for each OAuth provider.
type ProviderConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// State represents the OAuth state with provider and next URL information.
type State struct {
	Provider string `json:"provider"`
	Next     string `json:"next"`
}

// Profile represents the user profile information obtained from OAuth.
type Profile struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Thumbnail string `json:"thumbnail"`
}

// Action defines the methods required for OAuth actions for different providers.
type Action interface {
	Wechat() string
	QQ() string
	StackOverflow() string
	Github(id, redirectURL string) string
	Facebook(id, redirectURL string) string
	Google(id, redirectURL string) string
}

// OAuthState contains the state and implements the Action interface.
type OAuthState struct {
	State
}

// Wechat generates the OAuth URL for WeChat (not implemented).
func (s OAuthState) Wechat() string {
	return "OAuth for WeChat is not implemented."
}

// QQ generates the OAuth URL for QQ (not implemented).
func (s OAuthState) QQ() string {
	return "OAuth for QQ is not implemented."
}

// StackOverflow generates the OAuth URL for StackOverflow (not implemented).
func (s OAuthState) StackOverflow() string {
	return "OAuth for StackOverflow is not implemented."
}

// Google generates the OAuth URL for Google.
func (s OAuthState) Google(id, redirectURL string) string {
	redirectURL = redirectURL + "/v1/oauth/callback/google"
	state, err := json.Marshal(s)
	if err != nil {
		return ""
	}
	oauthConfig := &oauth2.Config{
		ClientID:     id,
		ClientSecret: "", // ClientSecret should be securely obtained
		RedirectURL:  redirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
	}
	return oauthConfig.AuthCodeURL(string(state))
}

// Facebook generates the OAuth URL for Facebook.
func (s OAuthState) Facebook(id, redirectURL string) string {
	redirectURL = redirectURL + "/v1/oauth/callback/facebook"
	state, err := json.Marshal(s)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("https://www.facebook.com/v4.0/dialog/oauth?client_id=%v&redirect_uri=%v&state=%v&scope=email,public_profile", id, redirectURL, state)
}

// Github generates the OAuth URL for GitHub.
func (s OAuthState) Github(id, redirectURL string) string {
	redirectURLWithNext := redirectURL + "/v1/oauth/callback/github?next=" + s.Next
	return fmt.Sprintf("https://github.com/login/oauth/authorize?scope=user:email&client_id=%v&redirect_uri=%v", id, redirectURLWithNext)
}

// OAuth initializes a new OAuthState with the given parameters.
func OAuth(provider, next string) *OAuthState {
	return &OAuthState{
		State: State{
			Provider: provider,
			Next:     next,
		},
	}
}

// GenerateOAuthLink generates an OAuth link for the specified provider.
func GenerateOAuthLink(provider, id, redirectURL, next string) string {
	state := OAuth(provider, next)
	switch provider {
	case "facebook":
		return state.Facebook(id, redirectURL)
	case "google":
		return state.Google(id, redirectURL)
	case "github":
		return state.Github(id, redirectURL)
	default:
		return ""
	}
}

// GetOAuthProfile retrieves the user's profile using the provided access token.
func GetOAuthProfile(provider, accessToken string) (*Profile, error) {
	switch provider {
	case "facebook":
		return GetFacebookProfile(accessToken)
	case "github":
		return GetGithubProfile(accessToken)
	default:
		return nil, errors.New("unsupported provider")
	}
}
