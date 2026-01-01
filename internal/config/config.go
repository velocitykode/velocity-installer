package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gofrs/flock"
	"gopkg.in/yaml.v3"
)

// Config represents global CLI configuration
type Config struct {
	Defaults DefaultConfig `yaml:"defaults"`
}

// DefaultConfig holds default values for project creation
type DefaultConfig struct {
	Database string `yaml:"database,omitempty"`
	Cache    string `yaml:"cache,omitempty"`
	Queue    string `yaml:"queue,omitempty"`
	Auth     bool   `yaml:"auth,omitempty"`
	API      bool   `yaml:"api,omitempty"`
}

var validDatabases = []string{"postgres", "mysql", "sqlite"}
var validCaches = []string{"redis", "memory"}
var validQueues = []string{"redis", "database"}

// ConfigDir returns the path to the .velocity directory
func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".velocity"), nil
}

// ConfigPath returns the path to the config.yaml file
func ConfigPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

// Load reads the configuration from disk
func Load() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	// Return empty config if file doesn't exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &Config{}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid config file: %w", err)
	}

	return &cfg, nil
}

// Save writes the configuration to disk with file locking
func (c *Config) Save() error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Acquire lock
	lockPath := filepath.Join(dir, ".config.lock")
	lock := flock.New(lockPath)
	if err := lock.Lock(); err != nil {
		return fmt.Errorf("config locked by another process: %w", err)
	}
	defer lock.Unlock()

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	// Write with user-only permissions
	return os.WriteFile(path, data, 0600)
}

// ValidateDatabase validates database driver value
func ValidateDatabase(db string) error {
	if db == "" {
		return nil
	}
	for _, valid := range validDatabases {
		if db == valid {
			return nil
		}
	}
	return fmt.Errorf("invalid database: %s (must be: postgres, mysql, sqlite)", db)
}

// ValidateCache validates cache driver value
func ValidateCache(cache string) error {
	if cache == "" {
		return nil
	}
	for _, valid := range validCaches {
		if cache == valid {
			return nil
		}
	}
	return fmt.Errorf("invalid cache: %s (must be: redis, memory)", cache)
}

// ValidateQueue validates queue driver value
func ValidateQueue(queue string) error {
	if queue == "" {
		return nil
	}
	for _, valid := range validQueues {
		if queue == valid {
			return nil
		}
	}
	return fmt.Errorf("invalid queue: %s (must be: redis, database)", queue)
}
