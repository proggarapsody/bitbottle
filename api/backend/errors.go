package backend

import "fmt"

// HTTPError represents an error HTTP response from the Bitbucket API.
type HTTPError struct {
	StatusCode int
	Message    string
	RequestURL string
}

// Error implements the error interface.
func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
}
