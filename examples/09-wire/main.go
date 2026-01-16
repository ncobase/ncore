package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	_ "github.com/ncobase/ncore/data/sqlite"
)

func main() {
	flag.Parse()

	app, cleanup, err := InitializeApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize app: %v\n", err)
		os.Exit(1)
	}
	defer cleanup()

	app.Logger.Info(context.Background(), "Wire example initialized", "app", app.Config.AppName)

	if err := initExtras(app); err != nil {
		app.Logger.Error(context.Background(), "Extras initialization failed", "error", err)
		os.Exit(1)
	}

	app.Logger.Info(context.Background(), "Wire example completed")
}

func initExtras(app *App) error {
	if _, err := InitializeTokenManager(); err != nil {
		return err
	}

	pool, cleanup, err := InitializeWorkerPool()
	if err != nil {
		return err
	}
	defer cleanup()

	if err := pool.Submit(func() error {
		app.Logger.Info(context.Background(), "Worker pool task executed")
		return nil
	}); err != nil {
		return err
	}

	return nil
}
