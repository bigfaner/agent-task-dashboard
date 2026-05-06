package model

// RecordContent holds the structured sections parsed from an execution record.
type RecordContent struct {
	Summary     string   `json:"summary"`
	Files       []string `json:"files"`
	Decisions   string   `json:"decisions"`
	TestResults string   `json:"testResults"`
	Raw         string   `json:"raw"`
}

// TaskFileContent holds parsed content from a task markdown file.
type TaskFileContent struct {
	AcceptanceCriteria []string `json:"acceptanceCriteria"`
	Scope              string   `json:"scope"`
	Description        string   `json:"description"`
}
