package server

import (
	"strconv"
	"strings"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// wireApplicationProperties is the /rest/api/1.0/application-properties response.
type wireApplicationProperties struct {
	Version string `json:"version"`
}

// GetServerVersion fetches and caches the Bitbucket Server version.
// The result is cached for the lifetime of the Client; subsequent calls
// return the cached value without an HTTP request.
// On error, the zero ServerVersion is returned (AtLeast always returns false).
func (c *Client) GetServerVersion() backend.ServerVersion {
	c.versionOnce.Do(func() {
		var props wireApplicationProperties
		if err := c.http.GetJSON("/application-properties", &props); err != nil {
			return
		}
		c.cachedVersion = parseServerVersion(props.Version)
	})
	return c.cachedVersion
}

// parseServerVersion parses a "major.minor.patch" string.
// Returns zero ServerVersion on any parse failure — be lenient.
func parseServerVersion(s string) backend.ServerVersion {
	parts := strings.SplitN(s, ".", 3)
	if len(parts) < 2 {
		return backend.ServerVersion{Raw: s}
	}
	major, err1 := strconv.Atoi(parts[0])
	minor, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return backend.ServerVersion{Raw: s}
	}
	patch := 0
	if len(parts) == 3 {
		patch, _ = strconv.Atoi(parts[2]) // ignore error, default 0
	}
	return backend.ServerVersion{Major: major, Minor: minor, Patch: patch, Raw: s}
}
