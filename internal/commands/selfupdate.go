package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/velocitykode/velocity-installer/internal/ui"
)

var SelfUpdateCmd = &cobra.Command{
	Use:   "self-update",
	Short: "Update velocity installer to the latest version",
	RunE:  runSelfUpdate,
}

// Version is set by main.go
var InstallerVersion = "0.0.1"

type githubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

func runSelfUpdate(cmd *cobra.Command, args []string) error {
	ui.Header("self-update")

	ui.Step("Checking for updates...")

	// Fetch latest release
	resp, err := http.Get("https://api.github.com/repos/velocitykode/velocity-installer/releases/latest")
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		ui.Info("No releases found yet")
		return nil
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to check for updates: HTTP %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return fmt.Errorf("failed to parse release info: %w", err)
	}

	// Compare versions
	latestVersion := strings.TrimPrefix(release.TagName, "v")
	currentVersion := strings.TrimPrefix(InstallerVersion, "v")

	if latestVersion == currentVersion {
		ui.Success(fmt.Sprintf("Already up to date (v%s)", currentVersion))
		return nil
	}

	ui.Info(fmt.Sprintf("New version available: v%s (current: v%s)", latestVersion, currentVersion))

	// Find asset for current OS/arch
	assetName := fmt.Sprintf("velocity-%s-%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		assetName += ".exe"
	}

	var downloadURL string
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		return fmt.Errorf("no binary available for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	// Download new binary
	ui.Step("Downloading update...")
	resp, err = http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to download update: HTTP %d", resp.StatusCode)
	}

	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Create temp file
	tmpFile, err := os.CreateTemp("", "velocity-update-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write new binary to temp file
	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write update: %w", err)
	}
	tmpFile.Close()

	// Make executable
	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	// Replace current binary
	ui.Step("Installing update...")
	if err := os.Rename(tmpFile.Name(), execPath); err != nil {
		// Try copying if rename fails (cross-device)
		if err := copyFile(tmpFile.Name(), execPath); err != nil {
			return fmt.Errorf("failed to install update: %w", err)
		}
	}

	ui.Success(fmt.Sprintf("Updated to v%s", latestVersion))
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	return out.Chmod(0755)
}
