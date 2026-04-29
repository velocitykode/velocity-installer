package generator

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	cli "github.com/velocitykode/velocity-cli"
	"github.com/velocitykode/velocity-installer/internal/ui"
)

// Fallback version if GitHub API is unavailable
const fallbackVelocityVersion = "v0.20.3"

// semverTagRE matches plain semver-style tags (vX.Y.Z) - pre-releases
// like vX.Y.Z-rc1 are excluded so the installer never picks an
// unfinished tag.
var semverTagRE = regexp.MustCompile(`^v(\d+)\.(\d+)\.(\d+)$`)

// getLatestVelocityVersion fetches the highest semver tag from the
// velocity repo. Tags are the source of truth - GitHub Releases are
// ceremonial and may lag (or never be created). Filtering to vX.Y.Z
// also guards against draft/pre-release tags.
func getLatestVelocityVersion() string {
	client := &http.Client{Timeout: 5 * time.Second}
	// 100 tags is the per-page max; the auto-release cadence makes
	// the latest plain semver tag fall comfortably inside the first
	// page, even when patch releases stack up.
	resp, err := client.Get("https://api.github.com/repos/velocitykode/velocity/tags?per_page=100")
	if err != nil {
		return fallbackVelocityVersion
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fallbackVelocityVersion
	}

	var tags []struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return fallbackVelocityVersion
	}

	names := make([]string, 0, len(tags))
	for _, t := range tags {
		names = append(names, t.Name)
	}
	if best := pickHighestSemverTag(names); best != "" {
		return best
	}
	return fallbackVelocityVersion
}

// pickHighestSemverTag returns the highest vX.Y.Z tag from names, or
// "" if none qualifies. Pre-release suffixes (-rc1, -beta) are
// excluded so an unfinished tag never wins.
func pickHighestSemverTag(names []string) string {
	var bestName string
	var bestMajor, bestMinor, bestPatch int
	for _, name := range names {
		m := semverTagRE.FindStringSubmatch(name)
		if m == nil {
			continue
		}
		major, minor, patch := atoi(m[1]), atoi(m[2]), atoi(m[3])
		if bestName == "" ||
			major > bestMajor ||
			(major == bestMajor && minor > bestMinor) ||
			(major == bestMajor && minor == bestMinor && patch > bestPatch) {
			bestName, bestMajor, bestMinor, bestPatch = name, major, minor, patch
		}
	}
	return bestName
}

// atoi parses a regex-captured digit group. The regex guarantees a
// digits-only match so strconv-style error handling is unnecessary.
func atoi(s string) int {
	n := 0
	for _, c := range s {
		n = n*10 + int(c-'0')
	}
	return n
}

// ProjectConfig holds the configuration for a new project
type ProjectConfig struct {
	Name     string
	Module   string
	Database string
	Cache    string
	Auth     bool
	API      bool
	// SSR toggles Inertia server-side rendering in the generated app.
	// When true the installer enables INERTIA_SSR_ENABLED=true and
	// points INERTIA_SSR_URL at Vite's /__inertia_ssr dev endpoint.
	// When false the installer adds ssr:false to vite.config.ts so the
	// @inertiajs/vite plugin stays quiet in dev.
	SSR bool
}

// CreateProject generates a new Velocity project from template
func CreateProject(config ProjectConfig) error {
	// Validate project name
	if err := validateProjectName(config.Name); err != nil {
		return err
	}

	// Determine module name
	moduleName := config.Module
	if moduleName == "" {
		moduleName = config.Name
	}

	cli.Info("Creating new Velocity project")

	// Clone template
	if err := cli.Spinner("Cloning template", func() error {
		return cloneTemplate(config.Name, config.API)
	}); err != nil {
		return fmt.Errorf("failed to clone template: %w", err)
	}
	cli.Success("Template cloned")

	// Replace module name in all files
	if err := cli.Spinner("Configuring module", func() error {
		return replaceModuleName(config.Name, moduleName)
	}); err != nil {
		return fmt.Errorf("failed to configure project: %w", err)
	}
	cli.Success("Module configured")

	// Remove template git history and initialize new repo
	if err := cli.Spinner("Initializing Git", func() error {
		return reinitGitRepo(config.Name)
	}); err != nil {
		return fmt.Errorf("failed to initialize git: %w", err)
	}
	cli.Success("Git initialized")

	// Create default migrations
	if err := createDefaultMigrations(config.Name); err != nil {
		return fmt.Errorf("failed to create migrations: %w", err)
	}
	cli.Success("Migrations created")

	// Create proper .env.example with database config
	if err := createEnvFiles(config); err != nil {
		return fmt.Errorf("failed to create env files: %w", err)
	}
	cli.Success("Environment configured")

	cli.Newline()
	cli.Info("Installing dependencies")
	// installDependencies now also runs `go build -o vel .` concurrently
	// with the JS install. APP_KEY is already written to .env by
	// createEnvFiles, so `./vel` can bootstrap cleanly.
	if err := installDependencies(config.Name, config.API); err != nil {
		return fmt.Errorf("failed to install dependencies: %w", err)
	}

	cli.Newline()
	ready, err := ensureDatabaseReady(config.Name)
	if err != nil {
		return fmt.Errorf("failed to prepare database: %w", err)
	}
	if !ready {
		// Preflight already printed a user-facing message. Installation is
		// otherwise complete - the caller decides whether to auto-start.
		return ErrMigrationsSkipped
	}

	cli.Newline()
	cli.Info("Running migrations")
	if err := runVel(config.Name, "migrate"); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	cli.Success("Migrations complete")

	return nil
}

// runVel invokes the project's own vel binary with the given args and
// streams its output through so the framework's console UI stays intact.
func runVel(projectPath string, args ...string) error {
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return err
	}
	cmd := exec.Command("./vel", args...)
	cmd.Dir = absPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// ErrMigrationsSkipped signals that project scaffolding succeeded but the
// database preflight chose to skip migrations (e.g. the DB server wasn't
// reachable). Callers should treat this as a non-fatal outcome.
var ErrMigrationsSkipped = errors.New("migrations skipped: database not ready")

// cloneTemplate downloads and extracts the appropriate velocity template
// as a tarball from GitHub. Faster than git clone (no .git/ payload, no
// pack-file assembly) and the template's git history is discarded by
// reinitGitRepo anyway, so nothing is lost. Falls back to git clone when
// the HTTP fetch fails (corporate proxy, offline mirror, etc.).
func cloneTemplate(projectName string, apiOnly bool) error {
	templateRepo := "velocity-template-react"
	if apiOnly {
		templateRepo = "velocity-template-api"
	}

	if err := downloadTemplateTarball(templateRepo, projectName); err == nil {
		return nil
	}

	// Fallback: git clone (keeps the tool working in environments that
	// block codeload.github.com but allow ssh/https git).
	cmd := exec.Command("git", "clone", "--depth=1", "git@github.com:velocitykode/"+templateRepo+".git", projectName)
	if err := cmd.Run(); err == nil {
		return nil
	}
	cmd = exec.Command("git", "clone", "--depth=1", "https://github.com/velocitykode/"+templateRepo+".git", projectName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to fetch template (tarball and git clone both failed): %w", err)
	}
	return nil
}

// downloadTemplateTarball streams codeload.github.com's .tar.gz of the
// template's main branch into projectName, stripping the single top-level
// directory that GitHub wraps every tarball in.
func downloadTemplateTarball(repo, projectName string) error {
	url := fmt.Sprintf("https://codeload.github.com/velocitykode/%s/tar.gz/refs/heads/main", repo)
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("tarball fetch: HTTP %d", resp.StatusCode)
	}

	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		return err
	}
	defer gz.Close()

	if err := os.MkdirAll(projectName, 0o755); err != nil {
		return err
	}
	root, err := filepath.Abs(projectName)
	if err != nil {
		return err
	}

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Strip the single leading directory GitHub wraps tarballs in
		// (e.g. "velocity-template-react-main/"). Skip the top-level entry.
		parts := strings.SplitN(hdr.Name, "/", 2)
		if len(parts) < 2 || parts[1] == "" {
			continue
		}
		dest := filepath.Join(root, parts[1])

		// Path traversal guard: every entry must stay inside root.
		if !strings.HasPrefix(dest, root+string(os.PathSeparator)) && dest != root {
			return fmt.Errorf("refusing tar entry outside project dir: %s", hdr.Name)
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(dest, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
				return err
			}
			f, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode)&0o777)
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return err
			}
			if err := f.Close(); err != nil {
				return err
			}
		}
		// Symlinks, char/block devices, etc. are intentionally skipped.
	}
	return nil
}

// replaceModuleName replaces {{MODULE_NAME}} in all files
func replaceModuleName(projectPath, moduleName string) error {
	// Get absolute path
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return fmt.Errorf("abs path: %w", err)
	}

	// Portable in-place edits: `sed PATTERN file > tmp && mv tmp file`.
	// Works identically on GNU (Linux) and BSD (macOS) sed without -i quirks.
	cmd := exec.Command("sh", "-c",
		fmt.Sprintf(`cd '%s' && find . -name '*.go' -type f -exec sh -c 'sed "s|{{MODULE_NAME}}|%s|g" "$1" > "$1.tmp" && mv "$1.tmp" "$1"' _ {} \;`, absPath, moduleName))
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("replace go files: %w: %s", err, string(output))
	}

	// Replace in go.mod (if exists)
	cmd = exec.Command("sh", "-c",
		fmt.Sprintf("cd '%s' && [ -f go.mod ] && sed 's|{{MODULE_NAME}}|%s|g' go.mod > go.mod.tmp && mv go.mod.tmp go.mod || true", absPath, moduleName))
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("replace go.mod: %w: %s", err, string(output))
	}

	// Replace in package.json (if exists)
	cmd = exec.Command("sh", "-c",
		fmt.Sprintf("cd '%s' && [ -f package.json ] && sed 's|{{MODULE_NAME}}|%s|g' package.json > package.json.tmp && mv package.json.tmp package.json || true", absPath, moduleName))
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("replace package.json: %w: %s", err, string(output))
	}

	// Replace in package-lock.json (if exists)
	cmd = exec.Command("sh", "-c",
		fmt.Sprintf("cd '%s' && [ -f package-lock.json ] && sed 's|{{MODULE_NAME}}|%s|g' package-lock.json > package-lock.json.tmp && mv package-lock.json.tmp package-lock.json || true", absPath, moduleName))
	cmd.Run() // Ignore error - file may not exist

	// Only process go.mod if it exists
	goModPath := filepath.Join(absPath, "go.mod")
	if _, err := os.Stat(goModPath); err == nil {
		// Remove replace directive
		cmd = exec.Command("sh", "-c",
			fmt.Sprintf("cd '%s' && sed '/^replace github.com\\/velocitykode\\/velocity/d' go.mod > go.mod.tmp && mv go.mod.tmp go.mod", absPath))
		if err := cmd.Run(); err != nil {
			return err
		}

		// Set pinned version of velocity framework (fetched from GitHub releases)
		velocityVersion := getLatestVelocityVersion()
		cmd = exec.Command("go", "mod", "edit", fmt.Sprintf("-require=github.com/velocitykode/velocity@%s", velocityVersion))
		cmd.Dir = absPath
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to set velocity framework version: %w", err)
		}
	}

	return nil
}

// reinitGitRepo removes template git history and creates new repo
func reinitGitRepo(projectPath string) error {
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return err
	}

	// Verify project directory exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("project directory does not exist: %s", absPath)
	}

	// Remove .git directory
	gitDir := filepath.Join(absPath, ".git")
	if err := os.RemoveAll(gitDir); err != nil {
		return err
	}

	// Initialize new git repo
	originalDir, _ := os.Getwd()
	if err := os.Chdir(absPath); err != nil {
		return fmt.Errorf("failed to change to project directory: %w", err)
	}
	defer os.Chdir(originalDir)

	exec.Command("git", "init").Run()
	exec.Command("git", "add", ".").Run()
	exec.Command("git", "commit", "-m", "Initial commit").Run()

	return nil
}

// depGroup is the live state of a single dependency install task, as
// displayed in the tree. tail holds the most recent N parsed package
// names so the user sees deps ticking past under each group header.
// All fields are protected by mu so the render loop and the streaming
// goroutines can touch them without racing.
type depGroup struct {
	mu    sync.Mutex
	name  string
	tail  []string
	count int
	done  bool
	err   error
}

const depTailMax = 5

func (g *depGroup) snapshot() (name string, tail []string, count int, done bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	t := make([]string, len(g.tail))
	copy(t, g.tail)
	return g.name, t, g.count, g.done
}

func (g *depGroup) push(pkg string) {
	g.mu.Lock()
	g.tail = append(g.tail, pkg)
	if len(g.tail) > depTailMax {
		g.tail = g.tail[len(g.tail)-depTailMax:]
	}
	g.count++
	g.mu.Unlock()
}

func (g *depGroup) markDone(success bool, err error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.err = err
	g.done = success
}

var (
	ansiRE       = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
	goDownloadRE = regexp.MustCompile(`^go: downloading (\S+) (\S+)`)
	bunAddRE     = regexp.MustCompile(`^\+\s+(\S+)`)
)

// parsePackageLine normalizes an output line for display. Known
// per-package formats (go: downloading ..., + pkg@ver) are cleaned
// up; unrecognized lines fall through as-is so the user still sees
// activity (e.g. bun's "Resolving dependencies" during slow phases).
// Returns "" only for blank lines.
func parsePackageLine(raw string) string {
	line := ansiRE.ReplaceAllString(raw, "")
	line = strings.TrimSpace(line)
	if line == "" {
		return ""
	}
	if m := goDownloadRE.FindStringSubmatch(line); m != nil {
		return m[1] + " " + m[2]
	}
	if m := bunAddRE.FindStringSubmatch(line); m != nil {
		return m[1]
	}
	return line
}

// runWithStreamedStatus runs cmd and streams its stdout+stderr through
// parsePackageLine into group.push. Blocks until the command exits.
func runWithStreamedStatus(cmd *exec.Cmd, group *depGroup) error {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	var wg sync.WaitGroup
	scan := func(r io.Reader) {
		defer wg.Done()
		sc := bufio.NewScanner(r)
		sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for sc.Scan() {
			if pkg := parsePackageLine(sc.Text()); pkg != "" {
				group.push(pkg)
			}
		}
	}
	wg.Add(2)
	go scan(stdout)
	go scan(stderr)

	waitErr := cmd.Wait()
	wg.Wait()
	return waitErr
}

// renderGroup prints a single group (header + tail sub-lines) and
// returns the number of terminal lines written. withPipe controls
// whether the sub-lines use a │ continuation (for non-last groups).
func renderGroup(prefix, name string, tail []string, count int, done bool, withPipe bool) int {
	status := "downloading…"
	if done {
		if count == 0 {
			status = "done"
		} else {
			status = fmt.Sprintf("done (%d packages)", count)
		}
	} else if count > 0 {
		status = fmt.Sprintf("downloading… (%d so far)", count)
	}
	ui.TreeItem(prefix, name, status, done)
	lines := 1

	if done {
		return lines
	}

	childPrefix := "    "
	if withPipe {
		childPrefix = "│   "
	}
	for _, pkg := range tail {
		fmt.Printf("  %s %s %s\n",
			cli.StyleMuted(childPrefix),
			cli.StyleMuted("•"),
			cli.StyleMuted(truncate(pkg, 68)),
		)
		lines++
	}
	return lines
}

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n-1]) + "…"
}

// installDependencies runs go mod tidy and bun/npm install in parallel,
// rendering each resolved package under its group in real time.
func installDependencies(projectPath string, apiOnly bool) error {
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return err
	}

	originalDir, _ := os.Getwd()
	os.Chdir(absPath)
	defer os.Chdir(originalDir)

	goGroup := &depGroup{name: "Go dependencies"}
	jsGroup := &depGroup{name: "JS dependencies"}
	velGroup := &depGroup{name: "Build vel"}

	// If we're going to fall back to npm, tell the user about bun first.
	// Printing before the tree so it stays above the in-place redraws.
	if !apiOnly {
		if _, err := exec.LookPath("bun"); err != nil {
			cli.Tip("Install bun for much faster JS installs → https://bun.sh")
		}
	}

	// Tree layout:
	//   api-only: Go deps + Build vel
	//   full:     Go deps + JS deps + Build vel
	// Go build has no filesystem dependency on node_modules, so the vel
	// build runs concurrently with bun/npm install once `go mod tidy`
	// has finished writing go.sum.
	printDepTree := func() int {
		lines := 0
		n1, t1, c1, d1 := goGroup.snapshot()
		if apiOnly {
			lines += renderGroup("├─", n1, t1, c1, d1, true)
			nV, tV, cV, dV := velGroup.snapshot()
			lines += renderGroup("└─", nV, tV, cV, dV, false)
			return lines
		}
		lines += renderGroup("├─", n1, t1, c1, d1, true)
		n2, t2, c2, d2 := jsGroup.snapshot()
		lines += renderGroup("├─", n2, t2, c2, d2, true)
		nV, tV, cV, dV := velGroup.snapshot()
		lines += renderGroup("└─", nV, tV, cV, dV, false)
		return lines
	}

	linesPrinted := printDepTree()

	// Tasks: Go deps, Build vel, (optionally) JS deps
	numTasks := 2
	if !apiOnly {
		numTasks = 3
	}

	done := make(chan bool, numTasks)
	goFinished := make(chan error, 1)

	// Go dependencies - signals goFinished so the vel build can start.
	go func() {
		cmd := exec.Command("go", "mod", "tidy")
		err := runWithStreamedStatus(cmd, goGroup)
		goGroup.markDone(err == nil, err)
		goFinished <- err
		done <- true
	}()

	// Build vel - waits for `go mod tidy` so go.sum is consistent, then
	// compiles concurrently with the still-running JS install.
	go func() {
		if err := <-goFinished; err != nil {
			velGroup.markDone(false, fmt.Errorf("skipped: go deps failed"))
			done <- true
			return
		}
		cmd := exec.Command("go", "build", "-o", "vel", ".")
		err := runWithStreamedStatus(cmd, velGroup)
		velGroup.markDone(err == nil, err)
		done <- true
	}()

	// JS dependencies (skip for API-only projects)
	if !apiOnly {
		go func() {
			var pm string
			if _, err := exec.LookPath("bun"); err == nil {
				pm = "bun"
			} else if _, err := exec.LookPath("npm"); err == nil {
				pm = "npm"
			} else {
				jsGroup.markDone(false, fmt.Errorf("neither bun nor npm was found in PATH"))
				done <- true
				return
			}
			cmd := exec.Command(pm, "install")
			err := runWithStreamedStatus(cmd, jsGroup)
			jsGroup.markDone(err == nil, err)
			done <- true
		}()
	}

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	completed := 0
	for completed < numTasks {
		select {
		case <-done:
			completed++
		case <-ticker.C:
		}
		ui.ClearLines(linesPrinted)
		linesPrinted = printDepTree()
	}

	if goGroup.err != nil {
		return goGroup.err
	}
	if !apiOnly && jsGroup.err != nil {
		return jsGroup.err
	}
	if velGroup.err != nil {
		return velGroup.err
	}

	return nil
}

func validateProjectName(name string) error {
	if name == "" {
		return fmt.Errorf("project name cannot be empty")
	}

	// Check for invalid characters
	if strings.ContainsAny(name, " !@#$%^&*()+=[]{}|\\;:'\",<>?/") {
		return fmt.Errorf("project name contains invalid characters")
	}

	// Check if directory already exists
	if _, err := os.Stat(name); err == nil {
		return fmt.Errorf("directory %s already exists", name)
	}

	return nil
}

func createDirectoryStructure(projectPath string) error {
	directories := []string{
		"app/http/controllers",
		"app/http/middleware",
		"app/models",
		"bootstrap",
		"config",
		"database/migrations",
		"database/factories",
		"public",
		"resources/views",
		"routes",
		"storage/logs",
		"tests",
	}

	for _, dir := range directories {
		path := filepath.Join(projectPath, dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			return err
		}
	}

	return nil
}

func initGoModule(config ProjectConfig) error {
	// Verify project directory exists
	if _, err := os.Stat(config.Name); os.IsNotExist(err) {
		return fmt.Errorf("project directory does not exist: %s", config.Name)
	}

	// Change to project directory
	originalDir, _ := os.Getwd()
	if err := os.Chdir(config.Name); err != nil {
		return fmt.Errorf("failed to change to project directory: %w", err)
	}
	defer os.Chdir(originalDir)

	// Initialize go module
	moduleName := config.Module
	if moduleName == "" {
		moduleName = config.Name
	}

	cmd := exec.Command("go", "mod", "init", moduleName)
	if err := cmd.Run(); err != nil {
		return err
	}

	cli.Step("Configuring dependencies...")

	// Check if local Velocity exists and use replace directive
	velocityPath := "/Users/ali/code/velocity"
	if _, err := os.Stat(velocityPath); err == nil {
		// Add replace directive for local development
		cmd = exec.Command("go", "mod", "edit", "-replace", "github.com/velocitykode/velocity="+velocityPath)
		cmd.Run()
		cli.Info("Using local Velocity framework")
	} else {
		// Try to get from GitHub (requires GOPRIVATE setup for private repos).
		// Use getLatestVelocityVersion so re-init/repair flows track the
		// same tag the new-project flow does, instead of pinning a stale
		// hardcoded version.
		velocityVersion := getLatestVelocityVersion()
		cmd = exec.Command("go", "get", "github.com/velocitykode/velocity@"+velocityVersion)
		if err := cmd.Run(); err != nil {
			cli.Warning("Note: Configure GOPRIVATE for private repo access")
		}
	}

	// Add other dependencies based on features
	if config.Database == "postgres" {
		cli.Info("PostgreSQL driver")
		exec.Command("go", "get", "github.com/lib/pq").Run()
	} else if config.Database == "mysql" {
		cli.Info("MySQL driver")
		exec.Command("go", "get", "github.com/go-sql-driver/mysql").Run()
	} else if config.Database == "sqlite" {
		cli.Info("SQLite driver")
		exec.Command("go", "get", "github.com/mattn/go-sqlite3").Run()
	}

	if config.Cache == "redis" {
		cli.Info("Redis client")
		exec.Command("go", "get", "github.com/redis/go-redis/v9").Run()
	}

	// Run go mod tidy
	cli.Step("Tidying up dependencies...")
	exec.Command("go", "mod", "tidy").Run()

	return nil
}

func initGitRepo(projectPath string) {
	originalDir, _ := os.Getwd()
	os.Chdir(projectPath)
	defer os.Chdir(originalDir)

	exec.Command("git", "init").Run()
	exec.Command("git", "add", ".").Run()
}

// InitProject adds Velocity structure to an existing Go project
func InitProject(config ProjectConfig, targetDir string) error {
	cli.Step("Setting up Velocity structure...")
	// Create directory structure in existing directory
	if err := createDirectoryStructure(targetDir); err != nil {
		return fmt.Errorf("failed to create directory structure: %w", err)
	}
	cli.Success("Velocity structure created")

	cli.Step("Generating application files...")
	// Generate files from stubs (skip if exists to preserve existing code)
	if err := generateFilesFromStubs(config); err != nil {
		return fmt.Errorf("failed to generate files: %w", err)
	}
	cli.Success("Application files generated")

	cli.Step("Creating configuration files...")
	// Generate config files if they don't exist
	if err := generateProjectFiles(config); err != nil {
		return fmt.Errorf("failed to generate project files: %w", err)
	}
	cli.Success("Configuration files created")

	cli.Step("Adding Velocity dependencies...")
	// Add dependencies to existing go.mod
	if err := addVelocityDependencies(config, targetDir); err != nil {
		return fmt.Errorf("failed to add dependencies: %w", err)
	}
	cli.Success("Dependencies added")

	return nil
}

// createEnvFiles copies .env.example to .env, writes a freshly generated
// APP_KEY, and patches the install-time choices (db driver, cache driver,
// SSR). The APP_KEY must be present before the app bootstraps - otherwise
// `./vel migrate` would fail in the session/cookie encryptor init.
func createEnvFiles(config ProjectConfig) error {
	absPath, err := filepath.Abs(config.Name)
	if err != nil {
		return err
	}

	// Copy .env.example to .env
	cmd := exec.Command("cp", filepath.Join(absPath, ".env.example"), filepath.Join(absPath, ".env"))
	if err := cmd.Run(); err != nil {
		return err
	}

	// Generate APP_KEY, QUEUE_SIGNING_KEY, and AUTH_JWT_SECRET using the
	// same scheme as the framework's console.KeyGenerate (crypto/rand 32
	// bytes, base64 encoded). Separate keys per domain so crypto, queue
	// signing, and JWT auth never share the same key material. Write via
	// a Go helper instead of sed because sed can't append a missing line.
	//
	// AUTH_JWT_SECRET must match the env var the framework reads
	// (velocity/config.go:288); an unconfigured JWT guard emits a boot
	// warning and is skipped, so shipping a scaffold with the key already
	// set makes the built-in guard work out of the box.
	envPath := filepath.Join(absPath, ".env")
	for _, envKey := range []string{"APP_KEY", "QUEUE_SIGNING_KEY", "AUTH_JWT_SECRET"} {
		value, err := generateKey()
		if err != nil {
			return fmt.Errorf("generate %s: %w", envKey, err)
		}
		if err := upsertEnvLine(envPath, envKey, value); err != nil {
			return fmt.Errorf("write %s: %w", envKey, err)
		}
	}

	// Update DB settings based on config
	if config.Database != "" && config.Database != "sqlite" {
		ports := map[string]string{"postgres": "5432", "mysql": "3306"}
		username := os.Getenv("USER")

		// Use base name for database name (not full path)
		dbName := filepath.Base(config.Name)

		subs := []string{
			fmt.Sprintf("s|^DB_CONNECTION=.*|DB_CONNECTION=%s|", config.Database),
			fmt.Sprintf("s|^DB_PORT=.*|DB_PORT=%s|", ports[config.Database]),
			fmt.Sprintf("s|^DB_DATABASE=.*|DB_DATABASE=%s|", dbName),
			fmt.Sprintf("s|^DB_USERNAME=.*|DB_USERNAME=%s|", username),
		}

		// Local Postgres installs don't have SSL enabled by default, but the
		// framework defaults to sslmode=require (secure-by-default). Uncomment
		// the template's DB_SSL_MODE=disable so fresh scaffolds can connect
		// out of the box. Production deployers override in their own .env.
		if config.Database == "postgres" {
			subs = append(subs, "s|^# *DB_SSL_MODE=.*|DB_SSL_MODE=disable|")
		}

		sedCmds := fmt.Sprintf("sed '%s' .env > .env.tmp && mv .env.tmp .env", strings.Join(subs, "; "))
		cmd = exec.Command("sh", "-c", fmt.Sprintf("cd '%s' && %s", absPath, sedCmds))
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	// Update CACHE_DRIVER based on config
	if config.Cache != "" && config.Cache != "memory" {
		cmd = exec.Command("sh", "-c",
			fmt.Sprintf("cd '%s' && sed 's|^CACHE_DRIVER=.*|CACHE_DRIVER=%s|' .env > .env.tmp && mv .env.tmp .env", absPath, config.Cache))
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	if err := applySSROption(config, absPath); err != nil {
		return err
	}

	return nil
}

// generateKey mirrors velocity/console.KeyGenerate - 32 crypto/rand
// bytes, standard-base64 encoded. Used for APP_KEY, QUEUE_SIGNING_KEY,
// and AUTH_JWT_SECRET.
func generateKey() (string, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(key), nil
}

// upsertEnvLine replaces `<key>=...` in envPath with `<key>=<value>`,
// prepending a new line if the key is absent.
func upsertEnvLine(envPath, key, value string) error {
	content, err := os.ReadFile(envPath)
	if err != nil {
		return err
	}
	prefix := key + "="
	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, prefix) {
			lines[i] = prefix + value
			return os.WriteFile(envPath, []byte(strings.Join(lines, "\n")), 0600)
		}
	}
	lines = append([]string{prefix + value}, lines...)
	return os.WriteFile(envPath, []byte(strings.Join(lines, "\n")), 0600)
}

// applySSROption wires the --ssr flag into the generated project.
//
// With --ssr: enable INERTIA_SSR_ENABLED in .env and point the URL at
// Vite's dev endpoint, so the scaffold serves SSR out of the box.
//
// Without --ssr: patch vite.config.ts so the @inertiajs/vite plugin
// skips its dev-server warmup. The plugin otherwise logs
// "Inertia SSR module graph warmed up" on every `vel serve` even
// when the Go backend has SSR disabled, which is confusing for users
// who explicitly opted out.
func applySSROption(config ProjectConfig, absPath string) error {
	if config.API {
		return nil
	}

	if config.SSR {
		cmd := exec.Command("sh", "-c", fmt.Sprintf(
			"cd '%s' && sed -E 's|^# *INERTIA_SSR_ENABLED=.*|INERTIA_SSR_ENABLED=true|; s|^# *INERTIA_SSR_URL=.*|INERTIA_SSR_URL=http://localhost:5173/__inertia_ssr|; s|^# *INERTIA_SSR_TIMEOUT=.*|INERTIA_SSR_TIMEOUT=3s|' .env > .env.tmp && mv .env.tmp .env",
			absPath,
		))
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("enable ssr in .env: %w", err)
		}
		return nil
	}

	// SSR off - splice `ssr: false,` into the inertia() plugin options.
	vitePath := filepath.Join(absPath, "vite.config.ts")
	src, err := os.ReadFile(vitePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read vite.config.ts: %w", err)
	}
	// Template style is 8-space indent for `inertia({` and 12-space
	// indent for options inside it (see velocity-template-react/vite.config.ts).
	out := strings.Replace(string(src), "inertia({", "inertia({\n            ssr: false,", 1)
	if out == string(src) {
		// No inertia() call - template structure changed, skip quietly.
		return nil
	}
	if err := os.WriteFile(vitePath, []byte(out), 0o644); err != nil {
		return fmt.Errorf("write vite.config.ts: %w", err)
	}
	return nil
}

// createDefaultMigrations creates default migration files only if the template
// didn't already provide them.
func createDefaultMigrations(projectPath string) error {
	absPath, err := filepath.Abs(projectPath)
	if err != nil {
		return err
	}

	migrationsDir := filepath.Join(absPath, "database", "migrations")

	// Skip if the template already has migration files
	if entries, err := filepath.Glob(filepath.Join(migrationsDir, "*.go")); err == nil && len(entries) > 0 {
		return nil
	}

	if err := os.MkdirAll(migrationsDir, 0755); err != nil {
		return err
	}

	// Migration 1: Create users table
	usersTable := `package migrations

import "github.com/velocitykode/velocity/orm/migrate"

func init() {
	migrate.Register(&migrate.Migration{
		Version:     "20010101000000",
		Description: "create users table",
		Up: func(m *migrate.Migrator) error {
			return m.CreateTable("users", func(t *migrate.TableBuilder) {
				t.ID()
				t.String("name")
				t.String("email").Unique()
				t.String("password")
				t.String("role").Default("user")
				t.Timestamps()
			})
		},
		Down: func(m *migrate.Migrator) error {
			return m.DropTable("users")
		},
	})
}
`

	// Migration 2: Create cache table
	cacheTable := `package migrations

import "github.com/velocitykode/velocity/orm/migrate"

func init() {
	migrate.Register(&migrate.Migration{
		Version:     "20010101000001",
		Description: "create cache table",
		Up: func(m *migrate.Migrator) error {
			return m.CreateTable("cache", func(t *migrate.TableBuilder) {
				t.String("key", 255).Unique()
				t.String("value", 10000)
				t.Integer("expiration")
			})
		},
		Down: func(m *migrate.Migrator) error {
			return m.DropTable("cache")
		},
	})
}
`

	// Migration 3: Create jobs table
	jobsTable := `package migrations

import "github.com/velocitykode/velocity/orm/migrate"

func init() {
	migrate.Register(&migrate.Migration{
		Version:     "20010101000002",
		Description: "create jobs table",
		Up: func(m *migrate.Migrator) error {
			if err := m.CreateTable("jobs", func(t *migrate.TableBuilder) {
				t.ID()
				t.String("queue", 255)
				t.String("payload", 10000)
				t.Integer("attempts").Default("0")
				t.String("scheduled_at", 50)
				t.String("reserved_at", 50).Nullable()
				t.String("reserved_by", 255).Nullable()
				t.String("failed_at", 50).Nullable()
				t.String("failed_reason", 5000).Nullable()
				t.Timestamps()
			}); err != nil {
				return err
			}

			return m.CreateTable("failed_jobs", func(t *migrate.TableBuilder) {
				t.ID()
				t.String("queue", 255)
				t.String("payload", 10000)
				t.String("exception", 10000)
				t.Timestamps()
			})
		},
		Down: func(m *migrate.Migrator) error {
			if err := m.DropTable("failed_jobs"); err != nil {
				return err
			}
			return m.DropTable("jobs")
		},
	})
}
`

	// Write migration files
	migrations := map[string]string{
		"0001_01_01_000000_create_users_table.go": usersTable,
		"0001_01_01_000001_create_cache_table.go": cacheTable,
		"0001_01_01_000002_create_jobs_table.go":  jobsTable,
	}

	for filename, content := range migrations {
		filePath := filepath.Join(migrationsDir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return err
		}
	}

	return nil
}

func addVelocityDependencies(config ProjectConfig, projectPath string) error {
	// Change to project directory
	originalDir, _ := os.Getwd()
	os.Chdir(projectPath)
	defer os.Chdir(originalDir)

	cli.Step("Configuring dependencies...")

	// Check if local Velocity exists and use replace directive
	velocityPath := "/Users/ali/code/velocity"
	if _, err := os.Stat(velocityPath); err == nil {
		// Add replace directive for local development
		cmd := exec.Command("go", "mod", "edit", "-replace", "github.com/velocitykode/velocity="+velocityPath)
		cmd.Run()
		cli.Info("Using local Velocity framework")
	} else {
		// Try to get from GitHub
		cmd := exec.Command("go", "get", "github.com/velocitykode/velocity")
		if err := cmd.Run(); err != nil {
			cli.Warning("Note: Configure GOPRIVATE for private repo access")
		}
	}

	// Add other dependencies based on features
	if config.Database == "postgres" {
		cli.Info("PostgreSQL driver")
		exec.Command("go", "get", "github.com/lib/pq").Run()
	} else if config.Database == "mysql" {
		cli.Info("MySQL driver")
		exec.Command("go", "get", "github.com/go-sql-driver/mysql").Run()
	} else if config.Database == "sqlite" {
		cli.Info("SQLite driver")
		exec.Command("go", "get", "github.com/mattn/go-sqlite3").Run()
	}

	if config.Cache == "redis" {
		cli.Info("Redis client")
		exec.Command("go", "get", "github.com/redis/go-redis/v9").Run()
	}

	// Run go mod tidy
	cli.Step("Tidying up dependencies...")
	exec.Command("go", "mod", "tidy").Run()

	return nil
}
