package detector

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDetect(t *testing.T) {
	tests := []struct {
		name       string
		setupFunc  func(dir string) error
		wantErr    bool
		wantModule string
		wantGoVer  string
		wantHasVel bool
	}{
		{
			name: "detects valid Go project without Velocity",
			setupFunc: func(dir string) error {
				content := `module github.com/example/testproject

go 1.21

require (
	github.com/spf13/cobra v1.7.0
)`
				return os.WriteFile(filepath.Join(dir, "go.mod"), []byte(content), 0644)
			},
			wantErr:    false,
			wantModule: "github.com/example/testproject",
			wantGoVer:  "1.21",
			wantHasVel: false,
		},
		{
			name: "detects Go project with Velocity directories",
			setupFunc: func(dir string) error {
				content := `module github.com/example/velproject

go 1.22

require (
	github.com/velocitykode/velocity v0.1.0
)`
				if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(content), 0644); err != nil {
					return err
				}
				// Create Velocity directories
				for _, d := range []string{"app", "config", "routes"} {
					if err := os.Mkdir(filepath.Join(dir, d), 0755); err != nil {
						return err
					}
				}
				return nil
			},
			wantErr:    false,
			wantModule: "github.com/example/velproject",
			wantGoVer:  "1.22",
			wantHasVel: true,
		},
		{
			name: "detects Go project with partial Velocity structure",
			setupFunc: func(dir string) error {
				content := `module myapp

go 1.20`
				if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(content), 0644); err != nil {
					return err
				}
				// Create only one Velocity directory
				return os.Mkdir(filepath.Join(dir, "app"), 0755)
			},
			wantErr:    false,
			wantModule: "myapp",
			wantGoVer:  "1.20",
			wantHasVel: true,
		},
		{
			name: "returns error when go.mod not found",
			setupFunc: func(dir string) error {
				// Don't create go.mod
				return nil
			},
			wantErr: true,
		},
		{
			name: "returns error for invalid go.mod",
			setupFunc: func(dir string) error {
				content := `this is not valid go.mod syntax
module missing
invalid content`
				return os.WriteFile(filepath.Join(dir, "go.mod"), []byte(content), 0644)
			},
			wantErr: true,
		},
		{
			name: "handles go.mod without Go version",
			setupFunc: func(dir string) error {
				content := `module simple-module`
				return os.WriteFile(filepath.Join(dir, "go.mod"), []byte(content), 0644)
			},
			wantErr:    false,
			wantModule: "simple-module",
			wantGoVer:  "",
			wantHasVel: false,
		},
		{
			name: "detects project with config directory only",
			setupFunc: func(dir string) error {
				content := `module configapp

go 1.21`
				if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(content), 0644); err != nil {
					return err
				}
				return os.Mkdir(filepath.Join(dir, "config"), 0755)
			},
			wantErr:    false,
			wantModule: "configapp",
			wantGoVer:  "1.21",
			wantHasVel: true,
		},
		{
			name: "detects project with routes directory only",
			setupFunc: func(dir string) error {
				content := `module routeapp

go 1.21`
				if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(content), 0644); err != nil {
					return err
				}
				return os.Mkdir(filepath.Join(dir, "routes"), 0755)
			},
			wantErr:    false,
			wantModule: "routeapp",
			wantGoVer:  "1.21",
			wantHasVel: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			if err := tt.setupFunc(tempDir); err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			got, err := Detect(tempDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("Detect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if got.Path != tempDir {
				t.Errorf("Detect() Path = %v, want %v", got.Path, tempDir)
			}
			if got.ModuleName != tt.wantModule {
				t.Errorf("Detect() ModuleName = %v, want %v", got.ModuleName, tt.wantModule)
			}
			if got.GoVersion != tt.wantGoVer {
				t.Errorf("Detect() GoVersion = %v, want %v", got.GoVersion, tt.wantGoVer)
			}
			if !got.HasGoMod {
				t.Errorf("Detect() HasGoMod = false, want true")
			}
			if got.HasVelocity != tt.wantHasVel {
				t.Errorf("Detect() HasVelocity = %v, want %v", got.HasVelocity, tt.wantHasVel)
			}
		})
	}
}

func TestDetect_ErrorCases(t *testing.T) {
	tests := []struct {
		name    string
		dir     string
		wantErr string
	}{
		{
			name:    "returns error for non-existent directory",
			dir:     "/nonexistent/path/that/does/not/exist",
			wantErr: "not a Go project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Detect(tt.dir)
			if err == nil {
				t.Errorf("Detect() expected error, got nil")
				return
			}
			if tt.wantErr != "" && !contains(err.Error(), tt.wantErr) {
				t.Errorf("Detect() error = %v, want substring %v", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestCheckVelocityDirs(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(dir string) error
		want      bool
	}{
		{
			name: "returns true when app directory exists",
			setupFunc: func(dir string) error {
				return os.Mkdir(filepath.Join(dir, "app"), 0755)
			},
			want: true,
		},
		{
			name: "returns true when config directory exists",
			setupFunc: func(dir string) error {
				return os.Mkdir(filepath.Join(dir, "config"), 0755)
			},
			want: true,
		},
		{
			name: "returns true when routes directory exists",
			setupFunc: func(dir string) error {
				return os.Mkdir(filepath.Join(dir, "routes"), 0755)
			},
			want: true,
		},
		{
			name: "returns true when multiple directories exist",
			setupFunc: func(dir string) error {
				for _, d := range []string{"app", "config", "routes"} {
					if err := os.Mkdir(filepath.Join(dir, d), 0755); err != nil {
						return err
					}
				}
				return nil
			},
			want: true,
		},
		{
			name: "returns false when no Velocity directories exist",
			setupFunc: func(dir string) error {
				return os.Mkdir(filepath.Join(dir, "other"), 0755)
			},
			want: false,
		},
		{
			name: "returns false for empty directory",
			setupFunc: func(dir string) error {
				return nil
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			if err := tt.setupFunc(tempDir); err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			got := checkVelocityDirs(tempDir)
			if got != tt.want {
				t.Errorf("checkVelocityDirs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsVelocityProject(t *testing.T) {
	// Create temp directory
	tempDir := t.TempDir()

	// Save current dir and change to temp
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tempDir)

	// Test non-Velocity project (empty dir)
	if IsVelocityProject() {
		t.Error("Empty directory should not be detected as Velocity project")
	}

	// Create go.mod without velocity dependency
	goModContent := `module testproject

go 1.21

require (
	github.com/spf13/cobra v1.7.0
)`
	err := os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goModContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	if IsVelocityProject() {
		t.Error("Project without velocity dependency should not be detected as Velocity project")
	}

	// Create go.mod with velocity dependency
	goModWithVelocity := `module testproject

go 1.21

require (
	github.com/velocitykode/velocity v0.1.0
	github.com/spf13/cobra v1.7.0
)`
	err = os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goModWithVelocity), 0644)
	if err != nil {
		t.Fatal(err)
	}

	if !IsVelocityProject() {
		t.Error("Project with velocity dependency should be detected as Velocity project")
	}
}

func TestIsVelocityProjectWithInvalidPath(t *testing.T) {
	// Save current dir
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)

	// Change to non-existent path - this will fail, so test empty temp dir instead
	tempDir := t.TempDir()
	os.Chdir(tempDir)

	if IsVelocityProject() {
		t.Error("Empty directory should not be detected as Velocity project")
	}
}

func TestIsVelocityProject_MarkerFiles(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(dir string) error
		want      bool
	}{
		{
			name: "detects .vel marker file",
			setupFunc: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, ".vel"), []byte(""), 0644)
			},
			want: true,
		},
		{
			name: "detects velocity.yaml config",
			setupFunc: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, "velocity.yaml"), []byte(""), 0644)
			},
			want: true,
		},
		{
			name: "detects velocity.toml config",
			setupFunc: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, "velocity.toml"), []byte(""), 0644)
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			oldDir, _ := os.Getwd()
			defer os.Chdir(oldDir)
			os.Chdir(tempDir)

			if err := tt.setupFunc(tempDir); err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			got := IsVelocityProject()
			if got != tt.want {
				t.Errorf("IsVelocityProject() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasVelocityModule(t *testing.T) {
	tests := []struct {
		name         string
		goModContent string
		want         bool
	}{
		{
			name: "detects velocity framework import",
			goModContent: `module myapp

go 1.21

require (
	github.com/velocitykode/velocity v0.1.0
)`,
			want: true,
		},
		{
			name: "detects module name containing velocity",
			goModContent: `module github.com/myorg/velocity-app

go 1.21`,
			want: true,
		},
		{
			name: "returns false for non-velocity project",
			goModContent: `module myapp

go 1.21

require (
	github.com/spf13/cobra v1.7.0
)`,
			want: false,
		},
		{
			name:         "returns false when go.mod missing",
			goModContent: "",
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			oldDir, _ := os.Getwd()
			defer os.Chdir(oldDir)
			os.Chdir(tempDir)

			if tt.goModContent != "" {
				if err := os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(tt.goModContent), 0644); err != nil {
					t.Fatalf("failed to write go.mod: %v", err)
				}
			}

			got := hasVelocityModule()
			if got != tt.want {
				t.Errorf("hasVelocityModule() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasVelocityStructure(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(dir string) error
		want      bool
	}{
		{
			name: "detects when app/controllers exists",
			setupFunc: func(dir string) error {
				return os.MkdirAll(filepath.Join(dir, "app/controllers"), 0755)
			},
			want: false, // Only 1 directory, needs 2+
		},
		{
			name: "detects when app/models and routes exist",
			setupFunc: func(dir string) error {
				if err := os.MkdirAll(filepath.Join(dir, "app/models"), 0755); err != nil {
					return err
				}
				return os.Mkdir(filepath.Join(dir, "routes"), 0755)
			},
			want: true,
		},
		{
			name: "detects when all Velocity directories exist",
			setupFunc: func(dir string) error {
				dirs := []string{"app/controllers", "app/models", "routes", "database/migrations"}
				for _, d := range dirs {
					if err := os.MkdirAll(filepath.Join(dir, d), 0755); err != nil {
						return err
					}
				}
				return nil
			},
			want: true,
		},
		{
			name: "returns false when no Velocity directories exist",
			setupFunc: func(dir string) error {
				return nil
			},
			want: false,
		},
		{
			name: "returns false when only one Velocity directory exists",
			setupFunc: func(dir string) error {
				return os.Mkdir(filepath.Join(dir, "routes"), 0755)
			},
			want: false,
		},
		{
			name: "detects when database/migrations and app/controllers exist",
			setupFunc: func(dir string) error {
				if err := os.MkdirAll(filepath.Join(dir, "database/migrations"), 0755); err != nil {
					return err
				}
				return os.MkdirAll(filepath.Join(dir, "app/controllers"), 0755)
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			oldDir, _ := os.Getwd()
			defer os.Chdir(oldDir)
			os.Chdir(tempDir)

			if err := tt.setupFunc(tempDir); err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			got := hasVelocityStructure()
			if got != tt.want {
				t.Errorf("hasVelocityStructure() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindProjectRoot(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(baseDir string) (startDir string, expectedRoot string, err error)
		wantFound bool
	}{
		{
			name: "finds project root in current directory",
			setupFunc: func(baseDir string) (string, string, error) {
				// Create a Velocity project in baseDir
				goMod := `module myapp

go 1.21

require (
	github.com/velocitykode/velocity v0.1.0
)`
				if err := os.WriteFile(filepath.Join(baseDir, "go.mod"), []byte(goMod), 0644); err != nil {
					return "", "", err
				}
				return baseDir, baseDir, nil
			},
			wantFound: true,
		},
		{
			name: "finds project root in parent directory",
			setupFunc: func(baseDir string) (string, string, error) {
				// Create a Velocity project in baseDir
				goMod := `module myapp

go 1.21

require (
	github.com/velocitykode/velocity v0.1.0
)`
				if err := os.WriteFile(filepath.Join(baseDir, "go.mod"), []byte(goMod), 0644); err != nil {
					return "", "", err
				}
				// Create a subdirectory and start from there
				subDir := filepath.Join(baseDir, "app", "controllers")
				if err := os.MkdirAll(subDir, 0755); err != nil {
					return "", "", err
				}
				return subDir, baseDir, nil
			},
			wantFound: true,
		},
		{
			name: "finds project root two levels up",
			setupFunc: func(baseDir string) (string, string, error) {
				// Create a Velocity project in baseDir
				goMod := `module myapp

go 1.21

require (
	github.com/velocitykode/velocity v0.1.0
)`
				if err := os.WriteFile(filepath.Join(baseDir, "go.mod"), []byte(goMod), 0644); err != nil {
					return "", "", err
				}
				// Create nested subdirectories
				deepDir := filepath.Join(baseDir, "app", "controllers", "api")
				if err := os.MkdirAll(deepDir, 0755); err != nil {
					return "", "", err
				}
				return deepDir, baseDir, nil
			},
			wantFound: true,
		},
		{
			name: "returns false when no project root found",
			setupFunc: func(baseDir string) (string, string, error) {
				// Create a non-Velocity directory
				return baseDir, "", nil
			},
			wantFound: false,
		},
		{
			name: "finds project with .vel marker file",
			setupFunc: func(baseDir string) (string, string, error) {
				// Create .vel marker
				if err := os.WriteFile(filepath.Join(baseDir, ".vel"), []byte(""), 0644); err != nil {
					return "", "", err
				}
				subDir := filepath.Join(baseDir, "subdir")
				if err := os.Mkdir(subDir, 0755); err != nil {
					return "", "", err
				}
				return subDir, baseDir, nil
			},
			wantFound: true,
		},
		{
			name: "finds project with velocity.yaml config",
			setupFunc: func(baseDir string) (string, string, error) {
				// Create velocity.yaml
				if err := os.WriteFile(filepath.Join(baseDir, "velocity.yaml"), []byte(""), 0644); err != nil {
					return "", "", err
				}
				subDir := filepath.Join(baseDir, "nested", "path")
				if err := os.MkdirAll(subDir, 0755); err != nil {
					return "", "", err
				}
				return subDir, baseDir, nil
			},
			wantFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseDir := t.TempDir()
			startDir, expectedRoot, err := tt.setupFunc(baseDir)
			if err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			oldDir, _ := os.Getwd()
			defer os.Chdir(oldDir)
			os.Chdir(startDir)

			gotRoot, gotFound := FindProjectRoot()
			if gotFound != tt.wantFound {
				t.Errorf("FindProjectRoot() found = %v, want %v", gotFound, tt.wantFound)
				return
			}

			if tt.wantFound {
				// Resolve symlinks in both paths for comparison (macOS has /var -> /private/var)
				gotRootResolved, _ := filepath.EvalSymlinks(gotRoot)
				expectedRootResolved, _ := filepath.EvalSymlinks(expectedRoot)
				if gotRootResolved != expectedRootResolved {
					t.Errorf("FindProjectRoot() root = %v, want %v", gotRoot, expectedRoot)
				}
			}

			if !tt.wantFound && gotRoot != "" {
				t.Errorf("FindProjectRoot() should return empty string when not found, got %v", gotRoot)
			}
		})
	}
}

func TestIsProjectRoot(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(dir string) error
		want      bool
	}{
		{
			name: "returns true for Velocity project directory",
			setupFunc: func(dir string) error {
				goMod := `module myapp

go 1.21

require (
	github.com/velocitykode/velocity v0.1.0
)`
				return os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0644)
			},
			want: true,
		},
		{
			name: "returns false for non-Velocity directory",
			setupFunc: func(dir string) error {
				goMod := `module regularapp

go 1.21

require (
	github.com/spf13/cobra v1.7.0
)`
				return os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0644)
			},
			want: false,
		},
		{
			name: "returns false for empty directory",
			setupFunc: func(dir string) error {
				return nil
			},
			want: false,
		},
		{
			name: "returns true when .vel marker exists",
			setupFunc: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, ".vel"), []byte(""), 0644)
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			if err := tt.setupFunc(tempDir); err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			// Save and restore current directory
			oldDir, _ := os.Getwd()
			defer os.Chdir(oldDir)

			got := isProjectRoot(tempDir)
			if got != tt.want {
				t.Errorf("isProjectRoot() = %v, want %v", got, tt.want)
			}

			// Verify we're back in the original directory
			currentDir, _ := os.Getwd()
			if currentDir != oldDir {
				t.Errorf("isProjectRoot() did not restore working directory, got %v, want %v", currentDir, oldDir)
			}
		})
	}
}

func TestProjectInfo_Fields(t *testing.T) {
	// Test that ProjectInfo struct is properly populated
	info := &ProjectInfo{
		Path:        "/test/path",
		ModuleName:  "github.com/test/module",
		GoVersion:   "1.21",
		HasGoMod:    true,
		HasVelocity: true,
	}

	if info.Path != "/test/path" {
		t.Errorf("ProjectInfo.Path = %v, want /test/path", info.Path)
	}
	if info.ModuleName != "github.com/test/module" {
		t.Errorf("ProjectInfo.ModuleName = %v, want github.com/test/module", info.ModuleName)
	}
	if info.GoVersion != "1.21" {
		t.Errorf("ProjectInfo.GoVersion = %v, want 1.21", info.GoVersion)
	}
	if !info.HasGoMod {
		t.Errorf("ProjectInfo.HasGoMod = false, want true")
	}
	if !info.HasVelocity {
		t.Errorf("ProjectInfo.HasVelocity = false, want true")
	}
}

func TestDetect_IntegrationScenarios(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(dir string) error
		validate  func(t *testing.T, info *ProjectInfo)
	}{
		{
			name: "complex Velocity project with nested structure",
			setupFunc: func(dir string) error {
				goMod := `module github.com/example/complex-app

go 1.22

require (
	github.com/velocitykode/velocity v0.2.0
	github.com/spf13/cobra v1.8.0
)

replace github.com/velocitykode/velocity => ../velocity`
				if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0644); err != nil {
					return err
				}
				// Create full Velocity structure
				dirs := []string{
					"app/controllers",
					"app/models",
					"app/views",
					"config",
					"routes",
					"database/migrations",
					"database/seeds",
				}
				for _, d := range dirs {
					if err := os.MkdirAll(filepath.Join(dir, d), 0755); err != nil {
						return err
					}
				}
				return nil
			},
			validate: func(t *testing.T, info *ProjectInfo) {
				if info.ModuleName != "github.com/example/complex-app" {
					t.Errorf("ModuleName = %v, want github.com/example/complex-app", info.ModuleName)
				}
				if info.GoVersion != "1.22" {
					t.Errorf("GoVersion = %v, want 1.22", info.GoVersion)
				}
				if !info.HasVelocity {
					t.Error("HasVelocity should be true for project with Velocity structure")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			if err := tt.setupFunc(tempDir); err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			info, err := Detect(tempDir)
			if err != nil {
				t.Fatalf("Detect() unexpected error: %v", err)
			}

			tt.validate(t, info)
		})
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Benchmark tests
func BenchmarkDetect(b *testing.B) {
	tempDir := b.TempDir()
	goMod := `module benchmarkapp

go 1.21

require (
	github.com/velocitykode/velocity v0.1.0
)`
	os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goMod), 0644)
	os.Mkdir(filepath.Join(tempDir, "app"), 0755)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Detect(tempDir)
	}
}

func BenchmarkIsVelocityProject(b *testing.B) {
	tempDir := b.TempDir()
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tempDir)

	goMod := `module benchmarkapp

go 1.21

require (
	github.com/velocitykode/velocity v0.1.0
)`
	os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goMod), 0644)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IsVelocityProject()
	}
}

func BenchmarkFindProjectRoot(b *testing.B) {
	baseDir := b.TempDir()
	goMod := `module benchmarkapp

go 1.21

require (
	github.com/velocitykode/velocity v0.1.0
)`
	os.WriteFile(filepath.Join(baseDir, "go.mod"), []byte(goMod), 0644)

	deepDir := filepath.Join(baseDir, "a", "b", "c", "d")
	os.MkdirAll(deepDir, 0755)

	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(deepDir)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = FindProjectRoot()
	}
}

// Example tests for documentation
func ExampleDetect() {
	// Create a temporary directory for the example
	tempDir := "/tmp/example-project"

	// Detect project information
	info, err := Detect(tempDir)
	if err != nil {
		// Handle error
		return
	}

	// Use project information
	_ = info.ModuleName
	_ = info.HasVelocity
}

func ExampleFindProjectRoot() {
	// Find the Velocity project root from current directory
	root, found := FindProjectRoot()
	if found {
		// Use project root
		_ = root
	}
}

// Table-driven test to verify reflect.DeepEqual usage
func TestDetect_DeepEqual(t *testing.T) {
	tempDir := t.TempDir()
	goMod := `module testapp

go 1.21`
	os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goMod), 0644)

	got, err := Detect(tempDir)
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}

	want := &ProjectInfo{
		Path:        tempDir,
		ModuleName:  "testapp",
		GoVersion:   "1.21",
		HasGoMod:    true,
		HasVelocity: false,
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Detect() = %+v, want %+v", got, want)
	}
}
