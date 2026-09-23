package commands

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func releaseArchive(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for name, body := range files {
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o755, Size: int64(len(body)), Typeflag: tar.TypeReg}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestInstallLauncher(t *testing.T) {
	dir := t.TempDir()
	archive := releaseArchive(t, map[string]string{"velocity": "installer", launcherName(): "launcher"})

	path, err := installLauncher(archive, dir)
	if err != nil {
		t.Fatalf("installLauncher: %v", err)
	}
	if path != filepath.Join(dir, launcherName()) {
		t.Errorf("path = %s", path)
	}
	got, err := os.ReadFile(path)
	if err != nil || string(got) != "launcher" {
		t.Fatalf("contents = %q, err %v", got, err)
	}
	if runtime.GOOS != "windows" {
		if info, _ := os.Stat(path); info.Mode().Perm()&0o111 == 0 {
			t.Errorf("not executable: %v", info.Mode())
		}
	}
}

func TestInstallLauncher_ReplacesExisting(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, launcherName()), []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	archive := releaseArchive(t, map[string]string{launcherName(): "new"})
	path, err := installLauncher(archive, dir)
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(path); string(got) != "new" {
		t.Errorf("contents = %q, want new", got)
	}
}

func TestInstallLauncher_MissingFromOlderRelease(t *testing.T) {
	dir := t.TempDir()
	archive := releaseArchive(t, map[string]string{"velocity": "installer"})
	if _, err := installLauncher(archive, dir); !errors.Is(err, errLauncherMissing) {
		t.Fatalf("err = %v, want errLauncherMissing", err)
	}
	if _, err := os.Stat(filepath.Join(dir, launcherName())); !os.IsNotExist(err) {
		t.Errorf("launcher written for archive without one")
	}
}

func TestLauncherPresent(t *testing.T) {
	dir := t.TempDir()
	if launcherPresent(dir) {
		t.Fatal("reported present in empty dir")
	}
	if err := os.WriteFile(filepath.Join(dir, launcherName()), []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}
	if !launcherPresent(dir) {
		t.Fatal("not reported present")
	}
}
