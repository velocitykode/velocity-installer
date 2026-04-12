package generator

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	cli "github.com/velocitykode/velocity-cli"
	"github.com/velocitykode/velocity-installer/internal/ui"
)

// Fallback version if GitHub API is unavailable
const fallbackVelocityVersion = "v0.20.3"

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
	// SSR toggles Inertia server-side rendering in the generated app.
	// When true the installer enables INERTIA_SSR_ENABLED=true and
	// points INERTIA_SSR_URL at Vite's /__inertia_ssr dev endpoint.
	// When false the installer adds ssr:false to vite.config.ts so the
	// @inertiajs/vite plugin stays quiet in dev.
	SSR bool
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

	cli.Info("Creating new Velocity project")

	// Clone template
	if err := cli.Spinner("Cloning template", func() error {
		return cloneTemplate(config.Name, config.API)
	}); err != nil {
		return fmt.Errorf("failed to clone template: %w", err)
	}
	cli.Success("Template cloned")

	// Replace module name in all files
	if err := cli.Spinner("Configuring module", func() error {
		return replaceModuleName(config.Name, moduleName)
	}); err != nil {
		return fmt.Errorf("failed to configure project: %w", err)
	}
	cli.Success("Module configured")

	// Remove template git history and initialize new repo
	if err := cli.Spinner("Initializing Git", func() error {
		return reinitGitRepo(config.Name)
	}); err != nil {
		return fmt.Errorf("failed to initialize git: %w", err)
	}
	cli.Success("Git initialized")

	// Create default migrations
	if err := createDefaultMigrations(config.Name); err != nil {
		return fmt.Errorf("failed to create migrations: %w", err)
	}
	cli.Success("Migrations created")

	// Create proper .env.example with database config
	if err := createEnvFiles(config); err != nil {
		return fmt.Errorf("failed to create env files: %w", err)
	}
	cli.Success("Environment configured")

	// Setup hot reload
	if err := setupTemplatesAndHotReload(config.Name); err != nil {
		return fmt.Errorf("failed to setup templates: %w", err)
	}
	cli.Success("Hot reload configured")

	cli.Newline()
	cli.Info("Installing dependencies")
	if err := installDependencies(config.Name, config.API); err != nil {
		return fmt.Errorf("failed to install dependencies: %w", err)
	}

	// Everything downstream (key:generate, migrate, serve) shells out to
	// the project-local vel binary so the framework stays authoritative —
	// the installer owns no runtime logic.
	cli.Newline()
	cli.Info("Building vel")
	if err := buildVel(config.Name); err != nil {
		return fmt.Errorf("failed to build vel: %w", err)
	}
	cli.Success("Built ./vel")

	cli.Newline()
	cli.Info("Generating application key")
	if err := runVel(config.Name, "key:generate"); err != nil {
		return fmt.Errorf("failed to generate app key: %w", err)
	}

	cli.Newline()
	ready, err := ensureDatabaseReady(config.Name)
	if err != nil {
		return fmt.Errorf("failed to prepare database: %w", err)
	}
	if !ready {
		// Preflight already printed a user-facing message. Installation is
		// otherwise complete — the caller decides whether to auto-start.
		return ErrMigrationsSkipped
	}

	cli.Newline()
	cli.Info("Running migrations")
	if err := runVel(config.Name, "migrate"); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	cli.Success("Migrations complete")

	return nil
}

// buildVel compiles the project's vel binary. Downstream steps (key:generate,
// migrate, serve) shell out to this binary, so every user-visible action
// flows through the framework's own console commands.
func buildVel(projectPath string) error {
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return err
	}
	cmd := exec.Command("go", "build", "-o", "vel", ".")
	cmd.Dir = absPath
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// runVel invokes the project's own vel binary with the given args and
// streams its output through so the framework's console UI stays intact.
func runVel(projectPath string, args ...string) error {
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return err
	}
	cmd := exec.Command("./vel", args...)
	cmd.Dir = absPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// ErrMigrationsSkipped signals that project scaffolding succeeded but the
// database preflight chose to skip migrations (e.g. the DB server wasn't
// reachable). Callers should treat this as a non-fatal outcome.
var ErrMigrationsSkipped = errors.New("migrations skipped: database not ready")

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

	// Portable in-place edits: `sed PATTERN file > tmp && mv tmp file`.
	// Works identically on GNU (Linux) and BSD (macOS) sed without -i quirks.
	cmd := exec.Command("sh", "-c",
		fmt.Sprintf(`cd '%s' && find . -name '*.go' -type f -exec sh -c 'sed "s|{{MODULE_NAME}}|%s|g" "$1" > "$1.tmp" && mv "$1.tmp" "$1"' _ {} \;`, absPath, moduleName))
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("replace go files: %w: %s", err, string(output))
	}

	// Replace in go.mod (if exists)
	cmd = exec.Command("sh", "-c",
		fmt.Sprintf("cd '%s' && [ -f go.mod ] && sed 's|{{MODULE_NAME}}|%s|g' go.mod > go.mod.tmp && mv go.mod.tmp go.mod || true", absPath, moduleName))
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("replace go.mod: %w: %s", err, string(output))
	}

	// Replace in package.json (if exists)
	cmd = exec.Command("sh", "-c",
		fmt.Sprintf("cd '%s' && [ -f package.json ] && sed 's|{{MODULE_NAME}}|%s|g' package.json > package.json.tmp && mv package.json.tmp package.json || true", absPath, moduleName))
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("replace package.json: %w: %s", err, string(output))
	}

	// Replace in package-lock.json (if exists)
	cmd = exec.Command("sh", "-c",
		fmt.Sprintf("cd '%s' && [ -f package-lock.json ] && sed 's|{{MODULE_NAME}}|%s|g' package-lock.json > package-lock.json.tmp && mv package-lock.json.tmp package-lock.json || true", absPath, moduleName))
	cmd.Run() // Ignore error - file may not exist

	// Only process go.mod if it exists
	goModPath := filepath.Join(absPath, "go.mod")
	if _, err := os.Stat(goModPath); err == nil {
		// Remove replace directive
		cmd = exec.Command("sh", "-c",
			fmt.Sprintf("cd '%s' && sed '/^replace github.com\\/velocitykode\\/velocity/d' go.mod > go.mod.tmp && mv go.mod.tmp go.mod", absPath))
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

	// Print initial tree (skip JS for API-only projects). Hot reload now ships
	// inside the vel binary via `./vel serve`, so no external watcher install.
	printDepTree := func() {
		if apiOnly {
			ui.TreeItem("└─", goStatus.name, goStatus.status, goStatus.done)
		} else {
			ui.TreeItem("├─", goStatus.name, goStatus.status, goStatus.done)
			ui.TreeItem("└─", jsStatus.name, jsStatus.status, jsStatus.done)
		}
	}

	printDepTree()

	numTasks := 2
	linesToClear := 2
	if apiOnly {
		numTasks = 1
		linesToClear = 1
	}

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

	// Wait for all to complete, updating display
	completed := 0
	for completed < numTasks {
		<-done
		completed++
		ui.ClearLines(linesToClear)
		printDepTree()
	}

	if goStatus.err != nil {
		return goStatus.err
	}
	if !apiOnly && jsStatus.err != nil {
		return jsStatus.err
	}

	return nil
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

	cli.Step("Configuring dependencies...")

	// Check if local Velocity exists and use replace directive
	velocityPath := "/Users/ali/code/velocity"
	if _, err := os.Stat(velocityPath); err == nil {
		// Add replace directive for local development
		cmd = exec.Command("go", "mod", "edit", "-replace", "github.com/velocitykode/velocity="+velocityPath)
		cmd.Run()
		cli.Info("Using local Velocity framework")
	} else {
		// Try to get from GitHub (requires GOPRIVATE setup for private repos)
		cmd = exec.Command("go", "get", "github.com/velocitykode/velocity@v0.20.3")
		if err := cmd.Run(); err != nil {
			cli.Warning("Note: Configure GOPRIVATE for private repo access")
		}
	}

	// Add other dependencies based on features
	if config.Database == "postgres" {
		cli.Info("PostgreSQL driver")
		exec.Command("go", "get", "github.com/lib/pq").Run()
	} else if config.Database == "mysql" {
		cli.Info("MySQL driver")
		exec.Command("go", "get", "github.com/go-sql-driver/mysql").Run()
	} else if config.Database == "sqlite" {
		cli.Info("SQLite driver")
		exec.Command("go", "get", "github.com/mattn/go-sqlite3").Run()
	}

	if config.Cache == "redis" {
		cli.Info("Redis client")
		exec.Command("go", "get", "github.com/redis/go-redis/v9").Run()
	}

	// Run go mod tidy
	cli.Step("Tidying up dependencies...")
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
	cli.Step("Setting up Velocity structure...")
	// Create directory structure in existing directory
	if err := createDirectoryStructure(targetDir); err != nil {
		return fmt.Errorf("failed to create directory structure: %w", err)
	}
	cli.Success("Velocity structure created")

	cli.Step("Generating application files...")
	// Generate files from stubs (skip if exists to preserve existing code)
	if err := generateFilesFromStubs(config); err != nil {
		return fmt.Errorf("failed to generate files: %w", err)
	}
	cli.Success("Application files generated")

	cli.Step("Creating configuration files...")
	// Generate config files if they don't exist
	if err := generateProjectFiles(config); err != nil {
		return fmt.Errorf("failed to generate project files: %w", err)
	}
	cli.Success("Configuration files created")

	cli.Step("Adding Velocity dependencies...")
	// Add dependencies to existing go.mod
	if err := addVelocityDependencies(config, targetDir); err != nil {
		return fmt.Errorf("failed to add dependencies: %w", err)
	}
	cli.Success("Dependencies added")

	return nil
}

// createEnvFiles copies .env.example to .env and patches the install-time
// choices (db driver, cache driver, SSR). Key material is left alone — the
// framework's `./vel key:generate` step writes APP_KEY later.
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

	// Update DB settings based on config
	if config.Database != "" && config.Database != "sqlite" {
		ports := map[string]string{"postgres": "5432", "mysql": "3306"}
		username := os.Getenv("USER")

		// Use base name for database name (not full path)
		dbName := filepath.Base(config.Name)

		// Update DB_CONNECTION, DB_PORT, DB_DATABASE, DB_USERNAME
		sedCmds := fmt.Sprintf(
			"sed 's|^DB_CONNECTION=.*|DB_CONNECTION=%s|; s|^DB_PORT=.*|DB_PORT=%s|; s|^DB_DATABASE=.*|DB_DATABASE=%s|; s|^DB_USERNAME=.*|DB_USERNAME=%s|' .env > .env.tmp && mv .env.tmp .env",
			config.Database, ports[config.Database], dbName, username)
		cmd = exec.Command("sh", "-c", fmt.Sprintf("cd '%s' && %s", absPath, sedCmds))
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	// Update CACHE_DRIVER based on config
	if config.Cache != "" && config.Cache != "memory" {
		cmd = exec.Command("sh", "-c",
			fmt.Sprintf("cd '%s' && sed 's|^CACHE_DRIVER=.*|CACHE_DRIVER=%s|' .env > .env.tmp && mv .env.tmp .env", absPath, config.Cache))
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	if err := applySSROption(config, absPath); err != nil {
		return err
	}

	return nil
}

// applySSROption wires the --ssr flag into the generated project.
// SSR is strictly opt-in: without --ssr the installer leaves every
// SSR-related config untouched (the template ships with Inertia Vite
// plugin's v3 defaults). With --ssr we enable INERTIA_SSR_ENABLED in
// .env and point the URL at Vite's dev endpoint, so the scaffold is
// serving SSR out of the box.
func applySSROption(config ProjectConfig, absPath string) error {
	if !config.SSR || config.API {
		return nil
	}

	cmd := exec.Command("sh", "-c", fmt.Sprintf(
		"cd '%s' && sed -E 's|^# *INERTIA_SSR_ENABLED=.*|INERTIA_SSR_ENABLED=true|; s|^# *INERTIA_SSR_URL=.*|INERTIA_SSR_URL=http://localhost:5173/__inertia_ssr|; s|^# *INERTIA_SSR_TIMEOUT=.*|INERTIA_SSR_TIMEOUT=3s|' .env > .env.tmp && mv .env.tmp .env",
		absPath,
	))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("enable ssr in .env: %w", err)
	}
	return nil
}

// createDefaultMigrations creates default migration files only if the template
// didn't already provide them.
func createDefaultMigrations(projectPath string) error {
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return err
	}

	migrationsDir := filepath.Join(absPath, "database", "migrations")

	// Skip if the template already has migration files
	if entries, err := filepath.Glob(filepath.Join(migrationsDir, "*.go")); err == nil && len(entries) > 0 {
		return nil
	}

	if err := os.MkdirAll(migrationsDir, 0755); err != nil {
		return err
	}

	// Migration 1: Create users table
	usersTable := `package migrations

import "github.com/velocitykode/velocity/orm/migrate"

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

import "github.com/velocitykode/velocity/orm/migrate"

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

import "github.com/velocitykode/velocity/orm/migrate"

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

// StartDevServers launches `./vel serve`, which handles hot reload and Vite
// itself (no separate air/npm processes needed).
func StartDevServers(projectPath string, apiOnly bool) {
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		cli.Error(fmt.Sprintf("Failed to resolve project path: %v", err))
		return
	}

	serveCmd := exec.Command("./vel", "serve")
	serveCmd.Dir = absPath
	serveCmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := serveCmd.Start(); err != nil {
		cli.Error(fmt.Sprintf("Failed to start dev server: %v", err))
		return
	}

	cli.Success("Dev server running")
	cli.Newline()

	cli.KeyValue("cd", projectPath)
	if !apiOnly {
		cli.KeyValue("Vite", cli.Highlight("http://localhost:5173"))
	}
	cli.KeyValue("Velocity", cli.Highlight("http://localhost:4000"))
	cli.Newline()

	cli.Muted("Tip: ./vel serve, ./vel migrate, ./vel route:list, ./vel make:handler")
	cli.Newline()
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

	cli.Step("Configuring dependencies...")

	// Check if local Velocity exists and use replace directive
	velocityPath := "/Users/ali/code/velocity"
	if _, err := os.Stat(velocityPath); err == nil {
		// Add replace directive for local development
		cmd := exec.Command("go", "mod", "edit", "-replace", "github.com/velocitykode/velocity="+velocityPath)
		cmd.Run()
		cli.Info("Using local Velocity framework")
	} else {
		// Try to get from GitHub
		cmd := exec.Command("go", "get", "github.com/velocitykode/velocity")
		if err := cmd.Run(); err != nil {
			cli.Warning("Note: Configure GOPRIVATE for private repo access")
		}
	}

	// Add other dependencies based on features
	if config.Database == "postgres" {
		cli.Info("PostgreSQL driver")
		exec.Command("go", "get", "github.com/lib/pq").Run()
	} else if config.Database == "mysql" {
		cli.Info("MySQL driver")
		exec.Command("go", "get", "github.com/go-sql-driver/mysql").Run()
	} else if config.Database == "sqlite" {
		cli.Info("SQLite driver")
		exec.Command("go", "get", "github.com/mattn/go-sqlite3").Run()
	}

	if config.Cache == "redis" {
		cli.Info("Redis client")
		exec.Command("go", "get", "github.com/redis/go-redis/v9").Run()
	}

	// Run go mod tidy
	cli.Step("Tidying up dependencies...")
	exec.Command("go", "mod", "tidy").Run()

	return nil
}
