// Package launcher implements the global `vel` command.
//
// vel is deliberately thin. It has exactly three jobs:
//
//  1. find the enclosing Velocity project (nearest go.mod requiring the framework),
//  2. build that project's own CLI binary (./vel) with the project's toolchain,
//  3. hand the command line to that binary.
//
// Every command (serve, migrate, routes, gen, package-provided commands) lives in
// the project's framework version. The launcher knows no command names, parses no
// flags, and prints no help of its own, so upgrading it can never change how a
// project behaves. It must not depend on the Velocity framework for the same reason.
package launcher

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"golang.org/x/mod/modfile"
)

// FrameworkModule is the module a go.mod must require to count as a Velocity project.
const FrameworkModule = "github.com/velocitykode/velocity"

// ErrNotInProject is returned when no enclosing Velocity project is found.
var ErrNotInProject = errors.New("not inside a Velocity project")

// FindProject walks up from dir to the nearest directory whose go.mod requires
// the Velocity framework. A go.mod that does not (a nested tool module, say) is
// skipped and the walk continues upward.
func FindProject(dir string) (string, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	for {
		gomod := filepath.Join(dir, "go.mod")
		if data, err := os.ReadFile(gomod); err == nil && requiresFramework(gomod, data) {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", ErrNotInProject
		}
		dir = parent
	}
}

func requiresFramework(path string, data []byte) bool {
	f, err := modfile.ParseLax(path, data, nil)
	if err != nil {
		return false
	}
	for _, r := range f.Require {
		if r.Mod.Path == FrameworkModule {
			return true
		}
	}
	return false
}

// BinaryName is the project CLI binary name for the current OS.
func BinaryName() string {
	if runtime.GOOS == "windows" {
		return "vel.exe"
	}
	return "vel"
}

// Build compiles the project's main package into root/vel. The user's
// environment passes through untouched: the project decides CGO, GOFLAGS, and
// toolchain, never the launcher. Compiler output is returned, not printed, so a
// successful (cached) build stays silent.
//
// The binary is built under .vel/tmp and renamed into place, so a concurrent
// build (another vel call, or `vel serve --watch` refreshing ./vel) never
// leaves a half-written ./vel for someone to execute.
func Build(root string) (output []byte, err error) {
	tmpDir := filepath.Join(root, ".vel", "tmp")
	if err := os.MkdirAll(tmpDir, 0o700); err != nil {
		return nil, err
	}
	buildDir, err := os.MkdirTemp(tmpDir, "vel-build-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(buildDir)
	tmpName := filepath.Join(buildDir, BinaryName())

	cmd := exec.Command("go", "build", "-o", tmpName, ".")
	cmd.Dir = root
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil {
		return buf.Bytes(), err
	}
	return buf.Bytes(), os.Rename(tmpName, filepath.Join(root, BinaryName()))
}

// Main runs the launcher and returns the process exit code. On Unix a
// successful hand-off never returns: the project binary replaces this process.
func Main(args []string, stderr io.Writer) int {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, "vel: %v\n", err)
		return 1
	}
	root, err := FindProject(cwd)
	if err != nil {
		fmt.Fprintf(stderr, "vel: %v (no go.mod requiring %s found in this directory or any parent).\n", err, FrameworkModule)
		fmt.Fprintln(stderr, "Create a project with: velocity new <name>")
		return 1
	}

	bin := filepath.Join(root, BinaryName())
	_, statErr := os.Stat(bin)
	haveLastGood := statErr == nil

	if out, err := Build(root); err != nil {
		stderr.Write(out)
		if !haveLastGood {
			fmt.Fprintf(stderr, "vel: build failed: %v\n", err)
			return 1
		}
		if len(out) == 0 {
			// No compiler output (go missing from PATH, rename failed): say why.
			fmt.Fprintf(stderr, "vel: %v\n", err)
		}
		fmt.Fprintln(stderr, "vel: build failed, running the last good build (output may reflect old code)")
	}

	// Commands expect to run from the project root (.env, relative paths),
	// exactly as `./vel` does today.
	if err := os.Chdir(root); err != nil {
		fmt.Fprintf(stderr, "vel: %v\n", err)
		return 1
	}
	return run(bin, args, stderr)
}
