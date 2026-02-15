// Package ecode defines standardized error codes for API responses and provides
// utilities for error code management and localization.
//
// This package provides:
//   - Predefined error codes for common scenarios
//   - Human-readable error messages
//   - Multi-language support (i18n)
//   - Error code to HTTP status mapping
//   - Custom error code registration
//
// # Error Code Convention
//
// Error codes follow a standardized numbering scheme:
//   - 0: Success (OK)
//   - -1 to -99: Application-level errors
//   - -100 to -199: Authentication/authorization errors
//   - -200 to -299: Request validation errors
//   - -300 to -399: Resource errors
//   - -400 to -499: Business logic errors
//   - -500+: Server errors
//
// # Common Error Codes
//
// Authentication errors:
//
//	ecode.NoLogin           // -101: Not logged in
//	ecode.UserDisabled      // -102: Account suspended
//	ecode.CaptchaErr        // -105: Captcha verification failed
//	ecode.UserInactive      // -106: Account not activated
//
// Request errors:
//
//	ecode.RequestErr        // -400: Invalid request
//	ecode.ParamErr          // -401: Invalid parameters
//	ecode.SignCheckErr      // -3: Signature verification failed
//
// Resource errors:
//
//	ecode.NotFound          // -404: Resource not found
//	ecode.Conflict          // -409: Resource conflict
//	ecode.AccessDenied      // -403: Access denied
//
// Server errors:
//
//	ecode.ServerErr         // -500: Internal server error
//	ecode.ServiceUnavailable // -503: Service unavailable
//	ecode.Deadline          // -504: Deadline exceeded
//
// # Getting Error Messages
//
// Retrieve human-readable error messages:
//
//	message := ecode.Text(ecode.NoLogin)
//	// Returns: "Account not logged in"
//
//	message := ecode.Text(ecode.ParamErr)
//	// Returns: "Invalid parameters"
//
// # Custom Error Codes
//
// Register custom error codes for your application:
//
//	const (
//	    InsufficientBalance = -1001
//	    OrderExpired        = -1002
//	)
//
//	ecode.Register(InsufficientBalance, "Insufficient account balance")
//	ecode.Register(OrderExpired, "Order has expired")
//
// # HTTP Status Mapping
//
// Error codes can be mapped to appropriate HTTP status codes:
//
//	httpStatus := ecode.ToHTTPStatus(ecode.NotFound)
//	// Returns: 404
//
//	httpStatus := ecode.ToHTTPStatus(ecode.NoLogin)
//	// Returns: 401
//
//	httpStatus := ecode.ToHTTPStatus(ecode.ServerErr)
//	// Returns: 500
//
// # Usage with Response Package
//
// Error codes integrate seamlessly with the resp package:
//
//	import (
//	    "github.com/ncobase/ncore/ecode"
//	    "github.com/ncobase/ncore/net/resp"
//	)
//
//	resp.Fail(w, &resp.Exception{
//	    Status:  http.StatusUnauthorized,
//	    Code:    ecode.NoLogin,
//	    Message: ecode.Text(ecode.NoLogin),
//	})
//
// # Localization
//
// Support for multiple languages:
//
//	// Set language for error messages
//	ecode.SetLanguage("zh-CN")
//
//	message := ecode.Text(ecode.NoLogin)
//	// Returns: "账号未登录" (Chinese)
//
//	ecode.SetLanguage("en-US")
//	message = ecode.Text(ecode.NoLogin)
//	// Returns: "Account not logged in" (English)
//
// # Best Practices
//
//   - Use predefined codes when possible
//   - Keep custom codes in application-specific ranges (-1000+)
//   - Provide clear, actionable error messages
//   - Map codes to appropriate HTTP statuses
//   - Document custom error codes
//   - Use consistent error code patterns
//   - Return codes in API responses
//   - Log error codes for debugging
package ecode
