package create

import (
	"github.com/ncobase/ncore/cmd/ncore/generator"

	"github.com/spf13/cobra"
)

func newPluginCommand() *cobra.Command {
	opts := &generator.Options{}

	cmd := &cobra.Command{
		Use:     "plugin [name]",
		Aliases: []string{"p"},
		Short:   "Create a new extension in plugin",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Name = args[0]
			opts.Type = "plugin"

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

	// Add flags
	cmd.Flags().StringVarP(&opts.OutputPath, "path", "p", "", "output path (defaults to current directory)")
	cmd.Flags().StringVarP(&opts.ModuleName, "module", "m", "", "Go module name (defaults to current module)")
	cmd.Flags().BoolVar(&opts.UseMongo, "use-mongo", false, "use MongoDB")
	cmd.Flags().BoolVar(&opts.UseEnt, "use-ent", false, "use Ent as ORM")
	cmd.Flags().BoolVar(&opts.UseGorm, "use-gorm", false, "use Gorm as ORM")
	cmd.Flags().BoolVar(&opts.WithCmd, "with-cmd", false, "generate cmd directory with main.go")
	cmd.Flags().BoolVar(&opts.WithTest, "with-test", false, "generate test files")
	cmd.Flags().BoolVar(&opts.Standalone, "standalone", false, "generate as standalone app without extension structure")
	cmd.Flags().StringVar(&opts.Group, "group", "", "belongs domain group (optional)")

	return cmd
}
