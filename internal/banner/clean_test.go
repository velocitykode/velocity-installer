package banner

import (
	"strings"
	"testing"
)

func TestCleanBannerFunctions(t *testing.T) {
	tests := []struct {
		name     string
		function func() string
		contains []string
	}{
		{
			name:     "Clean",
			function: Clean,
			contains: []string{"VELOCITY CLI", "The Go Web Framework"},
		},
		{
			name:     "CleanBox",
			function: CleanBox,
			contains: []string{"VELOCITY CLI", "The Go Web Framework", "┌", "└"},
		},
		{
			name:     "Title",
			function: Title,
			contains: []string{"VELOCITY CLI"},
		},
		{
			name:     "Divider",
			function: Divider,
			contains: []string{"────"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.function()
			if result == "" {
				t.Errorf("%s() returned empty string", tt.name)
			}
			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("%s() result does not contain %q", tt.name, expected)
				}
			}
		})
	}
}
