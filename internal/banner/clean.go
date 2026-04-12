package banner

import (
	cli "github.com/velocitykode/velocity-cli"
)

// Clean returns a clean, simple text banner
func Clean() string {
	return cli.StylePrimary(`
 VELOCITY CLI
 The Go Web Framework`)
}

// CleanBox returns a clean boxed banner
func CleanBox() string {
	return cli.StylePrimary(`
┌─────────────────────────────────────┐
│          VELOCITY CLI               │
│   The Go Web Framework              │
└─────────────────────────────────────┘`)
}

// Title returns just the title
func Title() string {
	return cli.StylePrimary("VELOCITY CLI")
}

// Divider returns a simple divider
func Divider() string {
	return cli.StyleMuted("────────────────────────────────────")
}
