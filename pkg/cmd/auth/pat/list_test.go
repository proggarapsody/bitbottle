package pat_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/auth/pat"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

const serverConfig = "git.example.com:\n  user: alice\n  oauth_token: tok\n"

func TestPATList_Success(t *testing.T) {
	t.Parallel()
	expiry := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	fake := &testhelpers.FakeClient{
		T: t,
		ListPATsFn: func(userSlug string, limit int) ([]backend.PAT, error) {
			assert.Equal(t, "alice", userSlug)
			return []backend.PAT{
				{ID: "1", Name: "CI Token", Permissions: []string{"REPO_READ"}, CreatedDate: time.Now(), ExpiryDate: &expiry},
				{ID: "2", Name: "Dev Token", Permissions: []string{"REPO_WRITE"}, CreatedDate: time.Now()},
			}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig})
	factorytest.UseBackend(f, fake)

	cmd := pat.NewCmdPAT(f)
	cmd.SetArgs([]string{"list", "--hostname", "git.example.com"})
	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, out.String(), "CI Token")
	assert.Contains(t, out.String(), "Dev Token")
}

func TestPATList_JSONOutput(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListPATsFn: func(userSlug string, limit int) ([]backend.PAT, error) {
			return []backend.PAT{
				{ID: "1", Name: "CI Token", Permissions: []string{"REPO_READ"}, CreatedDate: time.Now()},
			}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig})
	factorytest.UseBackend(f, fake)

	cmd := pat.NewCmdPAT(f)
	cmd.SetArgs([]string{"list", "--hostname", "git.example.com", "--json"})
	err := cmd.Execute()
	require.NoError(t, err)
	assert.True(t, strings.Contains(out.String(), `"id"`))
	assert.True(t, strings.Contains(out.String(), `"name"`))
}

func TestPATList_CloudUnsupported(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		ListPATsFn: func(userSlug string, limit int) ([]backend.PAT, error) {
			return nil, &backend.DomainError{
				Kind:    backend.ErrUnsupportedOnHost,
				Code:    backend.CodeHostUnsupported,
				Message: "personal access token management is not supported on bitbucket.org (Bitbucket Server / Data Center only)",
			}
		},
	}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: "bitbucket.org:\n  user: alice\n  oauth_token: tok\n"})
	factorytest.UseBackend(f, fake)

	cmd := pat.NewCmdPAT(f)
	cmd.SetArgs([]string{"list", "--hostname", "bitbucket.org"})
	err := cmd.Execute()
	require.Error(t, err)
}
