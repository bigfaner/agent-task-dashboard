package model

// DependencyGraph represents the task dependency graph for rendering.
type DependencyGraph struct {
	Nodes []GraphNode `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
}

// GraphNode represents a single task node in the dependency graph.
type GraphNode struct {
	ID      string `json:"id"`
	Key     string `json:"key"`
	Title   string `json:"title"`
	Status  string `json:"status"`
	Phase   int    `json:"phase"`
	Feature string `json:"feature"`
}

// GraphEdge represents a dependency edge between two tasks.
type GraphEdge struct {
	Source       string `json:"source"`
	Target       string `json:"target"`
	CrossFeature bool   `json:"crossFeature"`
}
