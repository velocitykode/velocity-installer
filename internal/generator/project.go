package generator

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/velocitykode/velocity-installer/internal/ui"
)

// Fallback version if GitHub API is unavailable
const fallbackVelocityVersion = "v0.6.2"

// getLatestVelocityVersion fetches the latest release tag from GitHub
func getLatestVelocityVersion() string {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("https://api.github.com/repos/velocitykode/velocity/releases/latest")
	if err != nil {
		return fallbackVelocityVersion
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fallbackVelocityVersion
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return fallbackVelocityVersion
	}

	if release.TagName == "" {
		return fallbackVelocityVersion
	}

	return release.TagName
}

// ProjectConfig holds the configuration for a new project
type ProjectConfig struct {
	Name     string
	Module   string
	Database string
	Cache    string
	Auth     bool
	API      bool
}

// CreateProject generates a new Velocity project from template
func CreateProject(config ProjectConfig) error {
	// Validate project name
	if err := validateProjectName(config.Name); err != nil {
		return err
	}

	// Determine module name
	moduleName := config.Module
	if moduleName == "" {
		moduleName = config.Name
	}

	ui.Info("Creating new Velocity project")

	// Clone template
	if err := ui.Spinner("Cloning template", func() error {
		return cloneTemplate(config.Name, config.API)
	}); err != nil {
		return fmt.Errorf("failed to clone template: %w", err)
	}
	ui.Success("Template cloned")

	// Replace module name in all files
	if err := ui.Spinner("Configuring module", func() error {
		return replaceModuleName(config.Name, moduleName)
	}); err != nil {
		return fmt.Errorf("failed to configure project: %w", err)
	}
	ui.Success("Module configured")

	// Remove template git history and initialize new repo
	if err := ui.Spinner("Initializing Git", func() error {
		return reinitGitRepo(config.Name)
	}); err != nil {
		return fmt.Errorf("failed to initialize git: %w", err)
	}
	ui.Success("Git initialized")

	// Create default migrations
	if err := createDefaultMigrations(config.Name); err != nil {
		return fmt.Errorf("failed to create migrations: %w", err)
	}
	ui.Success("Migrations created")

	// Create proper .env.example with database config
	if err := createEnvFiles(config); err != nil {
		return fmt.Errorf("failed to create env files: %w", err)
	}
	ui.Success("Environment configured")

	// Setup hot reload
	if err := setupTemplatesAndHotReload(config.Name); err != nil {
		return fmt.Errorf("failed to setup templates: %w", err)
	}
	ui.Success("Hot reload configured")

	// Install dependencies
	ui.Info("Installing dependencies")
	if err := installDependencies(config.Name, config.API); err != nil {
		return fmt.Errorf("failed to install dependencies: %w", err)
	}

	// Run migrations
	ui.Newline()
	ui.Info("Running migrations...")
	if err := runMigrations(config.Name); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	ui.Success("Database ready")

	return nil
}

// cloneTemplate clones the appropriate velocity template
func cloneTemplate(projectName string, apiOnly bool) error {
	templateRepo := "velocity-template"
	if apiOnly {
		templateRepo = "velocity-template-api"
	}

	// Use git clone directly (gh repo clone can use stale cache)
	cmd := exec.Command("git", "clone", "--depth=1", "git@github.com:velocitykode/"+templateRepo+".git", projectName)
	if err := cmd.Run(); err != nil {
		// Try HTTPS fallback
		cmd = exec.Command("git", "clone", "--depth=1", "https://github.com/velocitykode/"+templateRepo+".git", projectName)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to clone template: %w", err)
		}
	}
	return nil
}

// replaceModuleName replaces {{MODULE_NAME}} in all files
func replaceModuleName(projectPath, moduleName string) error {
	// Get absolute path
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return fmt.Errorf("abs path: %w", err)
	}

	// Use find and sed to replace in all Go files
	cmd := exec.Command("sh", "-c",
		fmt.Sprintf("cd '%s' && find . -name '*.go' -type f -exec sed -i '' 's|{{MODULE_NAME}}|%s|g' {} +", absPath, moduleName))
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("replace go files: %w: %s", err, string(output))
	}

	// Replace in go.mod (if exists)
	cmd = exec.Command("sh", "-c",
		fmt.Sprintf("cd '%s' && [ -f go.mod ] && sed -i '' 's|{{MODULE_NAME}}|%s|g' go.mod || true", absPath, moduleName))
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("replace go.mod: %w: %s", err, string(output))
	}

	// Replace in package.json (if exists)
	cmd = exec.Command("sh", "-c",
		fmt.Sprintf("cd '%s' && [ -f package.json ] && sed -i '' 's|{{MODULE_NAME}}|%s|g' package.json || true", absPath, moduleName))
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("replace package.json: %w: %s", err, string(output))
	}

	// Replace in package-lock.json (if exists)
	cmd = exec.Command("sh", "-c",
		fmt.Sprintf("cd '%s' && [ -f package-lock.json ] && sed -i '' 's|{{MODULE_NAME}}|%s|g' package-lock.json || true", absPath, moduleName))
	cmd.Run() // Ignore error - file may not exist

	// Only process go.mod if it exists
	goModPath := filepath.Join(absPath, "go.mod")
	if _, err := os.Stat(goModPath); err == nil {
		// Remove replace directive
		cmd = exec.Command("sh", "-c",
			fmt.Sprintf("cd '%s' && sed -i '' '/^replace github.com\\/velocitykode\\/velocity/d' go.mod", absPath))
		if err := cmd.Run(); err != nil {
			return err
		}

		// Set pinned version of velocity framework (fetched from GitHub releases)
		velocityVersion := getLatestVelocityVersion()
		cmd = exec.Command("go", "mod", "edit", fmt.Sprintf("-require=github.com/velocitykode/velocity@%s", velocityVersion))
		cmd.Dir = absPath
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to set velocity framework version: %w", err)
		}
	}

	return nil
}

// reinitGitRepo removes template git history and creates new repo
func reinitGitRepo(projectPath string) error {
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return err
	}

	// Verify project directory exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("project directory does not exist: %s", absPath)
	}

	// Remove .git directory
	gitDir := filepath.Join(absPath, ".git")
	if err := os.RemoveAll(gitDir); err != nil {
		return err
	}

	// Initialize new git repo
	originalDir, _ := os.Getwd()
	if err := os.Chdir(absPath); err != nil {
		return fmt.Errorf("failed to change to project directory: %w", err)
	}
	defer os.Chdir(originalDir)

	exec.Command("git", "init").Run()
	exec.Command("git", "add", ".").Run()
	exec.Command("git", "commit", "-m", "Initial commit").Run()

	return nil
}

// installDependencies runs go mod tidy and bun install in parallel
func installDependencies(projectPath string, apiOnly bool) error {
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return err
	}

	originalDir, _ := os.Getwd()
	os.Chdir(absPath)
	defer os.Chdir(originalDir)

	// Track status
	type depStatus struct {
		name   string
		status string
		done   bool
		err    error
	}

	goStatus := &depStatus{name: "Go dependencies", status: "downloading..."}
	jsStatus := &depStatus{name: "JS dependencies", status: "downloading..."}
	airStatus := &depStatus{name: "Air (hot reload)", status: "checking..."}

	// Check if air is already installed
	airInstalled := isAirInstalled()

	// Print initial tree (skip JS for API-only projects)
	printDepTree := func() {
		if apiOnly {
			ui.TreeItem("├─", goStatus.name, goStatus.status, goStatus.done)
			if airInstalled {
				ui.TreeItemSkipped("└─", airStatus.name, "already installed")
			} else {
				ui.TreeItem("└─", airStatus.name, airStatus.status, airStatus.done)
			}
		} else {
			ui.TreeItem("├─", goStatus.name, goStatus.status, goStatus.done)
			ui.TreeItem("├─", jsStatus.name, jsStatus.status, jsStatus.done)
			if airInstalled {
				ui.TreeItemSkipped("└─", airStatus.name, "already installed")
			} else {
				ui.TreeItem("└─", airStatus.name, airStatus.status, airStatus.done)
			}
		}
	}

	printDepTree()

	// Determine number of tasks
	numTasks := 3
	linesToClear := 3
	if apiOnly {
		numTasks = 2
		linesToClear = 2
	}

	// Run deps in parallel
	done := make(chan bool, numTasks)

	// Go dependencies
	go func() {
		goStatus.err = exec.Command("go", "mod", "tidy").Run()
		if goStatus.err == nil {
			goStatus.status = "done"
			goStatus.done = true
		} else {
			goStatus.status = "failed"
		}
		done <- true
	}()

	// JS dependencies (skip for API-only projects)
	if !apiOnly {
		go func() {
			if err := exec.Command("bun", "install").Run(); err != nil {
				// Try npm as fallback
				jsStatus.err = exec.Command("npm", "install").Run()
			}
			if jsStatus.err == nil {
				jsStatus.status = "done"
				jsStatus.done = true
			} else {
				jsStatus.status = "failed"
			}
			done <- true
		}()
	}

	// Air installation (only if not already installed)
	go func() {
		if !airInstalled {
			exec.Command("go", "install", "github.com/air-verse/air@latest").Run()
			airStatus.status = "done"
			airStatus.done = true
		}
		done <- true
	}()

	// Wait for all to complete, updating display
	completed := 0
	for completed < numTasks {
		<-done
		completed++
		// Clear and redraw tree
		ui.ClearLines(linesToClear)
		printDepTree()
	}

	// Return first error encountered
	if goStatus.err != nil {
		return goStatus.err
	}
	if !apiOnly && jsStatus.err != nil {
		return jsStatus.err
	}

	return nil
}

// isAirInstalled checks if air binary is available
func isAirInstalled() bool {
	_, err := exec.LookPath("air")
	return err == nil
}

func validateProjectName(name string) error {
	if name == "" {
		return fmt.Errorf("project name cannot be empty")
	}

	// Check for invalid characters
	if strings.ContainsAny(name, " !@#$%^&*()+=[]{}|\\;:'\",<>?/") {
		return fmt.Errorf("project name contains invalid characters")
	}

	// Check if directory already exists
	if _, err := os.Stat(name); err == nil {
		return fmt.Errorf("directory %s already exists", name)
	}

	return nil
}

func createDirectoryStructure(projectPath string) error {
	directories := []string{
		"app/http/controllers",
		"app/http/middleware",
		"app/models",
		"bootstrap",
		"cmd/vel",
		"config",
		"database/migrations",
		"database/factories",
		"public",
		"resources/views",
		"routes",
		"storage/logs",
		"tests",
	}

	for _, dir := range directories {
		path := filepath.Join(projectPath, dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			return err
		}
	}

	return nil
}

func initGoModule(config ProjectConfig) error {
	// Verify project directory exists
	if _, err := os.Stat(config.Name); os.IsNotExist(err) {
		return fmt.Errorf("project directory does not exist: %s", config.Name)
	}

	// Change to project directory
	originalDir, _ := os.Getwd()
	if err := os.Chdir(config.Name); err != nil {
		return fmt.Errorf("failed to change to project directory: %w", err)
	}
	defer os.Chdir(originalDir)

	// Initialize go module
	moduleName := config.Module
	if moduleName == "" {
		moduleName = config.Name
	}

	cmd := exec.Command("go", "mod", "init", moduleName)
	if err := cmd.Run(); err != nil {
		return err
	}

	ui.Step("Configuring dependencies...")

	// Check if local Velocity exists and use replace directive
	velocityPath := "/Users/ali/code/velocity"
	if _, err := os.Stat(velocityPath); err == nil {
		// Add replace directive for local development
		cmd = exec.Command("go", "mod", "edit", "-replace", "github.com/velocitykode/velocity="+velocityPath)
		cmd.Run()
		ui.Info("Using local Velocity framework")
	} else {
		// Try to get from GitHub (requires GOPRIVATE setup for private repos)
		cmd = exec.Command("go", "get", "github.com/velocitykode/velocity@v0.6.2")
		if err := cmd.Run(); err != nil {
			ui.Warning("Note: Configure GOPRIVATE for private repo access")
		}
	}

	// Add other dependencies based on features
	if config.Database == "postgres" {
		ui.Info("PostgreSQL driver")
		exec.Command("go", "get", "github.com/lib/pq").Run()
	} else if config.Database == "mysql" {
		ui.Info("MySQL driver")
		exec.Command("go", "get", "github.com/go-sql-driver/mysql").Run()
	} else if config.Database == "sqlite" {
		ui.Info("SQLite driver")
		exec.Command("go", "get", "github.com/mattn/go-sqlite3").Run()
	}

	if config.Cache == "redis" {
		ui.Info("Redis client")
		exec.Command("go", "get", "github.com/redis/go-redis/v9").Run()
	}

	// Run go mod tidy
	ui.Step("Tidying up dependencies...")
	exec.Command("go", "mod", "tidy").Run()

	return nil
}

func initGitRepo(projectPath string) {
	originalDir, _ := os.Getwd()
	os.Chdir(projectPath)
	defer os.Chdir(originalDir)

	exec.Command("git", "init").Run()
	exec.Command("git", "add", ".").Run()
}

// InitProject adds Velocity structure to an existing Go project
func InitProject(config ProjectConfig, targetDir string) error {
	ui.Step("Setting up Velocity structure...")
	// Create directory structure in existing directory
	if err := createDirectoryStructure(targetDir); err != nil {
		return fmt.Errorf("failed to create directory structure: %w", err)
	}
	ui.Success("Velocity structure created")

	ui.Step("Generating application files...")
	// Generate files from stubs (skip if exists to preserve existing code)
	if err := generateFilesFromStubs(config); err != nil {
		return fmt.Errorf("failed to generate files: %w", err)
	}
	ui.Success("Application files generated")

	ui.Step("Creating configuration files...")
	// Generate config files if they don't exist
	if err := generateProjectFiles(config); err != nil {
		return fmt.Errorf("failed to generate project files: %w", err)
	}
	ui.Success("Configuration files created")

	ui.Step("Adding Velocity dependencies...")
	// Add dependencies to existing go.mod
	if err := addVelocityDependencies(config, targetDir); err != nil {
		return fmt.Errorf("failed to add dependencies: %w", err)
	}
	ui.Success("Dependencies added")

	return nil
}

// addVelocityDependencies adds Velocity and feature dependencies to existing go.mod
// createEnvFiles copies .env.example to .env and generates a new crypto key
func createEnvFiles(config ProjectConfig) error {
	absPath, err := filepath.Abs(config.Name)
	if err != nil {
		return err
	}

	// Copy .env.example to .env
	cmd := exec.Command("cp", filepath.Join(absPath, ".env.example"), filepath.Join(absPath, ".env"))
	if err := cmd.Run(); err != nil {
		return err
	}

	// Generate new crypto key
	newKey, err := generateCryptoKey()
	if err != nil {
		return err
	}

	// Replace crypto key in .env using sed
	cmd = exec.Command("sh", "-c",
		fmt.Sprintf("cd '%s' && sed -i '' 's|^CRYPTO_KEY=.*|CRYPTO_KEY=base64:%s|' .env", absPath, newKey))
	if err := cmd.Run(); err != nil {
		return err
	}

	// Update DB settings based on config
	if config.Database != "" && config.Database != "sqlite" {
		ports := map[string]string{"postgres": "5432", "mysql": "3306"}
		username := os.Getenv("USER")

		// Use base name for database name (not full path)
		dbName := filepath.Base(config.Name)

		// Update DB_CONNECTION, DB_PORT, DB_DATABASE, DB_USERNAME
		sedCmds := fmt.Sprintf(
			"sed -i '' 's|^DB_CONNECTION=.*|DB_CONNECTION=%s|; s|^DB_PORT=.*|DB_PORT=%s|; s|^DB_DATABASE=.*|DB_DATABASE=%s|; s|^DB_USERNAME=.*|DB_USERNAME=%s|' .env",
			config.Database, ports[config.Database], dbName, username)
		cmd = exec.Command("sh", "-c", fmt.Sprintf("cd '%s' && %s", absPath, sedCmds))
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	// Update CACHE_DRIVER based on config
	if config.Cache != "" && config.Cache != "memory" {
		cmd = exec.Command("sh", "-c",
			fmt.Sprintf("cd '%s' && sed -i '' 's|^CACHE_DRIVER=.*|CACHE_DRIVER=%s|' .env", absPath, config.Cache))
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
}

// generateCryptoKey generates a new 32-byte base64 encoded key
func generateCryptoKey() (string, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(key), nil
}

// createDefaultMigrations creates the 3 default migration files
func createDefaultMigrations(projectPath string) error {
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return err
	}

	migrationsDir := filepath.Join(absPath, "database", "migrations")
	if err := os.MkdirAll(migrationsDir, 0755); err != nil {
		return err
	}

	// Migration 1: Create users table
	usersTable := `package migrations

import "github.com/velocitykode/velocity/pkg/orm/migrate"

func init() {
	migrate.Register(&migrate.Migration{
		Version:     "20010101000000",
		Description: "create users table",
		Up: func(m *migrate.Migrator) error {
			return m.CreateTable("users", func(t *migrate.TableBuilder) {
				t.ID()
				t.String("name")
				t.String("email").Unique()
				t.String("password")
				t.String("role").Default("user")
				t.Timestamps()
			})
		},
		Down: func(m *migrate.Migrator) error {
			return m.DropTable("users")
		},
	})
}
`

	// Migration 2: Create cache table
	cacheTable := `package migrations

import "github.com/velocitykode/velocity/pkg/orm/migrate"

func init() {
	migrate.Register(&migrate.Migration{
		Version:     "20010101000001",
		Description: "create cache table",
		Up: func(m *migrate.Migrator) error {
			return m.CreateTable("cache", func(t *migrate.TableBuilder) {
				t.String("key", 255).Unique()
				t.String("value", 10000)
				t.Integer("expiration")
			})
		},
		Down: func(m *migrate.Migrator) error {
			return m.DropTable("cache")
		},
	})
}
`

	// Migration 3: Create jobs table
	jobsTable := `package migrations

import "github.com/velocitykode/velocity/pkg/orm/migrate"

func init() {
	migrate.Register(&migrate.Migration{
		Version:     "20010101000002",
		Description: "create jobs table",
		Up: func(m *migrate.Migrator) error {
			if err := m.CreateTable("jobs", func(t *migrate.TableBuilder) {
				t.ID()
				t.String("queue", 255)
				t.String("payload", 10000)
				t.Integer("attempts").Default("0")
				t.String("scheduled_at", 50)
				t.String("reserved_at", 50).Nullable()
				t.String("reserved_by", 255).Nullable()
				t.String("failed_at", 50).Nullable()
				t.String("failed_reason", 5000).Nullable()
				t.Timestamps()
			}); err != nil {
				return err
			}

			return m.CreateTable("failed_jobs", func(t *migrate.TableBuilder) {
				t.ID()
				t.String("queue", 255)
				t.String("payload", 10000)
				t.String("exception", 10000)
				t.Timestamps()
			})
		},
		Down: func(m *migrate.Migrator) error {
			if err := m.DropTable("failed_jobs"); err != nil {
				return err
			}
			return m.DropTable("jobs")
		},
	})
}
`

	// Write migration files
	migrations := map[string]string{
		"0001_01_01_000000_create_users_table.go": usersTable,
		"0001_01_01_000001_create_cache_table.go": cacheTable,
		"0001_01_01_000002_create_jobs_table.go":  jobsTable,
	}

	for filename, content := range migrations {
		filePath := filepath.Join(migrationsDir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return err
		}
	}

	return nil
}

// runMigrations runs migrations directly without subprocess
func runMigrations(projectPath string) error {
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return err
	}

	originalDir, _ := os.Getwd()
	os.Chdir(absPath)
	defer os.Chdir(originalDir)

	// Create temporary migration runner script
	tmpDir := ".vel/tmp"
	os.MkdirAll(tmpDir, 0755)

	// Get module name
	moduleName, err := getProjectModuleName()
	if err != nil {
		return err
	}

	script := fmt.Sprintf(`
package main

import (
	"fmt"
	"os"

	_ "%s/database/migrations"
	"github.com/joho/godotenv"
	"github.com/velocitykode/velocity/pkg/orm"
	"github.com/velocitykode/velocity/pkg/orm/migrate"
)

const (
	// ANSI symbols (matching CLI style)
	checkSymbol = "\033[32m✓\033[0m"   // green checkmark
	warnSymbol  = "\033[33m!\033[0m"   // yellow warning
	crossSymbol = "\033[31m✗\033[0m"   // red cross
)

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Printf("%%s .env file not found\n", warnSymbol)
	}

	if err := orm.InitFromEnv(); err != nil {
		fmt.Printf("%%s Failed to initialize database: %%v\n", crossSymbol, err)
		os.Exit(1)
	}

	driver := orm.DB()
	if driver == nil {
		fmt.Printf("%%s Database driver not initialized\n", crossSymbol)
		os.Exit(1)
	}

	driverName := os.Getenv("DB_CONNECTION")
	migrator := migrate.NewMigrator(driver, driverName)

	registry := migrate.All()

	appliedVersions := make(map[string]bool)
	rows, err := driver.Query("SELECT version FROM migrations")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var version string
			if err := rows.Scan(&version); err == nil {
				appliedVersions[version] = true
			}
		}
	}

	pending := []migrate.Migration{}
	for _, m := range registry {
		if !appliedVersions[m.Version] {
			pending = append(pending, m)
		}
	}

	if len(pending) == 0 {
		os.Exit(0)
	}

	if err := migrator.Up(); err != nil {
		fmt.Printf("%%s Migration failed: %%v\n", crossSymbol, err)
		os.Exit(1)
	}

	for _, m := range pending {
		fmt.Printf("%%s \033[32;1m%%s_%%s\033[0m\n", checkSymbol, m.Version, m.Description)
	}
}
`, moduleName)

	tmpFile := fmt.Sprintf("%s/migrate_runner.go", tmpDir)
	if err := os.WriteFile(tmpFile, []byte(script), 0644); err != nil {
		return err
	}
	defer os.Remove(tmpFile)

	// Run with go run (uses module mode)
	runCmd := exec.Command("go", "run", tmpFile)
	runCmd.Stdout = os.Stdout
	runCmd.Stderr = os.Stderr

	return runCmd.Run()
}

func getProjectModuleName() (string, error) {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimPrefix(line, "module "), nil
		}
	}

	return "", fmt.Errorf("module name not found in go.mod")
}

// StartDevServers starts npm run dev and go run main.go in background
func StartDevServers(projectPath string, apiOnly bool) {
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		ui.Error(fmt.Sprintf("Failed to resolve project path: %v", err))
		return
	}

	// Start npm run dev in background (detached) - skip for API-only projects
	if !apiOnly {
		npmCmd := exec.Command("npm", "run", "dev")
		npmCmd.Dir = absPath
		npmCmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		if err := npmCmd.Start(); err != nil {
			ui.Error(fmt.Sprintf("Failed to start npm: %v", err))
			return
		}
	}

	// Start air for hot reloading (detached)
	goCmd := exec.Command("air")
	goCmd.Dir = absPath
	goCmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := goCmd.Start(); err != nil {
		ui.Error(fmt.Sprintf("Failed to start Go server: %v", err))
		return
	}

	// Show URLs
	ui.Step(fmt.Sprintf("cd %s", projectPath))
	if !apiOnly {
		ui.KeyValue("Vite", ui.Highlight("http://localhost:5173"))
	}
	ui.KeyValue("Velocity", ui.Highlight("http://localhost:4000"))

	ui.Newline()
	ui.Success("Build something great!")
	ui.Newline()

	// Show tip about vel shell function
	ui.Muted("Tip: Run this once to use 'vel' instead of './vel':")
	ui.Muted(`  grep -q "vel()" ~/.zshrc || echo 'vel() { [ -x ./vel ] && ./vel "$@" || echo "vel: not found"; }' >> ~/.zshrc && source ~/.zshrc`)
	ui.Newline()
}

func setupTemplatesAndHotReload(projectPath string) error {
	// .air.toml and tmp/ in .gitignore are now part of the template
	return nil
}

func addVelocityDependencies(config ProjectConfig, projectPath string) error {
	// Change to project directory
	originalDir, _ := os.Getwd()
	os.Chdir(projectPath)
	defer os.Chdir(originalDir)

	ui.Step("Configuring dependencies...")

	// Check if local Velocity exists and use replace directive
	velocityPath := "/Users/ali/code/velocity"
	if _, err := os.Stat(velocityPath); err == nil {
		// Add replace directive for local development
		cmd := exec.Command("go", "mod", "edit", "-replace", "github.com/velocitykode/velocity="+velocityPath)
		cmd.Run()
		ui.Info("Using local Velocity framework")
	} else {
		// Try to get from GitHub
		cmd := exec.Command("go", "get", "github.com/velocitykode/velocity")
		if err := cmd.Run(); err != nil {
			ui.Warning("Note: Configure GOPRIVATE for private repo access")
		}
	}

	// Add other dependencies based on features
	if config.Database == "postgres" {
		ui.Info("PostgreSQL driver")
		exec.Command("go", "get", "github.com/lib/pq").Run()
	} else if config.Database == "mysql" {
		ui.Info("MySQL driver")
		exec.Command("go", "get", "github.com/go-sql-driver/mysql").Run()
	} else if config.Database == "sqlite" {
		ui.Info("SQLite driver")
		exec.Command("go", "get", "github.com/mattn/go-sqlite3").Run()
	}

	if config.Cache == "redis" {
		ui.Info("Redis client")
		exec.Command("go", "get", "github.com/redis/go-redis/v9").Run()
	}

	// Run go mod tidy
	ui.Step("Tidying up dependencies...")
	exec.Command("go", "mod", "tidy").Run()

	return nil
}
