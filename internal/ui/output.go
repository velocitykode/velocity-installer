// Package ui holds installer-specific output helpers that have no SDK equivalent.
// All generic output (Header/Info/Success/...) comes from github.com/velocitykode/velocity-cli.
package ui

import (
	"fmt"

	cli "github.com/velocitykode/velocity-cli"
)

// TreeItem prints a tree-style item with status.
// prefix: "├─" for middle items, "└─" for last item.
func TreeItem(prefix, label, status string, done bool) {
	statusText := cli.StyleMuted(status)
	if done {
		statusText = cli.StyleSuccess("✓ " + status)
	}
	fmt.Printf("  %s %s %s\n",
		cli.StyleMuted(prefix),
		cli.StyleMuted(label),
		statusText,
	)
}

// TreeItemSkipped prints a skipped tree item.
func TreeItemSkipped(prefix, label, reason string) {
	fmt.Printf("  %s %s %s\n",
		cli.StyleMuted(prefix),
		cli.StyleMuted(label),
		cli.StyleWarning("skipped ("+reason+")"),
	)
}

// ClearLines clears n lines above the cursor.
func ClearLines(n int) {
	for i := 0; i < n; i++ {
		fmt.Print("\033[A\033[K")
	}
}

