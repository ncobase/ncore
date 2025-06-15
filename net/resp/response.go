package resp

import (
	"encoding/json"
	"net/http"

	"github.com/ncobase/ncore/ecode"
)

// Exception represents the response structure.
type Exception struct {
	Status  int    `json:"status,omitempty"`  // HTTP status
	Code    int    `json:"code,omitempty"`    // Business code
	Message string `json:"message,omitempty"` // Message
	Errors  any    `json:"errors,omitempty"`  // Validation errors
	Data    any    `json:"data,omitempty"`    // Response data
}

// newResponse creates a new response.
func newResponse(status, code int, message string, data ...any) *Exception {
	var responseData any
	if len(data) > 0 {
		responseData = data[0]
	}

	if status < 200 || status >= 400 || code != 0 {
		return &Exception{
			Status:  status,
			Code:    code,
			Message: message,
			Errors:  responseData,
		}
	}

	return &Exception{
		Status:  status,
		Code:    code,
		Message: message,
		Data:    responseData,
	}
}

// Success handles success responses.
func Success(w http.ResponseWriter, data ...any) {
	WithStatusCode(w, http.StatusOK, data...)
}

// WithStatusCode handles success responses with custom status code.
func WithStatusCode(w http.ResponseWriter, statusCode int, data ...any) {
	var message string
	var responseData any

	if len(data) > 0 {
		responseData = data[0]
		if strData, ok := responseData.(string); ok {
			message = strData
			responseData = nil
		}
	}

	r := newResponse(statusCode, 0, message, responseData)
	statusCode, result := buildSuccessResponse(r)
	writeResponse(w, "JSON", statusCode, result)
}

// buildSuccessResponse builds the success response.
func buildSuccessResponse(r *Exception) (int, any) {
	status := http.StatusOK

	if r != nil && r.Status != 0 {
		status = r.Status
	}

	if status < 200 || status >= 400 {
		return buildFailureResponse(r)
	}

	if r != nil && r.Data != nil {
		return status, r.Data
	}

	message := "ok"
	if r != nil && r.Message != "" {
		message = r.Message
	}

	return status, map[string]any{"message": message}
}

// Fail handles failure responses.
func Fail(w http.ResponseWriter, r *Exception, abort ...bool) {
	if r == nil {
		r = &Exception{
			Status:  http.StatusInternalServerError,
			Code:    ecode.ServerErr,
			Message: ecode.Text(ecode.ServerErr),
		}
	}
	statusCode, result := buildFailureResponse(r)
	writeResponse(w, "JSON", statusCode, result)

	if len(abort) > 0 && abort[0] {
		http.Error(w, "", statusCode)
	}
}

// buildFailureResponse builds the failure response.
func buildFailureResponse(r *Exception) (int, any) {
	status := http.StatusBadRequest
	code := ecode.RequestErr
	message := ecode.Text(code)

	if r.Status != 0 {
		status = r.Status
	}
	if r.Code != 0 {
		code = r.Code
	}
	if r.Message != "" {
		message = r.Message
	}

	return status, &Exception{
		Code:    code,
		Message: message,
		Errors:  r.Errors,
	}
}

// writeResponse writes the response based on the specified status code.
func writeResponse(w http.ResponseWriter, contextType string, code int, res any) {
	w.WriteHeader(code)
	switch contextType {
	case "JSON":
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(res)
		if err != nil {
			return
		}
	case "XML":
		w.Header().Set("Content-Type", "application/xml")
		// Implement XML encoding here
	case "Text":
		w.Header().Set("Content-Type", "text/plain")
		// Convert res to string if needed and write it to response writer
	default:
		// Default to JSON if no contextType matches
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(res)
		if err != nil {
			return
		}
	}
}
