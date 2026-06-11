// Velocity installer - creates and manages Velocity projects
package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"
	cli "github.com/velocitykode/velocity-cli"
	"github.com/velocitykode/velocity-installer/internal/commands"
	"github.com/velocitykode/velocity-installer/internal/generator"
	"github.com/velocitykode/velocity-installer/internal/version"
)

var Version = "0.21.24"

//go:embed cli-theme.toml
var themeConfig []byte

func main() {
	if err := cli.LoadConfig(bytes.NewReader(themeConfig)); err != nil {
		fmt.Fprintf(os.Stderr, "velocity-cli theme: %v\n", err)
	}

	if err := version.CheckGoVersion(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	rootCmd := &cobra.Command{
		Use:     "velocity",
		Short:   "Velocity installer - create and manage Velocity projects",
		Version: Version,
		Run: func(cmd *cobra.Command, args []string) {
			commands.RenderHome(cmd)
		},
	}

	rootCmd.CompletionOptions.DisableDefaultCmd = true
	// Show pinned template tags alongside the installer's own version.
	// Each installer release pins exact template tags, and the framework
	// version is whatever those templates' go.mod files require - so the
	// template tags are the relevant build coordinates for support and
	// reproducibility, not just the installer semver.
	rootCmd.SetVersionTemplate(buildVersionTemplate(Version))

	// Set version for self-update
	commands.InstallerVersion = Version

	rootCmd.AddCommand(commands.NewCmd)
	// rootCmd.AddCommand(commands.InitCmd) // TODO: Re-enable after fixing stub generation
	rootCmd.AddCommand(commands.ConfigCmd)
	rootCmd.AddCommand(commands.SelfUpdateCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// buildVersionTemplate renders the multi-line --version output. Lists
// the installer version followed by every entry in supportedTemplates,
// stably sorted so output is deterministic across runs.
func buildVersionTemplate(installerVersion string) string {
	var b bytes.Buffer
	fmt.Fprintf(&b, "velocity %s\n", installerVersion)
	fmt.Fprintln(&b, "templates:")
	tmpls := generator.SupportedTemplates()
	keys := make([]string, 0, len(tmpls))
	for k := range tmpls {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		tag := tmpls[k]
		if tag == "" {
			tag = "main"
		}
		fmt.Fprintf(&b, "  %s -> %s\n", k, tag)
	}
	return b.String()
}
