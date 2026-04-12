// Velocity installer - creates and manages Velocity projects
package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	cli "github.com/velocitykode/velocity-cli"
	"github.com/velocitykode/velocity-installer/internal/banner"
	"github.com/velocitykode/velocity-installer/internal/commands"
	"github.com/velocitykode/velocity-installer/internal/version"
)

var Version = "0.9.1"

//go:embed velocity-cli.toml
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
			fmt.Println(banner.Simple())
			fmt.Println(cli.StyleMuted("       The Official CLI for Velocity Web Framework"))
			fmt.Println()
			cmd.Help()
		},
	}

	rootCmd.CompletionOptions.DisableDefaultCmd = true

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
