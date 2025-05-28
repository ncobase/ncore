package cookie

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/ncobase/ncore/utils/convert"
)

// Cookie names
const (
	AccessTokenName   = "access_token"
	RefreshTokenName  = "refresh_token"
	RegisterTokenName = "register_token"
	SessionIDName     = "session_id"
	CSRFTokenName     = "csrf_token"
)

// Cookie max ages (in seconds)
const (
	AccessTokenMaxAge   = 60 * 60 * 24      // 24 hours
	RefreshTokenMaxAge  = 60 * 60 * 24 * 30 // 30 days
	RegisterTokenMaxAge = 60 * 60           // 1 hour
	SessionMaxAge       = 60 * 60 * 24      // 24 hours
	CSRFTokenMaxAge     = 60 * 60 * 24      // 24 hours
)

// formatDomain formats the domain
func formatDomain(domain string) string {
	if domain != "localhost" && !strings.HasPrefix(domain, ".") {
		return "." + domain
	}
	return domain
}

// Set sets cookies
func Set(w http.ResponseWriter, accessToken, refreshToken, domain string) {
	if accessToken != "" {
		SetAccessToken(w, accessToken, domain)
	}
	if refreshToken != "" {
		SetRefreshToken(w, refreshToken, domain)
	}
}

// SetAccessToken sets access token cookie
func SetAccessToken(w http.ResponseWriter, accessToken, domain string) {
	cookie := &http.Cookie{
		Name:     AccessTokenName,
		Value:    accessToken,
		MaxAge:   AccessTokenMaxAge,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}

	if domain != "" {
		cookie.Domain = formatDomain(domain)
	}

	http.SetCookie(w, cookie)
}

// SetRefreshToken sets refresh token cookie
func SetRefreshToken(w http.ResponseWriter, refreshToken string, domain ...string) {
	cookie := &http.Cookie{
		Name:     RefreshTokenName,
		Value:    refreshToken,
		MaxAge:   RefreshTokenMaxAge,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}

	if len(domain) > 0 {
		cookie.Domain = formatDomain(domain[0])
	}

	http.SetCookie(w, cookie)
}

// SetRegister sets registration token cookie
func SetRegister(w http.ResponseWriter, registerToken, domain string) {
	cookie := &http.Cookie{
		Name:     RegisterTokenName,
		Value:    registerToken,
		MaxAge:   RegisterTokenMaxAge,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}

	if domain != "" {
		cookie.Domain = formatDomain(domain)
	}

	http.SetCookie(w, cookie)
}

// SetSessionID sets session ID cookie for web authentication
func SetSessionID(w http.ResponseWriter, sessionID string, domain ...string) error {
	if sessionID == "" {
		return fmt.Errorf("session ID cannot be empty")
	}

	cookie := &http.Cookie{
		Name:     SessionIDName,
		Value:    sessionID,
		MaxAge:   SessionMaxAge,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}

	if len(domain) > 0 && domain[0] != "" {
		cookie.Domain = formatDomain(domain[0])
	}

	http.SetCookie(w, cookie)
	return nil
}

// SetCSRFToken sets CSRF token cookie
func SetCSRFToken(w http.ResponseWriter, csrfToken string, domain ...string) error {
	if csrfToken == "" {
		return fmt.Errorf("CSRF token cannot be empty")
	}

	cookie := &http.Cookie{
		Name:     CSRFTokenName,
		Value:    csrfToken,
		MaxAge:   CSRFTokenMaxAge,
		Path:     "/",
		Secure:   true,
		HttpOnly: false,
		SameSite: http.SameSiteStrictMode,
	}

	if len(domain) > 0 && domain[0] != "" {
		cookie.Domain = formatDomain(domain[0])
	}

	http.SetCookie(w, cookie)
	return nil
}

// GetSessionID gets session ID from cookie
func GetSessionID(r *http.Request) (string, error) {
	cookie, err := r.Cookie(SessionIDName)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// GetCSRFToken gets CSRF token from cookie
func GetCSRFToken(r *http.Request) (string, error) {
	cookie, err := r.Cookie(CSRFTokenName)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// Clear clears token cookies
func Clear(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:   AccessTokenName,
		MaxAge: -1,
		Path:   "/",
	})
	http.SetCookie(w, &http.Cookie{
		Name:   RefreshTokenName,
		MaxAge: -1,
		Path:   "/",
	})
}

// ClearRegister clears registration cookie
func ClearRegister(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:   RegisterTokenName,
		MaxAge: -1,
		Path:   "/",
	})
}

// ClearSessionID clears session ID cookie
func ClearSessionID(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionIDName,
		Value:    "",
		MaxAge:   -1,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

// ClearCSRFToken clears CSRF token cookie
func ClearCSRFToken(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     CSRFTokenName,
		Value:    "",
		MaxAge:   -1,
		Path:     "/",
		HttpOnly: false,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
}

// ClearAll clears all authentication cookies
func ClearAll(w http.ResponseWriter) {
	Clear(w)
	ClearRegister(w)
	ClearSessionID(w)
	ClearCSRFToken(w)
}

// Get gets cookie value by name
func Get(r *http.Request, key string) (string, error) {
	cookie, err := r.Cookie(key)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// GetRegister gets registration cookie
func GetRegister(r *http.Request, key string) (string, error) {
	cookie, err := r.Cookie(key)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// GetTokenFromResult retrieves a token from the result map
func GetTokenFromResult(result *map[string]any, key string) (string, error) {
	value, ok := convert.ToValue(result)[key]
	if !ok {
		return "", fmt.Errorf("key %s not found in result", key)
	}
	token, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("key %s is not a string", key)
	}
	return token, nil
}

// SetTokensFromResult sets access and refresh tokens from result map
func SetTokensFromResult(w http.ResponseWriter, r *http.Request, result *map[string]any, domain ...string) error {
	var dm string
	if len(domain) > 0 {
		dm = domain[0]
	} else {
		dm = r.Host
	}
	formattedDomain := formatDomain(dm)

	accessToken, _ := GetTokenFromResult(result, "access_token")
	refreshToken, _ := GetTokenFromResult(result, "refresh_token")

	if accessToken == "" && refreshToken == "" {
		return fmt.Errorf("both access_token and refresh_token are missing")
	}

	Set(w, accessToken, refreshToken, formattedDomain)
	return nil
}

// SetRegisterTokenFromResult sets registration token from result map
func SetRegisterTokenFromResult(w http.ResponseWriter, r *http.Request, result *map[string]any, domain ...string) error {
	var dm string
	if len(domain) > 0 {
		dm = domain[0]
	} else {
		dm = r.Host
	}
	formattedDomain := formatDomain(dm)
	token, _ := GetTokenFromResult(result, "register_token")
	SetRegister(w, token, formattedDomain)
	return nil
}

// SetSessionFromResult sets session ID cookie from result map
func SetSessionFromResult(w http.ResponseWriter, r *http.Request, result *map[string]any, domain ...string) error {
	var dm string
	if len(domain) > 0 {
		dm = domain[0]
	} else {
		dm = r.Host
	}

	sessionID, err := GetTokenFromResult(result, "session_id")
	if err != nil {
		return err // Session ID not found in result
	}

	return SetSessionID(w, sessionID, dm)
}

// SetSecureCookie sets a secure cookie with common security settings
func SetSecureCookie(w http.ResponseWriter, name, value string, maxAge int, domain ...string) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   maxAge,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}

	if len(domain) > 0 && domain[0] != "" {
		cookie.Domain = formatDomain(domain[0])
	}

	http.SetCookie(w, cookie)
}

// ClearCookie clears a specific cookie
func ClearCookie(w http.ResponseWriter, name string, domain ...string) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    "",
		MaxAge:   -1,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	if len(domain) > 0 && domain[0] != "" {
		cookie.Domain = formatDomain(domain[0])
	}

	http.SetCookie(w, cookie)
}
