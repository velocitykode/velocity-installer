package commands

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/velocitykode/velocity-installer/internal/generator"
	cli "github.com/velocitykode/velocity-cli"
)

var (
	database string
	cache    string
	api      bool
	ssr      bool
)

var NewCmd = &cobra.Command{
	Use:           "new [project-name]",
	Short:         "Create a new Velocity project",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			cli.Error("Project name is required")
			cli.Newline()
			cli.Muted("Usage: velocity new [project-name]")
			cli.Newline()
			cli.Muted("Flags:")
			cli.Muted("  --database    Database driver (postgres, sqlite)")
			cli.Muted("  --cache       Cache driver (redis, memory)")
			cli.Muted("  --api         Create API-only project (no frontend)")
			cli.Muted("  --ssr         Enable Inertia server-side rendering")
			return fmt.Errorf("")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]
		cli.Header("velocity new")

		if !cmd.Flags().Changed("database") {
			database = cli.Select(
				"Database:",
				[]string{"sqlite", "postgres"},
				cli.WithSelectDefault(database),
			)
		}

		if !api && !cmd.Flags().Changed("ssr") {
			ssr = cli.Confirm("Enable Inertia server-side rendering?", cli.WithDefaultNo())
		}

		config := generator.ProjectConfig{
			Name:     projectName,
			Module:   projectName,
			Database: database,
			Cache:    cache,
			API:      api,
			SSR:      ssr,
		}

		if err := generator.CreateProject(config); err != nil {
			cli.Newline()
			cli.Error(err.Error())
			return
		}

		// Build vel binary
		cli.Step("Building vel...")
		buildCmd := exec.Command("go", "build", "-o", "vel", ".")
		buildCmd.Dir = projectName
		if err := buildCmd.Run(); err != nil {
			cli.Warning("Failed to build vel: " + err.Error())
			cli.Muted("Run manually: go build -o vel .")
		} else {
			cli.Success("Built ./vel")
		}

		cli.Newline()
		cli.Info("Starting development servers")

		generator.StartDevServers(projectName, config.API)
	},
}

func init() {
	NewCmd.Flags().StringVar(&database, "database", "sqlite", "Database driver (postgres, sqlite)")
	NewCmd.Flags().StringVar(&cache, "cache", "memory", "Cache driver (redis, memory)")
	NewCmd.Flags().BoolVar(&api, "api", false, "Create API-only project (no frontend)")
	NewCmd.Flags().BoolVar(&ssr, "ssr", false, "Enable Inertia server-side rendering (sets INERTIA_SSR_ENABLED=true and wires Vite SSR)")
}
