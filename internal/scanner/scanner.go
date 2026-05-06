package scanner

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/panda/agent-task-center/internal/config"
	"github.com/panda/agent-task-center/internal/model"
)

// staleThreshold defines how long without activity before a project is considered stale.
const staleThreshold = 7 * 24 * time.Hour

// Scanner reads project directories, parses index.json files, and builds
// the in-memory data model. Uses fs.FS for testability.
type Scanner struct {
	config *config.Config
	fs     fs.FS
	cache  map[string]*model.ProjectData
}

// NewScanner creates a scanner for the given config. Production uses os.DirFS.
func NewScanner(cfg *config.Config) *Scanner {
	return &Scanner{
		config: cfg,
		fs:     dirFSWrapper{FS: os.DirFS("/")},
		cache:  make(map[string]*model.ProjectData),
	}
}

// dirFSWrapper wraps an fs.FS to mark it as an os.DirFS production filesystem.
type dirFSWrapper struct {
	fs.FS
}

func (dirFSWrapper) isDirFS() {}

// ScanAll reads all configured projects. Returns map keyed by project ID (lowercased name).
// Skips projects with invalid paths (adds warning).
func (s *Scanner) ScanAll() (map[string]*model.ProjectData, error) {
	result := make(map[string]*model.ProjectData)

	for _, pc := range s.config.Projects {
		id := strings.ToLower(pc.Name)
		pd := s.scanProjectConfig(id, pc)
		s.cache[id] = pd
		result[id] = pd
	}

	return result, nil
}

// ScanProject reads a single project by ID.
// Returns ErrProjectNotFound if ID not in config.
func (s *Scanner) ScanProject(id string) (*model.ProjectData, error) {
	// Check cache first
	if pd, ok := s.cache[id]; ok {
		return pd, nil
	}

	// Find project in config
	for _, pc := range s.config.Projects {
		projID := strings.ToLower(pc.Name)
		if projID == id {
			pd := s.scanProjectConfig(id, pc)
			s.cache[id] = pd
			return pd, nil
		}
	}

	return nil, model.ErrProjectNotFound(id)
}

// Invalidate clears the in-memory cache, forcing fresh filesystem reads.
func (s *Scanner) Invalidate() {
	s.cache = make(map[string]*model.ProjectData)
}

// dirFS is a marker interface for os.DirFS-based filesystems.
// Used to distinguish production filesystems from test filesystems.
type dirFS interface {
	fs.FS
	isDirFS()
}

// projectFS returns an fs.FS rooted at the given project path.
// For the production os.DirFS it creates a new os.DirFS at the project path.
// For test fstest.MapFS it uses fs.Sub to navigate to the project prefix.
func (s *Scanner) projectFS(path string) (fs.FS, error) {
	if _, ok := s.fs.(dirFS); ok {
		return os.DirFS(path), nil
	}
	// For test filesystems (fstest.MapFS), use fs.Sub with the path prefix.
	subPath := filepath.ToSlash(strings.TrimPrefix(path, "/"))
	return fs.Sub(s.fs, subPath)
}

// SortFeatures sorts features in-place by PRD rule:
//  1. Features with any blocked tasks come first
//  2. Within each group, sort by completion % ascending (most incomplete first)
//  3. Ties broken by slug alphabetically
func SortFeatures(features []model.FeatureData) {
	sort.SliceStable(features, func(i, j int) bool {
		fi, fj := features[i], features[j]

		// Blocked features first
		if fi.HasBlockedTasks != fj.HasBlockedTasks {
			return fi.HasBlockedTasks
		}

		// Then by completion % ascending
		if fi.CompletionPct != fj.CompletionPct {
			return fi.CompletionPct < fj.CompletionPct
		}

		// Then alphabetically by slug
		return fi.Slug < fj.Slug
	})
}

// scanProjectConfig scans a single project from its config entry.
func (s *Scanner) scanProjectConfig(id string, pc config.ProjectConfig) *model.ProjectData {
	pd := &model.ProjectData{
		ID:       id,
		Name:     pc.Name,
		Path:     pc.Path,
		Features: []model.FeatureData{},
	}

	// Get a filesystem rooted at the project directory.
	// For production (os.DirFS): create a new DirFS at the project path.
	// For tests (fstest.MapFS): use fs.Sub to navigate to the project prefix.
	projFS, err := s.projectFS(pc.Path)
	if err != nil {
		pd.Warnings = append(pd.Warnings, fmt.Sprintf("cannot access project path: %v", err))
		pd.HealthStatus = computeHealthStatus(pd.Features, time.Time{})
		return pd
	}

	featuresDir := "docs/features"

	entries, err := fs.ReadDir(projFS, featuresDir)
	if err != nil {
		// Path doesn't exist or can't be read - add warning
		pd.Warnings = append(pd.Warnings, fmt.Sprintf("cannot read features directory: %v", err))
		pd.HealthStatus = computeHealthStatus(pd.Features, time.Time{})
		return pd
	}

	var lastUpdated time.Time

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		featureSlug := entry.Name()
		featureData := s.scanFeature(projFS, featuresDir, featureSlug)
		if featureData != nil {
			pd.Features = append(pd.Features, *featureData)
			if featureData.LastUpdated.After(lastUpdated) {
				lastUpdated = featureData.LastUpdated
			}
		}
	}

	// Sort features by PRD rule
	SortFeatures(pd.Features)

	// Compute aggregated stats
	pd.LastUpdated = lastUpdated
	for i := range pd.Features {
		pd.TotalTasks += pd.Features[i].TotalTasks
		pd.CompletedTasks += pd.Features[i].CompletedTasks
	}

	pd.HealthStatus = computeHealthStatus(pd.Features, pd.LastUpdated)

	return pd
}

// scanFeature reads and parses a single feature's index.json using the given FS.
func (s *Scanner) scanFeature(projFS fs.FS, featuresPath, slug string) *model.FeatureData {
	indexPath := filepath.ToSlash(filepath.Join(featuresPath, slug, "tasks", "index.json"))

	data, err := fs.ReadFile(projFS, indexPath)
	if err != nil {
		// No index.json for this feature - skip silently
		return nil
	}

	var idx indexFile
	if err := json.Unmarshal(data, &idx); err != nil {
		log.Printf("warning: malformed index.json in feature %q: %v", slug, err)
		return nil
	}

	// Validate required fields
	if idx.Feature == "" {
		log.Printf("warning: missing 'feature' field in %s", indexPath)
		return nil
	}

	return buildFeatureData(idx, indexPath, projFS)
}

// indexFile represents the structure of an index.json file.
type indexFile struct {
	Feature string               `json:"feature"`
	PRD     string               `json:"prd"`
	Design  string               `json:"design"`
	Created string               `json:"created"`
	Status  string               `json:"status"`
	Tasks   map[string]indexTask `json:"tasks"`
}

// indexTask represents a single task entry in index.json.
type indexTask struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	Priority      string   `json:"priority"`
	EstimatedTime string   `json:"estimatedTime,omitempty"`
	Dependencies  []string `json:"dependencies"`
	Status        string   `json:"status"`
	File          string   `json:"file"`
	Record        string   `json:"record"`
	Breaking      bool     `json:"breaking"`
	Scope         string   `json:"scope"`
}

// buildFeatureData converts a parsed indexFile into a FeatureData model.
func buildFeatureData(idx indexFile, indexPath string, fsys fs.FS) *model.FeatureData {
	tasks := make(map[string]model.Task)
	phaseMap := make(map[int][]string)

	for key, it := range idx.Tasks {
		task := model.Task{
			ID:            it.ID,
			Key:           key,
			Title:         it.Title,
			Priority:      it.Priority,
			Status:        it.Status,
			Scope:         it.Scope,
			EstimatedTime: it.EstimatedTime,
			Dependencies:  it.Dependencies,
			Breaking:      it.Breaking,
			File:          it.File,
			Record:        it.Record,
			Phase:         model.DerivePhase(it.ID),
		}

		tasks[key] = task
		phase := task.Phase
		if phase > 0 {
			phaseMap[phase] = append(phaseMap[phase], key)
		}
	}

	// Expand wildcard dependencies
	expandWildcards(tasks)

	// Build phases
	phases := buildPhases(phaseMap)

	// Compute stats
	totalTasks := len(tasks)
	completedTasks := 0
	hasBlocked := false
	for _, task := range tasks {
		if task.Status == "completed" {
			completedTasks++
		}
		if task.Status == "blocked" {
			hasBlocked = true
		}
	}

	var completionPct float64
	if totalTasks > 0 {
		completionPct = (float64(completedTasks) / float64(totalTasks)) * 100
	}

	// Get file mtime for LastUpdated
	var lastUpdated time.Time
	if info, err := fs.Stat(fsys, indexPath); err == nil {
		lastUpdated = info.ModTime()
	}

	return &model.FeatureData{
		Slug:            idx.Feature,
		Status:          idx.Status,
		PRDPath:         idx.PRD,
		DesignPath:      idx.Design,
		Tasks:           tasks,
		Phases:          phases,
		LastUpdated:     lastUpdated,
		TotalTasks:      totalTasks,
		CompletedTasks:  completedTasks,
		HasBlockedTasks: hasBlocked,
		CompletionPct:   completionPct,
	}
}

// expandWildcards replaces wildcard dependencies ("1.x", "2.x") with actual task IDs
// from the same feature that match the phase number.
func expandWildcards(tasks map[string]model.Task) {
	// Build phase -> task IDs map
	phaseTasks := make(map[int][]string)
	for _, task := range tasks {
		if task.Phase > 0 {
			phaseTasks[task.Phase] = append(phaseTasks[task.Phase], task.ID)
		}
	}

	for key, task := range tasks {
		if len(task.Dependencies) == 0 {
			continue
		}

		expanded := make([]string, 0, len(task.Dependencies))
		changed := false

		for _, dep := range task.Dependencies {
			if strings.HasSuffix(dep, ".x") {
				// Wildcard: "1.x" -> all tasks in phase 1
				phaseStr := dep[:len(dep)-2]
				phaseNum := 0
				if _, err := fmt.Sscanf(phaseStr, "%d", &phaseNum); err == nil && phaseNum > 0 {
					if ids, ok := phaseTasks[phaseNum]; ok {
						expanded = append(expanded, ids...)
						changed = true
					}
				}
			} else {
				expanded = append(expanded, dep)
			}
		}

		if changed {
			task.Dependencies = expanded
			tasks[key] = task
		}
	}
}

// buildPhases converts a phase->taskKeys map into sorted PhaseInfo slice.
func buildPhases(phaseMap map[int][]string) []model.PhaseInfo {
	phases := make([]model.PhaseInfo, 0, len(phaseMap))
	for num, keys := range phaseMap {
		sort.Strings(keys)
		phases = append(phases, model.PhaseInfo{
			Number:   num,
			Label:    fmt.Sprintf("Phase %d", num),
			TaskKeys: keys,
		})
	}
	sort.Slice(phases, func(i, j int) bool {
		return phases[i].Number < phases[j].Number
	})
	return phases
}

// computeHealthStatus determines the project health based on task statuses and last update time.
// "active" - has in_progress tasks or was updated recently
// "complete" - all tasks completed
// "stale" - no updates in 7+ days
func computeHealthStatus(features []model.FeatureData, lastUpdated time.Time) string {
	if len(features) == 0 {
		if lastUpdated.IsZero() {
			return "stale"
		}
	}

	// Check if any task is in progress
	hasInProgress := false
	allCompleted := true

	for _, feat := range features {
		for _, task := range feat.Tasks {
			if task.Status == "in_progress" {
				hasInProgress = true
			}
			if task.Status != "completed" && task.Status != "skipped" {
				allCompleted = false
			}
		}
	}

	if allCompleted && len(features) > 0 {
		return "complete"
	}

	if hasInProgress {
		return "active"
	}

	// Check for recent activity
	if !lastUpdated.IsZero() && time.Since(lastUpdated) > staleThreshold {
		return "stale"
	}

	return "active"
}
