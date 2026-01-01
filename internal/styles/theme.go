package styles

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Velocity brand color
	Primary = lipgloss.Color("#0e87cd")
	Success = lipgloss.Color("#10b981")
	Warning = lipgloss.Color("#f59e0b")
	Error   = lipgloss.Color("#ef4444")
	Muted   = lipgloss.Color("#6b7280")

	// Text styles
	TitleStyle = lipgloss.NewStyle().
			Foreground(Primary).
			Bold(true).
			MarginBottom(1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(Muted).
			MarginBottom(1)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(Success).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(Error).
			Bold(true)

	// Box styles
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Primary).
			Padding(1, 2)

	// List styles
	SelectedItemStyle = lipgloss.NewStyle().
				Foreground(Primary).
				Bold(true).
				PaddingLeft(2)

	ItemStyle = lipgloss.NewStyle().
			PaddingLeft(4)

	// Progress bar styles
	ProgressBarStyle = lipgloss.NewStyle().
				Foreground(Primary)

	// Help styles
	HelpStyle = lipgloss.NewStyle().
			Foreground(Muted).
			MarginTop(1)
)
