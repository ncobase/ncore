package templates

import "fmt"

func HandlerTemplate(name, extType, moduleName string) string {
	return fmt.Sprintf(`package handler

import "{{ .PackagePath }}/service"

// Handler represents the %s handler.
type Handler struct {
	s *service.Service
}

// New creates a new handler.
func New(s *service.Service) *Handler {
	return &Handler{
		s: s,
	}
}

// Add your handler methods here
`, name)
}
