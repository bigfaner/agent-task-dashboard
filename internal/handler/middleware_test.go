package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// errorResponse is the expected JSON structure for error responses.
type errorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// makeGetRequest creates a GET request with the given path, safely handling
// characters that would break URL parsing (spaces, slashes, etc.).
func makeGetRequest(target string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/placeholder", nil)
	req.URL = &url.URL{Path: target}
	return req
}

// --- Slug Validation Tests ---

func TestValidateSlug_ValidSlugs(t *testing.T) {
	tests := []struct {
		name string
		slug string
	}{
		{"simple alphanumeric", "myfeature"},
		{"with hyphens", "my-feature"},
		{"alphanumeric mixed", "feature123"},
		{"all numbers", "123"},
		{"single char repeated", "aa"},
		{"complex slug", "my-cool-feature-v2"},
		{"uppercase allowed via regex", "MyFeature"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.Use(ValidateSlug())
			r.GET("/features/:slug", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"slug": c.Param("slug")})
			})

			w := httptest.NewRecorder()
			req := makeGetRequest("/features/" + tt.slug)
			r.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("expected 200, got %d; slug=%q; body=%s", w.Code, tt.slug, w.Body.String())
			}
		})
	}
}

func TestValidateSlug_InvalidSlugs(t *testing.T) {
	tests := []struct {
		name    string
		slug    string
		errCode string
	}{
		{"path traversal", "../etc", "invalid_slug"},
		{"single character", "a", "invalid_slug"},
		{"empty slug", "", "invalid_slug"},
		{"spaces", "my feature", "invalid_slug"},
		{"special chars", "feature!@#", "invalid_slug"},
		{"dot dot slash in middle", "feat../ure", "invalid_slug"},
		{"leading hyphen", "-feature", "invalid_slug"},
		{"trailing hyphen", "feature-", "invalid_slug"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.Use(ValidateSlug())
			r.GET("/features/:slug", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"slug": c.Param("slug")})
			})

			w := httptest.NewRecorder()
			req := makeGetRequest("/features/" + tt.slug)
			r.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("expected 400, got %d; slug=%q; body=%s", w.Code, tt.slug, w.Body.String())
			}

			var resp errorResponse
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("failed to parse error response: %v", err)
			}
			if resp.Error != tt.errCode {
				t.Errorf("expected error code %q, got %q", tt.errCode, resp.Error)
			}
		})
	}
}

func TestValidateSlug_SetsSlugInContext(t *testing.T) {
	r := gin.New()
	r.Use(ValidateSlug())
	r.GET("/features/:slug", func(c *gin.Context) {
		slug := c.Param("slug")
		c.JSON(http.StatusOK, gin.H{"slug": slug})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/features/my-feature", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["slug"] != "my-feature" {
		t.Errorf("expected slug 'my-feature', got %v", resp["slug"])
	}
}

// --- Task ID Validation Tests ---

func TestValidateTaskID_ValidIDs(t *testing.T) {
	tests := []struct {
		name   string
		taskID string
	}{
		{"simple numeric", "1.1"},
		{"multi dot", "1.2.3"},
		{"with letters", "T-test-1"},
		{"alphanumeric dot", "1.1-interfaces"},
		{"underscore", "task_1"},
		{"complex", "1.2.3-alpha"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.Use(ValidateTaskID())
			r.GET("/tasks/:taskId", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"taskId": c.Param("taskId")})
			})

			w := httptest.NewRecorder()
			req := makeGetRequest("/tasks/" + tt.taskID)
			r.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("expected 200, got %d; taskId=%q; body=%s", w.Code, tt.taskID, w.Body.String())
			}
		})
	}
}

func TestValidateTaskID_InvalidIDs(t *testing.T) {
	tests := []struct {
		name   string
		taskID string
	}{
		{"path traversal dot dot", ".."},
		{"path traversal prefix", "../1.1"},
		// Note: "1.1/.." is not tested here because Gin's router treats "/" as a
		// path separator, so "/tasks/1.1/.." results in a 404 from the router itself
		// (it doesn't match the single-param route). This is correct behavior.
		{"path traversal middle", "1../1"},
		{"spaces", "1 1"},
		{"empty", ""},
		{"only dots", "..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.Use(ValidateTaskID())
			r.GET("/tasks/:taskId", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"taskId": c.Param("taskId")})
			})

			w := httptest.NewRecorder()
			req := makeGetRequest("/tasks/" + tt.taskID)
			r.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("expected 400, got %d; taskId=%q; body=%s", w.Code, tt.taskID, w.Body.String())
			}

			var resp errorResponse
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("failed to parse error response: %v", err)
			}
			if resp.Error != "invalid_task_id" {
				t.Errorf("expected error code 'invalid_task_id', got %q", resp.Error)
			}
		})
	}
}

func TestValidateTaskID_PathTraversalRejected(t *testing.T) {
	r := gin.New()
	r.Use(ValidateTaskID())
	r.GET("/tasks/:taskId", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"taskId": c.Param("taskId")})
	})

	// Test URL-encoded path traversal
	w := httptest.NewRecorder()
	req := makeGetRequest("/tasks/../../etc/passwd")
	r.ServeHTTP(w, req)

	// Gin router will interpret ../../ as path segments, but the middleware
	// should still catch the ".." in the taskId parameter.
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for path traversal, got %d; body=%s", w.Code, w.Body.String())
	}
}

// --- Error Handling Middleware Tests ---

func TestErrorHandler_ScannerErrors(t *testing.T) {
	tests := []struct {
		name         string
		errorType    string
		expectedCode int
		expectedErr  string
	}{
		{"project not found -> 404", "ErrProjectNotFound", http.StatusNotFound, "not_found"},
		{"feature not found -> 404", "ErrFeatureNotFound", http.StatusNotFound, "not_found"},
		{"task not found -> 404", "ErrTaskNotFound", http.StatusNotFound, "not_found"},
		{"invalid slug -> 400", "ErrInvalidSlug", http.StatusBadRequest, "invalid_slug"},
		{"invalid task ID -> 400", "ErrInvalidTaskID", http.StatusBadRequest, "invalid_task_id"},
		{"fs read error -> 500", "ErrFSRead", http.StatusInternalServerError, "internal_error"},
		{"parse index error -> 500", "ErrParseIndex", http.StatusInternalServerError, "internal_error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.Use(ErrorHandler())

			r.GET("/test", func(c *gin.Context) {
				var err error
				switch tt.errorType {
				case "ErrProjectNotFound":
					err = createErrProjectNotFound("test")
				case "ErrFeatureNotFound":
					err = createErrFeatureNotFound("test")
				case "ErrTaskNotFound":
					err = createErrTaskNotFound("test")
				case "ErrInvalidSlug":
					err = createErrInvalidSlug("test")
				case "ErrInvalidTaskID":
					err = createErrInvalidTaskID("test")
				case "ErrFSRead":
					err = createErrFSRead("test")
				case "ErrParseIndex":
					err = createErrParseIndex("test")
				}
				_ = c.Error(err)
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("expected %d, got %d; body=%s", tt.expectedCode, w.Code, w.Body.String())
			}

			var resp errorResponse
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("failed to parse error response: %v", err)
			}
			if resp.Error != tt.expectedErr {
				t.Errorf("expected error code %q, got %q", tt.expectedErr, resp.Error)
			}
		})
	}
}

func TestErrorHandler_UnknownError(t *testing.T) {
	r := gin.New()
	r.Use(ErrorHandler())

	r.GET("/test", func(c *gin.Context) {
		_ = c.Error(errUnknown{msg: "some random error"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	var resp errorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse error response: %v", err)
	}
	if resp.Error != "internal_error" {
		t.Errorf("expected error code 'internal_error', got %q", resp.Error)
	}
}

func TestErrorHandler_NoError(t *testing.T) {
	r := gin.New()
	r.Use(ErrorHandler())

	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

// --- Error Response Format Tests ---

func TestErrorResponse_ConsistentFormat(t *testing.T) {
	r := gin.New()
	r.Use(ErrorHandler())

	r.GET("/test", func(c *gin.Context) {
		_ = c.Error(createErrProjectNotFound("nonexistent"))
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	// Verify Content-Type is JSON
	ct := w.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Errorf("expected Content-Type to contain 'application/json', got %q", ct)
	}

	// Verify exact structure: only "error" and "message" keys
	var raw map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &raw); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	keys := make(map[string]bool)
	for k := range raw {
		keys[k] = true
	}
	if !keys["error"] || !keys["message"] {
		t.Errorf("expected keys 'error' and 'message', got %v", keys)
	}
	if len(keys) != 2 {
		t.Errorf("expected exactly 2 keys, got %d: %v", len(keys), keys)
	}
}

// --- Combined Middleware Tests ---

func TestCombinedMiddleware_ValidationErrorAndErrorHandler(t *testing.T) {
	r := gin.New()
	r.Use(ErrorHandler())
	r.Use(ValidateSlug())
	r.GET("/features/:slug", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"slug": c.Param("slug")})
	})

	w := httptest.NewRecorder()
	req := makeGetRequest("/features/../hack")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d; body=%s", w.Code, w.Body.String())
	}

	var resp errorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse error response: %v", err)
	}
	if resp.Error != "invalid_slug" {
		t.Errorf("expected error code 'invalid_slug', got %q", resp.Error)
	}
}

// --- Regex Compilation Tests ---

func TestSlugRegexPatterns(t *testing.T) {
	valid := []string{"ab", "my-feature", "feature123", "My-Feature-V2", "a0", "0a"}
	for _, s := range valid {
		if !slugRegex.MatchString(s) {
			t.Errorf("slugRegex should match %q", s)
		}
	}

	invalid := []string{"a", "-feature", "feature-", "my feature", "../hack", "", "a!b"}
	for _, s := range invalid {
		if slugRegex.MatchString(s) {
			t.Errorf("slugRegex should NOT match %q", s)
		}
	}
}

func TestTaskIDRegexPatterns(t *testing.T) {
	valid := []string{"1.1", "1.2.3", "T-test-1", "task_1", "1.1-interfaces"}
	for _, s := range valid {
		if !taskIDRegex.MatchString(s) {
			t.Errorf("taskIDRegex should match %q", s)
		}
	}

	invalid := []string{"", "1 1", "1/1", "task!1"}
	for _, s := range invalid {
		if taskIDRegex.MatchString(s) {
			t.Errorf("taskIDRegex should NOT match %q", s)
		}
	}
}
