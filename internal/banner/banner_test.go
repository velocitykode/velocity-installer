package banner

import (
	"strings"
	"testing"
)

func TestBannerFunctions(t *testing.T) {
	tests := []struct {
		name     string
		function func() string
		contains []string
	}{
		{
			name:     "Simple",
			function: Simple,
			contains: []string{"██", "╗", "╚"},
		},
		{
			name:     "Block",
			function: Block,
			contains: []string{"█", "The Go Web Framework"},
		},
		{
			name:     "CompactBox",
			function: CompactBox,
			contains: []string{"╔", "╚", "║"},
		},
		{
			name:     "Retro",
			function: Retro,
			contains: []string{"▄", "▀", "██"},
		},
		{
			name:     "Shadow",
			function: Shadow,
			contains: []string{"██", "╗", "╚"},
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
