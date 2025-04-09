package generator

import (
	"fmt"
	"ncore/cmd/generator/templates"
	"ncore/pkg/utils"
	"os"
	"path/filepath"
	"strings"
)

// Options defines generation options
type Options struct {
	Name       string
	Type       string // core / business / plugin / custom
	CustomDir  string // Custom Directory, if Type is custom
	OutputPath string // Generated code output path
	ModuleName string // Module name
	UseMongo   bool
	UseEnt     bool
	UseGorm    bool
	WithTest   bool
	Group      string
}

// DefaultOptions returns default options
func DefaultOptions() *Options {
	return &Options{
		Type:       "custom",
		OutputPath: "",
		ModuleName: "",
		UseMongo:   false,
		UseEnt:     false,
		UseGorm:    false,
		WithTest:   false,
		Group:      "",
	}
}

var extDescriptions = map[string]string{
	"core":     "Core Domain",
	"business": "Business Domain",
	"plugin":   "Plugin Domain",
	"custom":   "Custom Directory",
}

// Generate generates code
func Generate(opts *Options) error {
	if !utils.ValidateName(opts.Name) {
		return fmt.Errorf("invalid name: %s", opts.Name)
	}

	// Determine output path
	if opts.OutputPath == "" {
		// Use current directory
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %v", err)
		}
		opts.OutputPath = cwd
	}

	// Determine module name if not provided
	if opts.ModuleName == "" {
		// Try to detect from go.mod file
		goModPath := filepath.Join(opts.OutputPath, "go.mod")
		if utils.FileExists(goModPath) {
			content, err := os.ReadFile(goModPath)
			if err == nil {
				lines := strings.Split(string(content), "\n")
				for _, line := range lines {
					if strings.HasPrefix(line, "module ") {
						opts.ModuleName = strings.TrimSpace(strings.TrimPrefix(line, "module "))
						break
					}
				}
			}
		}

		// If still empty, use a default
		if opts.ModuleName == "" {
			// Use current directory name as module name
			dirs := strings.Split(opts.OutputPath, string(os.PathSeparator))
			opts.ModuleName = dirs[len(dirs)-1]
		}
	}

	var basePath string
	var extType string
	var mainTemplate func(string) string

	// Determine base paths and templates based on type
	switch opts.Type {
	case "core":
		basePath = filepath.Join(opts.OutputPath, "core", opts.Name)
		extType = "core"
		mainTemplate = templates.CoreTemplate
	case "business":
		basePath = filepath.Join(opts.OutputPath, "business", opts.Name)
		extType = "business"
		mainTemplate = templates.BusinessTemplate
	case "plugin":
		basePath = filepath.Join(opts.OutputPath, "plugin", opts.Name)
		extType = "plugin"
		mainTemplate = templates.PluginTemplate
	case "direct":
		basePath = filepath.Join(opts.OutputPath, opts.Name)
		extType = "direct"
		// Use business template
		mainTemplate = templates.BusinessTemplate
	case "custom":
		basePath = filepath.Join(opts.OutputPath, opts.CustomDir, opts.Name)
		extType = "custom"
		// Use business template
		mainTemplate = templates.BusinessTemplate
	default:
		return fmt.Errorf("unknown type: %s", opts.Type)
	}

	// Check if component already exists
	if exists, err := utils.PathExists(basePath); err != nil {
		return fmt.Errorf("error checking existence: %v", err)
	} else if exists {
		return fmt.Errorf("'%s' already exists in %s", opts.Name, extDescriptions[extType])
	}

	// Prepare template data
	data := &templates.Data{
		Name:        opts.Name,
		Type:        opts.Type,
		UseMongo:    opts.UseMongo,
		UseEnt:      opts.UseEnt,
		UseGorm:     opts.UseGorm,
		WithTest:    opts.WithTest,
		Group:       opts.Group,
		ExtType:     extType,
		ModuleName:  opts.ModuleName,
		CustomDir:   opts.CustomDir,
		PackagePath: getPackagePath(opts),
	}

	return createStructure(basePath, data, mainTemplate)
}

// getPackagePath returns the package path based on options
func getPackagePath(opts *Options) string {
	switch opts.Type {
	case "custom":
		if opts.CustomDir == "" {
			return fmt.Sprintf("%s/%s", opts.ModuleName, opts.Name)
		}
		return fmt.Sprintf("%s/%s/%s", opts.ModuleName, opts.CustomDir, opts.Name)
	case "direct":
		return fmt.Sprintf("%s/%s", opts.ModuleName, opts.Name)
	default:
		return fmt.Sprintf("%s/%s/%s", opts.ModuleName, opts.Type, opts.Name)
	}
}

func createStructure(basePath string, data *templates.Data, mainTemplate func(string) string) error {
	// Create base directory
	if err := utils.EnsureDir(basePath); err != nil {
		return fmt.Errorf("failed to create base directory: %v", err)
	}

	// Create directory structure
	directories := []string{
		"data",
		"data/repository",
		"data/schema",
		"handler",
		"service",
		"structs",
	}

	if data.WithTest {
		directories = append(directories, "tests")
	}

	for _, dir := range directories {
		if err := utils.EnsureDir(filepath.Join(basePath, dir)); err != nil {
			return fmt.Errorf("failed to create directory %s: %v", dir, err)
		}
	}

	// Create files
	selectDataTemplate := func(data templates.Data) string {
		if data.UseEnt {
			return templates.DataTemplateWithEnt(data.Name, data.ExtType)
		}
		if data.UseGorm {
			return templates.DataTemplateWithGorm(data.Name, data.ExtType)
		}
		if data.UseMongo {
			return templates.DataTemplateWithMongo(data.Name, data.ExtType)
		}
		return templates.DataTemplate(data.Name, data.ExtType)
	}

	files := map[string]string{
		fmt.Sprintf("%s.go", data.Name): mainTemplate(data.Name),
		"data/data.go":                  selectDataTemplate(*data),
		"data/repository/provider.go":   templates.RepositoryTemplate(data.Name, data.ExtType, data.ModuleName),
		"data/schema/schema.go":         templates.SchemaTemplate(),
		"handler/provider.go":           templates.HandlerTemplate(data.Name, data.ExtType, data.ModuleName),
		"service/provider.go":           templates.ServiceTemplate(data.Name, data.ExtType, data.ModuleName),
		"structs/structs.go":            templates.StructsTemplate(),
	}

	// Add ent files if required
	if data.UseEnt {
		files["generate.go"] = templates.GeneraterTemplate(data.Name, data.ExtType, data.ModuleName)
	}

	// Add test files if required
	if data.WithTest {
		files["tests/ext_test.go"] = templates.ExtTestTemplate(data.Name, data.ExtType, data.ModuleName)
		files["tests/handler_test.go"] = templates.HandlerTestTemplate(data.Name, data.ExtType, data.ModuleName)
		files["tests/service_test.go"] = templates.ServiceTestTemplate(data.Name, data.ExtType, data.ModuleName)
	}

	// Write files
	for filePath, tmpl := range files {
		if err := utils.WriteTemplateFile(
			filepath.Join(basePath, filePath),
			tmpl,
			data,
		); err != nil {
			return fmt.Errorf("failed to create file %s: %v", filePath, err)
		}
	}

	fmt.Printf("Successfully generated '%s' in %s\n", data.Name, getDesc(data))
	return nil
}

// getDesc returns the description of the generated component
func getDesc(data *templates.Data) string {
	if data.Type == "custom" {
		return fmt.Sprintf("'%s' directory", data.CustomDir)
	}
	return extDescriptions[data.ExtType]
}
