package bbrepo_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/internal/bbrepo"
)

func TestParse_Valid(t *testing.T) {
	t.Parallel()
	ref, err := bbrepo.Parse("PROJ/repo")
	require.NoError(t, err)
	assert.Equal(t, "PROJ", ref.Project)
	assert.Equal(t, "repo", ref.Slug)
}

func TestParse_Empty(t *testing.T) {
	t.Parallel()
	_, err := bbrepo.Parse("")
	require.Error(t, err)
}

func TestParse_MissingSlash(t *testing.T) {
	t.Parallel()
	_, err := bbrepo.Parse("noslash")
	require.Error(t, err)
}

func TestParse_TooManyParts(t *testing.T) {
	t.Parallel()
	_, err := bbrepo.Parse("a/b/c")
	require.Error(t, err)
}

func TestInferFromRemote_SSHColon(t *testing.T) {
	t.Parallel()
	ref, err := bbrepo.InferFromRemote("git@git.example.com:PROJ/repo.git")
	require.NoError(t, err)
	assert.Equal(t, "git.example.com", ref.Host)
	assert.Equal(t, "PROJ", ref.Project)
	assert.Equal(t, "repo", ref.Slug)
}

func TestInferFromRemote_SSHScheme(t *testing.T) {
	t.Parallel()
	ref, err := bbrepo.InferFromRemote("ssh://git@git.example.com/PROJ/repo.git")
	require.NoError(t, err)
	assert.Equal(t, "git.example.com", ref.Host)
	assert.Equal(t, "PROJ", ref.Project)
	assert.Equal(t, "repo", ref.Slug)
}

func TestInferFromRemote_HTTPS(t *testing.T) {
	t.Parallel()
	ref, err := bbrepo.InferFromRemote("https://git.example.com/scm/PROJ/repo.git")
	require.NoError(t, err)
	assert.Equal(t, "git.example.com", ref.Host)
	assert.Equal(t, "PROJ", ref.Project)
	assert.Equal(t, "repo", ref.Slug)
}

func TestInferFromRemote_NotBitbucket(t *testing.T) {
	t.Parallel()
	_, err := bbrepo.InferFromRemote("https://github.com/user/repo.git")
	require.Error(t, err)
}

func TestInferFromRemote_Empty(t *testing.T) {
	t.Parallel()
	_, err := bbrepo.InferFromRemote("")
	require.Error(t, err)
}

// Bug: ssh://git@host:7999/PROJ/repo.git — u.Host includes the port suffix
// (:7999), causing API calls to use "host:7999" instead of "host".
// Bitbucket Server's default SSH port (7999) must be stripped from the hostname.
func TestInferFromRemote_SSHSchemeWithPort_StripPort(t *testing.T) {
	t.Parallel()
	ref, err := bbrepo.InferFromRemote("ssh://git@git.example.com:7999/PROJ/repo.git")
	require.NoError(t, err)
	assert.Equal(t, "git.example.com", ref.Host, "port must be stripped from ssh:// hostname")
	assert.Equal(t, "PROJ", ref.Project)
	assert.Equal(t, "repo", ref.Slug)
}
