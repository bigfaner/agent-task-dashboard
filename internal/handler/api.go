package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/panda/agent-task-center/internal/model"
	"github.com/panda/agent-task-center/internal/scanner"
)

// RegisterAPI registers JSON API routes under /api on the Gin engine.
func RegisterAPI(r *gin.Engine, s *scanner.Scanner) {
	api := r.Group("/api")
	api.Use(ErrorHandler())

	// GET /api/projects — list all projects with summaries
	api.GET("/projects", handleListProjects(s))

	// GET /api/projects/:id — single project with feature list
	api.GET("/projects/:id", handleGetProject(s))

	// GET /api/projects/:id/features — list features
	api.GET("/projects/:id/features", handleListFeatures(s))

	featureRoutes := api.Group("/projects/:id/features/:slug")
	featureRoutes.Use(ValidateSlug())
	{
		// GET /api/projects/:id/features/:slug — single feature with tasks
		featureRoutes.GET("", handleGetFeature(s))

		// GET /api/projects/:id/features/:slug/tasks — list tasks
		featureRoutes.GET("/tasks", handleListTasks(s))
	}

	taskRoutes := featureRoutes.Group("/tasks/:taskId")
	taskRoutes.Use(ValidateTaskID())
	{
		// GET /api/projects/:id/features/:slug/tasks/:taskId — single task
		taskRoutes.GET("", handleGetTask(s))
	}

	// GET /api/projects/:id/features/:slug/dependencies — dependency graph
	featureRoutes.GET("/dependencies", handleGetDependencies(s))
}

// ---- Response types ----

// metaResponse is included in every API response.
type metaResponse struct {
	LastUpdated string `json:"lastUpdated"`
}

func makeMeta(t time.Time) metaResponse {
	if t.IsZero() {
		return metaResponse{LastUpdated: time.Now().UTC().Format(time.RFC3339)}
	}
	return metaResponse{LastUpdated: t.UTC().Format(time.RFC3339)}
}

// projectSummary is the summary object returned in list projects.
type projectSummary struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	FeatureCount   int     `json:"featureCount"`
	CompletedTasks int     `json:"completedTasks"`
	TotalTasks     int     `json:"totalTasks"`
	CompletionPct  float64 `json:"completionPct"`
	LastUpdated    string  `json:"lastUpdated"`
	HealthStatus   string  `json:"healthStatus"`
}

// featureSummary is returned in the list features endpoint and within get project.
type featureSummary struct {
	Slug           string `json:"slug"`
	Status         string `json:"status"`
	CompletedTasks int    `json:"completedTasks"`
	TotalTasks     int    `json:"totalTasks"`
	LastUpdated    string `json:"lastUpdated"`
}

// taskSummary is the task object returned in list tasks.
type taskSummary struct {
	ID            string   `json:"id"`
	Key           string   `json:"key"`
	Title         string   `json:"title"`
	Priority      string   `json:"priority"`
	Status        string   `json:"status"`
	Scope         string   `json:"scope"`
	EstimatedTime string   `json:"estimatedTime"`
	Dependencies  []string `json:"dependencies"`
	Breaking      bool     `json:"breaking"`
	Phase         int      `json:"phase"`
	File          string   `json:"file"`
	Record        string   `json:"record"`
}

// ---- Handler implementations ----

func handleListProjects(s *scanner.Scanner) gin.HandlerFunc {
	return func(c *gin.Context) {
		all, err := s.ScanAll()
		if err != nil {
			_ = c.Error(err)
			return
		}

		var summaries []projectSummary
		var lastUpdated time.Time

		for _, pd := range all {
			pct := computePct(pd.CompletedTasks, pd.TotalTasks)
			lu := pd.LastUpdated
			if lu.After(lastUpdated) {
				lastUpdated = lu
			}

			summaries = append(summaries, projectSummary{
				ID:             pd.ID,
				Name:           pd.Name,
				FeatureCount:   len(pd.Features),
				CompletedTasks: pd.CompletedTasks,
				TotalTasks:     pd.TotalTasks,
				CompletionPct:  pct,
				LastUpdated:    lu.UTC().Format(time.RFC3339),
				HealthStatus:   pd.HealthStatus,
			})
		}

		if summaries == nil {
			summaries = []projectSummary{}
		}

		c.JSON(http.StatusOK, gin.H{
			"projects": summaries,
			"meta":     makeMeta(lastUpdated),
		})
	}
}

func handleGetProject(s *scanner.Scanner) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		pd, err := s.ScanProject(id)
		if err != nil {
			_ = c.Error(err)
			return
		}

		pct := computePct(pd.CompletedTasks, pd.TotalTasks)

		var features []featureSummary
		for _, f := range pd.Features {
			features = append(features, featureSummary{
				Slug:           f.Slug,
				Status:         f.Status,
				CompletedTasks: f.CompletedTasks,
				TotalTasks:     f.TotalTasks,
				LastUpdated:    f.LastUpdated.UTC().Format(time.RFC3339),
			})
		}
		if features == nil {
			features = []featureSummary{}
		}

		warnings := pd.Warnings
		if warnings == nil {
			warnings = []string{}
		}

		c.JSON(http.StatusOK, gin.H{
			"id":             pd.ID,
			"name":           pd.Name,
			"path":           pd.Path,
			"featureCount":   len(pd.Features),
			"completedTasks": pd.CompletedTasks,
			"totalTasks":     pd.TotalTasks,
			"completionPct":  pct,
			"lastUpdated":    pd.LastUpdated.UTC().Format(time.RFC3339),
			"healthStatus":   pd.HealthStatus,
			"features":       features,
			"warnings":       warnings,
			"meta":           makeMeta(pd.LastUpdated),
		})
	}
}

func handleListFeatures(s *scanner.Scanner) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		pd, err := s.ScanProject(id)
		if err != nil {
			_ = c.Error(err)
			return
		}

		var features []featureSummary
		var lastUpdated time.Time
		for _, f := range pd.Features {
			features = append(features, featureSummary{
				Slug:           f.Slug,
				Status:         f.Status,
				CompletedTasks: f.CompletedTasks,
				TotalTasks:     f.TotalTasks,
				LastUpdated:    f.LastUpdated.UTC().Format(time.RFC3339),
			})
			if f.LastUpdated.After(lastUpdated) {
				lastUpdated = f.LastUpdated
			}
		}
		if features == nil {
			features = []featureSummary{}
		}

		c.JSON(http.StatusOK, gin.H{
			"features": features,
			"meta":     makeMeta(lastUpdated),
		})
	}
}

func handleGetFeature(s *scanner.Scanner) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		slug := c.Param("slug")

		pd, err := s.ScanProject(id)
		if err != nil {
			_ = c.Error(err)
			return
		}

		feature, err := findFeature(pd, slug)
		if err != nil {
			_ = c.Error(err)
			return
		}

		// Build tasks map with summary info
		tasks := make(map[string]interface{})
		for key, t := range feature.Tasks {
			tasks[key] = gin.H{
				"id":     t.ID,
				"key":    t.Key,
				"title":  t.Title,
				"status": t.Status,
				"phase":  t.Phase,
			}
		}

		// Build phases
		var phases []interface{}
		for _, p := range feature.Phases {
			phases = append(phases, gin.H{
				"number":   p.Number,
				"label":    p.Label,
				"taskKeys": p.TaskKeys,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"slug":           feature.Slug,
			"status":         feature.Status,
			"prdPath":        feature.PRDPath,
			"designPath":     feature.DesignPath,
			"completedTasks": feature.CompletedTasks,
			"totalTasks":     feature.TotalTasks,
			"lastUpdated":    feature.LastUpdated.UTC().Format(time.RFC3339),
			"phases":         phases,
			"tasks":          tasks,
			"meta":           makeMeta(feature.LastUpdated),
		})
	}
}

func handleListTasks(s *scanner.Scanner) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		slug := c.Param("slug")

		pd, err := s.ScanProject(id)
		if err != nil {
			_ = c.Error(err)
			return
		}

		feature, err := findFeature(pd, slug)
		if err != nil {
			_ = c.Error(err)
			return
		}

		var tasks []taskSummary
		for _, t := range feature.Tasks {
			tasks = append(tasks, taskSummary{
				ID:            t.ID,
				Key:           t.Key,
				Title:         t.Title,
				Priority:      t.Priority,
				Status:        t.Status,
				Scope:         t.Scope,
				EstimatedTime: t.EstimatedTime,
				Dependencies:  t.Dependencies,
				Breaking:      t.Breaking,
				Phase:         t.Phase,
				File:          t.File,
				Record:        t.Record,
			})
		}
		if tasks == nil {
			tasks = []taskSummary{}
		}

		c.JSON(http.StatusOK, gin.H{
			"tasks": tasks,
			"meta":  makeMeta(feature.LastUpdated),
		})
	}
}

func handleGetTask(s *scanner.Scanner) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		slug := c.Param("slug")
		taskID := c.Param("taskId")

		pd, err := s.ScanProject(id)
		if err != nil {
			_ = c.Error(err)
			return
		}

		feature, err := findFeature(pd, slug)
		if err != nil {
			_ = c.Error(err)
			return
		}

		task, err := findTask(feature, taskID)
		if err != nil {
			_ = c.Error(err)
			return
		}

		// Parse acceptance criteria from task file
		var acceptanceCriteria []string
		if task.File != "" {
			tf, parseErr := scanner.ParseTaskFile(task.File)
			if parseErr == nil && tf != nil {
				acceptanceCriteria = tf.AcceptanceCriteria
			}
		}
		if acceptanceCriteria == nil {
			acceptanceCriteria = []string{}
		}

		// Parse execution record
		var executionRecord interface{}
		if task.Record != "" {
			rc, parseErr := scanner.ParseRecordFile(task.Record)
			if parseErr == nil && rc != nil {
				executionRecord = gin.H{
					"summary":     rc.Summary,
					"files":       rc.Files,
					"decisions":   rc.Decisions,
					"testResults": rc.TestResults,
					"raw":         rc.Raw,
				}
			}
		}

		// Determine lastUpdated for meta
		var metaTime time.Time
		if pd.LastUpdated.After(feature.LastUpdated) {
			metaTime = pd.LastUpdated
		} else {
			metaTime = feature.LastUpdated
		}

		c.JSON(http.StatusOK, gin.H{
			"id":                 task.ID,
			"key":                task.Key,
			"title":              task.Title,
			"priority":           task.Priority,
			"status":             task.Status,
			"scope":              task.Scope,
			"estimatedTime":      task.EstimatedTime,
			"dependencies":       task.Dependencies,
			"breaking":           task.Breaking,
			"phase":              task.Phase,
			"file":               task.File,
			"record":             task.Record,
			"acceptanceCriteria": acceptanceCriteria,
			"executionRecord":    executionRecord,
			"meta":               makeMeta(metaTime),
		})
	}
}

func handleGetDependencies(s *scanner.Scanner) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		slug := c.Param("slug")

		pd, err := s.ScanProject(id)
		if err != nil {
			_ = c.Error(err)
			return
		}

		feature, err := findFeature(pd, slug)
		if err != nil {
			_ = c.Error(err)
			return
		}

		graph := buildDependencyGraph(pd, feature)

		c.JSON(http.StatusOK, gin.H{
			"nodes": graph.Nodes,
			"edges": graph.Edges,
			"meta":  makeMeta(feature.LastUpdated),
		})
	}
}

// ---- Helper functions ----

// computePct returns the completion percentage, 0 if total is 0.
func computePct(completed, total int) float64 {
	if total == 0 {
		return 0
	}
	return (float64(completed) / float64(total)) * 100
}

// findFeature finds a feature by slug within a project.
func findFeature(pd *model.ProjectData, slug string) (*model.FeatureData, error) {
	for i := range pd.Features {
		if pd.Features[i].Slug == slug {
			return &pd.Features[i], nil
		}
	}
	return nil, model.ErrFeatureNotFound(slug)
}

// findTask finds a task by ID within a feature.
func findTask(feature *model.FeatureData, taskID string) (*model.Task, error) {
	for _, t := range feature.Tasks {
		if t.ID == taskID {
			return &t, nil
		}
	}
	return nil, model.ErrTaskNotFound(taskID)
}

// buildDependencyGraph constructs a dependency graph for a feature,
// resolving wildcard dependencies and marking cross-feature edges.
func buildDependencyGraph(pd *model.ProjectData, feature *model.FeatureData) *model.DependencyGraph {
	graph := &model.DependencyGraph{
		Nodes: []model.GraphNode{},
		Edges: []model.GraphEdge{},
	}

	// Build node list from the feature's tasks
	for _, t := range feature.Tasks {
		graph.Nodes = append(graph.Nodes, model.GraphNode{
			ID:      t.ID,
			Key:     t.Key,
			Title:   t.Title,
			Status:  t.Status,
			Phase:   t.Phase,
			Feature: feature.Slug,
		})
	}

	// Build a lookup of task ID -> feature slug across all features in the project
	taskFeatureMap := make(map[string]string)
	for _, f := range pd.Features {
		for _, t := range f.Tasks {
			taskFeatureMap[t.ID] = f.Slug
		}
	}

	// Build edges from dependencies (wildcards already expanded by scanner)
	for _, t := range feature.Tasks {
		for _, dep := range t.Dependencies {
			// Determine if this is a cross-feature edge
			depFeature, exists := taskFeatureMap[dep]
			crossFeature := exists && depFeature != feature.Slug

			graph.Edges = append(graph.Edges, model.GraphEdge{
				Source:       t.ID,
				Target:       dep,
				CrossFeature: crossFeature,
			})
		}
	}

	return graph
}
