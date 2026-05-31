package view_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/user/view"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

const serverConfig = "git.example.com:\n  oauth_token: tok\n"
const cloudConfig = "bitbucket.org:\n  oauth_token: tok\n  user: alice\n  git_protocol: https\n"

func TestNewCmdView_Flags(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig})
	cmd := view.NewCmdView(f, nil)
	format.RegisterOutputFlags(cmd)
	assert.NotNil(t, cmd.Flag("hostname"))
	assert.NotNil(t, cmd.Flag("json"))
	assert.NotNil(t, cmd.Flag("jq"))
}

func TestNewCmdView_RejectsArgs(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig})
	cmd := view.NewCmdView(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"unexpected"})
	require.Error(t, cmd.Execute())
}

func TestView_PrintsSlugAndName(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetCurrentUserFn: func() (backend.User, error) {
			return backend.User{Slug: "jdoe", DisplayName: "Jane Doe"}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig})
	factorytest.UseBackend(f, fake)

	cmd := view.NewCmdView(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "jdoe")
	assert.Contains(t, got, "Jane Doe")
}

func TestView_JSONOutput(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetCurrentUserFn: func() (backend.User, error) {
			return backend.User{Slug: "jdoe", DisplayName: "Jane Doe"}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig})
	factorytest.UseBackend(f, fake)

	cmd := view.NewCmdView(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"--json"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "jdoe")
	assert.Contains(t, got, "Jane Doe")
	assert.Contains(t, got, "slug")
	assert.Contains(t, got, "name")
}

// TestView_JSON_IncludesCloudIdentifiers is the MCP-15 regression: the
// machine-readable Cloud identifiers must appear in --json output (and be
// reachable by --jq for the nested links.html.href).
func TestView_JSON_IncludesCloudIdentifiers(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetCurrentUserFn: func() (backend.User, error) {
			return backend.User{
				Slug:        "proggarapsody",
				DisplayName: "Aleksey K",
				AccountID:   "557058:abc-123",
				UUID:        "{1234-uuid}",
				CreatedOn:   "2018-01-01T00:00:00Z",
				HTMLURL:     "https://bitbucket.org/proggarapsody/",
			}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := view.NewCmdView(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"--json"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "account_id")
	assert.Contains(t, got, "557058:abc-123")
	assert.Contains(t, got, "uuid")
	assert.Contains(t, got, "{1234-uuid}")
	assert.Contains(t, got, "created_on")
	assert.Contains(t, got, "2018-01-01T00:00:00Z")
	assert.Contains(t, got, "links")
	assert.Contains(t, got, "https://bitbucket.org/proggarapsody/")
}

// TestView_JQ_NestedLinks proves --jq can reach the nested links.html.href.
func TestView_JQ_NestedLinks(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetCurrentUserFn: func() (backend.User, error) {
			return backend.User{Slug: "p", HTMLURL: "https://bitbucket.org/p/"}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := view.NewCmdView(f, nil)
	format.RegisterOutputFlags(cmd)
	// user view emits a single-element array; reach into it with .[].
	cmd.SetArgs([]string{"--json", "--jq", ".[].links.html.href"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "https://bitbucket.org/p/")
}

// TestView_JSON_ServerOmitsCloudFields verifies the identifiers are omitted
// (not emitted as empty/null) on a backend that lacks them, e.g. Server/DC.
func TestView_JSON_ServerOmitsCloudFields(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetCurrentUserFn: func() (backend.User, error) {
			return backend.User{Slug: "jdoe", DisplayName: "Jane Doe"}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig})
	factorytest.UseBackend(f, fake)

	cmd := view.NewCmdView(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"--json"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.NotContains(t, got, "account_id")
	assert.NotContains(t, got, "uuid")
	assert.NotContains(t, got, "links")
}

func TestView_GetCurrentUserError(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetCurrentUserFn: func() (backend.User, error) {
			return backend.User{}, errors.New("500 internal server error")
		},
	}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: serverConfig})
	factorytest.UseBackend(f, fake)

	cmd := view.NewCmdView(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestView_CloudBackend_Succeeds(t *testing.T) {
	// Ensure UserGetter is satisfied by the FakeClient when using a Cloud backend.
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetCurrentUserFn: func() (backend.User, error) {
			return backend.User{Slug: "alice", DisplayName: "Alice Cloud"}, nil
		},
	}
	f, out, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)

	cmd := view.NewCmdView(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "alice")
	assert.Contains(t, got, "Alice Cloud")
}
