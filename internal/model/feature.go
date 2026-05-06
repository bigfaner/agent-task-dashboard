package model

import "time"

// FeatureData holds information about a single feature within a project.
type FeatureData struct {
	Slug            string          `json:"slug"`
	Status          string          `json:"status"`
	PRDPath         string          `json:"prdPath"`
	DesignPath      string          `json:"designPath"`
	Tasks           map[string]Task `json:"tasks"`
	Phases          []PhaseInfo     `json:"phases"`
	LastUpdated     time.Time       `json:"lastUpdated"`
	TotalTasks      int             `json:"totalTasks"`
	CompletedTasks  int             `json:"completedTasks"`
	HasBlockedTasks bool            `json:"hasBlockedTasks"`
	CompletionPct   float64         `json:"completionPct"`
}

// PhaseInfo holds metadata about a single phase within a feature.
type PhaseInfo struct {
	Number   int      `json:"number"`
	Label    string   `json:"label"`
	TaskKeys []string `json:"taskKeys"`
}
