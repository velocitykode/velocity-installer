package detector

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// IsVelocityProject checks if we're inside a Velocity project
func IsVelocityProject() bool {
	// Check for go.mod with velocity module
	if hasVelocityModule() {
		return true
	}

	// Check for .velocity marker file
	if _, err := os.Stat(".velocity"); err == nil {
		return true
	}

	// Check for velocity.yaml or velocity.toml config
	if _, err := os.Stat("velocity.yaml"); err == nil {
		return true
	}
	if _, err := os.Stat("velocity.toml"); err == nil {
		return true
	}

	// Check for typical Velocity project structure
	if hasVelocityStructure() {
		return true
	}

	return false
}

// hasVelocityModule checks if go.mod contains velocity module
func hasVelocityModule() bool {
	file, err := os.Open("go.mod")
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Check if it imports velocity framework
		if strings.Contains(line, "github.com/velocitykode/velocity") {
			return true
		}
		// Or if the module itself is a velocity app
		if strings.HasPrefix(line, "module ") && strings.Contains(line, "velocity") {
			// This might be too broad, but works for now
			return true
		}
	}

	return false
}

// hasVelocityStructure checks for typical Velocity project directories
func hasVelocityStructure() bool {
	// Check for Velocity-specific directories
	velocityDirs := []string{
		"app/controllers",
		"app/models",
		"routes",
		"database/migrations",
	}

	foundCount := 0
	for _, dir := range velocityDirs {
		if _, err := os.Stat(dir); err == nil {
			foundCount++
		}
	}

	// If we find at least 2 Velocity directories, assume it's a project
	return foundCount >= 2
}

// FindProjectRoot walks up the directory tree to find the project root
func FindProjectRoot() (string, bool) {
	dir, err := os.Getwd()
	if err != nil {
		return "", false
	}

	for {
		// Check if this is the project root
		if isProjectRoot(dir) {
			return dir, true
		}

		// Go up one directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			break
		}
		dir = parent
	}

	return "", false
}

func isProjectRoot(dir string) bool {
	oldDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(oldDir)

	return IsVelocityProject()
}
