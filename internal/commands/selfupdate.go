package commands

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/velocitykode/prism"
)

var SelfUpdateCmd = &cobra.Command{
	Use:   "self-update",
	Short: "Update velocity installer to the latest version",
	RunE:  runSelfUpdate,
}

var InstallerVersion = "0.0.1"

type githubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

var (
	httpClient        = &http.Client{Timeout: 120 * time.Second}
	errReleaseMissing = errors.New("no releases found yet")
)

func runSelfUpdate(cmd *cobra.Command, args []string) error {
	prism.Header("self-update")

	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	if resolved, resolveErr := filepath.EvalSymlinks(execPath); resolveErr == nil {
		execPath = resolved
	}

	if isHomebrewInstall(execPath) {
		prism.Warning("Detected Homebrew install at " + execPath)
		prism.Muted("Use `brew upgrade --cask velocity` to update.")
		return nil
	}

	prism.Step("Checking for updates...")
	release, err := fetchLatestRelease()
	if errors.Is(err, errReleaseMissing) {
		prism.Info("No releases found yet")
		return nil
	}
	if err != nil {
		return err
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")
	currentVersion := strings.TrimPrefix(InstallerVersion, "v")
	if latestVersion == currentVersion {
		prism.Success(fmt.Sprintf("Already up to date (v%s)", currentVersion))
		return nil
	}
	prism.Info(fmt.Sprintf("New version available: v%s (current: v%s)", latestVersion, currentVersion))

	assetName := fmt.Sprintf("velocity-%s-%s.tar.gz", runtime.GOOS, runtime.GOARCH)
	var archiveURL, checksumURL string
	for _, a := range release.Assets {
		switch a.Name {
		case assetName:
			archiveURL = a.BrowserDownloadURL
		case "checksums.txt":
			checksumURL = a.BrowserDownloadURL
		}
	}
	if archiveURL == "" {
		return fmt.Errorf("no binary available for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	prism.Step("Downloading update...")
	archiveBytes, err := downloadBytes(archiveURL)
	if err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}

	if checksumURL == "" {
		prism.Warning("checksums.txt missing in release; skipping verification")
	} else {
		prism.Step("Verifying checksum...")
		checksumBytes, err := downloadBytes(checksumURL)
		if err != nil {
			return fmt.Errorf("failed to download checksums: %w", err)
		}
		expected, ok := lookupChecksum(checksumBytes, assetName)
		if !ok {
			return fmt.Errorf("checksum for %s not found in checksums.txt", assetName)
		}
		sum := sha256.Sum256(archiveBytes)
		got := hex.EncodeToString(sum[:])
		if !strings.EqualFold(got, expected) {
			return fmt.Errorf("checksum mismatch: got %s, want %s", got, expected)
		}
	}

	prism.Step("Extracting binary...")
	binaryBytes, err := extractBinary(archiveBytes, "velocity")
	if err != nil {
		return fmt.Errorf("failed to extract binary: %w", err)
	}

	prism.Step("Installing update...")
	if err := installBinary(execPath, binaryBytes); err != nil {
		return fmt.Errorf("failed to install update: %w", err)
	}

	if runtime.GOOS == "darwin" {
		_ = exec.Command("xattr", "-dr", "com.apple.quarantine", execPath).Run()
	}

	prism.Success(fmt.Sprintf("Updated to v%s", latestVersion))
	return nil
}

func fetchLatestRelease() (*githubRelease, error) {
	resp, err := httpClient.Get("https://api.github.com/repos/velocitykode/velocity-installer/releases/latest")
	if err != nil {
		return nil, fmt.Errorf("failed to check for updates: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, errReleaseMissing
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to check for updates: HTTP %d", resp.StatusCode)
	}
	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse release info: %w", err)
	}
	return &release, nil
}

func downloadBytes(url string) ([]byte, error) {
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func lookupChecksum(checksums []byte, assetName string) (string, bool) {
	scanner := bufio.NewScanner(bytes.NewReader(checksums))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 2 {
			continue
		}
		if filepath.Base(fields[1]) == assetName {
			return fields[0], true
		}
	}
	return "", false
}

func extractBinary(archive []byte, binaryName string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return nil, err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		if filepath.Base(hdr.Name) != binaryName {
			continue
		}
		return io.ReadAll(tr)
	}
	return nil, fmt.Errorf("binary %q not found in archive", binaryName)
}

func installBinary(execPath string, contents []byte) error {
	dir := filepath.Dir(execPath)
	tmp, err := os.CreateTemp(dir, "velocity-update-*")
	crossDevice := false
	if err != nil {
		tmp, err = os.CreateTemp("", "velocity-update-*")
		if err != nil {
			return err
		}
		crossDevice = true
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	if _, err := tmp.Write(contents); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpName, 0755); err != nil {
		return err
	}
	if !crossDevice {
		if err := os.Rename(tmpName, execPath); err == nil {
			return nil
		}
	}
	return copyFile(tmpName, execPath)
}

func isHomebrewInstall(path string) bool {
	prefixes := []string{
		"/opt/homebrew/",
		"/usr/local/Cellar/",
		"/usr/local/Caskroom/",
		"/home/linuxbrew/.linuxbrew/",
	}
	for _, p := range prefixes {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return strings.Contains(path, "/Caskroom/velocity/")
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
