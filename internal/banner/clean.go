package banner

import (
	"github.com/velocitykode/prism"
)

// Clean returns a clean, simple text banner
func Clean() string {
	return prism.StylePrimary(`
 VELOCITY CLI
 The Go Web Framework`)
}

// CleanBox returns a clean boxed banner
func CleanBox() string {
	return prism.StylePrimary(`
┌─────────────────────────────────────┐
│          VELOCITY CLI               │
│   The Go Web Framework              │
└─────────────────────────────────────┘`)
}

// Title returns just the title
func Title() string {
	return prism.StylePrimary("VELOCITY CLI")
}

// Divider returns a simple divider
func Divider() string {
	return prism.StyleMuted("────────────────────────────────────")
}
