package templates

import "fmt"

func RepositoryTemplate(name, extType, moduleName string) string {
	return fmt.Sprintf(`package repository

import "{{ .ModuleName }}/{{ if eq .Type "custom" }}{{ .CustomDir }}{{ else }}{{ .ExtType }}{{ end }}/%s/data"

// Repository represents the %s repository.
type Repository struct {
	d *data.Data
}

// New creates a new repository.
func New(d *data.Data) *Repository {
	return &Repository{
		d: d,
	}
}

// Add your repository methods here
`, name, name)
}
