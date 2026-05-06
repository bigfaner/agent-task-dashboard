package handler

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/panda/agent-task-center/internal/model"
)

var (
	// slugRegex validates feature slugs: alphanumeric + hyphens, at least 2 chars,
	// must start and end with alphanumeric.
	slugRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9]$`)

	// taskIDRegex validates task IDs: alphanumeric, dots, hyphens, underscores.
	// No path traversal sequences allowed.
	taskIDRegex = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
)

// ValidateSlug returns Gin middleware that validates the :slug path parameter.
// Valid slugs match: ^[a-zA-Z0-9][a-zA-Z0-9-]*[a-zA-Z0-9]$ (alphanumeric + hyphens, 2+ chars).
// Invalid slugs receive HTTP 400 with {"error": "invalid_slug", "message": "..."}.
func ValidateSlug() gin.HandlerFunc {
	return func(c *gin.Context) {
		slug := c.Param("slug")
		if slug == "" || !slugRegex.MatchString(slug) {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error":   "invalid_slug",
				"message": "Feature slug contains invalid characters. Only alphanumeric characters and hyphens are allowed (minimum 2 characters).",
			})
			return
		}
		c.Next()
	}
}

// ValidateTaskID returns Gin middleware that validates the :taskId path parameter.
// Valid task IDs match: ^[a-zA-Z0-9._-]+$ and must not contain path traversal (..).
// Invalid task IDs receive HTTP 400 with {"error": "invalid_task_id", "message": "..."}.
func ValidateTaskID() gin.HandlerFunc {
	return func(c *gin.Context) {
		taskID := c.Param("taskId")
		if taskID == "" || !taskIDRegex.MatchString(taskID) || containsPathTraversal(taskID) {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error":   "invalid_task_id",
				"message": "Task ID contains invalid characters or path traversal sequences.",
			})
			return
		}
		c.Next()
	}
}

// ErrorHandler returns Gin middleware that catches errors added to the context
// and maps them to appropriate HTTP status codes with consistent JSON error responses.
// Error response format: {"error": "<code>", "message": "<description>"}
//
// Mapping:
//   - ErrProjectNotFound, ErrFeatureNotFound, ErrTaskNotFound -> 404 not_found
//   - ErrInvalidSlug -> 400 invalid_slug
//   - ErrInvalidTaskID -> 400 invalid_task_id
//   - ErrFSRead, ErrParseIndex, unknown errors -> 500 internal_error
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Only process if there are errors in the context
		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err

		statusCode, errorCode, message := mapError(err)

		c.JSON(statusCode, gin.H{
			"error":   errorCode,
			"message": message,
		})
	}
}

// mapError maps a model error to HTTP status code, error code, and message.
func mapError(err error) (int, string, string) {
	switch err.(type) {
	case model.ErrProjectNotFound:
		return http.StatusNotFound, "not_found", err.Error()
	case model.ErrFeatureNotFound:
		return http.StatusNotFound, "not_found", err.Error()
	case model.ErrTaskNotFound:
		return http.StatusNotFound, "not_found", err.Error()
	case model.ErrInvalidSlug:
		return http.StatusBadRequest, "invalid_slug", err.Error()
	case model.ErrInvalidTaskID:
		return http.StatusBadRequest, "invalid_task_id", err.Error()
	case model.ErrFSRead:
		return http.StatusInternalServerError, "internal_error", "An internal error occurred"
	case model.ErrParseIndex:
		return http.StatusInternalServerError, "internal_error", "An internal error occurred"
	case model.ErrConfigInvalid:
		return http.StatusInternalServerError, "internal_error", "An internal error occurred"
	default:
		return http.StatusInternalServerError, "internal_error", "An internal error occurred"
	}
}

// containsPathTraversal checks if a string contains ".." as a path traversal sequence.
func containsPathTraversal(s string) bool {
	return strings.Contains(s, "..")
}
