// Velocity installer - creates and manages Velocity projects
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/velocitykode/velocity-installer/internal/banner"
	"github.com/velocitykode/velocity-installer/internal/colors"
	"github.com/velocitykode/velocity-installer/internal/commands"
	"github.com/velocitykode/velocity-installer/internal/version"
)

var Version = "0.6.50"

func main() {
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
			fmt.Println(colors.MutedStyle.Render("       The Official CLI for Velocity Web Framework"))
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
