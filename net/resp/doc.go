// Package resp provides standardized HTTP response helpers for building
// consistent JSON, XML, and text responses in web applications.
//
// This package simplifies response handling by providing:
//   - Success and failure response builders
//   - Common HTTP error responses (404, 400, 500, etc.)
//   - Multiple content type support (JSON, XML, Text)
//   - Business error code integration
//   - Consistent response structure
//
// # Response Structure
//
// All responses follow a standard structure:
//
//	{
//	  "status": 200,           // HTTP status code
//	  "code": 0,               // Business error code (0 = success)
//	  "message": "ok",         // Human-readable message
//	  "data": {...},           // Response payload (on success)
//	  "errors": {...}          // Error details (on failure)
//	}
//
// # Success Responses
//
//	// Simple success with data
//	resp.Success(w, userData)
//
//	// Success with custom status code
//	resp.WithStatusCode(w, http.StatusCreated, newResource)
//
//	// Success with message only
//	resp.Success(w, "Operation completed")
//
// # Error Responses
//
//	// Pre-defined error responses
//	resp.NotFound(w, "User not found")
//	resp.BadRequest(w, "Invalid input", validationErrors)
//	resp.Unauthorized(w, "Authentication required")
//	resp.ServerError(w, "Internal error occurred")
//
//	// Custom error response
//	resp.Fail(w, &resp.Exception{
//	    Status:  http.StatusConflict,
//	    Code:    1001,
//	    Message: "Resource already exists",
//	    Errors:  conflictDetails,
//	})
//
// # Content Types
//
// The package supports JSON (default), XML, and plain text responses.
// Content type is automatically set based on the response format.
//
// # Error Codes
//
// Business error codes are defined in the ecode package and provide
// standardized error classification across the application.
package resp
