// Package main boots the basic REST API example.
package main

import (
	"fmt"
	"os"
)

func main() {
	// Initialize application with Wire
	app, cleanup, err := InitializeApp()
	if err != nil {
		fmt.Printf("Failed to initialize app: %v\n", err)
		os.Exit(1)
	}
	defer cleanup()

	// Run application
	if err := app.Run(); err != nil {
		fmt.Printf("Failed to run app: %v\n", err)
		os.Exit(1)
	}
}
