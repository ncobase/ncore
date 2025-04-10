package templates

import "fmt"

func ServiceTemplate(name, extType, moduleName string) string {
	return fmt.Sprintf(`package service

import (
	"github.com/ncobase/ncore/pkg/config"
	"{{ .PackagePath }}/data"
)

// Service represents the %s service.
type Service struct {
	conf *config.Config
	d    *data.Data
}

// New creates a new service.
func New(conf *config.Config, d *data.Data) *Service {
	return &Service{
		conf: conf,
		d:    d,
	}
}

// Add your service methods here
`, name)
}
