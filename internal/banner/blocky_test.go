package banner

import (
	"strings"
	"testing"
)

func TestBlockyFunctions(t *testing.T) {
	tests := []struct {
		name     string
		function func() string
		contains []string
	}{
		{
			name:     "BlockyText",
			function: BlockyText,
			contains: []string{"██", "╗", "╚"},
		},
		{
			name:     "MediumBlocky",
			function: MediumBlocky,
			contains: []string{"██", "The Official CLI for Velocity Web Framework"},
		},
		{
			name:     "CompactBlocky",
			function: CompactBlocky,
			contains: []string{"█", "╗"},
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
