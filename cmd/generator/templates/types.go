package templates

// Data represents template data structure
type Data struct {
	Name        string // Extension name
	Type        string // Extension type (core/business/plugin/custom)
	CustomDir   string // Custom directory name, if type is custom
	ModuleName  string // Go module name
	UseMongo    bool   // Whether to use MongoDB
	UseEnt      bool   // Whether to use Ent ORM
	UseGorm     bool   // Whether to use GORM
	WithCmd     bool   // Whether to generate cmd directory with main.go
	WithTest    bool   // Whether to generate test files
	Standalone  bool   // Whether to generate as standalone app without extension structure
	Group       string // Optional group name
	ExtType     string // Extension type in belongs domain path (core/business/plugin/custom)
	PackagePath string // Full package path
}
