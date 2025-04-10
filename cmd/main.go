package main

import (
	"fmt"
	"os"

	"github.com/ncobase/ncore/cmd/commands"
)

func main() {
	// Execute the root command
	rootCmd := commands.NewRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
