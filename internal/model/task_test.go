package model

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestDerivePhase(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"1.1", 1},
		{"2.3", 2},
		{"3.5-data-models", 3},
		{"10.2", 10},
		{"T-test-1", 0},
		{"abc", 0},
		{"", 0},
		{"1", 1},
		{"0.1", 0},
		{"12.4-refactor", 12},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := DerivePhase(tt.input)
			if got != tt.expected {
				t.Errorf("DerivePhase(%q) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}

func TestTaskJSONTags(t *testing.T) {
	task := Task{
		ID:            "1.1",
		Key:           "1.1-interfaces",
		Title:         "Define core interfaces",
		Priority:      "P0",
		Status:        "completed",
		Scope:         "all",
		EstimatedTime: "1h",
		Dependencies:  []string{"1.0"},
		Breaking:      true,
		File:          "1.1-interfaces.md",
		Record:        "records/1.1-interfaces.md",
		Phase:         1,
	}

	data, err := json.Marshal(task)
	if err != nil {
		t.Fatalf("failed to marshal Task: %v", err)
	}

	// Verify JSON keys match api-handbook.md
	expected := map[string]interface{}{
		"id":            "1.1",
		"key":           "1.1-interfaces",
		"title":         "Define core interfaces",
		"priority":      "P0",
		"status":        "completed",
		"scope":         "all",
		"estimatedTime": "1h",
		"dependencies":  []interface{}{"1.0"},
		"breaking":      true,
		"file":          "1.1-interfaces.md",
		"record":        "records/1.1-interfaces.md",
		"phase":         float64(1),
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	for key, want := range expected {
		got, ok := result[key]
		if !ok {
			t.Errorf("missing JSON key %q", key)
			continue
		}
		if fmt.Sprintf("%v", got) != fmt.Sprintf("%v", want) {
			t.Errorf("JSON key %q: got %v, want %v", key, got, want)
		}
	}
}
