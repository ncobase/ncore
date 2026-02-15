// Package cookie provides secure cookie management utilities for web applications,
// with specialized support for authentication tokens and session management.
//
// This package offers:
//   - Secure cookie creation and retrieval
//   - Built-in support for access/refresh tokens
//   - Automatic cookie attributes (HttpOnly, Secure, SameSite)
//   - Domain and path configuration
//   - Cookie deletion helpers
//
// # Predefined Cookie Names
//
// The package defines standard cookie names for common use cases:
//   - AccessTokenName: "access_token" - JWT access token
//   - RefreshTokenName: "refresh_token" - JWT refresh token
//   - RegisterTokenName: "register_token" - Registration verification
//   - DefaultName: "token" - General purpose token
//
// # Creating Cookies
//
//	// Set an access token cookie
//	cookie.Set(w, cookie.AccessTokenName, "jwt-token-here", 3600)
//
//	// Set a custom cookie
//	cookie.Set(w, "session_id", sessionID, 86400)
//
// # Retrieving Cookies
//
//	// Get access token from request
//	token, err := cookie.Get(r, cookie.AccessTokenName)
//	if err != nil {
//	    // Cookie not found or invalid
//	}
//
// # Deleting Cookies
//
//	// Remove access token
//	cookie.Delete(w, cookie.AccessTokenName)
//
//	// Remove custom cookie
//	cookie.Delete(w, "session_id")
//
// # Security Features
//
// All cookies created by this package automatically include:
//   - HttpOnly: true (prevents JavaScript access)
//   - Secure: true in production (HTTPS only)
//   - SameSite: Lax (CSRF protection)
//   - Path: / (accessible across entire site)
//
// Cookie values are automatically encoded and decoded for safe transmission.
package cookie
