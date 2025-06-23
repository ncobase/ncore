package oauth

import "fmt"

// OAuth specific errors
var (
	ErrProviderNotSupported = fmt.Errorf("OAuth provider not supported")
	ErrProviderNotEnabled   = fmt.Errorf("OAuth provider not enabled")
	ErrInvalidConfiguration = fmt.Errorf("invalid OAuth configuration")
	ErrInvalidState         = fmt.Errorf("invalid OAuth state parameter")
	ErrStateExpired         = fmt.Errorf("OAuth state parameter expired")
	ErrCodeExchangeFailed   = fmt.Errorf("OAuth code exchange failed")
	ErrTokenRefreshFailed   = fmt.Errorf("OAuth token refresh failed")
	ErrProfileFetchFailed   = fmt.Errorf("failed to fetch user profile")
	ErrInvalidToken         = fmt.Errorf("invalid OAuth token")
	ErrTokenRevokeFailed    = fmt.Errorf("OAuth token revocation failed")
	ErrMissingRequiredScope = fmt.Errorf("missing required OAuth scope")
)

// Error represents an OAuth specific error
type Error struct {
	Provider string
	Code     string
	Message  string
	Err      error
}

// Error implements the error interface
func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("OAuth error for provider %s [%s]: %s - %v", e.Provider, e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("OAuth error for provider %s [%s]: %s", e.Provider, e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *Error) Unwrap() error {
	return e.Err
}

// NewOAuthError creates a new OAuth error
func NewOAuthError(provider, code, message string, err error) *Error {
	return &Error{
		Provider: provider,
		Code:     code,
		Message:  message,
		Err:      err,
	}
}
