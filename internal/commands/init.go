package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/velocitykode/velocity-installer/internal/config"
	"github.com/velocitykode/velocity-installer/internal/detector"
	"github.com/velocitykode/velocity-installer/internal/generator"
	"github.com/velocitykode/velocity-installer/internal/ui"
)

var (
	initDatabase      string
	initCache         string
	initAuth          bool
	initAPI           bool
	initNoInteraction bool
)

var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Velocity in existing Go project",
	Long:  `Add Velocity framework structure to an existing Go project while preserving all existing files.`,
	RunE:  runInit,
}

func init() {
	InitCmd.Flags().StringVar(&initDatabase, "database", "", "Database driver (postgres, mysql, sqlite)")
	InitCmd.Flags().StringVar(&initCache, "cache", "", "Cache driver (redis, memory)")
	InitCmd.Flags().BoolVar(&initAuth, "auth", false, "Include authentication")
	InitCmd.Flags().BoolVar(&initAPI, "api", false, "API-only structure")
	InitCmd.Flags().BoolVar(&initNoInteraction, "no-interaction", false, "Non-interactive mode")
}

func runInit(cmd *cobra.Command, args []string) error {
	ui.Header("init")

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// Detect project
	ui.Info("Detecting Go project...")
	info, err := detector.Detect(cwd)
	if err != nil {
		return fmt.Errorf("not a Go project\n\nThis directory does not contain a go.mod file.\nInitialize a Go module first:\n  go mod init github.com/yourname/project\n\nThen run 'velocity init' again.")
	}

	if info.HasVelocity {
		return fmt.Errorf("Velocity already initialized\n\nThis project already has Velocity structure (app/, config/, routes/).\n\nIf you want to create a new project, use:\n  velocity new project-name")
	}

	ui.Success(fmt.Sprintf("Detected Go project: %s", info.ModuleName))

	// Load config defaults
	cfg, _ := config.Load()

	// Merge preferences: Flags > Config > Defaults
	database := initDatabase
	if !cmd.Flags().Changed("database") && cfg.Defaults.Database != "" {
		database = cfg.Defaults.Database
	}

	cache := initCache
	if !cmd.Flags().Changed("cache") && cfg.Defaults.Cache != "" {
		cache = cfg.Defaults.Cache
	}

	auth := initAuth
	if !cmd.Flags().Changed("auth") && cfg.Defaults.Auth {
		auth = cfg.Defaults.Auth
	}

	api := initAPI
	if !cmd.Flags().Changed("api") && cfg.Defaults.API {
		api = cfg.Defaults.API
	}

	// TODO: Interactive prompts if not --no-interaction and values missing
	// For now, we'll skip interactive mode implementation

	// Create project config
	// Use "." as name since we're working in current directory
	projectCfg := generator.ProjectConfig{
		Name:     ".",
		Module:   info.ModuleName,
		Database: database,
		Cache:    cache,
		Auth:     auth,
		API:      api,
	}

	// Initialize project
	if err := generator.InitProject(projectCfg, cwd); err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}

	// Success message
	ui.Newline()
	ui.Success("Velocity initialized successfully!")
	ui.NextSteps([]string{
		"go mod download",
		"./vel serve",
	})

	return nil
}
