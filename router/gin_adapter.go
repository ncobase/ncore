package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type GinRouter struct {
	engine *gin.Engine
	group  *gin.RouterGroup
}

func NewGinAdapter(engine *gin.Engine) Interface {
	return &GinRouter{engine: engine, group: &engine.RouterGroup}
}

func (r *GinRouter) Group(path string, middleware ...func(http.Handler) http.Handler) Interface {
	handlers := make([]gin.HandlerFunc, len(middleware))
	for i, m := range middleware {
		handlers[i] = func(c *gin.Context) {
			m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				c.Request = r
				c.Writer = ginResponseWriter{c.Writer}
			})).ServeHTTP(c.Writer, c.Request)
		}
	}
	return &GinRouter{group: r.engine.Group(path, handlers...)}
}

func (r *GinRouter) GET(path string, handler http.HandlerFunc) {
	r.group.GET(path, gin.WrapF(handler))
}

func (r *GinRouter) POST(path string, handler http.HandlerFunc) {
	r.group.POST(path, gin.WrapF(handler))
}

func (r *GinRouter) PUT(path string, handler http.HandlerFunc) {
	r.group.PUT(path, gin.WrapF(handler))
}

func (r *GinRouter) DELETE(path string, handler http.HandlerFunc) {
	r.group.DELETE(path, gin.WrapF(handler))
}

func (r *GinRouter) PATCH(path string, handler http.HandlerFunc) {
	r.group.PATCH(path, gin.WrapF(handler))
}

func (r *GinRouter) OPTIONS(path string, handler http.HandlerFunc) {
	r.group.OPTIONS(path, gin.WrapF(handler))
}

func (r *GinRouter) HEAD(path string, handler http.HandlerFunc) {
	r.group.HEAD(path, gin.WrapF(handler))
}

func (r *GinRouter) ServeFiles(path string, root http.FileSystem) {
	r.group.StaticFS(path, root)
}

type ginResponseWriter struct {
	gin.ResponseWriter
}

func (w ginResponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
}
