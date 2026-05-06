package model

import "time"

// ActivityEvent represents a single activity event for the sidebar.
type ActivityEvent struct {
	Timestamp time.Time `json:"timestamp"`
	TaskID    string    `json:"taskId"`
	TaskTitle string    `json:"taskTitle"`
	Feature   string    `json:"feature"`
	EventType string    `json:"eventType"`
}
