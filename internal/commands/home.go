package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	cli "github.com/velocitykode/velocity-cli"
	"github.com/velocitykode/velocity-installer/internal/banner"
)

// RenderHome prints the top-level help shown when `velocity` is invoked
// with no subcommand. Styled with the CLI theme so the entry point
// matches the rest of the installer's output instead of falling back
// to cobra's plain-text help.
func RenderHome(root *cobra.Command) {
	fmt.Println(banner.Simple())
	fmt.Println(cli.StyleMuted("       The Official CLI for Velocity Web Framework"))
	cli.Newline()

	section("USAGE")
	cli.Muted(fmt.Sprintf("%s <command> [flags]", root.Name()))
	cli.Newline()

	section("COMMANDS")
	cmds := visibleSubcommands(root)
	cmdPad := maxLabelWidth(cmds)
	for _, c := range cmds {
		row(c.Name(), c.Short, cmdPad)
	}
	cli.Newline()

	section("FLAGS")
	flags := []struct{ label, desc string }{
		{"-h, --help", "Show help"},
		{"-v, --version", "Show version"},
	}
	flagPad := 0
	for _, f := range flags {
		if w := len(f.label); w > flagPad {
			flagPad = w
		}
	}
	for _, f := range flags {
		row(f.label, f.desc, flagPad)
	}
	cli.Newline()

	cli.Muted(fmt.Sprintf(`Run "%s <command> --help" for command-specific options.`, root.Name()))
	cli.Newline()
}

func section(title string) {
	fmt.Println(cli.StylePrimary(title))
}

func row(label, desc string, pad int) {
	padding := strings.Repeat(" ", pad-len(label))
	fmt.Printf("  %s%s  %s\n", cli.StylePrimary(label), padding, cli.StyleMuted(desc))
}

func visibleSubcommands(root *cobra.Command) []*cobra.Command {
	out := make([]*cobra.Command, 0, len(root.Commands()))
	for _, c := range root.Commands() {
		if c.Hidden || c.Name() == "help" {
			continue
		}
		out = append(out, c)
	}
	return out
}

func maxLabelWidth(cmds []*cobra.Command) int {
	w := 0
	for _, c := range cmds {
		if n := len(c.Name()); n > w {
			w = n
		}
	}
	return w
}
