package backend

import (
	"errors"
	"fmt"
)

// ErrNotFound is returned when the backend responds 404.
var ErrNotFound = errors.New("resource not found")

// ErrUnauthorized is returned when the backend rejects the credential (401/403).
var ErrUnauthorized = errors.New("backend rejected credential")

// Error wraps a non-2xx backend response.
type Error struct {
	Status int
	Body   string
}

func (e *Error) Error() string {
	return fmt.Sprintf("backend status %d: %s", e.Status, e.Body)
}
