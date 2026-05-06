package scanner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/panda/agent-task-center/internal/model"
)

// ---- ParseTaskFile tests ----

func TestParseTaskFile_WithAcceptanceCriteria(t *testing.T) {
	content := `---
id: "2.1"
title: "Task Markdown Parser"
priority: "P0"
---

# 2.1: Task Markdown Parser

## Description

Implement parsers for task files.

## Acceptance Criteria

- [ ] ParseTaskFile extracts acceptance criteria
- [ ] ParseRecordFile parses structured sections
- [x] Already done criterion
- [ ] Returns error for missing files

## Implementation Notes

Use simple line-by-line parsing.
`
	tmpDir := t.TempDir()
	fPath := filepath.Join(tmpDir, "2.1-parser.md")
	if err := os.WriteFile(fPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result, err := ParseTaskFile(fPath)
	if err != nil {
		t.Fatalf("ParseTaskFile() returned error: %v", err)
	}
	if result == nil {
		t.Fatal("ParseTaskFile() returned nil")
	}

	if len(result.AcceptanceCriteria) != 4 {
		t.Fatalf("expected 4 acceptance criteria, got %d: %v", len(result.AcceptanceCriteria), result.AcceptanceCriteria)
	}

	expected := []string{
		"ParseTaskFile extracts acceptance criteria",
		"ParseRecordFile parses structured sections",
		"Already done criterion",
		"Returns error for missing files",
	}
	for i, ac := range result.AcceptanceCriteria {
		if ac != expected[i] {
			t.Errorf("criteria[%d] = %q, want %q", i, ac, expected[i])
		}
	}
}

func TestParseTaskFile_WithScope(t *testing.T) {
	content := `---
id: "2.1"
scope: "backend"
---

# 2.1: Some Task

## Description

Some description text.
`
	tmpDir := t.TempDir()
	fPath := filepath.Join(tmpDir, "2.1-task.md")
	if err := os.WriteFile(fPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result, err := ParseTaskFile(fPath)
	if err != nil {
		t.Fatalf("ParseTaskFile() returned error: %v", err)
	}

	if result.Scope != "backend" {
		t.Errorf("Scope = %q, want %q", result.Scope, "backend")
	}
}

func TestParseTaskFile_DescriptionExcludesAC(t *testing.T) {
	content := `---
id: "2.1"
---

# 2.1: Some Task

## Description

This is the description paragraph.

## Acceptance Criteria

- [ ] First criterion
- [ ] Second criterion

## Implementation Notes

Some notes here.
`
	tmpDir := t.TempDir()
	fPath := filepath.Join(tmpDir, "2.1-task.md")
	if err := os.WriteFile(fPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result, err := ParseTaskFile(fPath)
	if err != nil {
		t.Fatalf("ParseTaskFile() returned error: %v", err)
	}

	if !strings.Contains(result.Description, "This is the description paragraph.") {
		t.Errorf("Description should contain description text, got: %q", result.Description)
	}
	if strings.Contains(result.Description, "First criterion") {
		t.Errorf("Description should NOT contain acceptance criteria text, got: %q", result.Description)
	}
}

func TestParseTaskFile_NoAcceptanceCriteria(t *testing.T) {
	content := `---
id: "2.1"
---

# 2.1: Some Task

## Description

Just a description, no AC section.
`
	tmpDir := t.TempDir()
	fPath := filepath.Join(tmpDir, "2.1-task.md")
	if err := os.WriteFile(fPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result, err := ParseTaskFile(fPath)
	if err != nil {
		t.Fatalf("ParseTaskFile() returned error: %v", err)
	}

	if len(result.AcceptanceCriteria) != 0 {
		t.Errorf("expected 0 acceptance criteria, got %d", len(result.AcceptanceCriteria))
	}
}

func TestParseTaskFile_MissingFile(t *testing.T) {
	_, err := ParseTaskFile("/nonexistent/path/task.md")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestParseTaskFile_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	fPath := filepath.Join(tmpDir, "empty.md")
	if err := os.WriteFile(fPath, []byte(""), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result, err := ParseTaskFile(fPath)
	if err != nil {
		t.Fatalf("ParseTaskFile() returned error: %v", err)
	}
	if result == nil {
		t.Fatal("ParseTaskFile() returned nil for empty file")
	}
	if len(result.AcceptanceCriteria) != 0 {
		t.Errorf("expected 0 acceptance criteria for empty file, got %d", len(result.AcceptanceCriteria))
	}
}

// ---- ParseRecordFile tests ----

func TestParseRecordFile_AllSections(t *testing.T) {
	content := `---
status: "completed"
started: "2026-05-07 01:04"
completed: "2026-05-07 01:06"
time_spent: "~2m"
---

# Task Record: 1.2 Define Data Models

## Summary
Implemented all Go data model structs from the tech design.

## Changes

### Files Created
- internal/model/project.go
- internal/model/feature.go
- internal/model/task.go

### Files Modified
- internal/model/errors.go

### Key Decisions
- Error types implemented as string-based types
- DerivePhase handles edge cases

## Test Results
- **Passed**: 17
- **Failed**: 0
- **Coverage**: 89.3%
`
	tmpDir := t.TempDir()
	fPath := filepath.Join(tmpDir, "record.md")
	if err := os.WriteFile(fPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result, err := ParseRecordFile(fPath)
	if err != nil {
		t.Fatalf("ParseRecordFile() returned error: %v", err)
	}
	if result == nil {
		t.Fatal("ParseRecordFile() returned nil")
	}

	// Summary
	if !strings.Contains(result.Summary, "Implemented all Go data model structs") {
		t.Errorf("Summary = %q, should contain model structs text", result.Summary)
	}

	// Files
	if len(result.Files) != 4 {
		t.Fatalf("expected 4 files, got %d: %v", len(result.Files), result.Files)
	}
	expectedFiles := []string{
		"internal/model/project.go",
		"internal/model/feature.go",
		"internal/model/task.go",
		"internal/model/errors.go",
	}
	for i, f := range result.Files {
		if f != expectedFiles[i] {
			t.Errorf("Files[%d] = %q, want %q", i, f, expectedFiles[i])
		}
	}

	// Decisions
	if !strings.Contains(result.Decisions, "Error types implemented as string-based types") {
		t.Errorf("Decisions = %q, should contain error types decision", result.Decisions)
	}

	// TestResults
	if !strings.Contains(result.TestResults, "Passed") || !strings.Contains(result.TestResults, "17") {
		t.Errorf("TestResults = %q, should contain Passed count", result.TestResults)
	}

	// Raw should contain the full content
	if !strings.Contains(result.Raw, "# Task Record") {
		t.Errorf("Raw should contain full markdown content")
	}
}

func TestParseRecordFile_NoHeadings(t *testing.T) {
	content := `Just some random text without any markdown headings.
No structure here at all.
`
	tmpDir := t.TempDir()
	fPath := filepath.Join(tmpDir, "record.md")
	if err := os.WriteFile(fPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result, err := ParseRecordFile(fPath)
	if err != nil {
		t.Fatalf("ParseRecordFile() returned error: %v", err)
	}
	if result == nil {
		t.Fatal("ParseRecordFile() returned nil")
	}

	// All structured fields should be empty
	if result.Summary != "" {
		t.Errorf("Summary = %q, want empty for no headings", result.Summary)
	}
	if len(result.Files) != 0 {
		t.Errorf("Files = %v, want empty for no headings", result.Files)
	}
	if result.Decisions != "" {
		t.Errorf("Decisions = %q, want empty for no headings", result.Decisions)
	}
	if result.TestResults != "" {
		t.Errorf("TestResults = %q, want empty for no headings", result.TestResults)
	}

	// Raw should be populated
	if result.Raw != content {
		t.Errorf("Raw = %q, want %q", result.Raw, content)
	}
}

func TestParseRecordFile_MissingFile(t *testing.T) {
	_, err := ParseRecordFile("/nonexistent/path/record.md")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestParseRecordFile_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	fPath := filepath.Join(tmpDir, "empty.md")
	if err := os.WriteFile(fPath, []byte(""), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result, err := ParseRecordFile(fPath)
	if err != nil {
		t.Fatalf("ParseRecordFile() returned error: %v", err)
	}
	if result == nil {
		t.Fatal("ParseRecordFile() returned nil for empty file")
	}
	if result.Raw != "" {
		t.Errorf("Raw = %q, want empty for empty file", result.Raw)
	}
}

func TestParseRecordFile_PartialSections(t *testing.T) {
	content := `# Task Record: 2.1 Parser

## Summary
Only summary, no other sections.

## Acceptance Criteria
- [x] Some criterion
`
	tmpDir := t.TempDir()
	fPath := filepath.Join(tmpDir, "record.md")
	if err := os.WriteFile(fPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result, err := ParseRecordFile(fPath)
	if err != nil {
		t.Fatalf("ParseRecordFile() returned error: %v", err)
	}

	if !strings.Contains(result.Summary, "Only summary") {
		t.Errorf("Summary = %q, should contain summary text", result.Summary)
	}
	if len(result.Files) != 0 {
		t.Errorf("Files = %v, want empty when no Files section", result.Files)
	}
	if result.Decisions != "" {
		t.Errorf("Decisions = %q, want empty when no Decisions section", result.Decisions)
	}
	if result.TestResults != "" {
		t.Errorf("TestResults = %q, want empty when no Test Results section", result.TestResults)
	}
}

func TestParseRecordFile_FilesWithNoPaths(t *testing.T) {
	content := `# Task Record

## Summary
Some summary.

## Changes

### Files Created

No files created.

### Files Modified

No files modified.

## Test Results
All passing.
`
	tmpDir := t.TempDir()
	fPath := filepath.Join(tmpDir, "record.md")
	if err := os.WriteFile(fPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result, err := ParseRecordFile(fPath)
	if err != nil {
		t.Fatalf("ParseRecordFile() returned error: %v", err)
	}

	// Files should be empty since no lines start with "- "
	if len(result.Files) != 0 {
		t.Errorf("Files = %v, want empty when no file paths present", result.Files)
	}
}

func TestParseRecordFile_FilesParsePathsFromLines(t *testing.T) {
	content := `# Task Record

## Summary
Made changes.

## Changes

### Files Created
- cmd/app/main.go
- internal/parser.go

### Files Modified
- internal/model/errors.go

### Key Decisions
Used simple approach.

## Test Results
Passed: 5
`
	tmpDir := t.TempDir()
	fPath := filepath.Join(tmpDir, "record.md")
	if err := os.WriteFile(fPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result, err := ParseRecordFile(fPath)
	if err != nil {
		t.Fatalf("ParseRecordFile() returned error: %v", err)
	}

	expectedFiles := []string{
		"cmd/app/main.go",
		"internal/parser.go",
		"internal/model/errors.go",
	}
	if len(result.Files) != len(expectedFiles) {
		t.Fatalf("expected %d files, got %d: %v", len(expectedFiles), len(result.Files), result.Files)
	}
	for i, f := range result.Files {
		if f != expectedFiles[i] {
			t.Errorf("Files[%d] = %q, want %q", i, f, expectedFiles[i])
		}
	}
}

func TestParseRecordFile_DecisionsContent(t *testing.T) {
	content := `# Task Record

## Summary
Summary here.

## Key Decisions
- Decision one: because of reason A
- Decision two: for performance

## Test Results
- Passed: 10
- Coverage: 85%
`
	tmpDir := t.TempDir()
	fPath := filepath.Join(tmpDir, "record.md")
	if err := os.WriteFile(fPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result, err := ParseRecordFile(fPath)
	if err != nil {
		t.Fatalf("ParseRecordFile() returned error: %v", err)
	}

	if !strings.Contains(result.Decisions, "Decision one") {
		t.Errorf("Decisions = %q, should contain 'Decision one'", result.Decisions)
	}
	if !strings.Contains(result.Decisions, "Decision two") {
		t.Errorf("Decisions = %q, should contain 'Decision two'", result.Decisions)
	}
}

// ---- Integration test: parse a real task .md file ----

func TestIntegration_ParseRealTaskFile(t *testing.T) {
	// Use the actual task file from the project if it exists
	realPath := filepath.Join("..", "..", "docs", "features", "agent-task-dashboard", "tasks", "1.1-project-init.md")
	absPath, _ := filepath.Abs(realPath)

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		t.Skip("Skipping: real task file not found at", absPath)
	}

	result, err := ParseTaskFile(absPath)
	if err != nil {
		t.Fatalf("ParseTaskFile() returned error: %v", err)
	}

	if len(result.AcceptanceCriteria) == 0 {
		t.Error("expected at least some acceptance criteria from real task file")
	}
}

// ---- Integration test: parse a real record file ----

func TestIntegration_ParseRealRecordFile(t *testing.T) {
	realPath := filepath.Join("..", "..", "docs", "features", "agent-task-dashboard", "tasks", "records", "1.2-data-models.md")
	absPath, _ := filepath.Abs(realPath)

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		t.Skip("Skipping: real record file not found at", absPath)
	}

	result, err := ParseRecordFile(absPath)
	if err != nil {
		t.Fatalf("ParseRecordFile() returned error: %v", err)
	}

	if result.Summary == "" {
		t.Error("expected non-empty Summary from real record file")
	}
	if len(result.Files) == 0 {
		t.Error("expected at least some files from real record file")
	}
	if result.TestResults == "" {
		t.Error("expected non-empty TestResults from real record file")
	}
}

// ---- Edge case: Raw field always populated ----

func TestParseRecordFile_RawFieldAlwaysPopulated(t *testing.T) {
	content := `# Task Record

## Summary
Test summary.
`
	tmpDir := t.TempDir()
	fPath := filepath.Join(tmpDir, "record.md")
	if err := os.WriteFile(fPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	result, err := ParseRecordFile(fPath)
	if err != nil {
		t.Fatalf("ParseRecordFile() returned error: %v", err)
	}

	if result.Raw != content {
		t.Errorf("Raw = %q, want exact file content", result.Raw)
	}
}

// ---- Verify types from model package ----

func TestRecordContentType(t *testing.T) {
	r := &model.RecordContent{
		Summary:     "test",
		Files:       []string{"a.go", "b.go"},
		Decisions:   "decided X",
		TestResults: "passed 5",
		Raw:         "raw content",
	}
	if r.Summary != "test" {
		t.Errorf("Summary = %q, want %q", r.Summary, "test")
	}
}

func TestTaskFileContentType(t *testing.T) {
	tf := &model.TaskFileContent{
		AcceptanceCriteria: []string{"criterion 1", "criterion 2"},
		Scope:              "backend",
		Description:        "some description",
	}
	if len(tf.AcceptanceCriteria) != 2 {
		t.Errorf("expected 2 criteria, got %d", len(tf.AcceptanceCriteria))
	}
}
