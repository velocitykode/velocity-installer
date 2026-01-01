package version

import (
	"testing"
)

func TestGetSystemGoVersion(t *testing.T) {
	version, err := getSystemGoVersion()
	if err != nil {
		t.Skipf("Go not installed or not in PATH: %v", err)
	}

	if version == "" {
		t.Error("getSystemGoVersion() returned empty string")
	}

	// Should start with "go"
	if len(version) < 3 || version[:2] != "go" {
		t.Errorf("getSystemGoVersion() = %q, expected to start with 'go'", version)
	}
}

func TestParseGoVersion(t *testing.T) {
	tests := []struct {
		name      string
		version   string
		wantMajor int
		wantMinor int
		wantErr   bool
	}{
		{
			name:      "standard version with go prefix",
			version:   "go1.25",
			wantMajor: 1,
			wantMinor: 25,
			wantErr:   false,
		},
		{
			name:      "version with patch number",
			version:   "go1.25.1",
			wantMajor: 1,
			wantMinor: 25,
			wantErr:   false,
		},
		{
			name:      "version without go prefix",
			version:   "1.25",
			wantMajor: 1,
			wantMinor: 25,
			wantErr:   false,
		},
		{
			name:      "older version",
			version:   "go1.23.6",
			wantMajor: 1,
			wantMinor: 23,
			wantErr:   false,
		},
		{
			name:      "rc version",
			version:   "go1.25rc1",
			wantMajor: 1,
			wantMinor: 25,
			wantErr:   false,
		},
		{
			name:      "beta version",
			version:   "go1.25beta1",
			wantMajor: 1,
			wantMinor: 25,
			wantErr:   false,
		},
		{
			name:    "invalid format - single number",
			version: "1",
			wantErr: true,
		},
		{
			name:    "invalid format - empty",
			version: "",
			wantErr: true,
		},
		{
			name:    "invalid format - no numbers",
			version: "go.abc",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			major, minor, err := parseGoVersion(tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseGoVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if major != tt.wantMajor {
					t.Errorf("parseGoVersion() major = %v, want %v", major, tt.wantMajor)
				}
				if minor != tt.wantMinor {
					t.Errorf("parseGoVersion() minor = %v, want %v", minor, tt.wantMinor)
				}
			}
		})
	}
}

func TestCheckMinimumGoVersion(t *testing.T) {
	tests := []struct {
		name    string
		minimum string
		wantErr bool
	}{
		{
			name:    "current version meets minimum",
			minimum: "1.20",
			wantErr: false,
		},
		{
			name:    "minimum is 1.0",
			minimum: "1.0",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckMinimumGoVersion(tt.minimum)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckMinimumGoVersion() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVersionError(t *testing.T) {
	err := &VersionError{
		Current: "go1.23.6",
		Minimum: "1.25",
	}

	msg := err.Error()

	if msg == "" {
		t.Error("VersionError.Error() returned empty string")
	}

	// Check that the error message contains key information
	if !contains(msg, "go1.23.6") {
		t.Error("VersionError.Error() should contain current version")
	}
	if !contains(msg, "1.25") {
		t.Error("VersionError.Error() should contain minimum version")
	}
	if !contains(msg, "brew upgrade go") {
		t.Error("VersionError.Error() should contain upgrade instructions")
	}
}

func TestGoNotFoundError(t *testing.T) {
	err := &GoNotFoundError{}
	msg := err.Error()

	if msg == "" {
		t.Error("GoNotFoundError.Error() returned empty string")
	}

	if !contains(msg, "not installed") {
		t.Error("GoNotFoundError.Error() should mention Go not installed")
	}
	if !contains(msg, "brew install go") {
		t.Error("GoNotFoundError.Error() should contain install instructions")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
