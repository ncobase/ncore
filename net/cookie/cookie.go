package cookie

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/ncobase/ncore/types"
)

// formatDomain formats the domain.
func formatDomain(domain string) string {
	if domain != "localhost" && !strings.HasPrefix(domain, ".") {
		return "." + domain
	}
	return domain
}

// Set sets cookies.
func Set(w http.ResponseWriter, accessToken, refreshToken, domain string) {
	if accessToken != "" {
		SetAccessToken(w, accessToken, domain)
	}
	if refreshToken != "" {
		SetRefreshToken(w, refreshToken, domain)
	}
}

// SetAccessToken sets access token cookies.
func SetAccessToken(w http.ResponseWriter, accessToken, domain string) {
	formattedDomain := formatDomain(domain)
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		MaxAge:   60 * 60 * 24,
		Path:     "/",
		Domain:   formattedDomain,
		Secure:   true,
		HttpOnly: true,
	})
}

// SetRefreshToken sets refresh token cookies.
func SetRefreshToken(w http.ResponseWriter, refreshToken, domain string) {
	formattedDomain := formatDomain(domain)
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		MaxAge:   60 * 60 * 24 * 30,
		Path:     "/",
		Domain:   formattedDomain,
		Secure:   true,
		HttpOnly: true,
	})
}

// SetRegister sets registration cookies.
func SetRegister(w http.ResponseWriter, registerToken, domain string) {
	formattedDomain := formatDomain(domain)
	http.SetCookie(w, &http.Cookie{
		Name:     "register_token",
		Value:    registerToken,
		MaxAge:   60 * 60,
		Path:     "/",
		Domain:   formattedDomain,
		Secure:   true,
		HttpOnly: true,
	})
}

// Clear clears cookies.
func Clear(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:   "access_token",
		MaxAge: -1,
		Path:   "/",
	})
	http.SetCookie(w, &http.Cookie{
		Name:   "refresh_token",
		MaxAge: -1,
		Path:   "/",
	})
}

// ClearRegister clears registration cookies.
func ClearRegister(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:   "register_token",
		MaxAge: -1,
		Path:   "/",
	})
}

// ClearAll clears all cookies.
func ClearAll(w http.ResponseWriter) {
	Clear(w)
	ClearRegister(w)
}

// Get gets cookies.
func Get(r *http.Request, key string) (string, error) {
	cookie, err := r.Cookie(key)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// GetRegister gets registration cookies.
func GetRegister(r *http.Request, key string) (string, error) {
	cookie, err := r.Cookie(key)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// GetTokenFromResult retrieves a token from the result map.
func GetTokenFromResult(result *map[string]any, key string) (string, error) {
	value, ok := types.ToValue(result)[key]
	if !ok {
		return "", fmt.Errorf("key %s not found in result", key)
	}
	token, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("key %s is not a string", key)
	}
	return token, nil
}

// SetTokensFromResult sets access and refresh tokens from the result map.
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

// SetRegisterTokenFromResult sets registration token from the result map.
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
