package detector

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/mod/modfile"
)

// ProjectInfo contains information about an existing Go project
type ProjectInfo struct {
	Path        string
	ModuleName  string
	GoVersion   string
	HasGoMod    bool
	HasVelocity bool
}

// Detect analyzes a directory to determine if it's a valid Go project
func Detect(dir string) (*ProjectInfo, error) {
	info := &ProjectInfo{
		Path: dir,
	}

	// Check for go.mod
	goModPath := filepath.Join(dir, "go.mod")
	data, err := os.ReadFile(goModPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("not a Go project: go.mod not found")
		}
		return nil, err
	}

	info.HasGoMod = true

	// Parse go.mod
	modFile, err := modfile.Parse("go.mod", data, nil)
	if err != nil {
		return nil, fmt.Errorf("invalid go.mod: %w", err)
	}

	info.ModuleName = modFile.Module.Mod.Path
	if modFile.Go != nil {
		info.GoVersion = modFile.Go.Version
	}

	// Check if Velocity already initialized
	info.HasVelocity = checkVelocityDirs(dir)

	return info, nil
}

// checkVelocityDirs checks if the project already has Velocity structure
func checkVelocityDirs(dir string) bool {
	// Check for key Velocity directories
	dirs := []string{"app", "config", "routes"}
	for _, d := range dirs {
		path := filepath.Join(dir, d)
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}
	return false
}
