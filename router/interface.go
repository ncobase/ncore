package router

import "net/http"

// Interface abstracts the routing mechanism
type Interface interface {
	Group(path string, middleware ...func(http.Handler) http.Handler) Interface
	GET(path string, handler http.HandlerFunc)
	POST(path string, handler http.HandlerFunc)
	PUT(path string, handler http.HandlerFunc)
	DELETE(path string, handler http.HandlerFunc)
	PATCH(path string, handler http.HandlerFunc)
	OPTIONS(path string, handler http.HandlerFunc)
	HEAD(path string, handler http.HandlerFunc)
	ServeFiles(path string, root http.FileSystem)
}
