package generator

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/term"
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// runStep executes a function while showing a spinner animation.
// Returns the duration and any error from the function.
func runStep(text string, fn func() error) (time.Duration, error) {
	// Check if stdout is a TTY
	isTTY := term.IsTerminal(int(os.Stdout.Fd()))

	if !isTTY {
		// Non-TTY: just run the function without animation
		start := time.Now()
		err := fn()
		return time.Since(start), err
	}

	// Start spinner in background
	stop := make(chan struct{})
	done := make(chan struct{})

	// Add margin top
	fmt.Println()

	go func() {
		ticker := time.NewTicker(80 * time.Millisecond)
		defer ticker.Stop()
		frame := 0

		for {
			select {
			case <-stop:
				close(done)
				return
			case <-ticker.C:
				// Clear line and print spinner after text
				fmt.Printf("\r\033[K%s %s", text, spinnerFrames[frame])
				frame = (frame + 1) % len(spinnerFrames)
			}
		}
	}()

	// Run the actual function
	start := time.Now()
	err := fn()
	duration := time.Since(start)

	// Stop spinner
	close(stop)
	<-done

	// Clear the spinner line
	fmt.Print("\r\033[K")

	return duration, err
}
