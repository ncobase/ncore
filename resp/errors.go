package resp

import (
	"ncobase/common/ecode"
	"net/http"
)

// AlreadyExists indicates that the resource already exists.
func AlreadyExists(message string, data ...any) *Exception {
	return newResponse(http.StatusConflict, ecode.Conflict, message, data...)
}

// NotExists indicates that the resource does not exist.
func NotExists(message string, data ...any) *Exception {
	return newResponse(http.StatusNotFound, ecode.NothingFound, message, data...)
}

// DBQuery indicates a database query error.
func DBQuery(message string, data ...any) *Exception {
	return newResponse(http.StatusInternalServerError, ecode.ServerErr, message, data...)
}

// Transactions indicates a transaction processing failure.
func Transactions(message string, data ...any) *Exception {
	return newResponse(http.StatusInternalServerError, ecode.ServerErr, message, data...)
}

// UnAuthorized indicates that the request is unauthorized.
func UnAuthorized(message string, data ...any) *Exception {
	return newResponse(http.StatusUnauthorized, ecode.Unauthorized, message, data...)
}

// BadRequest indicates a bad request.
func BadRequest(message string, data ...any) *Exception {
	return newResponse(http.StatusBadRequest, ecode.RequestErr, message, data...)
}

// NotFound indicates that the requested resource is not found.
func NotFound(message string, data ...any) *Exception {
	return newResponse(http.StatusNotFound, ecode.NothingFound, message, data...)
}

// Forbidden indicates access is forbidden.
func Forbidden(message string, data ...any) *Exception {
	return newResponse(http.StatusForbidden, ecode.AccessDenied, message, data...)
}

// InternalServer indicates a server error.
func InternalServer(message string, data ...any) *Exception {
	return newResponse(http.StatusInternalServerError, ecode.ServerErr, message, data...)
}

// Conflict indicates a conflict error.
func Conflict(message string, data ...any) *Exception {
	return newResponse(http.StatusConflict, ecode.Conflict, message, data...)
}

// NotAllowed indicates a not allowed error.
func NotAllowed(message string, data ...any) *Exception {
	return newResponse(http.StatusMethodNotAllowed, ecode.MethodNotAllowed, message, data...)
}
