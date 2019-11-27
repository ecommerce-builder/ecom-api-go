package postgres

import (
	"errors"
)

var (
	// ErrPermission error
	ErrPermission = errors.New("permission denied")

	// ErrNotExist error
	ErrNotExist = errors.New("resource does not exist")

	// ErrInvalidText error
	ErrInvalidText = errors.New("invalid text representation")
)

// APIKeyError container
type APIKeyError struct {
	Op       string
	Resource string
	APIKey   string
	Err      error
}

func (e *APIKeyError) Error() string {
	return e.Op + " on " + e.Resource + " " + e.APIKey + ": " + e.Err.Error()
}

// ResourceError records an error.
type ResourceError struct {
	Op       string
	Resource string
	UUID     string
	Err      error
}

func (e *ResourceError) Error() string {
	return e.Op + " on " + e.Resource + " " + e.UUID + ": " + e.Err.Error()
}

type permission interface {
	Permission() bool
}

type notExists interface {
	NotExists() bool
}

type invalidText interface {
	InvalidText() bool
}

// Permission error
func (e *ResourceError) Permission() bool {
	return e.Err == ErrPermission
}

// NotExists error
func (e *ResourceError) NotExists() bool {
	return e.Err == ErrNotExist
}

// InvalidText error
func (e *ResourceError) InvalidText() bool {
	return e.Err == ErrInvalidText
}

// IsPermission error
func IsPermission(err error) bool {
	ip, ok := err.(permission)
	return ok && ip.Permission()
}

// IsNotExist error
func (m *PgModel) IsNotExist(err error) bool {
	ne, ok := err.(notExists)
	return ok && ne.NotExists()
}

// IsInvalidText error
func IsInvalidText(err error) bool {
	ie, ok := err.(invalidText)
	return ok && ie.InvalidText()
}
