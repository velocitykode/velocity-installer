package banner

import (
	"strings"
	"testing"
)

func TestSmallBannerFunctions(t *testing.T) {
	tests := []struct {
		name     string
		function func() string
		contains []string
	}{
		{
			name:     "SmallSimple",
			function: SmallSimple,
			contains: []string{"██", "╗", "╚"},
		},
		{
			name:     "Small",
			function: Small,
			contains: []string{"██", "╔", "╚", "CLI"},
		},
		{
			name:     "Minimal",
			function: Minimal,
			contains: []string{"VELOCITY CLI", "Web Framework"},
		},
		{
			name:     "Compact",
			function: Compact,
			contains: []string{"VELOCITY CLI", "The Go Web Framework"},
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
