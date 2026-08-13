package commands

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/velocitykode/prism"
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
			prism.Error("Project name is required")
			prism.Newline()
			prism.Muted("Usage: velocity new [project-name] [flags]")
			prism.Newline()
			prism.Muted("Flags:")
			prism.Muted("  --database         Database driver (postgres, mysql, sqlite)")
			prism.Muted("  --cache            Cache driver (redis, memory)")
			prism.Muted("  --api              Create API-only project (no frontend)")
			prism.Muted("  --stack            Frontend stack for full-stack projects (react, vue)")
			prism.Muted("  --ssr              Enable Inertia server-side rendering")
			prism.Muted("  -y, --non-interactive  Skip all prompts; use flags or defaults")
			return fmt.Errorf("")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]
		prism.Header("velocity new")

		if info, err := os.Stat(projectName); err == nil {
			kind := "file"
			if info.IsDir() {
				kind = "directory"
			}
			prism.Error(fmt.Sprintf("A %s named '%s' already exists here", kind, projectName))
			prism.Muted("Pick a different name, or remove the existing path and try again.")
			return
		} else if !os.IsNotExist(err) {
			prism.Error(fmt.Sprintf("Cannot check '%s': %v", projectName, err))
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
			api = prism.Select(
				"Project type:",
				[]string{labelFullStack, labelAPI},
				prism.WithSelectDefault(labelFullStack),
			) == labelAPI
		}

		// The frontend questions follow the project-type answer directly:
		// picking "Full stack" and then being asked about databases before
		// anything about the frontend reads like the stack was decided for
		// you.
		if !api && ask("stack") {
			stack = prism.Select(
				"Frontend stack:",
				generator.ValidStacks,
				prism.WithSelectDefault(stack),
			)
		}

		if !api && ask("ssr") {
			ssr = prism.Confirm("Enable Inertia server-side rendering?", prism.WithDefaultNo())
		}

		if ask("database") {
			database = prism.Select(
				"Database:",
				[]string{"sqlite", "postgres", "mysql"},
				prism.WithSelectDefault(database),
			)
		}

		if ask("cache") {
			cache = prism.Select(
				"Cache:",
				[]string{"memory", "redis"},
				prism.WithSelectDefault(cache),
			)
		}

		// Validate flags up-front so non-interactive mode fails fast.
		if !api {
			stack = strings.ToLower(strings.TrimSpace(stack))
			if !slices.Contains(generator.ValidStacks, stack) {
				prism.Error(fmt.Sprintf(
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
				prism.Newline()
				prism.Warning("Project ready - database setup pending")
				prism.NextSteps([]string{
					"Start your database server",
					fmt.Sprintf("cd %s", projectName),
					"./vel migrate",
					"./vel serve",
				})
				return
			}
			prism.Newline()
			prism.Error(err.Error())
			return
		}

		prism.Newline()
		prism.Success("Project ready")
		prism.NextSteps([]string{
			fmt.Sprintf("cd %s", projectName),
			"./vel serve",
		})
		prism.KeyValue("App", prism.Highlight("http://localhost:4000"))
		if !config.API {
			prism.KeyValue("Vite", prism.Highlight("http://localhost:5173"))
		}
		prism.Newline()
		prism.Muted("More: ./vel migrate, ./vel routes, ./vel gen handler")
		prism.Newline()
	},
}

func init() {
	NewCmd.Flags().StringVar(&database, "database", "sqlite", "Database driver (postgres, mysql, sqlite)")
	NewCmd.Flags().StringVar(&cache, "cache", "memory", "Cache driver (redis, memory)")
	NewCmd.Flags().BoolVar(&api, "api", false, "Create API-only project (no frontend)")
	NewCmd.Flags().StringVar(&stack, "stack", "react", "Frontend stack for full-stack projects (react, vue)")
	NewCmd.Flags().BoolVar(&ssr, "ssr", false, "Enable Inertia server-side rendering (sets VIEW_SSR_ENABLED=true and wires Vite SSR)")
	NewCmd.Flags().BoolVarP(&nonInteractive, "non-interactive", "y", false, "Skip all prompts; use flag values or defaults")
}
