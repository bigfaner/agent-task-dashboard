package model

import "testing"

func TestErrorTypes(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{"ErrProjectNotFound", ErrProjectNotFound("my-project"), "project not found: my-project"},
		{"ErrFeatureNotFound", ErrFeatureNotFound("my-feature"), "feature not found: my-feature"},
		{"ErrTaskNotFound", ErrTaskNotFound("1.1"), "task not found: 1.1"},
		{"ErrConfigInvalid", ErrConfigInvalid("bad yaml"), "config invalid: bad yaml"},
		{"ErrFSRead", ErrFSRead("read error"), "filesystem read error: read error"},
		{"ErrParseIndex", ErrParseIndex("parse error"), "index parse error: parse error"},
		{"ErrInvalidSlug", ErrInvalidSlug("bad slug!"), "invalid slug: bad slug!"},
		{"ErrInvalidTaskID", ErrInvalidTaskID("../traversal"), "invalid task id: ../traversal"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.expected {
				t.Errorf("Error() = %q, want %q", tt.err.Error(), tt.expected)
			}
		})
	}
}
