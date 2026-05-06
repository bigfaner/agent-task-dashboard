package handler

import (
	"encoding/json"
	"html/template"
	"net/http"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/panda/agent-task-center/internal/model"
	"github.com/panda/agent-task-center/internal/scanner"
	"github.com/panda/agent-task-center/web"
)

// parsePageTemplates creates and returns the parsed HTML templates from the embedded filesystem.
func parsePageTemplates() *template.Template {
	tmpl := template.New("").Funcs(template.FuncMap{
		"toJson": func(v interface{}) (template.JS, error) {
			b, err := json.Marshal(v)
			if err != nil {
				return "", err
			}
			return template.JS(b), nil
		},
	})
	return template.Must(tmpl.ParseFS(web.Assets, "templates/*.html"))
}

// RegisterPages registers HTML page routes on the Gin engine.
func RegisterPages(r *gin.Engine, s *scanner.Scanner) {
	// Load HTML templates from embedded filesystem
	tmpl := parsePageTemplates()
	r.SetHTMLTemplate(tmpl)

	r.GET("/", handleLanding(s))
	r.GET("/projects/:id", handleProject(s))
}

// handleLanding renders the landing page with project cards.
func handleLanding(s *scanner.Scanner) gin.HandlerFunc {
	return func(c *gin.Context) {
		all, err := s.ScanAll()
		if err != nil {
			renderErrorPage(c, http.StatusInternalServerError, "Failed to load projects")
			return
		}

		// Build project summaries for template
		type projectCard struct {
			ID             string   `json:"id"`
			Name           string   `json:"name"`
			FeatureCount   int      `json:"featureCount"`
			CompletedTasks int      `json:"completedTasks"`
			TotalTasks     int      `json:"totalTasks"`
			CompletionPct  float64  `json:"completionPct"`
			HealthStatus   string   `json:"healthStatus"`
			LastUpdated    string   `json:"lastUpdated"`
			Warnings       []string `json:"warnings,omitempty"`
		}

		var projects []projectCard
		for _, pd := range all {
			pc := projectCard{
				ID:             pd.ID,
				Name:           pd.Name,
				FeatureCount:   len(pd.Features),
				CompletedTasks: pd.CompletedTasks,
				TotalTasks:     pd.TotalTasks,
				CompletionPct:  computePct(pd.CompletedTasks, pd.TotalTasks),
				HealthStatus:   pd.HealthStatus,
				Warnings:       pd.Warnings,
			}
			if !pd.LastUpdated.IsZero() {
				pc.LastUpdated = pd.LastUpdated.Format(time.RFC3339)
			}
			projects = append(projects, pc)
		}

		if projects == nil {
			projects = []projectCard{}
		}

		c.HTML(http.StatusOK, "landing.html", gin.H{
			"Projects": projects,
			"Title":    "Task Dashboard",
		})
	}
}

// handleProject renders the swimlane view for a specific project.
func handleProject(s *scanner.Scanner) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		pd, err := s.ScanProject(id)
		if err != nil {
			renderErrorPage(c, http.StatusNotFound, "Project not found")
			return
		}

		// Build feature data for JS consumption
		type featureInfo struct {
			Slug           string  `json:"slug"`
			Status         string  `json:"status"`
			CompletedTasks int     `json:"completedTasks"`
			TotalTasks     int     `json:"totalTasks"`
			CompletionPct  float64 `json:"completionPct"`
			HasBlocked     bool    `json:"hasBlocked"`
		}

		var features []featureInfo
		for _, f := range pd.Features {
			features = append(features, featureInfo{
				Slug:           f.Slug,
				Status:         f.Status,
				CompletedTasks: f.CompletedTasks,
				TotalTasks:     f.TotalTasks,
				CompletionPct:  f.CompletionPct,
				HasBlocked:     f.HasBlockedTasks,
			})
		}

		if features == nil {
			features = []featureInfo{}
		}

		// Derive activity events from tasks across all features
		activityEvents := deriveActivityEvents(pd.Features)
		blockedCount := countBlockedTasks(pd.Features)

		c.HTML(http.StatusOK, "swimlane.html", gin.H{
			"ProjectID":      pd.ID,
			"ProjectName":    pd.Name,
			"Features":       features,
			"ActivityEvents": activityEvents,
			"BlockedCount":   blockedCount,
			"Title":          pd.Name + " - Swimlane View",
		})
	}
}

// statusToEventType maps task statuses to activity event types.
func statusToEventType(status string) string {
	switch status {
	case "in_progress":
		return "claimed"
	case "completed":
		return "completed"
	case "blocked":
		return "blocked"
	case "skipped":
		return "skipped"
	default:
		return ""
	}
}

// deriveActivityEvents creates a sorted list of activity events from all features' tasks.
// Events are derived from non-pending tasks. The timestamp comes from the feature's
// LastUpdated (index.json mtime). Events are sorted by timestamp descending,
// then by feature slug alphabetically for equal timestamps. Maximum 50 events.
func deriveActivityEvents(features []model.FeatureData) []model.ActivityEvent {
	events := make([]model.ActivityEvent, 0)

	for _, f := range features {
		for _, t := range f.Tasks {
			eventType := statusToEventType(t.Status)
			if eventType == "" {
				continue // skip pending tasks
			}
			events = append(events, model.ActivityEvent{
				Timestamp: f.LastUpdated,
				TaskID:    t.ID,
				TaskTitle: t.Title,
				Feature:   f.Slug,
				EventType: eventType,
			})
		}
	}

	// Sort: by timestamp descending, then by feature slug ascending
	sort.SliceStable(events, func(i, j int) bool {
		if !events[i].Timestamp.Equal(events[j].Timestamp) {
			return events[i].Timestamp.After(events[j].Timestamp)
		}
		return events[i].Feature < events[j].Feature
	})

	// Limit to 50
	if len(events) > 50 {
		events = events[:50]
	}

	return events
}

// countBlockedTasks counts the total number of blocked tasks across all features.
func countBlockedTasks(features []model.FeatureData) int {
	count := 0
	for _, f := range features {
		for _, t := range f.Tasks {
			if t.Status == "blocked" {
				count++
			}
		}
	}
	return count
}

// renderErrorPage renders an error page with the given status code and message.
func renderErrorPage(c *gin.Context, statusCode int, message string) {
	c.HTML(statusCode, "error.html", gin.H{
		"Title":        "Error",
		"StatusCode":   statusCode,
		"ErrorMessage": message,
	})
}
