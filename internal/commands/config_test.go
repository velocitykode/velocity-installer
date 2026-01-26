package commands

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/velocitykode/velocity-installer/internal/config"
	"gopkg.in/yaml.v3"
)

func TestRunConfigSet(t *testing.T) {
	// Save original HOME and stdout, restore after tests
	originalHome := os.Getenv("HOME")
	originalStdout := os.Stdout
	defer func() {
		os.Setenv("HOME", originalHome)
		os.Stdout = originalStdout
	}()

	tests := []struct {
		name        string
		args        []string
		setupFunc   func(t *testing.T, tempDir string)
		wantErr     bool
		errContains string
		validate    func(t *testing.T, tempDir string)
	}{
		{
			name: "sets valid database config",
			args: []string{"default.database", "postgres"},
			setupFunc: func(t *testing.T, tempDir string) {
				// No setup needed
			},
			wantErr: false,
			validate: func(t *testing.T, tempDir string) {
				cfg, err := config.Load()
				if err != nil {
					t.Fatalf("Failed to load config: %v", err)
				}
				if cfg.Defaults.Database != "postgres" {
					t.Errorf("Database = %s, want postgres", cfg.Defaults.Database)
				}
			},
		},
		{
			name: "sets valid cache config",
			args: []string{"default.cache", "redis"},
			setupFunc: func(t *testing.T, tempDir string) {
				// No setup needed
			},
			wantErr: false,
			validate: func(t *testing.T, tempDir string) {
				cfg, err := config.Load()
				if err != nil {
					t.Fatalf("Failed to load config: %v", err)
				}
				if cfg.Defaults.Cache != "redis" {
					t.Errorf("Cache = %s, want redis", cfg.Defaults.Cache)
				}
			},
		},
		{
			name: "sets valid queue config",
			args: []string{"default.queue", "database"},
			setupFunc: func(t *testing.T, tempDir string) {
				// No setup needed
			},
			wantErr: false,
			validate: func(t *testing.T, tempDir string) {
				cfg, err := config.Load()
				if err != nil {
					t.Fatalf("Failed to load config: %v", err)
				}
				if cfg.Defaults.Queue != "database" {
					t.Errorf("Queue = %s, want database", cfg.Defaults.Queue)
				}
			},
		},
		{
			name: "sets auth to true",
			args: []string{"default.auth", "true"},
			setupFunc: func(t *testing.T, tempDir string) {
				// No setup needed
			},
			wantErr: false,
			validate: func(t *testing.T, tempDir string) {
				cfg, err := config.Load()
				if err != nil {
					t.Fatalf("Failed to load config: %v", err)
				}
				if !cfg.Defaults.Auth {
					t.Errorf("Auth = %v, want true", cfg.Defaults.Auth)
				}
			},
		},
		{
			name: "sets auth to false when value is not true",
			args: []string{"default.auth", "false"},
			setupFunc: func(t *testing.T, tempDir string) {
				// No setup needed
			},
			wantErr: false,
			validate: func(t *testing.T, tempDir string) {
				cfg, err := config.Load()
				if err != nil {
					t.Fatalf("Failed to load config: %v", err)
				}
				if cfg.Defaults.Auth {
					t.Errorf("Auth = %v, want false", cfg.Defaults.Auth)
				}
			},
		},
		{
			name: "sets api to true",
			args: []string{"default.api", "true"},
			setupFunc: func(t *testing.T, tempDir string) {
				// No setup needed
			},
			wantErr: false,
			validate: func(t *testing.T, tempDir string) {
				cfg, err := config.Load()
				if err != nil {
					t.Fatalf("Failed to load config: %v", err)
				}
				if !cfg.Defaults.API {
					t.Errorf("API = %v, want true", cfg.Defaults.API)
				}
			},
		},
		{
			name: "sets api to false when value is not true",
			args: []string{"default.api", "anything"},
			setupFunc: func(t *testing.T, tempDir string) {
				// No setup needed
			},
			wantErr: false,
			validate: func(t *testing.T, tempDir string) {
				cfg, err := config.Load()
				if err != nil {
					t.Fatalf("Failed to load config: %v", err)
				}
				if cfg.Defaults.API {
					t.Errorf("API = %v, want false", cfg.Defaults.API)
				}
			},
		},
		{
			name: "updates existing config",
			args: []string{"default.database", "sqlite"},
			setupFunc: func(t *testing.T, tempDir string) {
				cfg := &config.Config{
					Defaults: config.DefaultConfig{
						Database: "postgres",
						Cache:    "redis",
					},
				}
				if err := cfg.Save(); err != nil {
					t.Fatalf("Failed to save initial config: %v", err)
				}
			},
			wantErr: false,
			validate: func(t *testing.T, tempDir string) {
				cfg, err := config.Load()
				if err != nil {
					t.Fatalf("Failed to load config: %v", err)
				}
				if cfg.Defaults.Database != "sqlite" {
					t.Errorf("Database = %s, want sqlite", cfg.Defaults.Database)
				}
				if cfg.Defaults.Cache != "redis" {
					t.Errorf("Cache = %s, want redis (should be preserved)", cfg.Defaults.Cache)
				}
			},
		},
		{
			name:        "returns error for unknown key",
			args:        []string{"unknown.key", "value"},
			setupFunc:   func(t *testing.T, tempDir string) {},
			wantErr:     true,
			errContains: "unknown configuration key",
		},
		{
			name:        "returns error for invalid database value",
			args:        []string{"default.database", "mongodb"},
			setupFunc:   func(t *testing.T, tempDir string) {},
			wantErr:     true,
			errContains: "invalid database",
		},
		{
			name:        "returns error for invalid cache value",
			args:        []string{"default.cache", "memcached"},
			setupFunc:   func(t *testing.T, tempDir string) {},
			wantErr:     true,
			errContains: "invalid cache",
		},
		{
			name:        "returns error for invalid queue value",
			args:        []string{"default.queue", "rabbitmq"},
			setupFunc:   func(t *testing.T, tempDir string) {},
			wantErr:     true,
			errContains: "invalid queue",
		},
		{
			name: "sets all database options",
			args: []string{"default.database", "mysql"},
			setupFunc: func(t *testing.T, tempDir string) {
				// No setup needed
			},
			wantErr: false,
			validate: func(t *testing.T, tempDir string) {
				cfg, err := config.Load()
				if err != nil {
					t.Fatalf("Failed to load config: %v", err)
				}
				if cfg.Defaults.Database != "mysql" {
					t.Errorf("Database = %s, want mysql", cfg.Defaults.Database)
				}
			},
		},
		{
			name: "sets all cache options",
			args: []string{"default.cache", "memory"},
			setupFunc: func(t *testing.T, tempDir string) {
				// No setup needed
			},
			wantErr: false,
			validate: func(t *testing.T, tempDir string) {
				cfg, err := config.Load()
				if err != nil {
					t.Fatalf("Failed to load config: %v", err)
				}
				if cfg.Defaults.Cache != "memory" {
					t.Errorf("Cache = %s, want memory", cfg.Defaults.Cache)
				}
			},
		},
		{
			name: "sets all queue options",
			args: []string{"default.queue", "redis"},
			setupFunc: func(t *testing.T, tempDir string) {
				// No setup needed
			},
			wantErr: false,
			validate: func(t *testing.T, tempDir string) {
				cfg, err := config.Load()
				if err != nil {
					t.Fatalf("Failed to load config: %v", err)
				}
				if cfg.Defaults.Queue != "redis" {
					t.Errorf("Queue = %s, want redis", cfg.Defaults.Queue)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory for this test
			tempDir := t.TempDir()
			os.Setenv("HOME", tempDir)

			// Redirect stdout to capture output
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Run setup if provided
			if tt.setupFunc != nil {
				tt.setupFunc(t, tempDir)
			}

			// Create a mock command
			cmd := &cobra.Command{}

			// Execute function
			err := runConfigSet(cmd, tt.args)

			// Restore stdout and read captured output
			w.Close()
			os.Stdout = originalStdout
			var buf bytes.Buffer
			io.Copy(&buf, r)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("runConfigSet() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" {
				if err == nil || !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("runConfigSet() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			// Run validation if provided
			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, tempDir)
			}
		})
	}
}

func TestRunConfigGet(t *testing.T) {
	// Save original HOME and stdout, restore after tests
	originalHome := os.Getenv("HOME")
	originalStdout := os.Stdout
	defer func() {
		os.Setenv("HOME", originalHome)
		os.Stdout = originalStdout
	}()

	tests := []struct {
		name        string
		args        []string
		setupFunc   func(t *testing.T, tempDir string)
		wantErr     bool
		errContains string
		wantOutput  string
	}{
		{
			name: "gets database config when set",
			args: []string{"default.database"},
			setupFunc: func(t *testing.T, tempDir string) {
				cfg := &config.Config{
					Defaults: config.DefaultConfig{
						Database: "postgres",
					},
				}
				if err := cfg.Save(); err != nil {
					t.Fatalf("Failed to save config: %v", err)
				}
			},
			wantErr:    false,
			wantOutput: "postgres",
		},
		{
			name: "gets cache config when set",
			args: []string{"default.cache"},
			setupFunc: func(t *testing.T, tempDir string) {
				cfg := &config.Config{
					Defaults: config.DefaultConfig{
						Cache: "redis",
					},
				}
				if err := cfg.Save(); err != nil {
					t.Fatalf("Failed to save config: %v", err)
				}
			},
			wantErr:    false,
			wantOutput: "redis",
		},
		{
			name: "gets queue config when set",
			args: []string{"default.queue"},
			setupFunc: func(t *testing.T, tempDir string) {
				cfg := &config.Config{
					Defaults: config.DefaultConfig{
						Queue: "database",
					},
				}
				if err := cfg.Save(); err != nil {
					t.Fatalf("Failed to save config: %v", err)
				}
			},
			wantErr:    false,
			wantOutput: "database",
		},
		{
			name: "gets auth config when true",
			args: []string{"default.auth"},
			setupFunc: func(t *testing.T, tempDir string) {
				cfg := &config.Config{
					Defaults: config.DefaultConfig{
						Auth: true,
					},
				}
				if err := cfg.Save(); err != nil {
					t.Fatalf("Failed to save config: %v", err)
				}
			},
			wantErr:    false,
			wantOutput: "true",
		},
		{
			name: "gets auth config when false",
			args: []string{"default.auth"},
			setupFunc: func(t *testing.T, tempDir string) {
				cfg := &config.Config{
					Defaults: config.DefaultConfig{
						Auth: false,
					},
				}
				if err := cfg.Save(); err != nil {
					t.Fatalf("Failed to save config: %v", err)
				}
			},
			wantErr:    false,
			wantOutput: "(not set)",
		},
		{
			name: "gets api config when true",
			args: []string{"default.api"},
			setupFunc: func(t *testing.T, tempDir string) {
				cfg := &config.Config{
					Defaults: config.DefaultConfig{
						API: true,
					},
				}
				if err := cfg.Save(); err != nil {
					t.Fatalf("Failed to save config: %v", err)
				}
			},
			wantErr:    false,
			wantOutput: "true",
		},
		{
			name: "gets api config when false",
			args: []string{"default.api"},
			setupFunc: func(t *testing.T, tempDir string) {
				cfg := &config.Config{
					Defaults: config.DefaultConfig{
						API: false,
					},
				}
				if err := cfg.Save(); err != nil {
					t.Fatalf("Failed to save config: %v", err)
				}
			},
			wantErr:    false,
			wantOutput: "(not set)",
		},
		{
			name:       "shows not set when database is empty",
			args:       []string{"default.database"},
			setupFunc:  func(t *testing.T, tempDir string) {},
			wantErr:    false,
			wantOutput: "(not set)",
		},
		{
			name:       "shows not set when cache is empty",
			args:       []string{"default.cache"},
			setupFunc:  func(t *testing.T, tempDir string) {},
			wantErr:    false,
			wantOutput: "(not set)",
		},
		{
			name:       "shows not set when queue is empty",
			args:       []string{"default.queue"},
			setupFunc:  func(t *testing.T, tempDir string) {},
			wantErr:    false,
			wantOutput: "(not set)",
		},
		{
			name:        "returns error for unknown key",
			args:        []string{"unknown.key"},
			setupFunc:   func(t *testing.T, tempDir string) {},
			wantErr:     true,
			errContains: "unknown configuration key",
		},
		{
			name: "gets config from existing file",
			args: []string{"default.database"},
			setupFunc: func(t *testing.T, tempDir string) {
				cfg := &config.Config{
					Defaults: config.DefaultConfig{
						Database: "sqlite",
						Cache:    "memory",
						Queue:    "redis",
						Auth:     true,
						API:      false,
					},
				}
				if err := cfg.Save(); err != nil {
					t.Fatalf("Failed to save config: %v", err)
				}
			},
			wantErr:    false,
			wantOutput: "sqlite",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory for this test
			tempDir := t.TempDir()
			os.Setenv("HOME", tempDir)

			// Redirect stdout to capture output
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Run setup if provided
			if tt.setupFunc != nil {
				tt.setupFunc(t, tempDir)
			}

			// Create a mock command
			cmd := &cobra.Command{}

			// Execute function
			err := runConfigGet(cmd, tt.args)

			// Restore stdout and read captured output
			w.Close()
			os.Stdout = originalStdout
			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("runConfigGet() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" {
				if err == nil || !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("runConfigGet() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			// Check output
			if !tt.wantErr && tt.wantOutput != "" {
				if !strings.Contains(output, tt.wantOutput) {
					t.Errorf("runConfigGet() output = %q, want to contain %q", output, tt.wantOutput)
				}
			}
		})
	}
}

func TestRunConfigList(t *testing.T) {
	// Save original HOME and stdout, restore after tests
	originalHome := os.Getenv("HOME")
	originalStdout := os.Stdout
	defer func() {
		os.Setenv("HOME", originalHome)
		os.Stdout = originalStdout
	}()

	tests := []struct {
		name         string
		setupFunc    func(t *testing.T, tempDir string)
		wantErr      bool
		wantContains []string
	}{
		{
			name: "lists all config values when set",
			setupFunc: func(t *testing.T, tempDir string) {
				cfg := &config.Config{
					Defaults: config.DefaultConfig{
						Database: "postgres",
						Cache:    "redis",
						Queue:    "database",
						Auth:     true,
						API:      true,
					},
				}
				if err := cfg.Save(); err != nil {
					t.Fatalf("Failed to save config: %v", err)
				}
			},
			wantErr: false,
			wantContains: []string{
				"default.database",
				"postgres",
				"default.cache",
				"redis",
				"default.queue",
				"database",
				"default.auth",
				"true",
				"default.api",
			},
		},
		{
			name: "lists partial config",
			setupFunc: func(t *testing.T, tempDir string) {
				cfg := &config.Config{
					Defaults: config.DefaultConfig{
						Database: "sqlite",
					},
				}
				if err := cfg.Save(); err != nil {
					t.Fatalf("Failed to save config: %v", err)
				}
			},
			wantErr: false,
			wantContains: []string{
				"default.database",
				"sqlite",
			},
		},
		{
			name: "lists config with only boolean values",
			setupFunc: func(t *testing.T, tempDir string) {
				cfg := &config.Config{
					Defaults: config.DefaultConfig{
						Auth: true,
						API:  true,
					},
				}
				if err := cfg.Save(); err != nil {
					t.Fatalf("Failed to save config: %v", err)
				}
			},
			wantErr: false,
			wantContains: []string{
				"default.auth",
				"default.api",
				"true",
			},
		},
		{
			name:         "lists empty config when file does not exist",
			setupFunc:    func(t *testing.T, tempDir string) {},
			wantErr:      false,
			wantContains: []string{"Configuration"},
		},
		{
			name: "lists config with false boolean values not shown",
			setupFunc: func(t *testing.T, tempDir string) {
				cfg := &config.Config{
					Defaults: config.DefaultConfig{
						Database: "postgres",
						Auth:     false,
						API:      false,
					},
				}
				if err := cfg.Save(); err != nil {
					t.Fatalf("Failed to save config: %v", err)
				}
			},
			wantErr: false,
			wantContains: []string{
				"default.database",
				"postgres",
			},
		},
		{
			name: "shows config file path",
			setupFunc: func(t *testing.T, tempDir string) {
				cfg := &config.Config{
					Defaults: config.DefaultConfig{
						Database: "postgres",
					},
				}
				if err := cfg.Save(); err != nil {
					t.Fatalf("Failed to save config: %v", err)
				}
			},
			wantErr: false,
			wantContains: []string{
				"Configuration",
				".vel",
				"config.yaml",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory for this test
			tempDir := t.TempDir()
			os.Setenv("HOME", tempDir)

			// Redirect stdout to capture output
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Run setup if provided
			if tt.setupFunc != nil {
				tt.setupFunc(t, tempDir)
			}

			// Create a mock command
			cmd := &cobra.Command{}

			// Execute function
			err := runConfigList(cmd, []string{})

			// Restore stdout and read captured output
			w.Close()
			os.Stdout = originalStdout
			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("runConfigList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check output contains expected strings
			if !tt.wantErr {
				for _, want := range tt.wantContains {
					if !strings.Contains(output, want) {
						t.Errorf("runConfigList() output does not contain %q\nGot: %s", want, output)
					}
				}
			}
		})
	}
}

func TestRunConfigReset(t *testing.T) {
	// Save original HOME and stdout, restore after tests
	originalHome := os.Getenv("HOME")
	originalStdout := os.Stdout
	defer func() {
		os.Setenv("HOME", originalHome)
		os.Stdout = originalStdout
	}()

	tests := []struct {
		name      string
		setupFunc func(t *testing.T, tempDir string)
		wantErr   bool
		validate  func(t *testing.T, tempDir string)
	}{
		{
			name: "deletes existing config file",
			setupFunc: func(t *testing.T, tempDir string) {
				cfg := &config.Config{
					Defaults: config.DefaultConfig{
						Database: "postgres",
						Cache:    "redis",
					},
				}
				if err := cfg.Save(); err != nil {
					t.Fatalf("Failed to save config: %v", err)
				}
			},
			wantErr: false,
			validate: func(t *testing.T, tempDir string) {
				configPath := filepath.Join(tempDir, ".vel", "config.yaml")
				if _, err := os.Stat(configPath); !os.IsNotExist(err) {
					t.Errorf("Config file still exists after reset")
				}
			},
		},
		{
			name:      "succeeds when config file does not exist",
			setupFunc: func(t *testing.T, tempDir string) {},
			wantErr:   false,
			validate: func(t *testing.T, tempDir string) {
				configPath := filepath.Join(tempDir, ".vel", "config.yaml")
				if _, err := os.Stat(configPath); !os.IsNotExist(err) {
					t.Errorf("Unexpected config file exists")
				}
			},
		},
		{
			name: "deletes config with all fields set",
			setupFunc: func(t *testing.T, tempDir string) {
				cfg := &config.Config{
					Defaults: config.DefaultConfig{
						Database: "postgres",
						Cache:    "redis",
						Queue:    "database",
						Auth:     true,
						API:      true,
					},
				}
				if err := cfg.Save(); err != nil {
					t.Fatalf("Failed to save config: %v", err)
				}
			},
			wantErr: false,
			validate: func(t *testing.T, tempDir string) {
				configPath := filepath.Join(tempDir, ".vel", "config.yaml")
				if _, err := os.Stat(configPath); !os.IsNotExist(err) {
					t.Errorf("Config file still exists after reset")
				}
				// Verify loading returns empty config
				cfg, err := config.Load()
				if err != nil {
					t.Fatalf("Failed to load config after reset: %v", err)
				}
				if !reflect.DeepEqual(cfg, &config.Config{}) {
					t.Errorf("Config after reset = %+v, want empty config", cfg)
				}
			},
		},
		{
			name: "can reset and create new config",
			setupFunc: func(t *testing.T, tempDir string) {
				cfg := &config.Config{
					Defaults: config.DefaultConfig{
						Database: "postgres",
					},
				}
				if err := cfg.Save(); err != nil {
					t.Fatalf("Failed to save config: %v", err)
				}
			},
			wantErr: false,
			validate: func(t *testing.T, tempDir string) {
				// After reset, create new config
				newCfg := &config.Config{
					Defaults: config.DefaultConfig{
						Database: "sqlite",
					},
				}
				if err := newCfg.Save(); err != nil {
					t.Fatalf("Failed to save new config: %v", err)
				}
				loaded, err := config.Load()
				if err != nil {
					t.Fatalf("Failed to load new config: %v", err)
				}
				if loaded.Defaults.Database != "sqlite" {
					t.Errorf("New config Database = %s, want sqlite", loaded.Defaults.Database)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory for this test
			tempDir := t.TempDir()
			os.Setenv("HOME", tempDir)

			// Redirect stdout to capture output
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Run setup if provided
			if tt.setupFunc != nil {
				tt.setupFunc(t, tempDir)
			}

			// Create a mock command
			cmd := &cobra.Command{}

			// Execute function
			err := runConfigReset(cmd, []string{})

			// Restore stdout and read captured output
			w.Close()
			os.Stdout = originalStdout
			var buf bytes.Buffer
			io.Copy(&buf, r)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("runConfigReset() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Run validation if provided
			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, tempDir)
			}
		})
	}
}

func TestRunConfigSetRoundTrip(t *testing.T) {
	// Save original HOME and stdout, restore after tests
	originalHome := os.Getenv("HOME")
	originalStdout := os.Stdout
	defer func() {
		os.Setenv("HOME", originalHome)
		os.Stdout = originalStdout
	}()

	tests := []struct {
		name   string
		sets   [][]string
		verify func(t *testing.T, cfg *config.Config)
	}{
		{
			name: "can set multiple values in sequence",
			sets: [][]string{
				{"default.database", "postgres"},
				{"default.cache", "redis"},
				{"default.queue", "database"},
			},
			verify: func(t *testing.T, cfg *config.Config) {
				if cfg.Defaults.Database != "postgres" {
					t.Errorf("Database = %s, want postgres", cfg.Defaults.Database)
				}
				if cfg.Defaults.Cache != "redis" {
					t.Errorf("Cache = %s, want redis", cfg.Defaults.Cache)
				}
				if cfg.Defaults.Queue != "database" {
					t.Errorf("Queue = %s, want database", cfg.Defaults.Queue)
				}
			},
		},
		{
			name: "can overwrite previous values",
			sets: [][]string{
				{"default.database", "postgres"},
				{"default.database", "sqlite"},
				{"default.database", "mysql"},
			},
			verify: func(t *testing.T, cfg *config.Config) {
				if cfg.Defaults.Database != "mysql" {
					t.Errorf("Database = %s, want mysql (last set value)", cfg.Defaults.Database)
				}
			},
		},
		{
			name: "can set all config options",
			sets: [][]string{
				{"default.database", "sqlite"},
				{"default.cache", "memory"},
				{"default.queue", "redis"},
				{"default.auth", "true"},
				{"default.api", "true"},
			},
			verify: func(t *testing.T, cfg *config.Config) {
				want := &config.Config{
					Defaults: config.DefaultConfig{
						Database: "sqlite",
						Cache:    "memory",
						Queue:    "redis",
						Auth:     true,
						API:      true,
					},
				}
				if !reflect.DeepEqual(cfg, want) {
					t.Errorf("Config = %+v, want %+v", cfg, want)
				}
			},
		},
		{
			name: "preserves other values when setting one",
			sets: [][]string{
				{"default.database", "postgres"},
				{"default.cache", "redis"},
				{"default.auth", "true"},
				{"default.database", "sqlite"},
			},
			verify: func(t *testing.T, cfg *config.Config) {
				if cfg.Defaults.Database != "sqlite" {
					t.Errorf("Database = %s, want sqlite", cfg.Defaults.Database)
				}
				if cfg.Defaults.Cache != "redis" {
					t.Errorf("Cache = %s, want redis (should be preserved)", cfg.Defaults.Cache)
				}
				if !cfg.Defaults.Auth {
					t.Errorf("Auth = %v, want true (should be preserved)", cfg.Defaults.Auth)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory for this test
			tempDir := t.TempDir()
			os.Setenv("HOME", tempDir)

			// Create a mock command
			cmd := &cobra.Command{}

			// Set all values
			for _, args := range tt.sets {
				// Redirect stdout to suppress output
				r, w, _ := os.Pipe()
				os.Stdout = w

				err := runConfigSet(cmd, args)

				// Restore stdout and discard output
				w.Close()
				os.Stdout = originalStdout
				io.Copy(io.Discard, r)

				if err != nil {
					t.Fatalf("runConfigSet(%v) error = %v", args, err)
				}
			}

			// Load and verify final config
			cfg, err := config.Load()
			if err != nil {
				t.Fatalf("Failed to load config: %v", err)
			}

			tt.verify(t, cfg)
		})
	}
}

func TestRunConfigResetAndSet(t *testing.T) {
	// Save original HOME and stdout, restore after tests
	originalHome := os.Getenv("HOME")
	originalStdout := os.Stdout
	defer func() {
		os.Setenv("HOME", originalHome)
		os.Stdout = originalStdout
	}()

	// Create temp directory for this test
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)

	// Create a mock command
	cmd := &cobra.Command{}

	// Set initial config
	r, w, _ := os.Pipe()
	os.Stdout = w
	err := runConfigSet(cmd, []string{"default.database", "postgres"})
	w.Close()
	os.Stdout = originalStdout
	io.Copy(io.Discard, r)
	if err != nil {
		t.Fatalf("Initial runConfigSet() error = %v", err)
	}

	// Verify initial config
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load initial config: %v", err)
	}
	if cfg.Defaults.Database != "postgres" {
		t.Fatalf("Initial database = %s, want postgres", cfg.Defaults.Database)
	}

	// Reset config
	r, w, _ = os.Pipe()
	os.Stdout = w
	err = runConfigReset(cmd, []string{})
	w.Close()
	os.Stdout = originalStdout
	io.Copy(io.Discard, r)
	if err != nil {
		t.Fatalf("runConfigReset() error = %v", err)
	}

	// Verify config is reset
	cfg, err = config.Load()
	if err != nil {
		t.Fatalf("Failed to load config after reset: %v", err)
	}
	if !reflect.DeepEqual(cfg, &config.Config{}) {
		t.Fatalf("Config after reset = %+v, want empty", cfg)
	}

	// Set new config
	r, w, _ = os.Pipe()
	os.Stdout = w
	err = runConfigSet(cmd, []string{"default.database", "sqlite"})
	w.Close()
	os.Stdout = originalStdout
	io.Copy(io.Discard, r)
	if err != nil {
		t.Fatalf("Final runConfigSet() error = %v", err)
	}

	// Verify new config
	cfg, err = config.Load()
	if err != nil {
		t.Fatalf("Failed to load final config: %v", err)
	}
	if cfg.Defaults.Database != "sqlite" {
		t.Errorf("Final database = %s, want sqlite", cfg.Defaults.Database)
	}
}

func TestRunConfigGetAfterSet(t *testing.T) {
	// Save original HOME and stdout, restore after tests
	originalHome := os.Getenv("HOME")
	originalStdout := os.Stdout
	defer func() {
		os.Setenv("HOME", originalHome)
		os.Stdout = originalStdout
	}()

	tests := []struct {
		name       string
		setArgs    []string
		getArg     string
		wantOutput string
	}{
		{
			name:       "get returns value that was set",
			setArgs:    []string{"default.database", "postgres"},
			getArg:     "default.database",
			wantOutput: "postgres",
		},
		{
			name:       "get returns last value when overwritten",
			setArgs:    []string{"default.cache", "redis"},
			getArg:     "default.cache",
			wantOutput: "redis",
		},
		{
			name:       "get returns true for boolean",
			setArgs:    []string{"default.auth", "true"},
			getArg:     "default.auth",
			wantOutput: "true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory for this test
			tempDir := t.TempDir()
			os.Setenv("HOME", tempDir)

			// Create a mock command
			cmd := &cobra.Command{}

			// Set value
			r, w, _ := os.Pipe()
			os.Stdout = w
			err := runConfigSet(cmd, tt.setArgs)
			w.Close()
			os.Stdout = originalStdout
			io.Copy(io.Discard, r)
			if err != nil {
				t.Fatalf("runConfigSet() error = %v", err)
			}

			// Get value
			r, w, _ = os.Pipe()
			os.Stdout = w
			err = runConfigGet(cmd, []string{tt.getArg})
			w.Close()
			os.Stdout = originalStdout
			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			if err != nil {
				t.Fatalf("runConfigGet() error = %v", err)
			}

			if !strings.Contains(output, tt.wantOutput) {
				t.Errorf("runConfigGet() output = %q, want to contain %q", output, tt.wantOutput)
			}
		})
	}
}

func TestRunConfigLoadError(t *testing.T) {
	// Save original HOME and stdout, restore after tests
	originalHome := os.Getenv("HOME")
	originalStdout := os.Stdout
	defer func() {
		os.Setenv("HOME", originalHome)
		os.Stdout = originalStdout
	}()

	tests := []struct {
		name        string
		setupFunc   func(t *testing.T, tempDir string)
		runFunc     func(cmd *cobra.Command) error
		wantErr     bool
		errContains string
	}{
		{
			name: "runConfigSet returns error when config file is invalid yaml",
			setupFunc: func(t *testing.T, tempDir string) {
				configDir := filepath.Join(tempDir, ".vel")
				if err := os.MkdirAll(configDir, 0755); err != nil {
					t.Fatalf("Failed to create config dir: %v", err)
				}
				invalidYAML := []byte("invalid: yaml: content: [")
				configPath := filepath.Join(configDir, "config.yaml")
				if err := os.WriteFile(configPath, invalidYAML, 0600); err != nil {
					t.Fatalf("Failed to write config file: %v", err)
				}
			},
			runFunc: func(cmd *cobra.Command) error {
				return runConfigSet(cmd, []string{"default.database", "postgres"})
			},
			wantErr:     true,
			errContains: "failed to load config",
		},
		{
			name: "runConfigGet returns error when config file is invalid yaml",
			setupFunc: func(t *testing.T, tempDir string) {
				configDir := filepath.Join(tempDir, ".vel")
				if err := os.MkdirAll(configDir, 0755); err != nil {
					t.Fatalf("Failed to create config dir: %v", err)
				}
				invalidYAML := []byte("invalid: yaml: content: [")
				configPath := filepath.Join(configDir, "config.yaml")
				if err := os.WriteFile(configPath, invalidYAML, 0600); err != nil {
					t.Fatalf("Failed to write config file: %v", err)
				}
			},
			runFunc: func(cmd *cobra.Command) error {
				return runConfigGet(cmd, []string{"default.database"})
			},
			wantErr:     true,
			errContains: "invalid config file",
		},
		{
			name: "runConfigList returns error when config file is invalid yaml",
			setupFunc: func(t *testing.T, tempDir string) {
				configDir := filepath.Join(tempDir, ".vel")
				if err := os.MkdirAll(configDir, 0755); err != nil {
					t.Fatalf("Failed to create config dir: %v", err)
				}
				invalidYAML := []byte("invalid: yaml: content: [")
				configPath := filepath.Join(configDir, "config.yaml")
				if err := os.WriteFile(configPath, invalidYAML, 0600); err != nil {
					t.Fatalf("Failed to write config file: %v", err)
				}
			},
			runFunc: func(cmd *cobra.Command) error {
				return runConfigList(cmd, []string{})
			},
			wantErr:     true,
			errContains: "invalid config file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory for this test
			tempDir := t.TempDir()
			os.Setenv("HOME", tempDir)

			// Redirect stdout to suppress output
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Run setup if provided
			if tt.setupFunc != nil {
				tt.setupFunc(t, tempDir)
			}

			// Create a mock command
			cmd := &cobra.Command{}

			// Execute function
			err := tt.runFunc(cmd)

			// Restore stdout and discard output
			w.Close()
			os.Stdout = originalStdout
			io.Copy(io.Discard, r)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" {
				if err == nil || !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error = %v, want error containing %q", err, tt.errContains)
				}
			}
		})
	}
}

func TestRunConfigSetSaveError(t *testing.T) {
	// Save original HOME and stdout, restore after tests
	originalHome := os.Getenv("HOME")
	originalStdout := os.Stdout
	defer func() {
		os.Setenv("HOME", originalHome)
		os.Stdout = originalStdout
	}()

	// Create temp directory for this test
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)

	// Create config directory with restrictive permissions
	configDir := filepath.Join(tempDir, ".vel")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	// Create config file with read-only permissions
	configPath := filepath.Join(configDir, "config.yaml")
	cfg := &config.Config{
		Defaults: config.DefaultConfig{
			Database: "postgres",
		},
	}
	data, _ := yaml.Marshal(cfg)
	if err := os.WriteFile(configPath, data, 0400); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Redirect stdout to suppress output
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Create a mock command
	cmd := &cobra.Command{}

	// Try to set config (should fail on save)
	err := runConfigSet(cmd, []string{"default.database", "sqlite"})

	// Restore stdout and discard output
	w.Close()
	os.Stdout = originalStdout
	io.Copy(io.Discard, r)

	// Check that we got a save error
	if err == nil {
		t.Error("runConfigSet() should return error when save fails, got nil")
	} else if !strings.Contains(err.Error(), "failed to save config") {
		t.Errorf("runConfigSet() error = %v, want error containing 'failed to save config'", err)
	}
}
