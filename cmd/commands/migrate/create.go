package migrate

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/ncobase/ncore/utils"

	"github.com/spf13/cobra"
)

func newCreateCommand() *cobra.Command {
	var migrationsPath string

	cmd := &cobra.Command{
		Use:   "create [name]",
		Short: "Create a new migration",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			timestamp := time.Now().Format("20060102150405")
			filename := fmt.Sprintf("%s_%s.sql", timestamp, name)

			if migrationsPath == "" {
				migrationsPath = "migrations" // default path
			}

			content := fmt.Sprintf(`-- migrate:up



-- migrate:down


`)

			if err := utils.EnsureDir(migrationsPath); err != nil {
				return fmt.Errorf("failed to create migrations directory: %v", err)
			}

			path := filepath.Join(migrationsPath, filename)
			if err := utils.WriteTemplateFile(path, content, nil); err != nil {
				return fmt.Errorf("failed to create migration file: %v", err)
			}

			fmt.Printf("Created migration file: %s\n", path)
			return nil
		},
	}

	cmd.Flags().StringVarP(&migrationsPath, "path", "p", "", "migrations directory path")
	return cmd
}
