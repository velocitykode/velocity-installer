package ui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors
	primaryColor = lipgloss.Color("#0e87cd")
	successColor = lipgloss.Color("#10b981")
	warningColor = lipgloss.Color("#f59e0b")
	errorColor   = lipgloss.Color("#ef4444")
	mutedColor   = lipgloss.Color("#6b7280")

	// Symbols
	arrowSymbol = lipgloss.NewStyle().Foreground(primaryColor).Bold(true).Render("→")
	checkSymbol = lipgloss.NewStyle().Foreground(successColor).Render("✓")
	warnSymbol  = lipgloss.NewStyle().Foreground(warningColor).Render("!")
	crossSymbol = lipgloss.NewStyle().Foreground(errorColor).Render("✗")

	// Text styles
	mutedStyle   = lipgloss.NewStyle().Foreground(mutedColor)
	primaryStyle = lipgloss.NewStyle().Foreground(primaryColor).Bold(true)
	successStyle = lipgloss.NewStyle().Foreground(successColor).Bold(true)
	warningStyle = lipgloss.NewStyle().Foreground(warningColor).Bold(true)
	errorStyle   = lipgloss.NewStyle().Foreground(errorColor).Bold(true)
)

// Header prints a styled command header (uppercase cyan)
func Header(command string) {
	fmt.Printf("\n%s\n\n", primaryStyle.Render(strings.ToUpper(command)))
}

// Info prints an info message with arrow symbol (muted text)
func Info(message string) {
	fmt.Printf("%s %s\n", arrowSymbol, mutedStyle.Render(message))
}

// Success prints a success message with checkmark
func Success(message string) {
	fmt.Printf("%s %s\n", checkSymbol, successStyle.Render(message))
}

// Warning prints a warning message
func Warning(message string) {
	fmt.Printf("%s %s\n", warnSymbol, warningStyle.Render(message))
}

// Error prints an error message
func Error(message string) {
	fmt.Printf("%s %s\n", crossSymbol, errorStyle.Render(message))
}

// Step prints a muted step message (indented)
func Step(message string) {
	fmt.Printf("  %s\n", mutedStyle.Render(message))
}

// Muted prints muted text
func Muted(message string) {
	fmt.Printf("  %s\n", mutedStyle.Render(message))
}

// Bold prints bold text
func Bold(message string) {
	fmt.Println(lipgloss.NewStyle().Bold(true).Render(message))
}

// Highlight returns highlighted text
func Highlight(text string) string {
	return primaryStyle.Render(text)
}

// Command returns styled command text
func Command(cmd string) string {
	return primaryStyle.Render(cmd)
}

// KeyValue prints a key-value pair
func KeyValue(key, value string) {
	fmt.Printf("  %s %s\n", mutedStyle.Render(key+":"), value)
}

// Newline prints an empty line
func Newline() {
	fmt.Println()
}

// NextSteps prints formatted next steps
func NextSteps(steps []string) {
	fmt.Println()
	fmt.Println(mutedStyle.Render("Next steps:"))
	for i, step := range steps {
		fmt.Printf("  %s %s\n", primaryStyle.Render(fmt.Sprintf("%d.", i+1)), step)
	}
}

// Task runs an action with step message, then shows success/error
func Task(stepMsg, successMsg string, action func() error) error {
	Info(stepMsg)
	fmt.Printf("  %s\n", mutedStyle.Render(stepMsg+"..."))
	if err := action(); err != nil {
		return err
	}
	Success(successMsg)
	return nil
}

// Spinner shows a single dot while action runs, then clears the line
func Spinner(message string, action func() error) error {
	done := make(chan bool)
	var err error

	go func() {
		err = action()
		done <- true
	}()

	ticker := time.NewTicker(300 * time.Millisecond)
	defer ticker.Stop()

	dots := 0
	for {
		select {
		case <-done:
			fmt.Fprintf(os.Stdout, "\r\033[K")
			return err
		case <-ticker.C:
			dots = (dots % 3) + 1
			fmt.Fprintf(os.Stdout, "\r  %s%s", mutedStyle.Render(message), mutedStyle.Render(strings.Repeat(".", dots)))
		}
	}
}

// TreeItem prints a tree-style item with status
// prefix: "├─" for middle items, "└─" for last item
func TreeItem(prefix, label, status string, done bool) {
	var statusText string
	if done {
		statusText = checkSymbol + " " + successStyle.Render(status)
	} else {
		statusText = mutedStyle.Render(status)
	}
	fmt.Printf("  %s %s %s\n", mutedStyle.Render(prefix), mutedStyle.Render(label), statusText)
}

// TreeItemSkipped prints a skipped tree item
func TreeItemSkipped(prefix, label, reason string) {
	fmt.Printf("  %s %s %s\n", mutedStyle.Render(prefix), mutedStyle.Render(label), warningStyle.Render("skipped ("+reason+")"))
}

// ClearLines clears n lines above cursor
func ClearLines(n int) {
	for i := 0; i < n; i++ {
		fmt.Print("\033[A\033[K")
	}
}
