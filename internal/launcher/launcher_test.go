package launcher

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// fakeProject lays out a Velocity project whose framework dependency is a local
// stub (via replace), so the test builds offline. Its main prints its args and cwd.
func fakeProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	write(t, filepath.Join(root, "stub", "go.mod"), "module github.com/velocitykode/velocity\n\ngo 1.22\n")
	write(t, filepath.Join(root, "stub", "stub.go"), "package velocity\n\nconst Name = \"stub\"\n")
	write(t, filepath.Join(root, "go.mod"), `module example.com/acme

go 1.22

require github.com/velocitykode/velocity v0.0.0

replace github.com/velocitykode/velocity => ./stub
`)
	write(t, filepath.Join(root, "main.go"), `package main

import (
	"fmt"
	"os"

	"github.com/velocitykode/velocity"
)

func main() {
	wd, _ := os.Getwd()
	fmt.Printf("%s v1 args=%v cwd=%s\n", velocity.Name, os.Args[1:], wd)
	if len(os.Args) > 1 && os.Args[1] == "fail" {
		os.Exit(3)
	}
}
`)
	return root
}

func TestFindProject(t *testing.T) {
	root := fakeProject(t)
	deep := filepath.Join(root, "internal", "models")
	if err := os.MkdirAll(deep, 0o755); err != nil {
		t.Fatal(err)
	}
	// A nested module that does not require the framework is skipped.
	write(t, filepath.Join(root, "tools", "go.mod"), "module example.com/acme/tools\n\ngo 1.22\n")

	want, _ := filepath.EvalSymlinks(root)
	for _, dir := range []string{root, deep, filepath.Join(root, "tools")} {
		got, err := FindProject(dir)
		if err != nil {
			t.Fatalf("FindProject(%s): %v", dir, err)
		}
		if g, _ := filepath.EvalSymlinks(got); g != want {
			t.Errorf("FindProject(%s) = %s, want %s", dir, got, root)
		}
	}
}

func TestFindProjectOutside(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "go.mod"), "module example.com/other\n\ngo 1.22\n")
	if _, err := FindProject(dir); !errors.Is(err, ErrNotInProject) {
		t.Fatalf("err = %v, want ErrNotInProject", err)
	}
}

// TestLauncher builds the real vel binary and drives it as a subprocess, since
// on Unix a successful hand-off replaces the process.
func TestLauncher(t *testing.T) {
	if testing.Short() {
		t.Skip("builds binaries")
	}
	velBin := filepath.Join(t.TempDir(), BinaryName())
	build := exec.Command("go", "build", "-o", velBin, "../../cmd/vel")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build vel: %v\n%s", err, out)
	}

	root := fakeProject(t)
	realRoot, _ := filepath.EvalSymlinks(root)
	sub := filepath.Join(root, "internal", "models")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}

	vel := func(dir string, args ...string) (string, int) {
		cmd := exec.Command(velBin, args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "GOWORK=off", "GOFLAGS=-mod=mod")
		out, err := cmd.CombinedOutput()
		code := 0
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			code = exitErr.ExitCode()
		} else if err != nil {
			t.Fatalf("run vel: %v", err)
		}
		return string(out), code
	}

	t.Run("builds and forwards from a subdirectory", func(t *testing.T) {
		out, code := vel(sub, "routes", "--json")
		if code != 0 {
			t.Fatalf("exit %d: %s", code, out)
		}
		if !strings.Contains(out, "v1 args=[routes --json]") {
			t.Errorf("args not forwarded: %q", out)
		}
		if !strings.Contains(out, "cwd="+realRoot) && !strings.Contains(out, "cwd="+root) {
			t.Errorf("not run from project root: %q", out)
		}
		if _, err := os.Stat(filepath.Join(root, BinaryName())); err != nil {
			t.Errorf("project binary not built: %v", err)
		}
	})

	t.Run("mirrors exit code", func(t *testing.T) {
		if _, code := vel(root, "fail"); code != 3 {
			t.Errorf("exit = %d, want 3", code)
		}
	})

	t.Run("rebuilds on source change", func(t *testing.T) {
		main := filepath.Join(root, "main.go")
		src, _ := os.ReadFile(main)
		write(t, main, strings.Replace(string(src), "v1", "v2", 1))
		out, _ := vel(root, "x")
		if !strings.Contains(out, "v2 args=[x]") {
			t.Errorf("stale binary ran: %q", out)
		}
	})

	t.Run("falls back to last good build on compile error", func(t *testing.T) {
		write(t, filepath.Join(root, "broken.go"), "package main\n\nfunc broken() { undefinedThing() }\n")
		defer os.Remove(filepath.Join(root, "broken.go"))
		out, code := vel(root, "x")
		if code != 0 {
			t.Fatalf("exit %d: %s", code, out)
		}
		if !strings.Contains(out, "running the last good build") || !strings.Contains(out, "v2 args=[x]") {
			t.Errorf("no fallback: %q", out)
		}
		if !strings.Contains(out, "undefinedThing") {
			t.Errorf("compiler output not shown: %q", out)
		}
	})

	t.Run("compile error with no previous build fails", func(t *testing.T) {
		os.Remove(filepath.Join(root, BinaryName()))
		write(t, filepath.Join(root, "broken.go"), "package main\n\nfunc broken() { undefinedThing() }\n")
		defer os.Remove(filepath.Join(root, "broken.go"))
		out, code := vel(root, "x")
		if code == 0 || !strings.Contains(out, "build failed") {
			t.Errorf("exit %d, out %q", code, out)
		}
	})

	t.Run("outside a project", func(t *testing.T) {
		out, code := vel(t.TempDir(), "serve")
		if code != 1 || !strings.Contains(out, "velocity new") {
			t.Errorf("exit %d, out %q", code, out)
		}
	})
}

func TestBuildLeavesNoTempFiles(t *testing.T) {
	if testing.Short() {
		t.Skip("builds binaries")
	}
	t.Setenv("GOWORK", "off")
	t.Setenv("GOFLAGS", "-mod=mod")
	root := fakeProject(t)
	if out, err := Build(root); err != nil {
		t.Fatalf("Build: %v\n%s", err, out)
	}
	if _, err := os.Stat(filepath.Join(root, BinaryName())); err != nil {
		t.Fatalf("binary not in place: %v", err)
	}
	left, _ := filepath.Glob(filepath.Join(root, ".vel", "tmp", "vel-build-*"))
	if len(left) != 0 {
		t.Errorf("temp builds left behind: %v", left)
	}

	write(t, filepath.Join(root, "broken.go"), "package main\n\nfunc broken() { undefinedThing() }\n")
	if _, err := Build(root); err == nil {
		t.Fatal("broken build succeeded")
	}
	left, _ = filepath.Glob(filepath.Join(root, ".vel", "tmp", "vel-build-*"))
	if len(left) != 0 {
		t.Errorf("temp builds left behind after failure: %v", left)
	}
}

func TestLauncherReportsMissingGo(t *testing.T) {
	if testing.Short() {
		t.Skip("builds binaries")
	}
	velBin := filepath.Join(t.TempDir(), BinaryName())
	if out, err := exec.Command("go", "build", "-o", velBin, "../../cmd/vel").CombinedOutput(); err != nil {
		t.Fatalf("build vel: %v\n%s", err, out)
	}
	root := fakeProject(t)
	cmd := exec.Command(velBin, "x")
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "PATH="+t.TempDir())
	out, _ := cmd.CombinedOutput()
	if !strings.Contains(string(out), "executable file not found") {
		t.Errorf("missing go not explained: %q", out)
	}
}
