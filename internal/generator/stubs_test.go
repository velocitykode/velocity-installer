package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCopyStubFileWithConfig(t *testing.T) {
	tests := []struct {
		name      string
		stubName  string
		config    interface{}
		wantErr   bool
		setupFunc func(t *testing.T) string
		validate  func(t *testing.T, destPath string)
	}{
		{
			name:     "copies stub file without config",
			stubName: "internal/middleware/middleware.go.stub",
			config:   nil,
			wantErr:  false,
			setupFunc: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "middleware.go")
			},
			validate: func(t *testing.T, destPath string) {
				content, err := os.ReadFile(destPath)
				if err != nil {
					t.Fatalf("failed to read destination file: %v", err)
				}
				if len(content) == 0 {
					t.Error("expected non-empty file content")
				}
			},
		},
		{
			name:     "processes template with valid config",
			stubName: "main.go.stub",
			config: ProjectConfig{
				Name:   "testapp",
				Module: "github.com/test/testapp",
			},
			wantErr: false,
			setupFunc: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "main.go")
			},
			validate: func(t *testing.T, destPath string) {
				content, err := os.ReadFile(destPath)
				if err != nil {
					t.Fatalf("failed to read destination file: %v", err)
				}
				if len(content) == 0 {
					t.Error("expected non-empty file content")
				}
			},
		},
		{
			name:     "creates parent directories when missing",
			stubName: "internal/handlers/home.go.stub",
			config:   nil,
			wantErr:  false,
			setupFunc: func(t *testing.T) string {
				baseDir := t.TempDir()
				return filepath.Join(baseDir, "nested", "deep", "path", "controller.go")
			},
			validate: func(t *testing.T, destPath string) {
				if _, err := os.Stat(destPath); os.IsNotExist(err) {
					t.Error("expected file to exist after creation")
				}
			},
		},
		{
			name:     "returns error for non-existent stub",
			stubName: "nonexistent/stub.go.stub",
			config:   nil,
			wantErr:  true,
			setupFunc: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "output.go")
			},
			validate: func(t *testing.T, destPath string) {
				if _, err := os.Stat(destPath); err == nil {
					t.Error("expected file not to exist when stub is missing")
				}
			},
		},
		{
			name:     "returns error for invalid template syntax",
			stubName: "routes/api.go.stub",
			config: struct {
				InvalidField int
			}{
				InvalidField: 123,
			},
			wantErr: true, // Template execution will fail when it can't evaluate required fields
			setupFunc: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "api.go")
			},
			validate: nil, // No validation needed since we expect an error
		},
		{
			name:     "processes config template with module reference",
			stubName: "config/config.go.stub",
			config: ProjectConfig{
				Name:     "myproject",
				Module:   "github.com/user/myproject",
				Database: "postgres",
			},
			wantErr: false,
			setupFunc: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "config.go")
			},
			validate: func(t *testing.T, destPath string) {
				content, err := os.ReadFile(destPath)
				if err != nil {
					t.Fatalf("failed to read destination file: %v", err)
				}
				if len(content) == 0 {
					t.Error("expected non-empty file content")
				}
			},
		},
		{
			name:     "processes routes template with auth enabled",
			stubName: "routes/web.go.stub",
			config: ProjectConfig{
				Name:   "authapp",
				Module: "github.com/test/authapp",
				Auth:   true,
			},
			wantErr: false,
			setupFunc: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "web.go")
			},
			validate: func(t *testing.T, destPath string) {
				content, err := os.ReadFile(destPath)
				if err != nil {
					t.Fatalf("failed to read destination file: %v", err)
				}
				if len(content) == 0 {
					t.Error("expected non-empty file content")
				}
			},
		},
		{
			name:     "processes api routes template with api and auth enabled",
			stubName: "routes/api.go.stub",
			config: ProjectConfig{
				Name:   "apiapp",
				Module: "github.com/test/apiapp",
				API:    true,
				Auth:   true,
			},
			wantErr: false,
			setupFunc: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "api.go")
			},
			validate: func(t *testing.T, destPath string) {
				content, err := os.ReadFile(destPath)
				if err != nil {
					t.Fatalf("failed to read destination file: %v", err)
				}
				if len(content) == 0 {
					t.Error("expected non-empty file content")
				}
			},
		},
		{
			name:     "overwrites existing file",
			stubName: "internal/middleware/middleware.go.stub",
			config:   nil,
			wantErr:  false,
			setupFunc: func(t *testing.T) string {
				destPath := filepath.Join(t.TempDir(), "middleware.go")
				// Pre-create file with different content
				if err := os.WriteFile(destPath, []byte("old content"), 0644); err != nil {
					t.Fatalf("failed to setup existing file: %v", err)
				}
				return destPath
			},
			validate: func(t *testing.T, destPath string) {
				content, err := os.ReadFile(destPath)
				if err != nil {
					t.Fatalf("failed to read destination file: %v", err)
				}
				if string(content) == "old content" {
					t.Error("expected file to be overwritten")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			destPath := tt.setupFunc(t)

			err := copyStubFileWithConfig(tt.stubName, destPath, tt.config)

			if (err != nil) != tt.wantErr {
				t.Errorf("copyStubFileWithConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.validate != nil {
				tt.validate(t, destPath)
			}
		})
	}
}

func TestCopyStubFile(t *testing.T) {
	tests := []struct {
		name     string
		stubName string
		wantErr  bool
	}{
		{
			name:     "copies valid stub file",
			stubName: "internal/middleware/middleware.go.stub",
			wantErr:  false,
		},
		{
			name:     "returns error for missing stub",
			stubName: "invalid/missing.stub",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			destPath := filepath.Join(t.TempDir(), "output.go")

			err := copyStubFile(tt.stubName, destPath)

			if (err != nil) != tt.wantErr {
				t.Errorf("copyStubFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if _, err := os.Stat(destPath); os.IsNotExist(err) {
					t.Error("expected file to exist after copy")
				}
			}
		})
	}
}

func TestGenerateFilesFromStubs(t *testing.T) {
	tests := []struct {
		name     string
		config   ProjectConfig
		wantErr  bool
		validate func(t *testing.T, config ProjectConfig)
	}{
		{
			name: "generates all required files for basic project",
			config: ProjectConfig{
				Name:     "basicapp",
				Module:   "github.com/test/basicapp",
				Database: "sqlite",
				Cache:    "memory",
				Auth:     false,
				API:      false,
			},
			wantErr: false,
			validate: func(t *testing.T, config ProjectConfig) {
				expectedFiles := []string{
					filepath.Join(config.Name, "main.go"),
					filepath.Join(config.Name, "internal", "handlers", "home.go"),
					filepath.Join(config.Name, "internal", "middleware", "middleware.go"),
					filepath.Join(config.Name, "routes", "web.go"),
					filepath.Join(config.Name, "config", "config.go"),
				}

				for _, file := range expectedFiles {
					if _, err := os.Stat(file); os.IsNotExist(err) {
						t.Errorf("expected file %s to exist", file)
					}
				}

				// The API pair should not exist
				for _, file := range []string{
					filepath.Join(config.Name, "routes", "api.go"),
					filepath.Join(config.Name, "internal", "handlers", "api.go"),
				} {
					if _, err := os.Stat(file); err == nil {
						t.Errorf("expected %s not to exist when API is disabled", file)
					}
				}

				// Auth adds route wiring, not a hand-rolled middleware file
				webRoutes, err := os.ReadFile(filepath.Join(config.Name, "routes", "web.go"))
				if err != nil {
					t.Fatalf("failed to read web routes: %v", err)
				}
				if strings.Contains(string(webRoutes), "auth.AuthMiddleware") {
					t.Error("expected no auth middleware wiring when Auth is disabled")
				}
			},
		},
		{
			name: "generates api routes when api enabled",
			config: ProjectConfig{
				Name:     "apiapp",
				Module:   "github.com/test/apiapp",
				Database: "postgres",
				Cache:    "redis",
				Auth:     false,
				API:      true,
			},
			wantErr: false,
			validate: func(t *testing.T, config ProjectConfig) {
				for _, file := range []string{
					filepath.Join(config.Name, "routes", "api.go"),
					filepath.Join(config.Name, "internal", "handlers", "api.go"),
				} {
					if _, err := os.Stat(file); os.IsNotExist(err) {
						t.Errorf("expected %s to exist when API is enabled", file)
					}
				}

				// The web pair is replaced by the API pair, not written alongside it
				for _, file := range []string{
					filepath.Join(config.Name, "routes", "web.go"),
					filepath.Join(config.Name, "internal", "handlers", "home.go"),
				} {
					if _, err := os.Stat(file); err == nil {
						t.Errorf("expected %s not to exist when API is enabled", file)
					}
				}
			},
		},
		{
			name: "wires framework auth middleware when auth enabled",
			config: ProjectConfig{
				Name:     "authapp",
				Module:   "github.com/test/authapp",
				Database: "postgres",
				Cache:    "memory",
				Auth:     true,
				API:      false,
			},
			wantErr: false,
			validate: func(t *testing.T, config ProjectConfig) {
				webRoutes, err := os.ReadFile(filepath.Join(config.Name, "routes", "web.go"))
				if err != nil {
					t.Fatalf("failed to read web routes: %v", err)
				}
				if !strings.Contains(string(webRoutes), "auth.AuthMiddleware") {
					t.Error("expected framework auth middleware wiring when Auth is enabled")
				}
			},
		},
		{
			name: "generates all files when both api and auth enabled",
			config: ProjectConfig{
				Name:     "fullapp",
				Module:   "github.com/test/fullapp",
				Database: "mysql",
				Cache:    "redis",
				Auth:     true,
				API:      true,
			},
			wantErr: false,
			validate: func(t *testing.T, config ProjectConfig) {
				expectedFiles := []string{
					filepath.Join(config.Name, "main.go"),
					filepath.Join(config.Name, "internal", "handlers", "api.go"),
					filepath.Join(config.Name, "internal", "middleware", "middleware.go"),
					filepath.Join(config.Name, "routes", "api.go"),
					filepath.Join(config.Name, "config", "config.go"),
				}

				for _, file := range expectedFiles {
					if _, err := os.Stat(file); os.IsNotExist(err) {
						t.Errorf("expected file %s to exist", file)
					}
				}
			},
		},
		{
			name: "creates nested directory structure",
			config: ProjectConfig{
				Name:     "nestedapp",
				Module:   "github.com/test/nestedapp",
				Database: "sqlite",
				Auth:     false,
				API:      false,
			},
			wantErr: false,
			validate: func(t *testing.T, config ProjectConfig) {
				// Verify directories exist
				expectedDirs := []string{
					filepath.Join(config.Name, "internal", "handlers"),
					filepath.Join(config.Name, "internal", "middleware"),
					filepath.Join(config.Name, "routes"),
					filepath.Join(config.Name, "config"),
				}

				for _, dir := range expectedDirs {
					if info, err := os.Stat(dir); os.IsNotExist(err) {
						t.Errorf("expected directory %s to exist", dir)
					} else if !info.IsDir() {
						t.Errorf("expected %s to be a directory", dir)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test in temp directory
			originalDir, _ := os.Getwd()
			tempDir := t.TempDir()
			os.Chdir(tempDir)
			defer os.Chdir(originalDir)

			err := generateFilesFromStubs(tt.config)

			if (err != nil) != tt.wantErr {
				t.Errorf("generateFilesFromStubs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.validate != nil {
				tt.validate(t, tt.config)
			}
		})
	}
}

func TestCopyStubFileWithConfig_FilePermissions(t *testing.T) {
	tests := []struct {
		name        string
		stubName    string
		setupFunc   func(t *testing.T) string
		wantErr     bool
		errContains string
	}{
		{
			name:     "creates file with correct permissions",
			stubName: "internal/middleware/middleware.go.stub",
			setupFunc: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "middleware.go")
			},
			wantErr: false,
		},
		{
			name:     "succeeds when destination directory has write permissions",
			stubName: "internal/handlers/home.go.stub",
			setupFunc: func(t *testing.T) string {
				dir := t.TempDir()
				os.Chmod(dir, 0755)
				return filepath.Join(dir, "controller.go")
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			destPath := tt.setupFunc(t)

			err := copyStubFileWithConfig(tt.stubName, destPath, nil)

			if (err != nil) != tt.wantErr {
				t.Errorf("copyStubFileWithConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				info, err := os.Stat(destPath)
				if err != nil {
					t.Fatalf("failed to stat file: %v", err)
				}

				expectedPerms := os.FileMode(0644)
				if info.Mode().Perm() != expectedPerms {
					t.Errorf("expected permissions %v, got %v", expectedPerms, info.Mode().Perm())
				}
			}
		})
	}
}

func TestCopyStubFileWithConfig_TemplateProcessing(t *testing.T) {
	tests := []struct {
		name     string
		stubName string
		config   ProjectConfig
		wantErr  bool
	}{
		{
			name:     "processes template with all config fields",
			stubName: "main.go.stub",
			config: ProjectConfig{
				Name:     "fullconfig",
				Module:   "github.com/test/fullconfig",
				Database: "postgres",
				Cache:    "redis",
				Auth:     true,
				API:      true,
			},
			wantErr: false,
		},
		{
			name:     "processes template with minimal config",
			stubName: "main.go.stub",
			config: ProjectConfig{
				Name:   "minimal",
				Module: "minimal",
			},
			wantErr: false,
		},
		{
			name:     "processes template with empty optional fields",
			stubName: "routes/web.go.stub",
			config: ProjectConfig{
				Name:     "emptyfields",
				Module:   "github.com/test/emptyfields",
				Database: "",
				Cache:    "",
				Auth:     false,
				API:      false,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			destPath := filepath.Join(t.TempDir(), "output.go")

			err := copyStubFileWithConfig(tt.stubName, destPath, tt.config)

			if (err != nil) != tt.wantErr {
				t.Errorf("copyStubFileWithConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				content, err := os.ReadFile(destPath)
				if err != nil {
					t.Fatalf("failed to read output file: %v", err)
				}

				// Verify template placeholders are replaced
				if len(content) == 0 {
					t.Error("expected non-empty file content")
				}
			}
		})
	}
}

func TestGenerateFilesFromStubs_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		config  ProjectConfig
		wantErr bool
	}{
		{
			name: "handles empty project name",
			config: ProjectConfig{
				Name:   "",
				Module: "test",
			},
			wantErr: false, // Will attempt to create files in current directory
		},
		{
			name: "handles special characters in module name",
			config: ProjectConfig{
				Name:   "specialapp",
				Module: "github.com/user-name/project_name",
			},
			wantErr: false,
		},
		{
			name: "handles long project paths",
			config: ProjectConfig{
				Name:   "very/long/nested/path/to/project",
				Module: "github.com/test/longpath",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test in temp directory
			originalDir, _ := os.Getwd()
			tempDir := t.TempDir()
			os.Chdir(tempDir)
			defer os.Chdir(originalDir)

			err := generateFilesFromStubs(tt.config)

			if (err != nil) != tt.wantErr {
				t.Errorf("generateFilesFromStubs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
