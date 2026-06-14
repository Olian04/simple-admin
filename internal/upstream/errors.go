package upstream

import (
	"errors"
	"fmt"
)

// ErrNotFound is returned when the upstream responds 404.
var ErrNotFound = errors.New("resource not found")

// ErrUnauthorized is returned when the upstream rejects the credential (401/403).
var ErrUnauthorized = errors.New("upstream rejected credential")

// UpstreamError wraps a non-2xx upstream response.
type UpstreamError struct {
	Status int
	Body   string
}

func (e *UpstreamError) Error() string {
	return fmt.Sprintf("upstream status %d: %s", e.Status, e.Body)
}
