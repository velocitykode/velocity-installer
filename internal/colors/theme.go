package colors

import (
	"github.com/charmbracelet/lipgloss"
)

func init() {
	// Skip heavy initialization - let lipgloss handle it automatically
	// This is much faster in tmux/ghostty environments
}

// Velocity Brand Colors - from velocity-docs/assets/css/custom.css
var (
	// Primary brand colors
	Primary      = lipgloss.Color("#0e87cd") // --velocity-primary
	PrimaryHover = lipgloss.Color("#0a6ba8") // --velocity-primary-hover

	// Status colors
	Success = lipgloss.Color("#10b981")
	Warning = lipgloss.Color("#f59e0b")
	Error   = lipgloss.Color("#ef4444")

	// Text colors
	Muted = lipgloss.Color("#6b7280")
	Dark  = lipgloss.Color("#1a1a1a")

	// Styles using brand colors
	BrandStyle = lipgloss.NewStyle().
			Foreground(Primary).
			Bold(true)

	BrandHoverStyle = lipgloss.NewStyle().
			Foreground(PrimaryHover).
			Bold(true)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(Success).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(Error).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(Warning)

	MutedStyle = lipgloss.NewStyle().
			Foreground(Muted)
)
