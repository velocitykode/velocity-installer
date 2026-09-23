//go:build windows

package launcher

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
)

// run starts the project binary as a child with inherited stdio and mirrors its
// exit code. Windows has no exec(2); Ctrl-C reaches the child through the shared
// console, so the launcher only ignores it while waiting.
func run(bin string, args []string, stderr io.Writer) int {
	signal.Ignore(os.Interrupt)
	cmd := exec.Command(bin, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.ExitCode()
		}
		fmt.Fprintf(stderr, "vel: run %s: %v\n", bin, err)
		return 1
	}
	return 0
}
