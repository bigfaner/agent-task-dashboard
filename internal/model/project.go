package model

import "time"

// ProjectData holds aggregated information about a project scanned from the filesystem.
type ProjectData struct {
	ID             string        `json:"id"`
	Name           string        `json:"name"`
	Path           string        `json:"path"`
	Features       []FeatureData `json:"features"`
	LastUpdated    time.Time     `json:"lastUpdated"`
	HealthStatus   string        `json:"healthStatus"`
	TotalTasks     int           `json:"totalTasks"`
	CompletedTasks int           `json:"completedTasks"`
	Warnings       []string      `json:"warnings"`
}
