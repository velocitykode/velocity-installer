//go:build !windows

package launcher

import (
	"fmt"
	"io"
	"os"
	"syscall"
)

// run replaces the launcher process with the project binary, so stdio, exit
// codes, and signals (Ctrl-C during `vel serve`) behave exactly as `./vel`.
func run(bin string, args []string, stderr io.Writer) int {
	err := syscall.Exec(bin, append([]string{bin}, args...), os.Environ())
	fmt.Fprintf(stderr, "vel: exec %s: %v\n", bin, err)
	return 1
}
