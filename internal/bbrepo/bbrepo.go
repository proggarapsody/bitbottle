package bbrepo

import (
	"fmt"
	"net/url"
	"strings"
)

// RepoRef identifies a repository on a specific Bitbucket host.
type RepoRef struct {
	Project string
	Slug    string
	Host    string
}

// Parse parses "PROJECT/REPO" into a RepoRef.
func Parse(ref string) (RepoRef, error) {
	if ref == "" {
		return RepoRef{}, fmt.Errorf("empty repo ref")
	}
	parts := strings.Split(ref, "/")
	if len(parts) != 2 {
		return RepoRef{}, fmt.Errorf("invalid repo ref %q: expected PROJECT/slug", ref)
	}
	return RepoRef{Project: parts[0], Slug: parts[1]}, nil
}

// InferFromRemote parses a git remote URL into a RepoRef.
// Supported formats:
//
//	ssh://git@HOST/PROJECT/REPO.git
//	git@HOST:PROJECT/REPO.git
//	https://HOST/scm/PROJECT/REPO.git
func InferFromRemote(remoteURL string) (RepoRef, error) {
	if remoteURL == "" {
		return RepoRef{}, fmt.Errorf("empty URL")
	}

	// SSH colon format: git@host:PROJ/repo.git
	if strings.HasPrefix(remoteURL, "git@") {
		s := strings.TrimPrefix(remoteURL, "git@")
		colonIdx := strings.Index(s, ":")
		if colonIdx < 0 {
			return RepoRef{}, fmt.Errorf("cannot parse SSH URL: %q", remoteURL)
		}
		host := s[:colonIdx]
		rest := strings.TrimSuffix(s[colonIdx+1:], ".git")
		parts := strings.Split(rest, "/")
		if len(parts) != 2 {
			return RepoRef{}, fmt.Errorf("unexpected path in SSH URL: %q", remoteURL)
		}
		return RepoRef{Host: host, Project: parts[0], Slug: parts[1]}, nil
	}

	u, err := url.Parse(remoteURL)
	if err != nil {
		return RepoRef{}, fmt.Errorf("parsing URL: %w", err)
	}

	path := strings.TrimSuffix(u.Path, ".git")
	path = strings.TrimPrefix(path, "/")

	switch u.Scheme {
	case "ssh":
		// ssh://git@host/PROJ/repo.git
		parts := strings.Split(path, "/")
		if len(parts) != 2 {
			return RepoRef{}, fmt.Errorf("unexpected path in ssh:// URL: %q", remoteURL)
		}
		return RepoRef{Host: u.Hostname(), Project: parts[0], Slug: parts[1]}, nil

	case "https", "http":
		// https://host/scm/PROJ/repo.git
		parts := strings.Split(path, "/")
		if len(parts) != 3 || parts[0] != "scm" {
			return RepoRef{}, fmt.Errorf("not a Bitbucket HTTPS clone URL: %q", remoteURL)
		}
		return RepoRef{Host: u.Host, Project: parts[1], Slug: parts[2]}, nil
	}

	return RepoRef{}, fmt.Errorf("unsupported URL scheme %q in: %q", u.Scheme, remoteURL)
}
