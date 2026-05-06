package model

import (
	"encoding/json"
	"testing"
	"time"
)

func TestProjectDataJSONTags(t *testing.T) {
	pd := ProjectData{
		ID:             "pm-work-tracker",
		Name:           "pm-work-tracker",
		Path:           "/some/path",
		Features:       []FeatureData{},
		LastUpdated:    time.Date(2026, 5, 6, 14, 30, 0, 0, time.UTC),
		HealthStatus:   "active",
		TotalTasks:     330,
		CompletedTasks: 312,
		Warnings:       []string{},
	}

	data, err := json.Marshal(pd)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	expectedKeys := []string{"id", "name", "path", "features", "lastUpdated", "healthStatus", "totalTasks", "completedTasks", "warnings"}
	for _, key := range expectedKeys {
		if _, ok := result[key]; !ok {
			t.Errorf("missing JSON key %q", key)
		}
	}
}

func TestFeatureDataJSONTags(t *testing.T) {
	fd := FeatureData{
		Slug:            "improve-ui",
		Status:          "in-progress",
		PRDPath:         "prd/prd-spec.md",
		DesignPath:      "design/tech-design.md",
		Tasks:           map[string]Task{},
		Phases:          []PhaseInfo{},
		LastUpdated:     time.Date(2026, 5, 6, 12, 0, 0, 0, time.UTC),
		TotalTasks:      18,
		CompletedTasks:  13,
		HasBlockedTasks: false,
		CompletionPct:   72.2,
	}

	data, err := json.Marshal(fd)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	expectedKeys := []string{"slug", "status", "prdPath", "designPath", "tasks", "phases", "lastUpdated", "totalTasks", "completedTasks", "hasBlockedTasks", "completionPct"}
	for _, key := range expectedKeys {
		if _, ok := result[key]; !ok {
			t.Errorf("missing JSON key %q", key)
		}
	}
}

func TestDependencyGraphJSONTags(t *testing.T) {
	graph := DependencyGraph{
		Nodes: []GraphNode{
			{ID: "1.1", Key: "1.1-interfaces", Title: "Define core interfaces", Status: "completed", Phase: 1, Feature: "improve-ui"},
		},
		Edges: []GraphEdge{
			{Source: "1.2", Target: "1.1", CrossFeature: false},
		},
	}

	data, err := json.Marshal(graph)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	nodes, ok := result["nodes"].([]interface{})
	if !ok || len(nodes) != 1 {
		t.Fatal("expected 1 node")
	}

	node := nodes[0].(map[string]interface{})
	nodeKeys := []string{"id", "key", "title", "status", "phase", "feature"}
	for _, key := range nodeKeys {
		if _, ok := node[key]; !ok {
			t.Errorf("missing node JSON key %q", key)
		}
	}

	edges, ok := result["edges"].([]interface{})
	if !ok || len(edges) != 1 {
		t.Fatal("expected 1 edge")
	}

	edge := edges[0].(map[string]interface{})
	edgeKeys := []string{"source", "target", "crossFeature"}
	for _, key := range edgeKeys {
		if _, ok := edge[key]; !ok {
			t.Errorf("missing edge JSON key %q", key)
		}
	}
}

func TestActivityEventJSONTags(t *testing.T) {
	evt := ActivityEvent{
		Timestamp: time.Date(2026, 5, 6, 14, 30, 0, 0, time.UTC),
		TaskID:    "1.1",
		TaskTitle: "Define core interfaces",
		Feature:   "improve-ui",
		EventType: "completed",
	}

	data, err := json.Marshal(evt)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	expectedKeys := []string{"timestamp", "taskId", "taskTitle", "feature", "eventType"}
	for _, key := range expectedKeys {
		if _, ok := result[key]; !ok {
			t.Errorf("missing JSON key %q", key)
		}
	}
}

func TestRecordContentJSONTags(t *testing.T) {
	rc := RecordContent{
		Summary:     "Implemented core interfaces",
		Files:       []string{"internal/model/types.go"},
		Decisions:   "Used fs.FS abstraction",
		TestResults: "All 12 tests passing",
		Raw:         "## Summary\nImplemented core interfaces...",
	}

	data, err := json.Marshal(rc)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	expectedKeys := []string{"summary", "files", "decisions", "testResults", "raw"}
	for _, key := range expectedKeys {
		if _, ok := result[key]; !ok {
			t.Errorf("missing JSON key %q", key)
		}
	}
}

func TestTaskFileContentJSONTags(t *testing.T) {
	tfc := TaskFileContent{
		AcceptanceCriteria: []string{"Interface defines all CRUD methods"},
		Scope:              "all",
		Description:        "Define core interfaces",
	}

	data, err := json.Marshal(tfc)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	expectedKeys := []string{"acceptanceCriteria", "scope", "description"}
	for _, key := range expectedKeys {
		if _, ok := result[key]; !ok {
			t.Errorf("missing JSON key %q", key)
		}
	}
}
