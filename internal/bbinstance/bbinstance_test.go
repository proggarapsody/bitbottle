package bbinstance_test

import (
	"testing"

	"github.com/aleksey/bitbottle/internal/bbinstance"
	"github.com/stretchr/testify/assert"
)

func TestSSHURL(t *testing.T) {
	t.Parallel()
	got := bbinstance.SSHURL("git.example.com", "PROJ", "repo")
	assert.Equal(t, "git@git.example.com:PROJ/repo.git", got)
}

func TestHTTPSURL(t *testing.T) {
	t.Parallel()
	got := bbinstance.HTTPSURL("git.example.com", "PROJ", "repo")
	assert.Equal(t, "https://git.example.com/scm/PROJ/repo.git", got)
}

func TestWebRepoURL(t *testing.T) {
	t.Parallel()
	got := bbinstance.WebRepoURL("git.example.com", "PROJ", "repo")
	assert.Equal(t, "https://git.example.com/projects/PROJ/repos/repo/browse", got)
}

func TestWebPRURL(t *testing.T) {
	t.Parallel()
	got := bbinstance.WebPRURL("git.example.com", "PROJ", "repo", 42)
	assert.Equal(t, "https://git.example.com/projects/PROJ/repos/repo/pull-requests/42", got)
}

func TestRESTBase(t *testing.T) {
	t.Parallel()
	got := bbinstance.RESTBase("git.example.com")
	assert.Equal(t, "https://git.example.com/rest/api/1.0", got)
}

func TestSupportsDraftPR_Above(t *testing.T) {
	t.Parallel()
	assert.True(t, bbinstance.SupportsDraftPR("8.9.1"))
}

func TestSupportsDraftPR_Below(t *testing.T) {
	t.Parallel()
	assert.False(t, bbinstance.SupportsDraftPR("7.0.0"))
}

func TestSupportsDraftPR_Exact(t *testing.T) {
	t.Parallel()
	assert.True(t, bbinstance.SupportsDraftPR("7.17.0"))
}
