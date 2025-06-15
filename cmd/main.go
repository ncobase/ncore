package main

import (
	"fmt"
	"os"

	"github.com/ncobase/ncore/cmd/commands"
)

func main() {
	// Create root command
	rootCmd := commands.NewRootCmd()

	// Disable completion
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	// Execute
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
