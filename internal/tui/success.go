package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/velocitykode/velocity-installer/internal/styles"
)

// ShowProjectCreated displays a success message for non-interactive mode
func ShowProjectCreated(projectName string) {
	// Mock creating project with some delay
	steps := []string{
		"Creating directory structure",
		"Generating application files",
		"Setting up configuration",
		"Installing dependencies",
		"Finalizing project",
	}

	for _, step := range steps {
		fmt.Println(styles.SubtitleStyle.Render(step + "..."))
		time.Sleep(200 * time.Millisecond)
	}

	fmt.Println()

	// Success box
	successBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Success).
		Padding(1, 2).
		Width(60)

	successMsg := fmt.Sprintf(`✨ Project '%s' created successfully!

Get started with:
  cd %s
  go run main.go serve

Visit http://localhost:4000`, projectName, projectName)

	fmt.Println(successBox.Render(successMsg))
}
