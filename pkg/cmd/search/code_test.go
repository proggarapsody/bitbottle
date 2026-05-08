package search_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/run"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/search"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

const cloudConfig = "bitbucket.org:\n  oauth_token: tok\n  user: alice\n  git_protocol: https\n"

// runner is the standard remote-resolving runner used by cloud-targeted
// commands; search code defaults to BaseRepo for the workspace inference
// path so it shares the same scaffolding as `issue list` tests.
func runner() *testhelpers.FakeRunner {
	return testhelpers.NewFakeRunner(testhelpers.RunResponse{
		Stdout: "https://bitbucket.org/acme/repo.git\n",
	})
}

func newFactory(t *testing.T, fake backend.Client) (*factory.Factory, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	f, out, errOut := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)
	r := runner()
	f.GitRunner = func() run.Runner { return r }
	return f, out, errOut
}

// noSearchFake wraps backend.Client without satisfying CodeSearcher. The
// embedding (not the concrete struct) prevents method promotion, so the
// type-assertion in AsCodeSearcher fails as it would for a Server backend.
type noSearchFake struct {
	backend.Client
}

func TestCode_Cloud_PassesQueryAndWorkspace(t *testing.T) {
	t.Parallel()
	var gotWS, gotQuery string
	var gotLimit int
	fake := &testhelpers.FakeClient{
		T: t,
		SearchCodeFn: func(ws, q string, limit int) ([]backend.CodeSearchHit, error) {
			gotWS, gotQuery, gotLimit = ws, q, limit
			return []backend.CodeSearchHit{
				{Repository: "acme/widgets", Path: "src/README.md", ContentMatchCount: 2},
			}, nil
		},
	}
	f, out, _ := newFactory(t, fake)
	cmd := search.NewCmdSearchCode(f)
	cmd.SetArgs([]string{"TODO", "--workspace", "acme", "--limit", "50"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "acme", gotWS)
	assert.Equal(t, "TODO", gotQuery)
	assert.Equal(t, 50, gotLimit)
	got := out.String()
	assert.Contains(t, got, "acme/widgets")
	assert.Contains(t, got, "src/README.md")
}

func TestCode_Cloud_DefaultsWorkspaceFromPinnedRepo(t *testing.T) {
	t.Parallel()
	// No --workspace flag: when the current checkout has a pinned default
	// repo (`bitbottle repo set-default`), search code reuses its
	// project/workspace as the workspace.
	var gotWS string
	fake := &testhelpers.FakeClient{
		T: t,
		SearchCodeFn: func(ws, q string, limit int) ([]backend.CodeSearchHit, error) {
			gotWS = ws
			return nil, nil
		},
	}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, fake)
	r := testhelpers.NewFakeRunner()
	r.BitbottleConfig = map[string]string{
		"bitbottle.host":    "bitbucket.org",
		"bitbottle.project": "acme",
		"bitbottle.slug":    "widgets",
	}
	f.GitRunner = func() run.Runner { return r }

	cmd := search.NewCmdSearchCode(f)
	cmd.SetArgs([]string{"TODO"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "acme", gotWS, "workspace must default to the pinned repo's project")
}

func TestCode_Cloud_RejectsEmptyQuery(t *testing.T) {
	t.Parallel()
	f, _, _ := newFactory(t, &testhelpers.FakeClient{T: t})
	cmd := search.NewCmdSearchCode(f)
	cmd.SetArgs([]string{}) // no QUERY positional
	err := cmd.Execute()
	require.Error(t, err)
}

func TestCode_ServerBackend_ReturnsUnsupported(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cloudConfig, BackendType: "cloud"})
	factorytest.UseBackend(f, noSearchFake{Client: &testhelpers.FakeClient{T: t}})
	r := runner()
	f.GitRunner = func() run.Runner { return r }

	cmd := search.NewCmdSearchCode(f)
	cmd.SetArgs([]string{"TODO", "--workspace", "acme"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Cloud only")
}

func TestCode_JSONOutputCarriesShape(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		SearchCodeFn: func(ws, q string, limit int) ([]backend.CodeSearchHit, error) {
			return []backend.CodeSearchHit{
				{
					Repository:        "acme/widgets",
					Path:              "src/README.md",
					ContentMatchCount: 1,
					ContentMatches: []backend.ContentMatch{
						{Line: 7, Segments: []backend.SearchSegment{{Text: "TODO", Match: true}}},
					},
				},
			}, nil
		},
	}
	f, out, _ := newFactory(t, fake)
	cmd := search.NewCmdSearchCode(f)
	cmd.SetArgs([]string{"TODO", "--workspace", "acme", "--json", "repository,path,contentMatchCount"})
	require.NoError(t, cmd.Execute())

	var got []map[string]any
	require.NoError(t, json.Unmarshal(out.Bytes(), &got))
	require.Len(t, got, 1)
	assert.Equal(t, "acme/widgets", got[0]["repository"])
	assert.Equal(t, "src/README.md", got[0]["path"])
	// JSON-decoded numbers are float64 by default.
	assert.InDelta(t, 1.0, got[0]["contentMatchCount"], 0)
}

func TestCode_JQFiltersJSONOutput(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		SearchCodeFn: func(ws, q string, limit int) ([]backend.CodeSearchHit, error) {
			return []backend.CodeSearchHit{
				{Repository: "acme/widgets", Path: "src/a.go"},
				{Repository: "acme/widgets", Path: "src/b.go"},
			}, nil
		},
	}
	f, out, _ := newFactory(t, fake)
	cmd := search.NewCmdSearchCode(f)
	cmd.SetArgs([]string{
		"TODO", "--workspace", "acme",
		"--json", "path",
		"--jq", ".[].path",
	})
	require.NoError(t, cmd.Execute())
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	require.Len(t, lines, 2)
	assert.Contains(t, lines[0], "src/a.go")
	assert.Contains(t, lines[1], "src/b.go")
}
