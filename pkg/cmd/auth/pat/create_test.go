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

func TestPATCreate_Success(t *testing.T) {
	t.Parallel()
	var gotInput backend.CreatePATInput
	fake := &testhelpers.FakeClient{
		T: t,
		CreatePATFn: func(userSlug string, in backend.CreatePATInput) (backend.PATWithSecret, error) {
			assert.Equal(t, "alice", userSlug)
			gotInput = in
			return backend.PATWithSecret{
				PAT:   backend.PAT{ID: "42", Name: in.Name, CreatedDate: time.Now()},
				Token: "BBDC-supersecret",
			}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig})
	factorytest.UseBackend(f, fake)

	cmd := pat.NewCmdPAT(f)
	cmd.SetArgs([]string{"create", "--hostname", "git.example.com", "--name", "CI Token", "--scopes", "repo:read,repo:write"})
	err := cmd.Execute()
	require.NoError(t, err)
	assert.Equal(t, "CI Token", gotInput.Name)
	assert.Contains(t, gotInput.Permissions, "REPO_READ")
	assert.Contains(t, gotInput.Permissions, "REPO_WRITE")
	assert.Nil(t, gotInput.ExpiryDays)
	assert.Contains(t, out.String(), "BBDC-supersecret")
	assert.Contains(t, out.String(), "Store this token now")
}

func TestPATCreate_WithExpiry(t *testing.T) {
	t.Parallel()
	var gotInput backend.CreatePATInput
	fake := &testhelpers.FakeClient{
		T: t,
		CreatePATFn: func(userSlug string, in backend.CreatePATInput) (backend.PATWithSecret, error) {
			gotInput = in
			return backend.PATWithSecret{
				PAT:   backend.PAT{ID: "7", Name: in.Name, CreatedDate: time.Now()},
				Token: "BBDC-xyz",
			}, nil
		},
	}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig})
	factorytest.UseBackend(f, fake)

	cmd := pat.NewCmdPAT(f)
	cmd.SetArgs([]string{"create", "--hostname", "git.example.com", "--name", "Expiring Token", "--scopes", "repo:read", "--expires-in", "30"})
	err := cmd.Execute()
	require.NoError(t, err)
	require.NotNil(t, gotInput.ExpiryDays)
	assert.Equal(t, 30, *gotInput.ExpiryDays)
}

func TestPATCreate_MissingName(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig})
	factorytest.UseBackend(f, fake)

	cmd := pat.NewCmdPAT(f)
	cmd.SetArgs([]string{"create", "--hostname", "git.example.com", "--scopes", "repo:read"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--name is required")
}

func TestPATCreate_InvalidScope(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig})
	factorytest.UseBackend(f, fake)

	cmd := pat.NewCmdPAT(f)
	cmd.SetArgs([]string{"create", "--hostname", "git.example.com", "--name", "T", "--scopes", "invalid:scope"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown scope")
}

func TestPATCreate_CanonicalScope(t *testing.T) {
	t.Parallel()
	var gotInput backend.CreatePATInput
	fake := &testhelpers.FakeClient{
		T: t,
		CreatePATFn: func(userSlug string, in backend.CreatePATInput) (backend.PATWithSecret, error) {
			gotInput = in
			return backend.PATWithSecret{
				PAT:   backend.PAT{ID: "1", Name: in.Name, CreatedDate: time.Now()},
				Token: "BBDC-tok",
			}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig})
	factorytest.UseBackend(f, fake)

	cmd := pat.NewCmdPAT(f)
	// Pass canonical name directly.
	cmd.SetArgs([]string{"create", "--hostname", "git.example.com", "--name", "T", "--scopes", "REPO_READ"})
	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, gotInput.Permissions, "REPO_READ")
	// Token must appear on stdout, not stderr.
	assert.Contains(t, out.String(), "BBDC-tok")
}

func TestPATCreate_TokenNotOnStderr(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		CreatePATFn: func(userSlug string, in backend.CreatePATInput) (backend.PATWithSecret, error) {
			return backend.PATWithSecret{
				PAT:   backend.PAT{ID: "1", Name: in.Name, CreatedDate: time.Now()},
				Token: "BBDC-secret",
			}, nil
		},
	}
	f, _, errOut := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig})
	factorytest.UseBackend(f, fake)

	cmd := pat.NewCmdPAT(f)
	cmd.SetArgs([]string{"create", "--hostname", "git.example.com", "--name", "T", "--scopes", "repo:read"})
	err := cmd.Execute()
	require.NoError(t, err)
	// Token MUST NOT appear on stderr.
	assert.NotContains(t, errOut.String(), "BBDC-secret")
}

func TestPATCreate_ScopeAliases(t *testing.T) {
	tests := []struct {
		alias  string
		expect string
	}{
		{"repo:read", "REPO_READ"},
		{"repo:write", "REPO_WRITE"},
		{"pr:read", "PR_READ"},
		{"pr:write", "PR_WRITE"},
		{"project:read", "PROJECT_READ"},
		{"project:write", "PROJECT_WRITE"},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.alias, func(t *testing.T) {
			t.Parallel()
			var gotPerms []string
			fake := &testhelpers.FakeClient{
				T: t,
				CreatePATFn: func(userSlug string, in backend.CreatePATInput) (backend.PATWithSecret, error) {
					gotPerms = in.Permissions
					return backend.PATWithSecret{
						PAT:   backend.PAT{ID: "1", Name: "T", CreatedDate: time.Now()},
						Token: "t",
					}, nil
				},
			}
			f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig})
			factorytest.UseBackend(f, fake)
			cmd := pat.NewCmdPAT(f)
			cmd.SetArgs([]string{"create", "--hostname", "git.example.com", "--name", "T", "--scopes", tc.alias})
			require.NoError(t, cmd.Execute())
			assert.Contains(t, gotPerms, tc.expect)
			assert.False(t, strings.Contains(strings.Join(gotPerms, ","), tc.alias),
				"alias form must be converted to canonical form")
		})
	}
}
