package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestValidateDatabase(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"empty string returns nil", "", false},
		{"postgres is valid", "postgres", false},
		{"mysql is valid", "mysql", false},
		{"sqlite is valid", "sqlite", false},
		{"returns error when value is invalid", "mongodb", true},
		{"returns error for uppercase valid value", "Postgres", true},
		{"returns error for uppercase invalid value", "POSTGRES", true},
		{"returns error for value with whitespace", " postgres", true},
		{"returns error for value with trailing space", "postgres ", true},
		{"returns error for mixed case", "PostgreSQL", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDatabase(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDatabase(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateCache(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"empty string returns nil", "", false},
		{"redis is valid", "redis", false},
		{"memory is valid", "memory", false},
		{"returns error when value is invalid", "memcached", true},
		{"returns error for uppercase valid value", "Redis", true},
		{"returns error for uppercase invalid value", "REDIS", true},
		{"returns error for value with whitespace", " redis", true},
		{"returns error for value with trailing space", "redis ", true},
		{"returns error for mixed case", "Memory", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCache(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCache(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateQueue(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"empty string returns nil", "", false},
		{"redis is valid", "redis", false},
		{"database is valid", "database", false},
		{"returns error when value is invalid", "rabbitmq", true},
		{"returns error for uppercase valid value", "Redis", true},
		{"returns error for uppercase invalid value", "REDIS", true},
		{"returns error for value with whitespace", " redis", true},
		{"returns error for value with trailing space", "redis ", true},
		{"returns error for mixed case", "Database", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateQueue(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateQueue(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestConfigDir(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"returns path to .vel directory", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConfigDir()
			if (err != nil) != tt.wantErr {
				t.Errorf("ConfigDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				home, _ := os.UserHomeDir()
				want := filepath.Join(home, ".vel")
				if got != want {
					t.Errorf("ConfigDir() = %v, want %v", got, want)
				}
			}
		})
	}
}

func TestConfigPath(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"returns path to config.yaml file", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConfigPath()
			if (err != nil) != tt.wantErr {
				t.Errorf("ConfigPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				home, _ := os.UserHomeDir()
				want := filepath.Join(home, ".vel", "config.yaml")
				if got != want {
					t.Errorf("ConfigPath() = %v, want %v", got, want)
				}
			}
		})
	}
}

func TestLoad(t *testing.T) {
	// Save original HOME and restore after tests
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	tests := []struct {
		name       string
		setupFunc  func(t *testing.T, tempDir string)
		want       *Config
		wantErr    bool
		errContains string
	}{
		{
			name: "returns empty config when file does not exist",
			setupFunc: func(t *testing.T, tempDir string) {
				// No setup needed - file should not exist
			},
			want:    &Config{},
			wantErr: false,
		},
		{
			name: "loads valid config file",
			setupFunc: func(t *testing.T, tempDir string) {
				configDir := filepath.Join(tempDir, ".vel")
				if err := os.MkdirAll(configDir, 0755); err != nil {
					t.Fatalf("Failed to create config dir: %v", err)
				}
				cfg := &Config{
					Defaults: DefaultConfig{
						Database: "postgres",
						Cache:    "redis",
						Queue:    "database",
						Auth:     true,
						API:      true,
					},
				}
				data, err := yaml.Marshal(cfg)
				if err != nil {
					t.Fatalf("Failed to marshal config: %v", err)
				}
				if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), data, 0600); err != nil {
					t.Fatalf("Failed to write config file: %v", err)
				}
			},
			want: &Config{
				Defaults: DefaultConfig{
					Database: "postgres",
					Cache:    "redis",
					Queue:    "database",
					Auth:     true,
					API:      true,
				},
			},
			wantErr: false,
		},
		{
			name: "loads config with partial defaults",
			setupFunc: func(t *testing.T, tempDir string) {
				configDir := filepath.Join(tempDir, ".vel")
				if err := os.MkdirAll(configDir, 0755); err != nil {
					t.Fatalf("Failed to create config dir: %v", err)
				}
				cfg := &Config{
					Defaults: DefaultConfig{
						Database: "sqlite",
					},
				}
				data, err := yaml.Marshal(cfg)
				if err != nil {
					t.Fatalf("Failed to marshal config: %v", err)
				}
				if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), data, 0600); err != nil {
					t.Fatalf("Failed to write config file: %v", err)
				}
			},
			want: &Config{
				Defaults: DefaultConfig{
					Database: "sqlite",
				},
			},
			wantErr: false,
		},
		{
			name: "loads empty config file",
			setupFunc: func(t *testing.T, tempDir string) {
				configDir := filepath.Join(tempDir, ".vel")
				if err := os.MkdirAll(configDir, 0755); err != nil {
					t.Fatalf("Failed to create config dir: %v", err)
				}
				if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte{}, 0600); err != nil {
					t.Fatalf("Failed to write config file: %v", err)
				}
			},
			want:    &Config{},
			wantErr: false,
		},
		{
			name: "returns error when yaml is invalid",
			setupFunc: func(t *testing.T, tempDir string) {
				configDir := filepath.Join(tempDir, ".vel")
				if err := os.MkdirAll(configDir, 0755); err != nil {
					t.Fatalf("Failed to create config dir: %v", err)
				}
				invalidYAML := []byte("invalid: yaml: content: [")
				if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), invalidYAML, 0600); err != nil {
					t.Fatalf("Failed to write config file: %v", err)
				}
			},
			want:        nil,
			wantErr:     true,
			errContains: "invalid config file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory for this test
			tempDir := t.TempDir()
			os.Setenv("HOME", tempDir)

			// Run setup if provided
			if tt.setupFunc != nil {
				tt.setupFunc(t, tempDir)
			}

			got, err := Load()
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" {
				if err == nil || !contains(err.Error(), tt.errContains) {
					t.Errorf("Load() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Load() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestConfig_Save(t *testing.T) {
	// Save original HOME and restore after tests
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	tests := []struct {
		name       string
		config     *Config
		setupFunc  func(t *testing.T, tempDir string)
		wantErr    bool
		errContains string
	}{
		{
			name: "creates directory and saves config",
			config: &Config{
				Defaults: DefaultConfig{
					Database: "postgres",
					Cache:    "redis",
					Queue:    "database",
					Auth:     true,
					API:      false,
				},
			},
			setupFunc: func(t *testing.T, tempDir string) {
				// No setup needed - directory should be created
			},
			wantErr: false,
		},
		{
			name: "saves config when directory exists",
			config: &Config{
				Defaults: DefaultConfig{
					Database: "sqlite",
					Cache:    "memory",
				},
			},
			setupFunc: func(t *testing.T, tempDir string) {
				configDir := filepath.Join(tempDir, ".vel")
				if err := os.MkdirAll(configDir, 0755); err != nil {
					t.Fatalf("Failed to create config dir: %v", err)
				}
			},
			wantErr: false,
		},
		{
			name:   "saves empty config",
			config: &Config{},
			setupFunc: func(t *testing.T, tempDir string) {
				// No setup needed
			},
			wantErr: false,
		},
		{
			name: "saves config with partial defaults",
			config: &Config{
				Defaults: DefaultConfig{
					Auth: true,
				},
			},
			setupFunc: func(t *testing.T, tempDir string) {
				// No setup needed
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory for this test
			tempDir := t.TempDir()
			os.Setenv("HOME", tempDir)

			// Run setup if provided
			if tt.setupFunc != nil {
				tt.setupFunc(t, tempDir)
			}

			err := tt.config.Save()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Save() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" {
				if err == nil || !contains(err.Error(), tt.errContains) {
					t.Errorf("Config.Save() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if !tt.wantErr {
				// Verify file was created
				configPath := filepath.Join(tempDir, ".vel", "config.yaml")
				if _, err := os.Stat(configPath); os.IsNotExist(err) {
					t.Errorf("Config.Save() did not create config file at %s", configPath)
					return
				}

				// Verify directory permissions
				configDir := filepath.Join(tempDir, ".vel")
				info, err := os.Stat(configDir)
				if err != nil {
					t.Errorf("Config.Save() failed to stat directory: %v", err)
					return
				}
				if info.Mode().Perm() != 0755 {
					t.Errorf("Config.Save() directory permissions = %o, want 0755", info.Mode().Perm())
				}

				// Verify file permissions
				fileInfo, err := os.Stat(configPath)
				if err != nil {
					t.Errorf("Config.Save() failed to stat file: %v", err)
					return
				}
				if fileInfo.Mode().Perm() != 0600 {
					t.Errorf("Config.Save() file permissions = %o, want 0600", fileInfo.Mode().Perm())
				}

				// Verify content can be read back
				data, err := os.ReadFile(configPath)
				if err != nil {
					t.Errorf("Config.Save() failed to read saved file: %v", err)
					return
				}

				var loaded Config
				if err := yaml.Unmarshal(data, &loaded); err != nil {
					t.Errorf("Config.Save() saved invalid YAML: %v", err)
					return
				}

				if !reflect.DeepEqual(&loaded, tt.config) {
					t.Errorf("Config.Save() saved config = %+v, want %+v", &loaded, tt.config)
				}
			}
		})
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	// Save original HOME and restore after tests
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	tests := []struct {
		name   string
		config *Config
	}{
		{
			name: "round trip with full config",
			config: &Config{
				Defaults: DefaultConfig{
					Database: "postgres",
					Cache:    "redis",
					Queue:    "database",
					Auth:     true,
					API:      true,
				},
			},
		},
		{
			name: "round trip with partial config",
			config: &Config{
				Defaults: DefaultConfig{
					Database: "sqlite",
				},
			},
		},
		{
			name:   "round trip with empty config",
			config: &Config{},
		},
		{
			name: "round trip with boolean values",
			config: &Config{
				Defaults: DefaultConfig{
					Auth: false,
					API:  true,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory for this test
			tempDir := t.TempDir()
			os.Setenv("HOME", tempDir)

			// Save config
			if err := tt.config.Save(); err != nil {
				t.Fatalf("Save() error = %v", err)
			}

			// Load config
			loaded, err := Load()
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			// Compare
			if !reflect.DeepEqual(loaded, tt.config) {
				t.Errorf("round trip: got %+v, want %+v", loaded, tt.config)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && stringContains(s, substr)))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
