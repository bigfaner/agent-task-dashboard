package handler

import "github.com/panda/agent-task-center/internal/model"

// These helpers create model error types for use in tests.
// They bridge the handler test package to the model error types.

func createErrProjectNotFound(id string) error { return model.ErrProjectNotFound(id) }
func createErrFeatureNotFound(id string) error { return model.ErrFeatureNotFound(id) }
func createErrTaskNotFound(id string) error    { return model.ErrTaskNotFound(id) }
func createErrInvalidSlug(slug string) error   { return model.ErrInvalidSlug(slug) }
func createErrInvalidTaskID(id string) error   { return model.ErrInvalidTaskID(id) }
func createErrFSRead(msg string) error         { return model.ErrFSRead(msg) }
func createErrParseIndex(msg string) error     { return model.ErrParseIndex(msg) }

// errUnknown is a non-model error used for testing unknown error handling.
type errUnknown struct {
	msg string
}

func (e errUnknown) Error() string { return e.msg }
