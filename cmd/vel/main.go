// vel runs a command in the enclosing Velocity project: it builds the project's
// own CLI and hands the command line to it. See internal/launcher.
package main

import (
	"os"

	"github.com/velocitykode/velocity-installer/internal/launcher"
)

func main() {
	os.Exit(launcher.Main(os.Args[1:], os.Stderr))
}
