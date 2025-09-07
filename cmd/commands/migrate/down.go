package migrate

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newDownCommand() *cobra.Command {
	var migrationsPath string

	cmd := &cobra.Command{
		Use:   "down",
		Short: "Rollback the last migration",
		RunE: func(cmd *cobra.Command, args []string) error {
			if migrationsPath == "" {
				migrationsPath = "migrations" // default path
			}
			fmt.Printf("Rolling back last migration from %s...\n", migrationsPath)
			return nil
		},
	}

	cmd.Flags().StringVarP(&migrationsPath, "path", "p", "", "migrations directory path")
	return cmd
}
