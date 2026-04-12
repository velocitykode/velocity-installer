package ui

import (
	"testing"
)

// TreeItem/TreeItemSkipped/ClearLines write directly to stdout; these smoke tests
// just ensure they don't panic.

func TestTreeItem_DoesNotPanic(t *testing.T) {
	TreeItem("├─", "item", "done", true)
	TreeItem("└─", "item", "pending", false)
}

func TestTreeItemSkipped_DoesNotPanic(t *testing.T) {
	TreeItemSkipped("└─", "item", "already installed")
}

func TestClearLines_DoesNotPanic(t *testing.T) {
	ClearLines(0)
	ClearLines(2)
}
