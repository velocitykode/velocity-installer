package commands

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	cli "github.com/velocitykode/velocity-cli"
	"github.com/velocitykode/velocity-installer/internal/generator"
)

var (
	database       string
	cache          string
	stack          string
	api            bool
	ssr            bool
	nonInteractive bool
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
			cli.Muted("Usage: velocity new [project-name] [flags]")
			cli.Newline()
			cli.Muted("Flags:")
			cli.Muted("  --database         Database driver (postgres, mysql, sqlite)")
			cli.Muted("  --cache            Cache driver (redis, memory)")
			cli.Muted("  --api              Create API-only project (no frontend)")
			cli.Muted("  --stack            Frontend stack for full-stack projects (react, vue)")
			cli.Muted("  --ssr              Enable Inertia server-side rendering")
			cli.Muted("  -y, --non-interactive  Skip all prompts; use flags or defaults")
			return fmt.Errorf("")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]
		cli.Header("velocity new")

		if info, err := os.Stat(projectName); err == nil {
			kind := "file"
			if info.IsDir() {
				kind = "directory"
			}
			cli.Error(fmt.Sprintf("A %s named '%s' already exists here", kind, projectName))
			cli.Muted("Pick a different name, or remove the existing path and try again.")
			return
		} else if !os.IsNotExist(err) {
			cli.Error(fmt.Sprintf("Cannot check '%s': %v", projectName, err))
			return
		}

		// Per-flag prompt: ask only when the flag was not explicitly set
		// AND the user did not opt into non-interactive mode.
		ask := func(name string) bool {
			return !nonInteractive && !cmd.Flags().Changed(name)
		}

		if ask("api") {
			const (
				labelFullStack = "Full stack (Inertia + Vite)"
				labelAPI       = "API only (no frontend)"
			)
			api = cli.Select(
				"Project type:",
				[]string{labelFullStack, labelAPI},
				cli.WithSelectDefault(labelFullStack),
			) == labelAPI
		}

		if ask("database") {
			database = cli.Select(
				"Database:",
				[]string{"sqlite", "postgres", "mysql"},
				cli.WithSelectDefault(database),
			)
		}

		if ask("cache") {
			cache = cli.Select(
				"Cache:",
				[]string{"memory", "redis"},
				cli.WithSelectDefault(cache),
			)
		}

		if !api && ask("stack") {
			stack = cli.Select(
				"Frontend stack:",
				generator.ValidStacks,
				cli.WithSelectDefault(stack),
			)
		}

		if !api && ask("ssr") {
			ssr = cli.Confirm("Enable Inertia server-side rendering?", cli.WithDefaultNo())
		}

		// Validate flags up-front so non-interactive mode fails fast.
		if !api {
			stack = strings.ToLower(strings.TrimSpace(stack))
			if !slices.Contains(generator.ValidStacks, stack) {
				cli.Error(fmt.Sprintf(
					"Invalid --stack %q. Valid: %s",
					stack, strings.Join(generator.ValidStacks, ", "),
				))
				return
			}
		}

		config := generator.ProjectConfig{
			Name:     projectName,
			Module:   projectName,
			Database: database,
			Cache:    cache,
			API:      api,
			Stack:    stack,
			SSR:      ssr,
		}

		if err := generator.CreateProject(config); err != nil {
			if errors.Is(err, generator.ErrMigrationsSkipped) {
				cli.Newline()
				cli.Warning("Project ready - database setup pending")
				cli.NextSteps([]string{
					"Start your database server",
					fmt.Sprintf("cd %s", projectName),
					"./vel migrate",
					"./vel serve",
				})
				return
			}
			cli.Newline()
			cli.Error(err.Error())
			return
		}

		cli.Newline()
		cli.Success("Project ready")
		cli.NextSteps([]string{
			fmt.Sprintf("cd %s", projectName),
			"./vel serve",
		})
		cli.KeyValue("App", cli.Highlight("http://localhost:4000"))
		if !config.API {
			cli.KeyValue("Vite", cli.Highlight("http://localhost:5173"))
		}
		cli.Newline()
		cli.Muted("More: ./vel migrate, ./vel route:list, ./vel make:handler")
		cli.Newline()
	},
}

func init() {
	NewCmd.Flags().StringVar(&database, "database", "sqlite", "Database driver (postgres, mysql, sqlite)")
	NewCmd.Flags().StringVar(&cache, "cache", "memory", "Cache driver (redis, memory)")
	NewCmd.Flags().BoolVar(&api, "api", false, "Create API-only project (no frontend)")
	NewCmd.Flags().StringVar(&stack, "stack", "react", "Frontend stack for full-stack projects (react, vue)")
	NewCmd.Flags().BoolVar(&ssr, "ssr", false, "Enable Inertia server-side rendering (sets INERTIA_SSR_ENABLED=true and wires Vite SSR)")
	NewCmd.Flags().BoolVarP(&nonInteractive, "non-interactive", "y", false, "Skip all prompts; use flag values or defaults")
}
