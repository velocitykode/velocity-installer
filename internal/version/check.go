package version

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

const MinimumGoVersion = "1.25"

// CheckGoVersion verifies that the system's installed Go version meets the minimum requirement.
// Returns an error if Go is not installed or version is too old.
func CheckGoVersion() error {
	return CheckMinimumGoVersion(MinimumGoVersion)
}

// CheckMinimumGoVersion checks if the system's installed Go version meets the specified minimum.
// The minimum should be in the format "1.25" (major.minor).
func CheckMinimumGoVersion(minimum string) error {
	current, err := getSystemGoVersion()
	if err != nil {
		return &GoNotFoundError{}
	}

	currentMajor, currentMinor, err := parseGoVersion(current)
	if err != nil {
		// If we can't parse the version, assume it's fine (development builds, etc.)
		return nil
	}

	minMajor, minMinor, err := parseGoVersion(minimum)
	if err != nil {
		return fmt.Errorf("invalid minimum version format: %s", minimum)
	}

	if currentMajor < minMajor || (currentMajor == minMajor && currentMinor < minMinor) {
		return &VersionError{
			Current: current,
			Minimum: minimum,
		}
	}

	return nil
}

// getSystemGoVersion runs "go version" and extracts the version string.
// Returns the version like "go1.25.1" or an error if Go is not installed.
func getSystemGoVersion() (string, error) {
	cmd := exec.Command("go", "version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// Output format: "go version go1.25.1 darwin/arm64"
	parts := strings.Fields(string(output))
	if len(parts) < 3 {
		return "", fmt.Errorf("unexpected go version output: %s", output)
	}

	return parts[2], nil // Returns "go1.25.1"
}

// parseGoVersion extracts major and minor version numbers from a Go version string.
// Handles formats like "go1.25", "go1.25.1", "1.25", "1.25.1"
func parseGoVersion(version string) (major, minor int, err error) {
	// Remove "go" prefix if present
	version = strings.TrimPrefix(version, "go")

	// Split by dots
	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		return 0, 0, fmt.Errorf("invalid version format: %s", version)
	}

	major, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid major version: %s", parts[0])
	}

	// Handle minor version that might have additional info (e.g., "25rc1", "25beta1")
	minorStr := parts[1]
	// Extract only the numeric part
	numericMinor := ""
	for _, c := range minorStr {
		if c >= '0' && c <= '9' {
			numericMinor += string(c)
		} else {
			break
		}
	}

	if numericMinor == "" {
		return 0, 0, fmt.Errorf("invalid minor version: %s", minorStr)
	}

	minor, err = strconv.Atoi(numericMinor)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid minor version: %s", numericMinor)
	}

	return major, minor, nil
}

// VersionError represents an error when the Go version is too old.
type VersionError struct {
	Current string
	Minimum string
}

func (e *VersionError) Error() string {
	return fmt.Sprintf(
		"Go version %s is not supported. Velocity requires Go %s or higher.\n\n"+
			"Please upgrade Go:\n"+
			"  brew upgrade go\n\n"+
			"Or download from: https://go.dev/dl/",
		e.Current, e.Minimum,
	)
}

// GoNotFoundError represents an error when Go is not installed.
type GoNotFoundError struct{}

func (e *GoNotFoundError) Error() string {
	return "Go is not installed or not in PATH. Velocity requires Go 1.25 or higher.\n\n" +
		"Please install Go:\n" +
		"  brew install go\n\n" +
		"Or download from: https://go.dev/dl/"
}
