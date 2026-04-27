package bbrepo_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/internal/bbrepo"
)

// TestInferFromRemote_SSHNoGitSuffix verifies an SSH URL without .git suffix
// is parsed correctly.
func TestInferFromRemote_SSHNoGitSuffix(t *testing.T) {
	t.Parallel()
	ref, err := bbrepo.InferFromRemote("git@git.example.com:PROJ/repo")
	require.NoError(t, err)
	assert.Equal(t, "git.example.com", ref.Host)
	assert.Equal(t, "PROJ", ref.Project)
	assert.Equal(t, "repo", ref.Slug)
}

// TestInferFromRemote_HTTPSPort verifies an HTTPS URL with a custom port.
func TestInferFromRemote_HTTPSPort(t *testing.T) {
	t.Parallel()
	ref, err := bbrepo.InferFromRemote("https://git.example.com:8443/scm/PROJ/repo.git")
	require.NoError(t, err)
	assert.Equal(t, "git.example.com:8443", ref.Host)
	assert.Equal(t, "PROJ", ref.Project)
	assert.Equal(t, "repo", ref.Slug)
}

// TestInferFromRemote_SSHPort verifies an ssh:// URL with a custom port strips the port from Host.
func TestInferFromRemote_SSHPort(t *testing.T) {
	t.Parallel()
	ref, err := bbrepo.InferFromRemote("ssh://git@git.example.com:7999/PROJ/repo.git")
	require.NoError(t, err)
	assert.Equal(t, "git.example.com", ref.Host)
	assert.Equal(t, "PROJ", ref.Project)
	assert.Equal(t, "repo", ref.Slug)
}

// TestInferFromRemote_HTTPSMissingScmPrefix verifies that an HTTPS URL without
// /scm/ prefix returns an error.
func TestInferFromRemote_HTTPSMissingScmPrefix(t *testing.T) {
	t.Parallel()
	_, err := bbrepo.InferFromRemote("https://git.example.com/PROJ/repo.git")
	require.Error(t, err)
}

// TestParse_SlugWithDots verifies that a slug containing dots is accepted.
func TestParse_SlugWithDots(t *testing.T) {
	t.Parallel()
	ref, err := bbrepo.Parse("PROJ/my.service")
	require.NoError(t, err)
	assert.Equal(t, "PROJ", ref.Project)
	assert.Equal(t, "my.service", ref.Slug)
}

// TestParse_ProjectWithNumbers verifies a project key with numbers is accepted.
func TestParse_ProjectWithNumbers(t *testing.T) {
	t.Parallel()
	ref, err := bbrepo.Parse("PROJ123/repo-name")
	require.NoError(t, err)
	assert.Equal(t, "PROJ123", ref.Project)
	assert.Equal(t, "repo-name", ref.Slug)
}

// TestInferFromRemote_UnsupportedScheme verifies that an unsupported URL scheme
// returns an error.
func TestInferFromRemote_UnsupportedScheme(t *testing.T) {
	t.Parallel()
	_, err := bbrepo.InferFromRemote("ftp://git.example.com/scm/PROJ/repo.git")
	require.Error(t, err)
}
