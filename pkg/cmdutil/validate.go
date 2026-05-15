package cmdutil

import "fmt"

// ValidatePositiveLimit returns an error if limit is less than 1.
// Use for --limit flags where 0 is not a documented "no-cap" value.
// For commands that explicitly document 0 as "no limit", skip this check.
func ValidatePositiveLimit(limit int) error {
	if limit < 1 {
		return fmt.Errorf("--limit must be at least 1, got %d", limit)
	}
	return nil
}
