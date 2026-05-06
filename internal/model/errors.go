package model

// ErrProjectNotFound returns an error indicating the project was not found.
type ErrProjectNotFound string

func (e ErrProjectNotFound) Error() string { return "project not found: " + string(e) }

// ErrFeatureNotFound returns an error indicating the feature was not found.
type ErrFeatureNotFound string

func (e ErrFeatureNotFound) Error() string { return "feature not found: " + string(e) }

// ErrTaskNotFound returns an error indicating the task was not found.
type ErrTaskNotFound string

func (e ErrTaskNotFound) Error() string { return "task not found: " + string(e) }

// ErrConfigInvalid returns an error indicating invalid configuration.
type ErrConfigInvalid string

func (e ErrConfigInvalid) Error() string { return "config invalid: " + string(e) }

// ErrFSRead returns an error indicating a filesystem read failure.
type ErrFSRead string

func (e ErrFSRead) Error() string { return "filesystem read error: " + string(e) }

// ErrParseIndex returns an error indicating an index.json parse failure.
type ErrParseIndex string

func (e ErrParseIndex) Error() string { return "index parse error: " + string(e) }

// ErrInvalidSlug returns an error indicating an invalid feature slug.
type ErrInvalidSlug string

func (e ErrInvalidSlug) Error() string { return "invalid slug: " + string(e) }

// ErrInvalidTaskID returns an error indicating an invalid task ID.
type ErrInvalidTaskID string

func (e ErrInvalidTaskID) Error() string { return "invalid task id: " + string(e) }
