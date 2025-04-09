package migrate

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newUpCommand() *cobra.Command {
	var migrationsPath string

	cmd := &cobra.Command{
		Use:   "up",
		Short: "Run all pending migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			if migrationsPath == "" {
				migrationsPath = "migrations" // default path
			}
			// TODO: Implement migration execution
			fmt.Printf("Running migrations from %s...\n", migrationsPath)
			return nil
		},
	}

	cmd.Flags().StringVarP(&migrationsPath, "path", "p", "", "migrations directory path")
	return cmd
}
