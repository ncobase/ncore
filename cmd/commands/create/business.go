package create

import (
	"ncore/cmd/generator"

	"github.com/spf13/cobra"
)

func newBusinessCommand() *cobra.Command {
	opts := &generator.Options{}

	cmd := &cobra.Command{
		Use:     "business [name]",
		Aliases: []string{"b"},
		Short:   "Create a new extension in business domain",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Name = args[0]
			opts.Type = "business"

			// Get flags
			opts.ModuleName, _ = cmd.Flags().GetString("module")
			opts.OutputPath, _ = cmd.Flags().GetString("path")
			opts.UseMongo, _ = cmd.Flags().GetBool("use-mongo")
			opts.UseEnt, _ = cmd.Flags().GetBool("use-ent")
			opts.UseGorm, _ = cmd.Flags().GetBool("use-gorm")
			opts.WithTest, _ = cmd.Flags().GetBool("with-test")
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
	cmd.Flags().BoolVar(&opts.WithTest, "with-test", false, "generate test files")
	cmd.Flags().StringVar(&opts.Group, "group", "", "belongs domain group (optional)")

	return cmd
}
