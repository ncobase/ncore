package commands

import (
	"github.com/spf13/cobra"
)

// NewRootCmd creates the root command
func NewRootCmd() *cobra.Command {
	// Create root command
	rootCmd := &cobra.Command{
		Use:   "ncore",
		Short: "A set of reusable components for Go applications",
	}

	// Add subcommands
	rootCmd.AddCommand(
		NewStartCommand(),
		NewPluginCommand(),
		NewDocsCommand(),
		NewVersionCommand(),
	)

	return rootCmd
}
