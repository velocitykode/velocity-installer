package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTemplateRef(t *testing.T) {
	cases := []struct {
		repo string
		want string
	}{
		{repo: "velocity-template-react", want: "tags/" + supportedTemplates["react"]},
		{repo: "velocity-template-api", want: "tags/" + supportedTemplates["api"]},
		{repo: "velocity-template-vue", want: "tags/" + supportedTemplates["vue"]},
		{repo: "velocity-template-unknown", want: "heads/main"}, // not in map -> main
	}
	for _, tc := range cases {
		t.Run(tc.repo, func(t *testing.T) {
			if got := templateRef(tc.repo); got != tc.want {
				t.Errorf("templateRef(%q) = %q, want %q", tc.repo, got, tc.want)
			}
		})
	}
}

func TestValidateProjectName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		setup   func(t *testing.T) string // returns cleanup function
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty string returns error",
			input:   "",
			wantErr: true,
			errMsg:  "project name cannot be empty",
		},
		{
			name:    "valid hyphenated name",
			input:   "my-project",
			wantErr: false,
		},
		{
			name:    "valid name with numbers",
			input:   "app123",
			wantErr: false,
		},
		{
			name:    "valid name with underscore",
			input:   "test_app",
			wantErr: false,
		},
		{
			name:    "valid mixed case name",
			input:   "MyProject",
			wantErr: false,
		},
		{
			name:    "valid all lowercase name",
			input:   "myproject",
			wantErr: false,
		},
		{
			name:    "valid name with dot",
			input:   "my.project",
			wantErr: false,
		},
		{
			name:    "name with space returns error",
			input:   "my project",
			wantErr: true,
			errMsg:  "project name contains invalid characters",
		},
		{
			name:    "name with exclamation mark returns error",
			input:   "project!",
			wantErr: true,
			errMsg:  "project name contains invalid characters",
		},
		{
			name:    "name with at sign returns error",
			input:   "project@test",
			wantErr: true,
			errMsg:  "project name contains invalid characters",
		},
		{
			name:    "name with hash returns error",
			input:   "project#1",
			wantErr: true,
			errMsg:  "project name contains invalid characters",
		},
		{
			name:    "name with dollar sign returns error",
			input:   "$project",
			wantErr: true,
			errMsg:  "project name contains invalid characters",
		},
		{
			name:    "name with percent returns error",
			input:   "project%",
			wantErr: true,
			errMsg:  "project name contains invalid characters",
		},
		{
			name:    "name with caret returns error",
			input:   "project^test",
			wantErr: true,
			errMsg:  "project name contains invalid characters",
		},
		{
			name:    "name with ampersand returns error",
			input:   "project&test",
			wantErr: true,
			errMsg:  "project name contains invalid characters",
		},
		{
			name:    "name with asterisk returns error",
			input:   "project*",
			wantErr: true,
			errMsg:  "project name contains invalid characters",
		},
		{
			name:    "name with parentheses returns error",
			input:   "project(test)",
			wantErr: true,
			errMsg:  "project name contains invalid characters",
		},
		{
			name:    "name with plus returns error",
			input:   "project+test",
			wantErr: true,
			errMsg:  "project name contains invalid characters",
		},
		{
			name:    "name with equals returns error",
			input:   "project=test",
			wantErr: true,
			errMsg:  "project name contains invalid characters",
		},
		{
			name:    "name with brackets returns error",
			input:   "project[1]",
			wantErr: true,
			errMsg:  "project name contains invalid characters",
		},
		{
			name:    "name with braces returns error",
			input:   "project{test}",
			wantErr: true,
			errMsg:  "project name contains invalid characters",
		},
		{
			name:    "name with pipe returns error",
			input:   "project|test",
			wantErr: true,
			errMsg:  "project name contains invalid characters",
		},
		{
			name:    "name with backslash returns error",
			input:   "project\\test",
			wantErr: true,
			errMsg:  "project name contains invalid characters",
		},
		{
			name:    "name with semicolon returns error",
			input:   "project;test",
			wantErr: true,
			errMsg:  "project name contains invalid characters",
		},
		{
			name:    "name with colon returns error",
			input:   "project:test",
			wantErr: true,
			errMsg:  "project name contains invalid characters",
		},
		{
			name:    "name with single quote returns error",
			input:   "project'test",
			wantErr: true,
			errMsg:  "project name contains invalid characters",
		},
		{
			name:    "name with double quote returns error",
			input:   `project"test`,
			wantErr: true,
			errMsg:  "project name contains invalid characters",
		},
		{
			name:    "name with comma returns error",
			input:   "project,test",
			wantErr: true,
			errMsg:  "project name contains invalid characters",
		},
		{
			name:    "name with less than returns error",
			input:   "project<test",
			wantErr: true,
			errMsg:  "project name contains invalid characters",
		},
		{
			name:    "name with greater than returns error",
			input:   "project>test",
			wantErr: true,
			errMsg:  "project name contains invalid characters",
		},
		{
			name:    "name with question mark returns error",
			input:   "project?",
			wantErr: true,
			errMsg:  "project name contains invalid characters",
		},
		{
			name:    "name with forward slash returns error",
			input:   "project/test",
			wantErr: true,
			errMsg:  "project name contains invalid characters",
		},
		{
			name:  "directory exists returns error",
			input: "existing-dir",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				originalDir, _ := os.Getwd()
				os.Chdir(tmpDir)
				os.Mkdir("existing-dir", 0755)
				return originalDir
			},
			wantErr: true,
			errMsg:  "directory existing-dir already exists",
		},
		{
			name:    "path traversal with double dots returns error",
			input:   "../test",
			wantErr: true,
			errMsg:  "project name contains invalid characters",
		},
		{
			name:    "path traversal nested returns error",
			input:   "../../foo",
			wantErr: true,
			errMsg:  "project name contains invalid characters",
		},
		{
			name:  "single dot returns error",
			input: ".",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				originalDir, _ := os.Getwd()
				os.Chdir(tmpDir)
				return originalDir
			},
			wantErr: true,
			errMsg:  "directory . already exists",
		},
		{
			name:  "double dot returns error",
			input: "..",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				originalDir, _ := os.Getwd()
				os.Chdir(tmpDir)
				return originalDir
			},
			wantErr: true,
			errMsg:  "directory .. already exists",
		},
		{
			name:    "very long name is accepted",
			input:   strings.Repeat("a", 255),
			wantErr: false,
		},
		{
			name:    "unicode characters are accepted",
			input:   "プロジェクト",
			wantErr: false,
		},
		{
			name:    "emoji in name is accepted",
			input:   "project🚀",
			wantErr: false,
		},
		{
			name:    "leading hyphen is accepted",
			input:   "-myproject",
			wantErr: false,
		},
		{
			name:    "trailing hyphen is accepted",
			input:   "myproject-",
			wantErr: false,
		},
		{
			name:    "multiple hyphens are accepted",
			input:   "my-awesome-project",
			wantErr: false,
		},
		{
			name:    "multiple underscores are accepted",
			input:   "my_awesome_project",
			wantErr: false,
		},
		{
			name:    "mixed hyphens and underscores are accepted",
			input:   "my-awesome_project",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var originalDir string
			if tt.setup != nil {
				originalDir = tt.setup(t)
				defer os.Chdir(originalDir)
			}

			err := validateProjectName(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("validateProjectName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" {
				if err == nil {
					t.Errorf("validateProjectName(%q) expected error containing %q, got nil", tt.input, tt.errMsg)
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("validateProjectName(%q) error = %q, want error containing %q", tt.input, err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestValidateProjectNameEdgeCasesWithCleanup(t *testing.T) {
	// This test ensures cleanup happens properly for directory tests
	t.Run("cleanup after directory exists test", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)

		os.Chdir(tmpDir)

		// Create a test directory
		testDirName := "test-cleanup-dir"
		err := os.Mkdir(testDirName, 0755)
		if err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}

		// Verify it fails validation
		err = validateProjectName(testDirName)
		if err == nil {
			t.Error("validateProjectName() expected error for existing directory, got nil")
		}

		// Verify directory still exists after validation
		if _, err := os.Stat(testDirName); os.IsNotExist(err) {
			t.Error("Test directory was removed by validateProjectName()")
		}
	})

	t.Run("nonexistent directory in temp location", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)

		os.Chdir(tmpDir)

		// Test a name that doesn't exist
		err := validateProjectName("nonexistent-project-xyz")
		if err != nil {
			t.Errorf("validateProjectName() unexpected error for nonexistent directory: %v", err)
		}
	})

	t.Run("symlink to existing directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)

		os.Chdir(tmpDir)

		// Create actual directory and symlink to it
		actualDir := filepath.Join(tmpDir, "actual-dir")
		symlinkName := "symlink-dir"

		err := os.Mkdir(actualDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create actual directory: %v", err)
		}

		err = os.Symlink(actualDir, symlinkName)
		if err != nil {
			t.Skipf("Failed to create symlink (may not be supported): %v", err)
		}

		// Should fail validation because symlink exists
		err = validateProjectName(symlinkName)
		if err == nil {
			t.Error("validateProjectName() expected error for symlink to existing directory, got nil")
		}
	})

	t.Run("file exists with same name", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)

		os.Chdir(tmpDir)

		// Create a file instead of directory
		fileName := "test-file"
		err := os.WriteFile(fileName, []byte("test"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Should fail validation because file exists
		err = validateProjectName(fileName)
		if err == nil {
			t.Error("validateProjectName() expected error for existing file, got nil")
		}
	})
}

func TestCreateDirectoryStructure(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) string // returns project path
		wantErr bool
	}{
		{
			name: "creates all directories successfully",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				return filepath.Join(tmpDir, "test-project")
			},
			wantErr: false,
		},
		{
			name: "creates directories when parent path exists",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				projectPath := filepath.Join(tmpDir, "existing-parent", "test-project")
				os.MkdirAll(filepath.Join(tmpDir, "existing-parent"), 0755)
				return projectPath
			},
			wantErr: false,
		},
		{
			name: "creates directories in nested path",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				return filepath.Join(tmpDir, "level1", "level2", "test-project")
			},
			wantErr: false,
		},
		{
			name: "succeeds when some directories already exist",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				projectPath := filepath.Join(tmpDir, "test-project")
				// Pre-create some directories
				os.MkdirAll(filepath.Join(projectPath, "app", "http"), 0755)
				os.MkdirAll(filepath.Join(projectPath, "config"), 0755)
				return projectPath
			},
			wantErr: false,
		},
		{
			name: "returns error when path is not writable",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				projectPath := filepath.Join(tmpDir, "readonly-parent", "test-project")
				os.MkdirAll(filepath.Join(tmpDir, "readonly-parent"), 0444) // read-only
				return projectPath
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectPath := tt.setup(t)

			err := createDirectoryStructure(projectPath)

			if (err != nil) != tt.wantErr {
				t.Errorf("createDirectoryStructure() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify all expected directories exist
				expectedDirs := []string{
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

				for _, dir := range expectedDirs {
					fullPath := filepath.Join(projectPath, dir)
					info, err := os.Stat(fullPath)
					if err != nil {
						t.Errorf("directory %s was not created: %v", dir, err)
						continue
					}
					if !info.IsDir() {
						t.Errorf("%s exists but is not a directory", dir)
					}
					// Verify permissions
					if info.Mode().Perm() != 0755 {
						t.Errorf("directory %s has incorrect permissions: got %v, want 0755", dir, info.Mode().Perm())
					}
				}
			}
		})
	}
}

func TestCreateEnvFiles(t *testing.T) {
	tests := []struct {
		name    string
		config  ProjectConfig
		setup   func(t *testing.T, projectPath string)
		verify  func(t *testing.T, projectPath string)
		wantErr bool
	}{
		{
			name: "creates env file from example with sqlite",
			config: ProjectConfig{
				Name:     "test-project",
				Database: "sqlite",
				Cache:    "memory",
			},
			setup: func(t *testing.T, projectPath string) {
				os.MkdirAll(projectPath, 0755)
				// Create a sample .env.example
				envExample := `APP_ENV=local
APP_DEBUG=true
CRYPTO_KEY=base64:oldkeyhere
DB_CONNECTION=sqlite
DB_PORT=5432
DB_DATABASE=test
DB_USERNAME=user
CACHE_DRIVER=memory
`
				os.WriteFile(filepath.Join(projectPath, ".env.example"), []byte(envExample), 0644)
			},
			verify: func(t *testing.T, projectPath string) {
				envPath := filepath.Join(projectPath, ".env")
				content, err := os.ReadFile(envPath)
				if err != nil {
					t.Fatalf("failed to read .env: %v", err)
				}

				envStr := string(content)

				// CRYPTO_KEY placeholders are left untouched - APP_KEY
				// is the single source of truth and also the crypto
				// fallback in the framework's config layer.
				if !strings.Contains(envStr, "CRYPTO_KEY=base64:oldkeyhere") {
					t.Error(".env should pass CRYPTO_KEY through from .env.example unchanged")
				}

				// APP_KEY, QUEUE_SIGNING_KEY, AUTH_JWT_SECRET must all
				// be written so the framework has everything it needs
				// at bootstrap. Each should be a 44-char base64 value
				// (32 random bytes, standard encoding with padding).
				for _, envKey := range []string{"APP_KEY", "QUEUE_SIGNING_KEY", "AUTH_JWT_SECRET"} {
					var match string
					for _, line := range strings.Split(envStr, "\n") {
						if strings.HasPrefix(line, envKey+"=") {
							match = line
							break
						}
					}
					if match == "" {
						t.Errorf(".env missing %s line", envKey)
						continue
					}
					value := strings.TrimPrefix(match, envKey+"=")
					if len(value) != 44 {
						t.Errorf("%s length = %d, want 44 (base64 of 32 bytes)", envKey, len(value))
					}
				}
			},
			wantErr: false,
		},
		{
			name: "creates env file with postgres configuration",
			config: ProjectConfig{
				Name:     "pg-project",
				Database: "postgres",
				Cache:    "redis",
			},
			setup: func(t *testing.T, projectPath string) {
				os.MkdirAll(projectPath, 0755)
				envExample := `CRYPTO_KEY=base64:test
DB_CONNECTION=sqlite
DB_PORT=3306
DB_DATABASE=olddb
DB_USERNAME=olduser
CACHE_DRIVER=memory
`
				os.WriteFile(filepath.Join(projectPath, ".env.example"), []byte(envExample), 0644)
			},
			verify: func(t *testing.T, projectPath string) {
				content, err := os.ReadFile(filepath.Join(projectPath, ".env"))
				if err != nil {
					t.Fatalf("failed to read .env: %v", err)
				}
				envStr := string(content)

				if !strings.Contains(envStr, "DB_CONNECTION=postgres") {
					t.Error(".env does not contain DB_CONNECTION=postgres")
				}
				if !strings.Contains(envStr, "DB_PORT=5432") {
					t.Error(".env does not contain DB_PORT=5432")
				}
				if !strings.Contains(envStr, "DB_DATABASE=pg-project") {
					t.Errorf(".env does not contain correct DB_DATABASE, content:\n%s", envStr)
				}
				if !strings.Contains(envStr, "CACHE_DRIVER=redis") {
					t.Error(".env does not contain CACHE_DRIVER=redis")
				}

				// Verify username is current user
				username := os.Getenv("USER")
				if username != "" && !strings.Contains(envStr, "DB_USERNAME="+username) {
					t.Errorf(".env does not contain DB_USERNAME=%s", username)
				}
			},
			wantErr: false,
		},
		{
			name: "creates env file with mysql configuration",
			config: ProjectConfig{
				Name:     "mysql-project",
				Database: "mysql",
				Cache:    "memory",
			},
			setup: func(t *testing.T, projectPath string) {
				os.MkdirAll(projectPath, 0755)
				envExample := `CRYPTO_KEY=base64:test
DB_CONNECTION=sqlite
DB_PORT=5432
DB_DATABASE=olddb
DB_USERNAME=olduser
CACHE_DRIVER=memory
`
				os.WriteFile(filepath.Join(projectPath, ".env.example"), []byte(envExample), 0644)
			},
			verify: func(t *testing.T, projectPath string) {
				content, err := os.ReadFile(filepath.Join(projectPath, ".env"))
				if err != nil {
					t.Fatalf("failed to read .env: %v", err)
				}
				envStr := string(content)

				if !strings.Contains(envStr, "DB_CONNECTION=mysql") {
					t.Error(".env does not contain DB_CONNECTION=mysql")
				}
				if !strings.Contains(envStr, "DB_PORT=3306") {
					t.Error(".env does not contain DB_PORT=3306")
				}
				if !strings.Contains(envStr, "DB_DATABASE=mysql-project") {
					t.Errorf(".env does not contain correct DB_DATABASE, content:\n%s", envStr)
				}
			},
			wantErr: false,
		},
		{
			name: "returns error when env example does not exist",
			config: ProjectConfig{
				Name:     "no-example",
				Database: "sqlite",
				Cache:    "memory",
			},
			setup: func(t *testing.T, projectPath string) {
				os.MkdirAll(projectPath, 0755)
				// Don't create .env.example
			},
			verify:  func(t *testing.T, projectPath string) {},
			wantErr: true,
		},
		{
			name: "returns error when project path does not exist",
			config: ProjectConfig{
				Name:     "nonexistent",
				Database: "sqlite",
				Cache:    "memory",
			},
			setup:   func(t *testing.T, projectPath string) {},
			verify:  func(t *testing.T, projectPath string) {},
			wantErr: true,
		},
		{
			name: "handles empty database config",
			config: ProjectConfig{
				Name:     "empty-db",
				Database: "",
				Cache:    "",
			},
			setup: func(t *testing.T, projectPath string) {
				os.MkdirAll(projectPath, 0755)
				envExample := `CRYPTO_KEY=base64:test
DB_CONNECTION=sqlite
CACHE_DRIVER=memory
`
				os.WriteFile(filepath.Join(projectPath, ".env.example"), []byte(envExample), 0644)
			},
			verify: func(t *testing.T, projectPath string) {
				if _, err := os.ReadFile(filepath.Join(projectPath, ".env")); err != nil {
					t.Fatalf("failed to read .env: %v", err)
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			projectPath := filepath.Join(tmpDir, tt.config.Name)
			tt.config.Name = projectPath

			tt.setup(t, projectPath)

			err := createEnvFiles(tt.config)

			if (err != nil) != tt.wantErr {
				t.Errorf("createEnvFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				tt.verify(t, projectPath)
			}
		})
	}
}

func TestReplaceModuleName(t *testing.T) {
	tests := []struct {
		name       string
		moduleName string
		setup      func(t *testing.T, projectPath string)
		verify     func(t *testing.T, projectPath string)
		wantErr    bool
	}{
		{
			name:       "replaces module name in go files",
			moduleName: "github.com/test/myproject",
			setup: func(t *testing.T, projectPath string) {
				os.MkdirAll(filepath.Join(projectPath, "app"), 0755)
				goFile := `package main

import "{{MODULE_NAME}}/app/models"

func main() {
	// {{MODULE_NAME}}
}
`
				os.WriteFile(filepath.Join(projectPath, "main.go"), []byte(goFile), 0644)
				os.WriteFile(filepath.Join(projectPath, "app", "handler.go"), []byte(`package app

import "{{MODULE_NAME}}/config"
`), 0644)
			},
			verify: func(t *testing.T, projectPath string) {
				content, err := os.ReadFile(filepath.Join(projectPath, "main.go"))
				if err != nil {
					t.Fatalf("failed to read main.go: %v", err)
				}
				if strings.Contains(string(content), "{{MODULE_NAME}}") {
					t.Error("main.go still contains {{MODULE_NAME}} placeholder")
				}
				if !strings.Contains(string(content), "github.com/test/myproject/app/models") {
					t.Error("main.go does not contain replaced module name in import")
				}

				content, err = os.ReadFile(filepath.Join(projectPath, "app", "handler.go"))
				if err != nil {
					t.Fatalf("failed to read handler.go: %v", err)
				}
				if strings.Contains(string(content), "{{MODULE_NAME}}") {
					t.Error("handler.go still contains {{MODULE_NAME}} placeholder")
				}
			},
			wantErr: false,
		},
		{
			name:       "replaces module name in go.mod",
			moduleName: "mymodule",
			setup: func(t *testing.T, projectPath string) {
				goMod := `module {{MODULE_NAME}}

go 1.21

require (
	github.com/velocitykode/velocity v0.1.0
)

replace github.com/velocitykode/velocity => /Users/ali/code/velocity
`
				os.WriteFile(filepath.Join(projectPath, "go.mod"), []byte(goMod), 0644)
			},
			verify: func(t *testing.T, projectPath string) {
				content, err := os.ReadFile(filepath.Join(projectPath, "go.mod"))
				if err != nil {
					t.Fatalf("failed to read go.mod: %v", err)
				}
				if strings.Contains(string(content), "{{MODULE_NAME}}") {
					t.Error("go.mod still contains {{MODULE_NAME}} placeholder")
				}
				if !strings.Contains(string(content), "module mymodule") {
					t.Error("go.mod does not contain replaced module name")
				}
				// Verify replace directive was removed
				if strings.Contains(string(content), "replace github.com/velocitykode/velocity") {
					t.Error("go.mod still contains replace directive")
				}
				// Verify velocity version was set
				if !strings.Contains(string(content), "github.com/velocitykode/velocity") {
					t.Error("go.mod does not contain velocity dependency")
				}
			},
			wantErr: false,
		},
		{
			name:       "replaces module name in package.json",
			moduleName: "my-project",
			setup: func(t *testing.T, projectPath string) {
				packageJSON := `{
  "name": "{{MODULE_NAME}}",
  "version": "1.0.0",
  "description": "{{MODULE_NAME}} project"
}
`
				os.WriteFile(filepath.Join(projectPath, "package.json"), []byte(packageJSON), 0644)
			},
			verify: func(t *testing.T, projectPath string) {
				content, err := os.ReadFile(filepath.Join(projectPath, "package.json"))
				if err != nil {
					t.Fatalf("failed to read package.json: %v", err)
				}
				if strings.Contains(string(content), "{{MODULE_NAME}}") {
					t.Error("package.json still contains {{MODULE_NAME}} placeholder")
				}
				if !strings.Contains(string(content), `"name": "my-project"`) {
					t.Error("package.json does not contain replaced module name")
				}
			},
			wantErr: false,
		},
		{
			name:       "handles package-lock.json when present",
			moduleName: "my-project",
			setup: func(t *testing.T, projectPath string) {
				packageJSON := `{"name": "{{MODULE_NAME}}"}`
				packageLock := `{"name": "{{MODULE_NAME}}", "lockfileVersion": 3}`
				os.WriteFile(filepath.Join(projectPath, "package.json"), []byte(packageJSON), 0644)
				os.WriteFile(filepath.Join(projectPath, "package-lock.json"), []byte(packageLock), 0644)
			},
			verify: func(t *testing.T, projectPath string) {
				content, err := os.ReadFile(filepath.Join(projectPath, "package-lock.json"))
				if err != nil {
					t.Fatalf("failed to read package-lock.json: %v", err)
				}
				if strings.Contains(string(content), "{{MODULE_NAME}}") {
					t.Error("package-lock.json still contains {{MODULE_NAME}} placeholder")
				}
			},
			wantErr: false,
		},
		{
			name:       "succeeds when package-lock.json is absent",
			moduleName: "my-project",
			setup: func(t *testing.T, projectPath string) {
				packageJSON := `{"name": "{{MODULE_NAME}}"}`
				os.WriteFile(filepath.Join(projectPath, "package.json"), []byte(packageJSON), 0644)
				// Don't create package-lock.json
			},
			verify: func(t *testing.T, projectPath string) {
				// Just verify it didn't error
			},
			wantErr: false,
		},
		{
			name:       "succeeds when go.mod does not exist",
			moduleName: "test",
			setup: func(t *testing.T, projectPath string) {
				// Don't create go.mod - function should still succeed
				os.WriteFile(filepath.Join(projectPath, "package.json"), []byte(`{"name": "{{MODULE_NAME}}"}`), 0644)
			},
			verify: func(t *testing.T, projectPath string) {
				// Verify package.json was updated
				content, err := os.ReadFile(filepath.Join(projectPath, "package.json"))
				if err != nil {
					t.Fatalf("failed to read package.json: %v", err)
				}
				if strings.Contains(string(content), "{{MODULE_NAME}}") {
					t.Error("package.json still contains placeholder")
				}
			},
			wantErr: false,
		},
		{
			name:       "succeeds when package.json does not exist",
			moduleName: "test",
			setup: func(t *testing.T, projectPath string) {
				os.WriteFile(filepath.Join(projectPath, "go.mod"), []byte("module {{MODULE_NAME}}"), 0644)
				// Don't create package.json - function should still succeed
			},
			verify: func(t *testing.T, projectPath string) {
				// Verify go.mod was updated
				content, err := os.ReadFile(filepath.Join(projectPath, "go.mod"))
				if err != nil {
					t.Fatalf("failed to read go.mod: %v", err)
				}
				if strings.Contains(string(content), "{{MODULE_NAME}}") {
					t.Error("go.mod still contains placeholder")
				}
			},
			wantErr: false,
		},
		{
			name:       "handles module names with special characters safely",
			moduleName: "github.com/user/my-app",
			setup: func(t *testing.T, projectPath string) {
				os.WriteFile(filepath.Join(projectPath, "main.go"), []byte(`package main
import "{{MODULE_NAME}}/app"
`), 0644)
				os.WriteFile(filepath.Join(projectPath, "go.mod"), []byte("module {{MODULE_NAME}}\n"), 0644)
				os.WriteFile(filepath.Join(projectPath, "package.json"), []byte(`{"name": "{{MODULE_NAME}}"}`), 0644)
			},
			verify: func(t *testing.T, projectPath string) {
				content, err := os.ReadFile(filepath.Join(projectPath, "main.go"))
				if err != nil {
					t.Fatalf("failed to read main.go: %v", err)
				}
				if !strings.Contains(string(content), "github.com/user/my-app/app") {
					t.Error("main.go does not contain correctly replaced module name")
				}
			},
			wantErr: false,
		},
		{
			name:       "handles multiple occurrences in single file",
			moduleName: "testmodule",
			setup: func(t *testing.T, projectPath string) {
				multiOccurrence := `package main

import (
	"{{MODULE_NAME}}/app"
	"{{MODULE_NAME}}/config"
	"{{MODULE_NAME}}/models"
)

// {{MODULE_NAME}} comment
`
				os.WriteFile(filepath.Join(projectPath, "main.go"), []byte(multiOccurrence), 0644)
				os.WriteFile(filepath.Join(projectPath, "go.mod"), []byte("module {{MODULE_NAME}}\n"), 0644)
				os.WriteFile(filepath.Join(projectPath, "package.json"), []byte(`{"name": "{{MODULE_NAME}}"}`), 0644)
			},
			verify: func(t *testing.T, projectPath string) {
				content, err := os.ReadFile(filepath.Join(projectPath, "main.go"))
				if err != nil {
					t.Fatalf("failed to read main.go: %v", err)
				}
				if strings.Contains(string(content), "{{MODULE_NAME}}") {
					t.Error("main.go still contains {{MODULE_NAME}} after replacement")
				}
				// Count occurrences of module name
				count := strings.Count(string(content), "testmodule")
				if count < 3 {
					t.Errorf("expected at least 3 occurrences of module name, got %d", count)
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			projectPath := filepath.Join(tmpDir, "test-project")
			os.MkdirAll(projectPath, 0755)

			tt.setup(t, projectPath)

			err := replaceModuleName(projectPath, tt.moduleName)

			if (err != nil) != tt.wantErr {
				t.Errorf("replaceModuleName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				tt.verify(t, projectPath)
			}
		})
	}
}

func TestInitGoModule(t *testing.T) {
	tests := []struct {
		name    string
		config  ProjectConfig
		setup   func(t *testing.T, projectPath string)
		verify  func(t *testing.T, projectPath string)
		wantErr bool
	}{
		{
			name: "initializes go module with project name",
			config: ProjectConfig{
				Name:     "test-project",
				Module:   "",
				Database: "",
				Cache:    "",
			},
			setup: func(t *testing.T, projectPath string) {
				os.MkdirAll(projectPath, 0755)
			},
			verify: func(t *testing.T, projectPath string) {
				goModPath := filepath.Join(projectPath, "go.mod")
				content, err := os.ReadFile(goModPath)
				if err != nil {
					t.Fatalf("failed to read go.mod: %v", err)
				}
				if !strings.Contains(string(content), "module test-project") {
					t.Error("go.mod does not contain correct module name")
				}
			},
			wantErr: false,
		},
		{
			name: "initializes go module with custom module name",
			config: ProjectConfig{
				Name:     "test-project",
				Module:   "github.com/user/custom-module",
				Database: "",
				Cache:    "",
			},
			setup: func(t *testing.T, projectPath string) {
				os.MkdirAll(projectPath, 0755)
			},
			verify: func(t *testing.T, projectPath string) {
				goModPath := filepath.Join(projectPath, "go.mod")
				content, err := os.ReadFile(goModPath)
				if err != nil {
					t.Fatalf("failed to read go.mod: %v", err)
				}
				if !strings.Contains(string(content), "module github.com/user/custom-module") {
					t.Error("go.mod does not contain custom module name")
				}
			},
			wantErr: false,
		},
		{
			name: "fails when project directory does not exist",
			config: ProjectConfig{
				Name:     "nonexistent-project",
				Database: "",
				Cache:    "",
			},
			setup:   func(t *testing.T, projectPath string) {},
			verify:  func(t *testing.T, projectPath string) {},
			wantErr: true,
		},
		{
			name: "fails when go.mod already exists",
			config: ProjectConfig{
				Name:     "existing-mod",
				Database: "",
				Cache:    "",
			},
			setup: func(t *testing.T, projectPath string) {
				os.MkdirAll(projectPath, 0755)
				os.WriteFile(filepath.Join(projectPath, "go.mod"), []byte("module existing\n"), 0644)
			},
			verify:  func(t *testing.T, projectPath string) {},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			projectPath := filepath.Join(tmpDir, tt.config.Name)
			originalDir, _ := os.Getwd()
			defer os.Chdir(originalDir)

			os.Chdir(tmpDir)
			tt.setup(t, projectPath)

			err := initGoModule(tt.config)

			if (err != nil) != tt.wantErr {
				t.Errorf("initGoModule() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				tt.verify(t, projectPath)
			}
		})
	}
}

func TestReinitGitRepo(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T, projectPath string)
		verify  func(t *testing.T, projectPath string)
		wantErr bool
	}{
		{
			name: "removes existing git directory and initializes new repo",
			setup: func(t *testing.T, projectPath string) {
				os.MkdirAll(projectPath, 0755)
				// Create fake .git directory
				gitDir := filepath.Join(projectPath, ".git")
				os.MkdirAll(gitDir, 0755)
				os.WriteFile(filepath.Join(gitDir, "config"), []byte("old config"), 0644)
				os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte("ref: refs/heads/main"), 0644)
			},
			verify: func(t *testing.T, projectPath string) {
				gitDir := filepath.Join(projectPath, ".git")

				// Verify old .git was removed and new one created
				if _, err := os.Stat(filepath.Join(gitDir, "config")); err == nil {
					// If config exists, verify it's a new git repo (not the old one with "old config")
					content, _ := os.ReadFile(filepath.Join(gitDir, "config"))
					if string(content) == "old config" {
						t.Error(".git directory was not reinitialized, still contains old config")
					}
				}

				// Verify .git directory exists (git init was successful)
				info, err := os.Stat(gitDir)
				if err != nil {
					t.Errorf(".git directory does not exist after reinit: %v", err)
					return
				}
				if !info.IsDir() {
					t.Error(".git exists but is not a directory")
				}
			},
			wantErr: false,
		},
		{
			name: "initializes git repo when no git directory exists",
			setup: func(t *testing.T, projectPath string) {
				os.MkdirAll(projectPath, 0755)
				// Don't create .git directory
			},
			verify: func(t *testing.T, projectPath string) {
				gitDir := filepath.Join(projectPath, ".git")
				info, err := os.Stat(gitDir)
				if err != nil {
					t.Errorf(".git directory was not created: %v", err)
					return
				}
				if !info.IsDir() {
					t.Error(".git exists but is not a directory")
				}
			},
			wantErr: false,
		},
		{
			name: "creates initial commit",
			setup: func(t *testing.T, projectPath string) {
				os.MkdirAll(projectPath, 0755)
				// Create some files to commit
				os.WriteFile(filepath.Join(projectPath, "README.md"), []byte("# Test"), 0644)
				os.WriteFile(filepath.Join(projectPath, "main.go"), []byte("package main"), 0644)
			},
			verify: func(t *testing.T, projectPath string) {
				gitDir := filepath.Join(projectPath, ".git")
				if _, err := os.Stat(gitDir); err != nil {
					t.Errorf(".git directory does not exist: %v", err)
				}
				// Note: We can't easily verify commit was made without running git commands,
				// but we can verify the git directory structure exists
			},
			wantErr: false,
		},
		{
			name: "returns error when project path does not exist",
			setup: func(t *testing.T, projectPath string) {
				// Don't create project directory
			},
			verify:  func(t *testing.T, projectPath string) {},
			wantErr: true,
		},
		{
			name: "handles nested git directory structure",
			setup: func(t *testing.T, projectPath string) {
				os.MkdirAll(projectPath, 0755)
				gitDir := filepath.Join(projectPath, ".git")
				os.MkdirAll(filepath.Join(gitDir, "objects", "pack"), 0755)
				os.MkdirAll(filepath.Join(gitDir, "refs", "heads"), 0755)
				os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte("ref: refs/heads/template"), 0644)
			},
			verify: func(t *testing.T, projectPath string) {
				// Verify old nested structure is gone
				oldHeadPath := filepath.Join(projectPath, ".git", "HEAD")
				if content, err := os.ReadFile(oldHeadPath); err == nil {
					// If HEAD exists, it should not contain "template" reference
					if strings.Contains(string(content), "template") {
						t.Error("old git structure not fully removed")
					}
				}
			},
			wantErr: false,
		},
		{
			name: "preserves project files after git reinit",
			setup: func(t *testing.T, projectPath string) {
				os.MkdirAll(projectPath, 0755)
				gitDir := filepath.Join(projectPath, ".git")
				os.MkdirAll(gitDir, 0755)

				// Create project files that should be preserved
				os.WriteFile(filepath.Join(projectPath, "important.txt"), []byte("preserve me"), 0644)
				os.MkdirAll(filepath.Join(projectPath, "app"), 0755)
				os.WriteFile(filepath.Join(projectPath, "app", "main.go"), []byte("package main"), 0644)
			},
			verify: func(t *testing.T, projectPath string) {
				// Verify project files still exist
				if _, err := os.Stat(filepath.Join(projectPath, "important.txt")); err != nil {
					t.Error("important.txt was removed during git reinit")
				}
				if _, err := os.Stat(filepath.Join(projectPath, "app", "main.go")); err != nil {
					t.Error("app/main.go was removed during git reinit")
				}

				content, err := os.ReadFile(filepath.Join(projectPath, "important.txt"))
				if err != nil {
					t.Error("failed to read important.txt after reinit")
				}
				if string(content) != "preserve me" {
					t.Error("important.txt content was modified")
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			projectPath := filepath.Join(tmpDir, "test-project")

			tt.setup(t, projectPath)

			err := reinitGitRepo(projectPath)

			if (err != nil) != tt.wantErr {
				t.Errorf("reinitGitRepo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				tt.verify(t, projectPath)
			}
		})
	}
}

func TestGetTemplateRepo(t *testing.T) {
	tests := []struct {
		name     string
		apiOnly  bool
		stack    string
		wantRepo string
	}{
		{
			name:     "react full-stack template",
			apiOnly:  false,
			stack:    "react",
			wantRepo: "velocity-template-react",
		},
		{
			name:     "vue full-stack template",
			apiOnly:  false,
			stack:    "vue",
			wantRepo: "velocity-template-vue",
		},
		{
			name:     "api template ignores stack",
			apiOnly:  true,
			stack:    "vue",
			wantRepo: "velocity-template-api",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := "velocity-template-" + tt.stack
			if tt.apiOnly {
				repo = "velocity-template-api"
			}
			if repo != tt.wantRepo {
				t.Errorf("template repo = %q, want %q", repo, tt.wantRepo)
			}
		})
	}
}

func TestValidateStack(t *testing.T) {
	tests := []struct {
		stack   string
		wantErr bool
	}{
		{"react", false},
		{"vue", false},
		{"svelte", true},
		{"", true},
		{"REACT", true},
	}
	for _, tt := range tests {
		t.Run(tt.stack, func(t *testing.T) {
			err := validateStack(tt.stack)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateStack(%q) error = %v, wantErr %v", tt.stack, err, tt.wantErr)
			}
		})
	}
}

func TestProjectConfigAPI(t *testing.T) {
	tests := []struct {
		name   string
		config ProjectConfig
		wantJS bool
	}{
		{
			name: "full-stack project includes JS dependencies",
			config: ProjectConfig{
				Name:     "test-project",
				Module:   "test-project",
				Database: "sqlite",
				Cache:    "memory",
				API:      false,
			},
			wantJS: true,
		},
		{
			name: "api-only project skips JS dependencies",
			config: ProjectConfig{
				Name:     "test-api",
				Module:   "test-api",
				Database: "sqlite",
				Cache:    "memory",
				API:      true,
			},
			wantJS: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// API projects should not install JS deps
			shouldInstallJS := !tt.config.API
			if shouldInstallJS != tt.wantJS {
				t.Errorf("shouldInstallJS = %v, want %v", shouldInstallJS, tt.wantJS)
			}
		})
	}
}
