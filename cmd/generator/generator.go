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
	WithCmd    bool
	WithTest   bool
	Standalone bool
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
		WithCmd:    false,
		WithTest:   false,
		Standalone: false,
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

	// Handle standalone mode differently
	if opts.Standalone {
		// In standalone mode, generate only the cmd directory
		if opts.Type == "direct" {
			basePath = filepath.Join(opts.OutputPath, opts.Name)
		} else if opts.Type == "custom" {
			basePath = filepath.Join(opts.OutputPath, opts.CustomDir, opts.Name)
		} else {
			basePath = filepath.Join(opts.OutputPath, opts.Type, opts.Name)
		}

		extType = opts.Type

		// Create base directory
		if err := utils.EnsureDir(basePath); err != nil {
			return fmt.Errorf("failed to create base directory: %v", err)
		}

		// Prepare template data
		data := &templates.Data{
			Name:        opts.Name,
			Type:        opts.Type,
			UseMongo:    opts.UseMongo,
			UseEnt:      opts.UseEnt,
			UseGorm:     opts.UseGorm,
			WithTest:    opts.WithTest,
			WithCmd:     true, // Standalone always includes cmd
			Standalone:  opts.Standalone,
			Group:       opts.Group,
			ExtType:     extType,
			ModuleName:  opts.ModuleName,
			CustomDir:   opts.CustomDir,
			PackagePath: getPackagePath(opts),
		}

		// Create standalone structure
		if err := createStandaloneStructure(basePath, data); err != nil {
			return err
		}

		fmt.Printf("Successfully generated standalone application '%s' in %s\n", data.Name, getDesc(data))
		return nil
	}

	// Regular extension generation (not standalone)
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
		WithCmd:     opts.WithCmd,
		Standalone:  opts.Standalone,
		Group:       opts.Group,
		ExtType:     extType,
		ModuleName:  opts.ModuleName,
		CustomDir:   opts.CustomDir,
		PackagePath: getPackagePath(opts),
	}

	// Create the main extension structure
	err := createStructure(basePath, data, mainTemplate)
	if err != nil {
		return err
	}

	// Generate cmd directory if WithCmd is true
	if opts.WithCmd {
		// Create cmd directory inside the extension directory
		cmdPath := filepath.Join(basePath, "cmd")
		if err := utils.EnsureDir(cmdPath); err != nil {
			return fmt.Errorf("failed to create cmd directory: %v", err)
		}

		// Create files in cmd directory
		files := map[string]string{
			"main.go": templates.CmdMainTemplate(data.Name, data.ExtType, data.ModuleName),
		}

		// Create provider directory
		providerPath := filepath.Join(cmdPath, "provider")
		if err := utils.EnsureDir(providerPath); err != nil {
			return fmt.Errorf("failed to create provider directory: %v", err)
		}

		// Add provider files
		files[filepath.Join("provider", "server.go")] = templates.CmdServerTemplate(data.Name, data.ExtType, data.ModuleName)
		files[filepath.Join("provider", "extension.go")] = templates.CmdExtensionTemplate(data.Name, data.ExtType, data.ModuleName)
		files[filepath.Join("provider", "gin.go")] = templates.CmdGinTemplate(data.Name, data.ExtType, data.ModuleName)
		files[filepath.Join("provider", "rest.go")] = templates.CmdRestTemplate(data.Name, data.ExtType, data.ModuleName)

		// Write files
		for filePath, tmpl := range files {
			if err := utils.WriteTemplateFile(
				filepath.Join(cmdPath, filePath),
				tmpl,
				data,
			); err != nil {
				return fmt.Errorf("failed to create file %s: %v", filePath, err)
			}
		}
	}

	fmt.Printf("Successfully generated '%s' in %s\n", data.Name, getDesc(data))
	return nil
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

	return nil
}

// createStandaloneStructure creates the structure for a standalone application
func createStandaloneStructure(basePath string, data *templates.Data) error {
	// Create essential directories
	directories := []string{
		"cmd",
		"cmd/provider",
		"config",
		"handler",
		"model",
		"service",
	}

	if data.WithTest {
		directories = append(directories, "tests")
	}

	for _, dir := range directories {
		if err := utils.EnsureDir(filepath.Join(basePath, dir)); err != nil {
			return fmt.Errorf("failed to create directory %s: %v", dir, err)
		}
	}

	// Create cmd files
	cmdFiles := map[string]string{
		"cmd/main.go":            templates.StandaloneMainTemplate(data.Name, data.ModuleName),
		"cmd/provider/server.go": templates.StandaloneServerTemplate(data.Name, data.ModuleName),
		"cmd/provider/gin.go":    templates.StandaloneGinTemplate(data.Name, data.ModuleName),
		"cmd/provider/rest.go":   templates.StandaloneRestTemplate(data.Name, data.ModuleName),
	}

	// Create files
	projectFiles := map[string]string{
		"config/config.go":   templates.StandaloneConfigTemplate(data.Name, data.ModuleName),
		"handler/handler.go": templates.StandaloneHandlerTemplate(data.Name, data.ModuleName),
		"model/model.go":     templates.StandaloneModelTemplate(data.Name, data.ModuleName),
		"service/service.go": templates.StandaloneServiceTemplate(data.Name, data.ModuleName),
	}

	// Merge the maps
	files := make(map[string]string)
	for k, v := range cmdFiles {
		files[k] = v
	}
	for k, v := range projectFiles {
		files[k] = v
	}

	// Add test files if required
	if data.WithTest {
		files["tests/handler_test.go"] = templates.StandaloneHandlerTestTemplate(data.Name, data.ModuleName)
		files["tests/service_test.go"] = templates.StandaloneServiceTestTemplate(data.Name, data.ModuleName)
	}

	// Add database related files if required
	if data.UseMongo || data.UseEnt || data.UseGorm {
		files["repository/repository.go"] = templates.StandaloneRepositoryTemplate(data.Name, data.ModuleName, data.UseMongo, data.UseEnt, data.UseGorm)

		if err := utils.EnsureDir(filepath.Join(basePath, "repository")); err != nil {
			return fmt.Errorf("failed to create directory %s: %v", "repository", err)
		}
	}

	// Write all files
	for filePath, tmpl := range files {
		if err := utils.WriteTemplateFile(
			filepath.Join(basePath, filePath),
			tmpl,
			data,
		); err != nil {
			return fmt.Errorf("failed to create file %s: %v", filePath, err)
		}
	}

	return nil
}

// getDesc returns the description of the generated component
func getDesc(data *templates.Data) string {
	if data.Type == "custom" {
		return fmt.Sprintf("'%s' directory", data.CustomDir)
	}
	return extDescriptions[data.ExtType]
}
