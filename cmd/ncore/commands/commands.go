package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ncobase/ncore/cmd/ncore/commands/create"
	"github.com/ncobase/ncore/cmd/ncore/commands/migrate"
	"github.com/ncobase/ncore/config"
	extm "github.com/ncobase/ncore/extension/manager"
	"github.com/ncobase/ncore/utils"
	"github.com/ncobase/ncore/version"

	"github.com/spf13/cobra"
)

// NewStartCommand creates the start command
func NewStartCommand() *cobra.Command {
	var configFile string

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the NCore server",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig(configFile)
			if err != nil {
				return fmt.Errorf("failed to load config: %v", err)
			}

			m, err := extm.NewManager(cfg)
			if err != nil {
				return fmt.Errorf("failed to create manager: %v", err)
			}

			if err := m.LoadPlugins(); err != nil {
				return fmt.Errorf("failed to load plugins: %v", err)
			}

			fmt.Println("Starting NCore server...")
			return nil
		},
	}

	cmd.Flags().StringVarP(&configFile, "config", "c", "config.yaml", "config file path")
	return cmd
}

// NewPluginCommand creates the plugin management command
func NewPluginCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plugin",
		Short: "Plugin management commands",
	}

	cmd.AddCommand(
		newPluginListCommand(),
		newPluginInstallCommand(),
	)

	return cmd
}

func newPluginListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List installed plugins",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig("config.yaml")
			if err != nil {
				return err
			}

			m, err := extm.NewManager(cfg)
			if err != nil {
				return err
			}

			plugins := m.GetExtensions()
			for name, plugin := range plugins {
				fmt.Printf("Plugin: %s\n", name)
				metadata := plugin.Instance.GetMetadata()
				fmt.Printf("  Version: %s\n", metadata.Version)
				fmt.Printf("  Type: %s\n", metadata.Type)
				fmt.Printf("  Status: %s\n\n", plugin.Instance.Status())
			}

			return nil
		},
	}
}

func newPluginInstallCommand() *cobra.Command {
	var source string

	cmd := &cobra.Command{
		Use:   "install [name]",
		Short: "Install a plugin",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			cfg, err := config.LoadConfig("config.yaml")
			if err != nil {
				return err
			}

			destPath := filepath.Join(cfg.Extension.Path, name+utils.GetPlatformExt())
			if err := os.Link(source, destPath); err != nil {
				return fmt.Errorf("failed to install plugin: %v", err)
			}

			fmt.Printf("Plugin %s installed successfully\n", name)
			return nil
		},
	}

	cmd.Flags().StringVarP(&source, "source", "s", "", "source path of the plugin")
	_ = cmd.MarkFlagRequired("source")
	return cmd
}

// NewDocsCommand creates the documentation command
func NewDocsCommand() *cobra.Command {
	var format string
	var output string

	cmd := &cobra.Command{
		Use:   "docs",
		Short: "Generate documentation",
		RunE: func(cmd *cobra.Command, args []string) error {
			var content string
			switch format {
			case "markdown":
				content = "# NCore API Documentation\n\n"
			case "json":
				content = `{"swagger": "2.0", "info": {"title": "NCore API", "version": "1.0"}}`
			default:
				return fmt.Errorf("unsupported format: %s", format)
			}

			if output == "" {
				fmt.Println(content)
				return nil
			}

			return os.WriteFile(output, []byte(content), 0644)
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "markdown", "documentation format (markdown or json)")
	cmd.Flags().StringVarP(&output, "output", "o", "", "output file path")
	return cmd
}

// NewVersionCommand creates the version command
func NewVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			info := version.GetVersionInfo()
			fmt.Println("Version:", info.Version)
			fmt.Println("Built At:", info.BuiltAt)
		},
	}
}

// NewCreateCommand creates the extension generation command
func NewCreateCommand() *cobra.Command {
	return create.NewCommand()
}

// NewMigrateCommand creates the migrate command
func NewMigrateCommand() *cobra.Command {
	return migrate.NewCommand()
}
