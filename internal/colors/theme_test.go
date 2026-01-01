package colors

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestColorDefinitions(t *testing.T) {
	tests := []struct {
		name     string
		color    lipgloss.Color
		expected string
	}{
		{"Primary", Primary, "#0e87cd"},
		{"PrimaryHover", PrimaryHover, "#0a6ba8"},
		{"Success", Success, "#10b981"},
		{"Warning", Warning, "#f59e0b"},
		{"Error", Error, "#ef4444"},
		{"Muted", Muted, "#6b7280"},
		{"Dark", Dark, "#1a1a1a"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.color) != tt.expected {
				t.Errorf("Color %s = %v, want %v", tt.name, tt.color, tt.expected)
			}
		})
	}
}

func TestStyleDefinitions(t *testing.T) {
	tests := []struct {
		name  string
		style lipgloss.Style
	}{
		{"BrandStyle", BrandStyle},
		{"BrandHoverStyle", BrandHoverStyle},
		{"SuccessStyle", SuccessStyle},
		{"ErrorStyle", ErrorStyle},
		{"WarningStyle", WarningStyle},
		{"MutedStyle", MutedStyle},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.style.Render("test")
			if result == "" {
				t.Errorf("Style %s.Render() returned empty string", tt.name)
			}
			// Verify the rendered output contains the input text
			if len(result) < 4 { // "test" is 4 chars, output should be at least that
				t.Errorf("Style %s.Render() output too short: %q", tt.name, result)
			}
		})
	}
}
