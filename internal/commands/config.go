package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/velocitykode/velocity-installer/internal/config"
	"github.com/velocitykode/velocity-installer/internal/ui"
)

var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage global CLI configuration",
	Long:  `Set, get, list, or reset global configuration defaults for velocity commands.`,
}

var configSetCmd = &cobra.Command{
	Use:           "set <key> <value>",
	Short:         "Set a configuration value",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			ui.Error("Key and value are required")
			ui.Newline()
			ui.Muted("Usage: velocity config set <key> <value>")
			return fmt.Errorf("")
		}
		return nil
	},
	RunE: runConfigSet,
}

var configGetCmd = &cobra.Command{
	Use:           "get <key>",
	Short:         "Get a configuration value",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			ui.Error("Key is required")
			ui.Newline()
			ui.Muted("Usage: velocity config get <key>")
			return fmt.Errorf("")
		}
		return nil
	},
	RunE: runConfigGet,
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration",
	Args:  cobra.NoArgs,
	RunE:  runConfigList,
}

var configResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset all configuration",
	Args:  cobra.NoArgs,
	RunE:  runConfigReset,
}

func init() {
	ConfigCmd.AddCommand(configSetCmd)
	ConfigCmd.AddCommand(configGetCmd)
	ConfigCmd.AddCommand(configListCmd)
	ConfigCmd.AddCommand(configResetCmd)
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	key, value := args[0], args[1]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Parse key and validate
	switch key {
	case "default.database":
		if err := config.ValidateDatabase(value); err != nil {
			return err
		}
		cfg.Defaults.Database = value
	case "default.cache":
		if err := config.ValidateCache(value); err != nil {
			return err
		}
		cfg.Defaults.Cache = value
	case "default.queue":
		if err := config.ValidateQueue(value); err != nil {
			return err
		}
		cfg.Defaults.Queue = value
	case "default.auth":
		cfg.Defaults.Auth = value == "true"
	case "default.api":
		cfg.Defaults.API = value == "true"
	default:
		return fmt.Errorf("unknown configuration key: %s", key)
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	ui.Success(fmt.Sprintf("Set %s = %s", key, value))

	path, _ := config.ConfigPath()
	ui.Muted(fmt.Sprintf("Configuration saved to %s", path))

	return nil
}

func runConfigGet(cmd *cobra.Command, args []string) error {
	key := args[0]

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	var value string
	switch key {
	case "default.database":
		value = cfg.Defaults.Database
	case "default.cache":
		value = cfg.Defaults.Cache
	case "default.queue":
		value = cfg.Defaults.Queue
	case "default.auth":
		if cfg.Defaults.Auth {
			value = "true"
		} else {
			value = "false"
		}
	case "default.api":
		if cfg.Defaults.API {
			value = "true"
		} else {
			value = "false"
		}
	default:
		return fmt.Errorf("unknown configuration key: %s", key)
	}

	if value == "" || value == "false" {
		ui.Muted("(not set)")
	} else {
		ui.Info(value)
	}

	return nil
}

func runConfigList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	path, _ := config.ConfigPath()
	ui.Info(fmt.Sprintf("Configuration (%s)", path))

	if cfg.Defaults.Database != "" {
		ui.KeyValue("default.database", cfg.Defaults.Database)
	}
	if cfg.Defaults.Cache != "" {
		ui.KeyValue("default.cache", cfg.Defaults.Cache)
	}
	if cfg.Defaults.Queue != "" {
		ui.KeyValue("default.queue", cfg.Defaults.Queue)
	}
	if cfg.Defaults.Auth {
		ui.KeyValue("default.auth", "true")
	}
	if cfg.Defaults.API {
		ui.KeyValue("default.api", "true")
	}

	return nil
}

func runConfigReset(cmd *cobra.Command, args []string) error {
	path, err := config.ConfigPath()
	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to reset config: %w", err)
	}

	ui.Success("Configuration reset")
	ui.Muted(fmt.Sprintf("Deleted %s", path))

	return nil
}
