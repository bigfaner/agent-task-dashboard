package scanner

import (
	"bufio"
	"os"
	"strings"

	"github.com/panda/agent-task-center/internal/model"
)

// ParseTaskFile reads a task .md file and extracts acceptance criteria,
// scope, and description. Uses simple line-by-line parsing without a
// markdown AST library.
func ParseTaskFile(filePath string) (*model.TaskFileContent, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	content := string(data)
	result := &model.TaskFileContent{
		AcceptanceCriteria: []string{},
	}

	// Extract scope from frontmatter
	result.Scope = extractScopeFromFrontmatter(content)

	// Parse sections by headings
	sections := parseSections(content)

	// Extract acceptance criteria from "## Acceptance Criteria" section
	if acSection, ok := sections["acceptance criteria"]; ok {
		result.AcceptanceCriteria = extractAcceptanceCriteria(acSection)
	}

	// Extract description from "## Description" section
	if descSection, ok := sections["description"]; ok {
		result.Description = strings.TrimSpace(descSection)
	}

	return result, nil
}

// ParseRecordFile reads a record .md file and extracts structured sections.
// Sections are identified by ## and ### headings. Returns empty RecordContent with
// Raw field populated if no recognized headings are found.
func ParseRecordFile(filePath string) (*model.RecordContent, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	content := string(data)
	result := &model.RecordContent{
		Raw:   content,
		Files: []string{},
	}

	sections := parseSections(content)

	// Map section headings to RecordContent fields
	if s, ok := sections["summary"]; ok {
		result.Summary = strings.TrimSpace(s)
	}

	// Files are extracted from "### Files Created" and "### Files Modified" sub-sections
	result.Files = extractFiles(content)

	if s, ok := sections["key decisions"]; ok {
		result.Decisions = strings.TrimSpace(s)
	}

	if s, ok := sections["test results"]; ok {
		result.TestResults = strings.TrimSpace(s)
	}

	return result, nil
}

// parseSections splits markdown content into sections keyed by ## and ### heading names.
// The heading prefix is stripped and the key is lowercased for matching.
// When a ## heading is encountered, any ### sub-section is collected under its
// own key as well as continuing to collect content under the parent ## section.
func parseSections(content string) map[string]string {
	sections := make(map[string]string)

	scanner := bufio.NewScanner(strings.NewReader(content))
	var currentHeading string
	var currentContent strings.Builder

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "## ") {
			// Save previous section
			if currentHeading != "" {
				sections[currentHeading] = currentContent.String()
			}
			// Start new section
			currentHeading = strings.TrimSpace(strings.TrimPrefix(line, "## "))
			currentHeading = strings.ToLower(currentHeading)
			currentContent.Reset()
		} else if strings.HasPrefix(line, "### ") {
			// Sub-heading: save as its own section key AND continue under parent
			subKey := strings.TrimSpace(strings.TrimPrefix(line, "### "))
			subKey = strings.ToLower(subKey)
			// Save accumulated content to parent before starting sub-section
			if currentHeading != "" {
				sections[currentHeading] = currentContent.String()
			}
			// Replace current heading with sub-heading
			currentHeading = subKey
			currentContent.Reset()
		} else if currentHeading != "" {
			currentContent.WriteString(line)
			currentContent.WriteString("\n")
		}
	}

	// Save last section
	if currentHeading != "" {
		sections[currentHeading] = currentContent.String()
	}

	return sections
}

// extractAcceptanceCriteria finds lines starting with "- [ ]" or "- [x]"
// and returns the text after the checkbox marker.
func extractAcceptanceCriteria(section string) []string {
	var criteria []string

	scanner := bufio.NewScanner(strings.NewReader(section))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		text := ""
		if strings.HasPrefix(line, "- [ ] ") {
			text = strings.TrimPrefix(line, "- [ ] ")
		} else if strings.HasPrefix(line, "- [x] ") {
			text = strings.TrimPrefix(line, "- [x] ")
		} else if strings.HasPrefix(line, "- [ ]") && len(line) == 5 {
			text = ""
		} else if strings.HasPrefix(line, "- [x]") && len(line) == 5 {
			text = ""
		}

		if text != "" {
			criteria = append(criteria, strings.TrimSpace(text))
		}
	}

	return criteria
}

// extractScopeFromFrontmatter extracts the "scope" field from YAML frontmatter.
// Frontmatter is delimited by "---" lines at the start of the file.
func extractScopeFromFrontmatter(content string) string {
	if !strings.HasPrefix(content, "---") {
		return ""
	}

	// Find the closing ---
	end := strings.Index(content[3:], "---")
	if end < 0 {
		return ""
	}

	fm := content[3 : end+3]

	scanner := bufio.NewScanner(strings.NewReader(fm))
	for scanner.Scan() {
		l := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(l, "scope:") {
			val := strings.TrimSpace(strings.TrimPrefix(l, "scope:"))
			return unquote(val)
		}
		if strings.HasPrefix(l, "scope :") {
			val := strings.TrimSpace(strings.TrimPrefix(l, "scope :"))
			return unquote(val)
		}
	}

	return ""
}

// unquote strips surrounding double or single quotes from a string.
func unquote(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

// extractFiles extracts file paths from "### Files Created" and "### Files Modified"
// sub-sections. Each path line starts with "- ".
func extractFiles(content string) []string {
	var files []string
	inFilesSection := false

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()

		// Detect ### Files Created or ### Files Modified sub-sections
		if strings.HasPrefix(line, "### Files ") {
			inFilesSection = true
			continue
		}

		// Exit file sub-section on any other heading
		if strings.HasPrefix(line, "## ") || strings.HasPrefix(line, "# ") {
			inFilesSection = false
			continue
		}

		if strings.HasPrefix(line, "### ") && !strings.HasPrefix(line, "### Files ") {
			inFilesSection = false
			continue
		}

		if inFilesSection {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "- ") {
				path := strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
				// Skip empty lines and non-path text
				if path != "" && !strings.Contains(path, "No files") {
					files = append(files, path)
				}
			}
		}
	}

	return files
}
