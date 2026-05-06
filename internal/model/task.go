package model

import "strconv"

// Task represents a single task within a feature.
type Task struct {
	ID            string   `json:"id"`
	Key           string   `json:"key"`
	Title         string   `json:"title"`
	Priority      string   `json:"priority"`
	Status        string   `json:"status"`
	Scope         string   `json:"scope"`
	EstimatedTime string   `json:"estimatedTime,omitempty"`
	Dependencies  []string `json:"dependencies"`
	Breaking      bool     `json:"breaking"`
	File          string   `json:"file"`
	Record        string   `json:"record"`
	Phase         int      `json:"phase"`
}

// DerivePhase extracts the leading phase number from a task ID.
// "1.1" -> 1, "2.3" -> 2, "10.2" -> 10.
// Non-numeric prefixes (e.g., "T-test-1") return 0.
// Empty strings return 0.
func DerivePhase(id string) int {
	if id == "" {
		return 0
	}
	for i, ch := range id {
		if ch == '.' {
			if i == 0 {
				return 0
			}
			n, err := strconv.Atoi(id[:i])
			if err != nil {
				return 0
			}
			if n > 0 {
				return n
			}
			return 0
		}
		if ch < '0' || ch > '9' {
			break
		}
	}
	// No dot found — try parsing entire string as number
	n, err := strconv.Atoi(id)
	if err != nil {
		return 0
	}
	if n > 0 {
		return n
	}
	return 0
}
