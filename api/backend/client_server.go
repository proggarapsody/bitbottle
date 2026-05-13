package backend

import "fmt"

// VersionedServer is implemented by backends that report their server version.
// Currently Server/DC only — Cloud has no equivalent endpoint.
type VersionedServer interface {
	GetServerVersion() ServerVersion
}

// ServerVersion holds a parsed Bitbucket Server version number.
type ServerVersion struct {
	Major int
	Minor int
	Patch int
	Raw   string // original string, e.g. "8.5.0"
}

// AtLeast reports whether the version is ≥ major.minor.0.
func (v ServerVersion) AtLeast(major, minor int) bool {
	if v.Major != major {
		return v.Major > major
	}
	return v.Minor >= minor
}

// FeatureServerVersion names the server-version capability for typed-error reporting.
const FeatureServerVersion Feature = "server-version"

// AsVersionedServer returns the VersionedServer accessor for c if the backend
// supports it, or ErrUnsupportedOnHost if not.
func AsVersionedServer(c Client, host string) (VersionedServer, error) {
	if vs, ok := c.(VersionedServer); ok {
		return vs, nil
	}
	return nil, &DomainError{
		Kind:    ErrUnsupportedOnHost,
		Code:    CodeHostUnsupported,
		Host:    host,
		Feature: string(FeatureServerVersion),
		Message: fmt.Sprintf("server version is not supported on %s (Bitbucket Server / Data Center only)", host),
	}
}
