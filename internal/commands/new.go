package commands

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/velocitykode/velocity-installer/internal/generator"
	"github.com/velocitykode/velocity-installer/internal/ui"
)

var (
	database string
	cache    string
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
		}

		if err := generator.CreateProject(config); err != nil {
			ui.Newline()
			ui.Error(err.Error())
			return
		}

		// Build vel binary
		ui.Step("Building project CLI...")
		buildCmd := exec.Command("go", "build", "-o", "vel", "./cmd/vel")
		buildCmd.Dir = projectName
		if err := buildCmd.Run(); err != nil {
			ui.Warning("Failed to build vel: " + err.Error())
			ui.Muted("Run manually: go build -o vel ./cmd/vel")
		} else {
			ui.Success("Built ./vel")
		}

		ui.Newline()
		ui.Info("Starting development servers")

		generator.StartDevServers(projectName)
	},
}

func init() {
	NewCmd.Flags().StringVar(&database, "database", "sqlite", "Database driver (postgres, sqlite)")
	NewCmd.Flags().StringVar(&cache, "cache", "memory", "Cache driver (redis, memory)")
}
