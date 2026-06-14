// Package ui holds installer-specific output helpers that have no SDK equivalent.
// All generic output (Header/Info/Success/...) comes from github.com/velocitykode/prism.
package ui

import (
	"fmt"

	"github.com/velocitykode/prism"
)

// TreeItem prints a tree-style item with status.
// prefix: "├─" for middle items, "└─" for last item.
func TreeItem(prefix, label, status string, done bool) {
	statusText := prism.StyleMuted(status)
	if done {
		statusText = prism.StyleSuccess("✓ " + status)
	}
	fmt.Printf("  %s %s %s\n",
		prism.StyleMuted(prefix),
		prism.StyleMuted(label),
		statusText,
	)
}

// TreeItemSkipped prints a skipped tree item.
func TreeItemSkipped(prefix, label, reason string) {
	fmt.Printf("  %s %s %s\n",
		prism.StyleMuted(prefix),
		prism.StyleMuted(label),
		prism.StyleWarning("skipped ("+reason+")"),
	)
}

// ClearLines clears n lines above the cursor.
func ClearLines(n int) {
	for i := 0; i < n; i++ {
		fmt.Print("\033[A\033[K")
	}
}
