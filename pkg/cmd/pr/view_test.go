package pr_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pr"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdPRView_HasWebFlag(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := pr.NewCmdPRView(f)
	assert.NotNil(t, cmd.Flag("web"))
}

func TestNewCmdPRView_RequiresArg(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := pr.NewCmdPRView(f)
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	require.Error(t, err)
}

func TestPRView_PrintsDetails(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		GetPRFn: func(ns, slug string, id int) (backend.PullRequest, error) {
			return testhelpers.BackendPRFactory(
				testhelpers.BackendPRWithID(42),
				testhelpers.BackendPRWithTitle("Hello PR"),
				testhelpers.BackendPRWithFromBranch("feat/x"),
			), nil
		},
	}
	f, out, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRView(f)
	cmd.SetArgs([]string{"42"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "Hello PR")
	assert.Contains(t, got, "feat/x")
}

func TestPRView_WebFlag_OpensBrowser(t *testing.T) {
	t.Parallel()

	url := "https://bb.example.com/projects/MYPROJ/repos/my-service/pull-requests/42"
	fake := &testhelpers.FakeClient{
		T: t,
		GetPRFn: func(ns, slug string, id int) (backend.PullRequest, error) {
			return testhelpers.BackendPRFactory(
				testhelpers.BackendPRWithID(42),
				testhelpers.BackendPRWithWebURL(url),
			), nil
		},
	}
	browser := &testhelpers.FakeBrowserLauncher{}
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig:   prConfig,
		BackendOverride: fake,
		GitRunner:       newPRRunner(),
		Browser:         browser,
	})
	cmd := pr.NewCmdPRView(f)
	cmd.SetArgs([]string{"42", "--web"})
	require.NoError(t, cmd.Execute())

	require.Len(t, browser.URLs, 1)
	assert.Equal(t, url, browser.URLs[0])
}

func TestPRView_APIError_PropagatesError(t *testing.T) {
	t.Parallel()

	apiErr := errors.New("404 not found")
	fake := &testhelpers.FakeClient{
		T: t,
		GetPRFn: func(ns, slug string, id int) (backend.PullRequest, error) {
			return backend.PullRequest{}, apiErr
		},
	}
	f, _, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRView(f)
	cmd.SetArgs([]string{"42"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}

func TestNewCmdPRView_HasJSONAndJQFlags(t *testing.T) {
	t.Parallel()
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{})
	cmd := pr.NewCmdPRView(f)
	assert.NotNil(t, cmd.Flag("json"))
	assert.NotNil(t, cmd.Flag("jq"))
}

func TestPRView_JSON_EmitsObject(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		GetPRFn: func(ns, slug string, id int) (backend.PullRequest, error) {
			return testhelpers.BackendPRFactory(
				testhelpers.BackendPRWithID(42),
				testhelpers.BackendPRWithTitle("Hello PR"),
			), nil
		},
	}
	f, out, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRView(f)
	cmd.SetArgs([]string{"42", "--json", "id,title"})
	require.NoError(t, cmd.Execute())

	got := strings.TrimSpace(out.String())
	assert.True(t, strings.HasPrefix(got, "{"), "expected JSON object, got: %s", got)
	assert.Contains(t, got, `"id":42`)
	assert.Contains(t, got, `"title":"Hello PR"`)
}

func TestPRView_JSON_IncludesDescription(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		GetPRFn: func(ns, slug string, id int) (backend.PullRequest, error) {
			p := testhelpers.BackendPRFactory(testhelpers.BackendPRWithID(1))
			p.Description = "Some description"
			return p, nil
		},
	}
	f, out, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRView(f)
	cmd.SetArgs([]string{"1", "--json", "id,description"})
	require.NoError(t, cmd.Execute())

	assert.Contains(t, out.String(), `"description":"Some description"`)
}
