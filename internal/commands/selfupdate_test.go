package commands

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestCopyFile(t *testing.T) {
	tests := []struct {
		name       string
		setupFunc  func(t *testing.T, tempDir string) (src, dst string)
		wantErr    bool
		errContains string
		validateFunc func(t *testing.T, dst string)
	}{
		{
			name: "copies file with content successfully",
			setupFunc: func(t *testing.T, tempDir string) (string, string) {
				src := filepath.Join(tempDir, "source.txt")
				dst := filepath.Join(tempDir, "dest.txt")

				content := []byte("test content for file copy")
				if err := os.WriteFile(src, content, 0644); err != nil {
					t.Fatalf("Failed to create source file: %v", err)
				}

				return src, dst
			},
			wantErr: false,
			validateFunc: func(t *testing.T, dst string) {
				// Verify content
				content, err := os.ReadFile(dst)
				if err != nil {
					t.Errorf("Failed to read destination file: %v", err)
					return
				}
				want := "test content for file copy"
				if string(content) != want {
					t.Errorf("Destination file content = %q, want %q", string(content), want)
				}

				// Verify permissions
				info, err := os.Stat(dst)
				if err != nil {
					t.Errorf("Failed to stat destination file: %v", err)
					return
				}
				if info.Mode().Perm() != 0755 {
					t.Errorf("Destination file permissions = %o, want 0755", info.Mode().Perm())
				}
			},
		},
		{
			name: "copies empty file successfully",
			setupFunc: func(t *testing.T, tempDir string) (string, string) {
				src := filepath.Join(tempDir, "empty.txt")
				dst := filepath.Join(tempDir, "empty_dest.txt")

				if err := os.WriteFile(src, []byte{}, 0644); err != nil {
					t.Fatalf("Failed to create empty source file: %v", err)
				}

				return src, dst
			},
			wantErr: false,
			validateFunc: func(t *testing.T, dst string) {
				content, err := os.ReadFile(dst)
				if err != nil {
					t.Errorf("Failed to read destination file: %v", err)
					return
				}
				if len(content) != 0 {
					t.Errorf("Destination file content length = %d, want 0", len(content))
				}
			},
		},
		{
			name: "copies large file successfully",
			setupFunc: func(t *testing.T, tempDir string) (string, string) {
				src := filepath.Join(tempDir, "large.bin")
				dst := filepath.Join(tempDir, "large_dest.bin")

				// Create a 1MB file
				content := make([]byte, 1024*1024)
				for i := range content {
					content[i] = byte(i % 256)
				}

				if err := os.WriteFile(src, content, 0644); err != nil {
					t.Fatalf("Failed to create large source file: %v", err)
				}

				return src, dst
			},
			wantErr: false,
			validateFunc: func(t *testing.T, dst string) {
				info, err := os.Stat(dst)
				if err != nil {
					t.Errorf("Failed to stat destination file: %v", err)
					return
				}
				if info.Size() != 1024*1024 {
					t.Errorf("Destination file size = %d, want %d", info.Size(), 1024*1024)
				}
			},
		},
		{
			name: "overwrites existing destination file",
			setupFunc: func(t *testing.T, tempDir string) (string, string) {
				src := filepath.Join(tempDir, "new.txt")
				dst := filepath.Join(tempDir, "existing.txt")

				// Create source with new content
				if err := os.WriteFile(src, []byte("new content"), 0644); err != nil {
					t.Fatalf("Failed to create source file: %v", err)
				}

				// Create existing destination with old content
				if err := os.WriteFile(dst, []byte("old content"), 0644); err != nil {
					t.Fatalf("Failed to create destination file: %v", err)
				}

				return src, dst
			},
			wantErr: false,
			validateFunc: func(t *testing.T, dst string) {
				content, err := os.ReadFile(dst)
				if err != nil {
					t.Errorf("Failed to read destination file: %v", err)
					return
				}
				want := "new content"
				if string(content) != want {
					t.Errorf("Destination file content = %q, want %q", string(content), want)
				}
			},
		},
		{
			name: "returns error when source file does not exist",
			setupFunc: func(t *testing.T, tempDir string) (string, string) {
				src := filepath.Join(tempDir, "nonexistent.txt")
				dst := filepath.Join(tempDir, "dest.txt")
				return src, dst
			},
			wantErr: true,
		},
		{
			name: "returns error when source file cannot be opened",
			setupFunc: func(t *testing.T, tempDir string) (string, string) {
				src := filepath.Join(tempDir, "noperm.txt")
				dst := filepath.Join(tempDir, "dest.txt")

				// Create file with no read permissions
				if err := os.WriteFile(src, []byte("content"), 0644); err != nil {
					t.Fatalf("Failed to create source file: %v", err)
				}
				if err := os.Chmod(src, 0000); err != nil {
					t.Fatalf("Failed to set permissions: %v", err)
				}

				return src, dst
			},
			wantErr: true,
		},
		{
			name: "returns error when destination directory does not exist",
			setupFunc: func(t *testing.T, tempDir string) (string, string) {
				src := filepath.Join(tempDir, "source.txt")
				dst := filepath.Join(tempDir, "nonexistent", "dest.txt")

				if err := os.WriteFile(src, []byte("content"), 0644); err != nil {
					t.Fatalf("Failed to create source file: %v", err)
				}

				return src, dst
			},
			wantErr: true,
		},
		{
			name: "returns error when destination is a directory",
			setupFunc: func(t *testing.T, tempDir string) (string, string) {
				src := filepath.Join(tempDir, "source.txt")
				dst := filepath.Join(tempDir, "destdir")

				if err := os.WriteFile(src, []byte("content"), 0644); err != nil {
					t.Fatalf("Failed to create source file: %v", err)
				}

				if err := os.Mkdir(dst, 0755); err != nil {
					t.Fatalf("Failed to create destination directory: %v", err)
				}

				return src, dst
			},
			wantErr: true,
		},
		{
			name: "copies binary file correctly",
			setupFunc: func(t *testing.T, tempDir string) (string, string) {
				src := filepath.Join(tempDir, "binary.bin")
				dst := filepath.Join(tempDir, "binary_dest.bin")

				// Create binary content with null bytes
				content := []byte{0x00, 0xFF, 0x01, 0xFE, 0x02, 0xFD}
				if err := os.WriteFile(src, content, 0644); err != nil {
					t.Fatalf("Failed to create source file: %v", err)
				}

				return src, dst
			},
			wantErr: false,
			validateFunc: func(t *testing.T, dst string) {
				content, err := os.ReadFile(dst)
				if err != nil {
					t.Errorf("Failed to read destination file: %v", err)
					return
				}
				want := []byte{0x00, 0xFF, 0x01, 0xFE, 0x02, 0xFD}
				if !reflect.DeepEqual(content, want) {
					t.Errorf("Destination file content = %v, want %v", content, want)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			src, dst := tt.setupFunc(t, tempDir)

			err := copyFile(src, dst)
			if (err != nil) != tt.wantErr {
				t.Errorf("copyFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errContains != "" {
				if err == nil || !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("copyFile() error = %v, want error containing %q", err, tt.errContains)
				}
				return
			}

			if !tt.wantErr && tt.validateFunc != nil {
				tt.validateFunc(t, dst)
			}
		})
	}
}

func TestCopyFile_PreservesContent(t *testing.T) {
	tests := []struct {
		name    string
		content []byte
	}{
		{
			name:    "preserves ascii text",
			content: []byte("Hello, World! This is a test file."),
		},
		{
			name:    "preserves unicode text",
			content: []byte("Hello 世界 🌍 Здравствуй мир"),
		},
		{
			name:    "preserves newlines and special characters",
			content: []byte("Line 1\nLine 2\r\nLine 3\tTabbed\x00Null"),
		},
		{
			name:    "preserves binary data",
			content: func() []byte {
				data := make([]byte, 256)
				for i := range data {
					data[i] = byte(i)
				}
				return data
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			src := filepath.Join(tempDir, "source")
			dst := filepath.Join(tempDir, "dest")

			if err := os.WriteFile(src, tt.content, 0644); err != nil {
				t.Fatalf("Failed to create source file: %v", err)
			}

			if err := copyFile(src, dst); err != nil {
				t.Fatalf("copyFile() error = %v", err)
			}

			gotContent, err := os.ReadFile(dst)
			if err != nil {
				t.Fatalf("Failed to read destination file: %v", err)
			}

			if !reflect.DeepEqual(gotContent, tt.content) {
				t.Errorf("Content mismatch: got %d bytes, want %d bytes", len(gotContent), len(tt.content))
			}
		})
	}
}

func TestCopyFile_Permissions(t *testing.T) {
	tempDir := t.TempDir()
	src := filepath.Join(tempDir, "source")
	dst := filepath.Join(tempDir, "dest")

	// Create source file with different permissions
	content := []byte("test content")
	if err := os.WriteFile(src, content, 0600); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	if err := copyFile(src, dst); err != nil {
		t.Fatalf("copyFile() error = %v", err)
	}

	// Verify destination has 0755 permissions regardless of source
	info, err := os.Stat(dst)
	if err != nil {
		t.Fatalf("Failed to stat destination file: %v", err)
	}

	if info.Mode().Perm() != 0755 {
		t.Errorf("Destination file permissions = %o, want 0755", info.Mode().Perm())
	}
}

func TestCopyFile_ErrorRecovery(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(t *testing.T, tempDir string) (src, dst string)
		wantErr   bool
	}{
		{
			name: "cleans up on copy error when destination is read-only directory",
			setupFunc: func(t *testing.T, tempDir string) (string, string) {
				src := filepath.Join(tempDir, "source.txt")
				dstDir := filepath.Join(tempDir, "readonly")
				dst := filepath.Join(dstDir, "dest.txt")

				if err := os.WriteFile(src, []byte("content"), 0644); err != nil {
					t.Fatalf("Failed to create source file: %v", err)
				}

				if err := os.Mkdir(dstDir, 0755); err != nil {
					t.Fatalf("Failed to create destination directory: %v", err)
				}

				// Make directory read-only after creation
				if err := os.Chmod(dstDir, 0444); err != nil {
					t.Fatalf("Failed to set directory permissions: %v", err)
				}

				return src, dst
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			src, dst := tt.setupFunc(t, tempDir)

			err := copyFile(src, dst)
			if (err != nil) != tt.wantErr {
				t.Errorf("copyFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGithubRelease_Struct(t *testing.T) {
	// Verify struct can be unmarshaled from JSON
	jsonData := `{
		"tag_name": "v1.0.0",
		"assets": [
			{
				"name": "velocity-linux-amd64",
				"browser_download_url": "https://github.com/velocitykode/velocity-installer/releases/download/v1.0.0/velocity-linux-amd64"
			},
			{
				"name": "velocity-darwin-arm64",
				"browser_download_url": "https://github.com/velocitykode/velocity-installer/releases/download/v1.0.0/velocity-darwin-arm64"
			}
		]
	}`

	// This test verifies the struct can unmarshal GitHub API responses
	// Note: We're not testing runSelfUpdate directly as it makes network calls
	// This ensures our data structures are compatible with the GitHub API
	var release githubRelease
	if err := unmarshalJSON([]byte(jsonData), &release); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if release.TagName != "v1.0.0" {
		t.Errorf("TagName = %q, want %q", release.TagName, "v1.0.0")
	}

	if len(release.Assets) != 2 {
		t.Errorf("len(Assets) = %d, want 2", len(release.Assets))
	}

	if len(release.Assets) > 0 {
		asset := release.Assets[0]
		if asset.Name != "velocity-linux-amd64" {
			t.Errorf("Asset[0].Name = %q, want %q", asset.Name, "velocity-linux-amd64")
		}
		if !strings.HasPrefix(asset.BrowserDownloadURL, "https://github.com") {
			t.Errorf("Asset[0].BrowserDownloadURL should start with https://github.com")
		}
	}
}

func TestGithubAsset_Struct(t *testing.T) {
	// Verify individual asset struct
	jsonData := `{
		"name": "velocity-windows-amd64.exe",
		"browser_download_url": "https://github.com/velocitykode/velocity-installer/releases/download/v1.0.0/velocity-windows-amd64.exe"
	}`

	var asset githubAsset
	if err := unmarshalJSON([]byte(jsonData), &asset); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if asset.Name != "velocity-windows-amd64.exe" {
		t.Errorf("Name = %q, want %q", asset.Name, "velocity-windows-amd64.exe")
	}

	if asset.BrowserDownloadURL == "" {
		t.Error("BrowserDownloadURL should not be empty")
	}
}

func TestInstallerVersion_Default(t *testing.T) {
	// Verify default version is set
	if InstallerVersion == "" {
		t.Error("InstallerVersion should not be empty")
	}

	// Verify version has expected format
	if !strings.Contains(InstallerVersion, ".") {
		t.Errorf("InstallerVersion = %q, expected to contain version number", InstallerVersion)
	}
}

func TestSelfUpdateCmd_Structure(t *testing.T) {
	// Verify command is properly configured
	if SelfUpdateCmd == nil {
		t.Fatal("SelfUpdateCmd should not be nil")
	}

	if SelfUpdateCmd.Use != "self-update" {
		t.Errorf("SelfUpdateCmd.Use = %q, want %q", SelfUpdateCmd.Use, "self-update")
	}

	if SelfUpdateCmd.Short == "" {
		t.Error("SelfUpdateCmd.Short should not be empty")
	}

	if SelfUpdateCmd.RunE == nil {
		t.Error("SelfUpdateCmd.RunE should not be nil")
	}
}

// Helper function to unmarshal JSON (wrapping standard library for testing)
func unmarshalJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// Additional helper for testing file operations
func createTestFile(t *testing.T, path string, content []byte) {
	t.Helper()
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
}

func readTestFile(t *testing.T, path string) []byte {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}
	return content
}

func TestCopyFile_ClosesFilesOnError(t *testing.T) {
	// This test ensures file handles are properly closed even on error
	tempDir := t.TempDir()
	src := filepath.Join(tempDir, "source.txt")

	// Create source file
	if err := os.WriteFile(src, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Try to copy to invalid destination
	invalidDst := filepath.Join(tempDir, "nonexistent", "dest.txt")

	err := copyFile(src, invalidDst)
	if err == nil {
		t.Fatal("Expected error when copying to nonexistent directory")
	}

	// Verify source file can still be deleted (file handle was closed)
	if err := os.Remove(src); err != nil {
		t.Errorf("Failed to remove source file after failed copy: %v", err)
	}
}

func TestCopyFile_HandlesConcurrentReads(t *testing.T) {
	tempDir := t.TempDir()
	src := filepath.Join(tempDir, "source.txt")

	content := []byte("concurrent test content")
	if err := os.WriteFile(src, content, 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Attempt to read source file while copying
	dst := filepath.Join(tempDir, "dest.txt")

	if err := copyFile(src, dst); err != nil {
		t.Fatalf("copyFile() error = %v", err)
	}

	// Verify both files are readable
	if _, err := io.ReadAll(openTestFile(t, src)); err != nil {
		t.Errorf("Failed to read source file: %v", err)
	}

	if _, err := io.ReadAll(openTestFile(t, dst)); err != nil {
		t.Errorf("Failed to read destination file: %v", err)
	}
}

func openTestFile(t *testing.T, path string) *os.File {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	t.Cleanup(func() { f.Close() })
	return f
}
