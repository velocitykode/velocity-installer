package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestInitialModel(t *testing.T) {
	m := initialModel("")

	// Test initial state
	if m.projectName != "" {
		t.Error("Project name should be empty initially")
	}

	if m.step != stepProjectName {
		t.Error("Should start at project name step")
	}

	if m.err != nil {
		t.Error("Should not have error initially")
	}
}

func TestInitialModelWithName(t *testing.T) {
	m := initialModel("myproject")

	// Test initial state with provided name
	if m.textInput.Value() != "myproject" {
		t.Error("Text input should have the provided project name")
	}
}

func TestInit(t *testing.T) {
	m := initialModel("")
	cmd := m.Init()

	// Init should return a command to blink text input
	if cmd == nil {
		t.Error("Init should return a command")
	}
}

func TestView(t *testing.T) {
	m := initialModel("")
	view := m.View()

	// View should return non-empty string
	if view == "" {
		t.Error("View should return non-empty string")
	}
}

func TestUpdateQuit(t *testing.T) {
	m := initialModel("")

	// Test quit command with ctrl+c
	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	resultModel := newModel.(model)

	// Should return quit command
	if cmd == nil {
		t.Error("Should return quit command")
	}

	// Model should remain unchanged
	if resultModel.step != m.step {
		t.Error("Step should not change on quit")
	}
}

func TestUpdateNavigation(t *testing.T) {
	m := initialModel("")
	m.textInput.SetValue("testproject")

	// Test moving forward with enter
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	resultModel := newModel.(model)

	if resultModel.step != stepDatabase {
		t.Error("Should move to database step on enter after project name")
	}

	if resultModel.projectName != "testproject" {
		t.Error("Project name should be set")
	}
}

func TestUpdateDatabaseSelection(t *testing.T) {
	m := initialModel("")
	m.step = stepDatabase
	m.choices = []string{"PostgreSQL", "MySQL", "SQLite", "None"}
	m.currentChoice = 0

	// Test selecting PostgreSQL
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	resultModel := newModel.(model)

	if resultModel.database != "postgres" {
		t.Errorf("Database should be 'postgres', got '%s'", resultModel.database)
	}

	if resultModel.step != stepCache {
		t.Error("Should move to cache step after database selection")
	}
}

func TestUpdateArrowKeys(t *testing.T) {
	m := initialModel("")
	m.step = stepDatabase
	m.choices = []string{"PostgreSQL", "MySQL", "SQLite", "None"}
	m.currentChoice = 0

	// Test down arrow
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	resultModel := newModel.(model)

	if resultModel.currentChoice != 1 {
		t.Error("Should move down in choices")
	}

	// Test up arrow
	newModel2, _ := resultModel.Update(tea.KeyMsg{Type: tea.KeyUp})
	resultModel2 := newModel2.(model)

	if resultModel2.currentChoice != 0 {
		t.Error("Should move up in choices")
	}
}
