package banner

import (
	"github.com/velocitykode/velocity-installer/internal/colors"
)

// Clean returns a clean, simple text banner
func Clean() string {
	return colors.BrandStyle.Render(`
 VELOCITY CLI
 The Go Web Framework`)
}

// CleanBox returns a clean boxed banner
func CleanBox() string {
	return colors.BrandStyle.Render(`
┌─────────────────────────────────────┐
│          VELOCITY CLI               │
│   The Go Web Framework              │
└─────────────────────────────────────┘`)
}

// Title returns just the title
func Title() string {
	return colors.BrandStyle.Bold(true).Render("VELOCITY CLI")
}

// Divider returns a simple divider
func Divider() string {
	return colors.MutedStyle.Render("────────────────────────────────────")
}
