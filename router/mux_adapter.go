package router

import (
	"net/http"

	"github.com/gorilla/mux"
)

type MuxRouter struct {
	router *mux.Router
	group  *mux.Router
}

func NewMuxAdapter(router *mux.Router) Interface {
	return &MuxRouter{router: router, group: router}
}

func (r *MuxRouter) Group(path string, middleware ...func(http.Handler) http.Handler) Interface {
	group := r.router.PathPrefix(path).Subrouter()
	for _, m := range middleware {
		group.Use(m)
	}
	return &MuxRouter{group: group}
}

func (r *MuxRouter) GET(path string, handler http.HandlerFunc) {
	r.group.HandleFunc(path, handler).Methods("GET")
}

func (r *MuxRouter) POST(path string, handler http.HandlerFunc) {
	r.group.HandleFunc(path, handler).Methods("POST")
}

func (r *MuxRouter) PUT(path string, handler http.HandlerFunc) {
	r.group.HandleFunc(path, handler).Methods("PUT")
}

func (r *MuxRouter) DELETE(path string, handler http.HandlerFunc) {
	r.group.HandleFunc(path, handler).Methods("DELETE")
}

func (r *MuxRouter) PATCH(path string, handler http.HandlerFunc) {
	r.group.HandleFunc(path, handler).Methods("PATCH")
}

func (r *MuxRouter) OPTIONS(path string, handler http.HandlerFunc) {
	r.group.HandleFunc(path, handler).Methods("OPTIONS")
}

func (r *MuxRouter) HEAD(path string, handler http.HandlerFunc) {
	r.group.HandleFunc(path, handler).Methods("HEAD")
}

func (r *MuxRouter) ServeFiles(path string, root http.FileSystem) {
	r.group.PathPrefix(path).Handler(http.StripPrefix(path, http.FileServer(root)))
}
