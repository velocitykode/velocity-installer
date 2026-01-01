package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/velocitykode/velocity-installer/internal/generator"
	"github.com/velocitykode/velocity-installer/internal/styles"
	"github.com/velocitykode/velocity-installer/internal/ui"
)

type step int

const (
	stepProjectName step = iota
	stepDatabase
	stepCache
	stepFeatures
	stepConfirm
	stepCreating
	stepDone
)

type model struct {
	step        step
	projectName string
	database    string
	cache       string
	features    map[string]bool

	// UI components
	textInput     textinput.Model
	progress      progress.Model
	currentChoice int
	choices       []string
	width         int
	height        int
	creating      bool
	done          bool
	progressPct   float64
	err           error
}

func initialModel(projectName string) model {
	ti := textinput.New()
	ti.Placeholder = "my-awesome-app"
	ti.Focus()
	ti.CharLimit = 50
	ti.Width = 40
	if projectName != "" {
		ti.SetValue(projectName)
	}

	return model{
		step:        stepProjectName,
		projectName: projectName,
		textInput:   ti,
		progress:    progress.New(progress.WithDefaultGradient()),
		features: map[string]bool{
			"auth": false,
			"api":  false,
		},
		width:  80,
		height: 24,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "enter":
			switch m.step {
			case stepProjectName:
				if m.textInput.Value() != "" {
					m.projectName = m.textInput.Value()
					m.step = stepDatabase
					m.choices = []string{"PostgreSQL", "MySQL", "SQLite", "None"}
					m.currentChoice = 0
				}

			case stepDatabase:
				m.database = strings.ToLower(m.choices[m.currentChoice])
				if m.database == "none" {
					m.database = ""
				} else if m.database == "postgresql" {
					m.database = "postgres"
				}
				m.step = stepCache
				m.choices = []string{"Redis", "Memory", "None"}
				m.currentChoice = 0

			case stepCache:
				m.cache = strings.ToLower(m.choices[m.currentChoice])
				if m.cache == "none" {
					m.cache = ""
				}
				m.step = stepFeatures
				m.choices = []string{"Authentication", "API-only mode"}
				m.currentChoice = 0

			case stepFeatures:
				// Don't toggle on enter, just move to next step
				m.step = stepConfirm

			case stepConfirm:
				// Mark as done and quit TUI
				m.step = stepDone
				return m, tea.Quit
			}

		case "tab":
			if m.step == stepFeatures {
				// Tab to move to confirm step
				m.step = stepConfirm
			}

		case "up", "k":
			if m.currentChoice > 0 {
				m.currentChoice--
			}

		case "down", "j":
			if m.currentChoice < len(m.choices)-1 {
				m.currentChoice++
			}

		case " ": // Space key
			if m.step == stepFeatures {
				// Toggle feature
				if m.currentChoice == 0 {
					m.features["auth"] = !m.features["auth"]
				} else if m.currentChoice == 1 {
					m.features["api"] = !m.features["api"]
				}
				return m, nil // Return updated model
			}
		} // Close the switch for tea.KeyMsg
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.progress.Width = msg.Width - 20

	}

	// Update text input
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m model) View() string {
	var b strings.Builder

	// Header
	header := styles.TitleStyle.Render("🚀 Create New Velocity Project")
	b.WriteString(header + "\n\n")

	switch m.step {
	case stepProjectName:
		b.WriteString(styles.SubtitleStyle.Render("Project Name"))
		b.WriteString("\n")
		b.WriteString(m.textInput.View())
		b.WriteString("\n\n")
		b.WriteString(styles.HelpStyle.Render("Press Enter to continue"))

	case stepDatabase:
		b.WriteString(styles.SubtitleStyle.Render("Choose Database Driver"))
		b.WriteString("\n\n")
		for i, choice := range m.choices {
			cursor := "  "
			if i == m.currentChoice {
				cursor = styles.SelectedItemStyle.Render("▸ ")
				b.WriteString(cursor + styles.SelectedItemStyle.Render(choice))
			} else {
				b.WriteString(cursor + styles.ItemStyle.Render(choice))
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
		b.WriteString(styles.HelpStyle.Render("↑/↓ to navigate, Enter to select"))

	case stepCache:
		b.WriteString(styles.SubtitleStyle.Render("Choose Cache Driver"))
		b.WriteString("\n\n")
		for i, choice := range m.choices {
			cursor := "  "
			if i == m.currentChoice {
				cursor = styles.SelectedItemStyle.Render("▸ ")
				b.WriteString(cursor + styles.SelectedItemStyle.Render(choice))
			} else {
				b.WriteString(cursor + styles.ItemStyle.Render(choice))
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
		b.WriteString(styles.HelpStyle.Render("↑/↓ to navigate, Enter to select"))

	case stepFeatures:
		b.WriteString(styles.SubtitleStyle.Render("Select Features"))
		b.WriteString("\n\n")

		// Authentication
		checkbox := "[ ]"
		if m.features["auth"] {
			checkbox = "[✓]"
		}
		if m.currentChoice == 0 {
			b.WriteString(styles.SelectedItemStyle.Render("▸ " + checkbox + " Authentication"))
		} else {
			b.WriteString(styles.ItemStyle.Render(checkbox + " Authentication"))
		}
		b.WriteString("\n")

		// API-only
		checkbox = "[ ]"
		if m.features["api"] {
			checkbox = "[✓]"
		}
		if m.currentChoice == 1 {
			b.WriteString(styles.SelectedItemStyle.Render("▸ " + checkbox + " API-only mode"))
		} else {
			b.WriteString(styles.ItemStyle.Render(checkbox + " API-only mode"))
		}
		b.WriteString("\n\n")
		b.WriteString(styles.HelpStyle.Render("Space to toggle, Enter to continue"))

	case stepConfirm:
		b.WriteString(styles.SubtitleStyle.Render("Ready to create project with:"))
		b.WriteString("\n\n")

		summary := fmt.Sprintf("  Project: %s\n", m.projectName)
		if m.database != "" {
			summary += fmt.Sprintf("  Database: %s\n", m.database)
		}
		if m.cache != "" {
			summary += fmt.Sprintf("  Cache: %s\n", m.cache)
		}
		if m.features["auth"] {
			summary += "  Authentication: Yes\n"
		}
		if m.features["api"] {
			summary += "  API-only: Yes\n"
		}

		b.WriteString(styles.BoxStyle.Render(summary))
		b.WriteString("\n\n")
		b.WriteString(styles.HelpStyle.Render("Press Enter to create, q to quit"))

	case stepCreating:
		b.WriteString("Creating project...\n")

	case stepDone:
		// Don't show anything when exiting
		return ""
	}

	// Return content directly without centering
	return b.String()
}

func LaunchNewProjectWizard(projectName string) {
	m := initialModel(projectName)
	p := tea.NewProgram(m)
	finalModel, err := p.Run()

	if err != nil {
		// If TUI fails (e.g., no TTY), fall back to creating with defaults
		if strings.Contains(err.Error(), "tty") || strings.Contains(err.Error(), "TTY") {
			CreateProjectWithDefaults(projectName)
		} else {
			fmt.Printf("Error: %v\n", err)
		}
		return
	}

	// If user completed the wizard, create the project
	if finalM, ok := finalModel.(model); ok && finalM.step == stepDone {
		config := generator.ProjectConfig{
			Name:     finalM.projectName,
			Module:   finalM.projectName,
			Database: finalM.database,
			Cache:    finalM.cache,
			Auth:     finalM.features["auth"],
			API:      finalM.features["api"],
		}

		// Don't clear screen - just continue with output
		fmt.Println() // Add a blank line for spacing

		if err := generator.CreateProject(config); err != nil {
			ui.Error(fmt.Sprintf("Error creating project: %v", err))
			return
		}

		ui.Newline()
		ui.Success("Project created successfully!")
		ui.Newline()
		ui.Bold(fmt.Sprintf("Your Velocity project '%s' is ready.", config.Name))
		ui.Newline()
		ui.Bold("Get started:")
		fmt.Printf("  %s %s\n", ui.Highlight("cd"), config.Name)
		fmt.Printf("  %s\n", ui.Highlight("npm run dev"))
		fmt.Printf("  %s\n", ui.Highlight("go run main.go"))
		ui.Newline()
		ui.Muted("Default port: 4000 (set PORT env to change)")
		ui.Newline()
	}
}

// CreateProjectWithDefaults creates a project with sensible defaults when TUI is not available
func CreateProjectWithDefaults(projectName string) {
	config := generator.ProjectConfig{
		Name:     projectName,
		Module:   projectName,
		Database: "",
		Cache:    "",
		Auth:     false,
		API:      false,
	}

	if err := generator.CreateProject(config); err != nil {
		ui.Error(fmt.Sprintf("Error creating project: %v", err))
		return
	}

	ui.Newline()
	ui.Success("Project created successfully!")
	ui.Newline()
	ui.Bold(fmt.Sprintf("Your Velocity project '%s' is ready.", projectName))
	ui.Newline()
	ui.Bold("Get started:")
	fmt.Printf("  %s %s\n", ui.Highlight("cd"), projectName)
	fmt.Printf("  %s\n", ui.Highlight("npm run dev"))
	fmt.Printf("  %s\n", ui.Highlight("go run main.go"))
	ui.Newline()
	ui.Muted("Default port: 4000 (set PORT env to change)")
	ui.Newline()
}
