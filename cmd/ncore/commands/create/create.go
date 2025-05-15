package create

import (
	"strings"

	"github.com/ncobase/ncore/cmd/ncore/generator"

	"github.com/spf13/cobra"
)

// knownTypes is a map of known extension types
var knownTypes = map[string]string{
	"core":     "core",
	"business": "business",
	"plugin":   "plugin",
}

// NewCommand creates a new create command
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create [extension type or custom directory] name",
		Aliases: []string{"gen", "generate"},
		Short:   "Generate new extension components",
		Long:    `Generate new extensions (core, business, plugin, or custom directory).`,
		Args:    cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := generator.DefaultOptions()

			var dir, name string
			if len(args) == 1 {
				// If only one argument, check if it's a known type
				firstArg := strings.ToLower(args[0])
				if _, ok := knownTypes[firstArg]; ok {
					// Is a known type, show help
					cmd.Help()
					return nil
				}

				// Not a known type, assume it's the name and create directly in current directory
				name = args[0]

				// Set options
				opts.Name = name
				opts.Type = "direct" // New type for direct creation

				// Get flags
				opts.ModuleName, _ = cmd.Flags().GetString("module")
				opts.OutputPath, _ = cmd.Flags().GetString("path")
				opts.UseMongo, _ = cmd.Flags().GetBool("use-mongo")
				opts.UseEnt, _ = cmd.Flags().GetBool("use-ent")
				opts.UseGorm, _ = cmd.Flags().GetBool("use-gorm")
				opts.WithTest, _ = cmd.Flags().GetBool("with-test")
				opts.Group, _ = cmd.Flags().GetString("group")
				opts.WithCmd, _ = cmd.Flags().GetBool("with-cmd")
				opts.Standalone, _ = cmd.Flags().GetBool("standalone")

				return generator.Generate(opts)
			}

			// If two arguments, use the first as the directory and the second as the name
			dir = args[0]
			name = args[1]

			// Check if the directory is a known type
			if extType, ok := knownTypes[strings.ToLower(dir)]; ok {
				// Is a known type
				switch extType {
				case "core":
					return newCoreCommand().RunE(cmd, []string{name})
				case "business":
					return newBusinessCommand().RunE(cmd, []string{name})
				case "plugin":
					return newPluginCommand().RunE(cmd, []string{name})
				}
			}

			// Not a known type, assume it's a custom directory
			opts.Name = name
			opts.Type = "custom"
			opts.CustomDir = dir

			// Get flags
			opts.ModuleName, _ = cmd.Flags().GetString("module")
			opts.OutputPath, _ = cmd.Flags().GetString("path")
			opts.UseMongo, _ = cmd.Flags().GetBool("use-mongo")
			opts.UseEnt, _ = cmd.Flags().GetBool("use-ent")
			opts.UseGorm, _ = cmd.Flags().GetBool("use-gorm")
			opts.WithCmd, _ = cmd.Flags().GetBool("with-cmd")
			opts.WithTest, _ = cmd.Flags().GetBool("with-test")
			opts.Standalone, _ = cmd.Flags().GetBool("standalone")
			opts.Group, _ = cmd.Flags().GetString("group")

			return generator.Generate(opts)
		},
	}

	// add subcommands
	cmd.AddCommand(
		newCoreCommand(),
		newBusinessCommand(),
		newPluginCommand(),
	)

	// add flags
	cmd.Flags().StringP("path", "p", "", "output path (defaults to current directory)")
	cmd.Flags().StringP("module", "m", "", "Go module name (defaults to current module)")
	cmd.Flags().Bool("use-mongo", false, "use MongoDB")
	cmd.Flags().Bool("use-ent", false, "use Ent as ORM")
	cmd.Flags().Bool("use-gorm", false, "use Gorm as ORM")
	cmd.Flags().Bool("with-test", false, "generate test files")
	cmd.Flags().Bool("with-cmd", false, "generate cmd directory with main.go")
	cmd.Flags().Bool("standalone", false, "generate as standalone app without extension structure")
	cmd.Flags().String("group", "", "belongs domain group (optional)")

	return cmd
}
