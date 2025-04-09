package templates

import "fmt"

// ExtTestTemplate generates extension test template
func ExtTestTemplate(name, extType, moduleName string) string {
	return fmt.Sprintf(`package tests

import (
	"testing"
	"ncore/pkg/config"
	"ncore/extension"
	"{{ .ModuleName }}/{{ if eq .Type "custom" }}{{ .CustomDir }}{{ else }}{{ .ExtType }}{{ end }}/%s"
)

func TestModuleLifecycle(t *testing.T) {
	m := %s.New()

	t.Run("initialization", func(t *testing.T) {
		// Test Pre-Init
		if err := m.PreInit(); err != nil {
			t.Errorf("PreInit failed: %%v", err)
		}

		// Test Init
		conf := &config.Config{}
		em := &extension.Manager{}
		if err := m.Init(conf, em); err != nil {
			t.Errorf("Init failed: %%v", err)
		}

		// Test Post-Init
		if err := m.PostInit(); err != nil {
			t.Errorf("PostInit failed: %%v", err)
		}
	})

	t.Run("metadata", func(t *testing.T) {
		meta := m.GetMetadata()
		if meta.Name != "%s" {
			t.Errorf("want name %%s, got %%s", "%s", meta.Name)
		}
	})

	t.Run("cleanup", func(t *testing.T) {
		if err := m.Cleanup(); err != nil {
			t.Errorf("Cleanup failed: %%v", err)
		}
	})
}`, name, name, name, name)
}

// HandlerTestTemplate generates handler test template
func HandlerTestTemplate(name, extType, moduleName string) string {
	return fmt.Sprintf(`package tests

import (
	"testing"
	"net/http"
	"net/http/httptest"
	"github.com/gin-gonic/gin"
	"{{ .ModuleName }}/{{ if eq .Type "custom" }}{{ .CustomDir }}{{ else }}{{ .ExtType }}{{ end }}/%s/handler"
	"{{ .ModuleName }}/{{ if eq .Type "custom" }}{{ .CustomDir }}{{ else }}{{ .ExtType }}{{ end }}/%s/service"
)

func TestHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	s := service.New(nil, nil)
	h := handler.New(s)

	t.Run("health check", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/health", nil)

		r.GET("/health", func(c *gin.Context) {
			// Replace with your actual health check handler
			c.Status(http.StatusOK)
		})
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("want status 200, got %%d", w.Code)
		}
	})
}`, name, name)
}

// ServiceTestTemplate generates service test template
func ServiceTestTemplate(name, extType, moduleName string) string {
	return fmt.Sprintf(`package tests

import (
	"testing"
	"context"
	"{{ .ModuleName }}/{{ if eq .Type "custom" }}{{ .CustomDir }}{{ else }}{{ .ExtType }}{{ end }}/%s/service"
)

func TestService(t *testing.T) {
	ctx := context.Background()
	s := service.New(nil, nil)

	t.Run("service initialization", func(t *testing.T) {
		// Add your service tests here
		if s == nil {
			t.Error("service should not be nil")
		}
	})
}`, name)
}
