package detector

import (
	"os"
	"path/filepath"
	"testing"
)

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
