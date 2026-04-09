package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/velocitykode/velocity-installer/internal/generator"
	"github.com/velocitykode/velocity-installer/internal/ui"
)

var (
	database string
	cache    string
	api      bool
)

var NewCmd = &cobra.Command{
	Use:           "new [project-name]",
	Short:         "Create a new Velocity project",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			ui.Error("Project name is required")
			ui.Newline()
			ui.Muted("Usage: velocity new [project-name]")
			ui.Newline()
			ui.Muted("Flags:")
			ui.Muted("  --database    Database driver (postgres, sqlite)")
			ui.Muted("  --cache       Cache driver (redis, memory)")
			ui.Muted("  --api         Create API-only project (no frontend)")
			return fmt.Errorf("")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]
		ui.Header("velocity new")

		// Create project with flags (defaults to sqlite if not specified)
		config := generator.ProjectConfig{
			Name:     projectName,
			Module:   projectName,
			Database: database,
			Cache:    cache,
			API:      api,
		}

		if err := generator.CreateProject(config); err != nil {
			ui.Newline()
			ui.Error(err.Error())
			return
		}

		ui.Newline()
		ui.Info("Starting development servers")

		generator.StartDevServers(projectName, config.API)
	},
}

func init() {
	NewCmd.Flags().StringVar(&database, "database", "sqlite", "Database driver (postgres, sqlite)")
	NewCmd.Flags().StringVar(&cache, "cache", "memory", "Cache driver (redis, memory)")
	NewCmd.Flags().BoolVar(&api, "api", false, "Create API-only project (no frontend)")
}
