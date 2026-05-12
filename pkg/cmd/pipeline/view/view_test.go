package view_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/internal/run"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pipeline/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pipeline/view"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdView_HasFlags(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := view.NewCmdView(f, nil)
	format.RegisterOutputFlags(cmd)
	assert.NotNil(t, cmd.Flag("web"))
	assert.NotNil(t, cmd.Flag("json"))
	assert.NotNil(t, cmd.Flag("jq"))
}

func TestNewCmdView_RequiresTwoArgs(t *testing.T) {
	t.Parallel()
	f, _, _ := cmdtest.NewFactory(t, &testhelpers.FakeClient{T: t}, cmdtest.NewRunner())
	cmd := view.NewCmdView(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service"}) // missing UUID
	require.Error(t, cmd.Execute())
}

func TestView_PrintsDetails(t *testing.T) {
	t.Parallel()
	uuid := "{aabbccdd-1234-5678-abcd-000000000001}"
	fake := &testhelpers.FakeClient{
		T: t,
		GetPipelineFn: func(ns, slug, u string) (backend.Pipeline, error) {
			return backend.Pipeline{UUID: uuid, BuildNumber: 42, State: "SUCCESSFUL", RefName: "main", Duration: 120}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := view.NewCmdView(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service", uuid})
	require.NoError(t, cmd.Execute())
	got := out.String()
	assert.Contains(t, got, "42")
	assert.Contains(t, got, "SUCCESSFUL")
	assert.Contains(t, got, "main")
}

func TestView_WebFlag_OpensBrowser(t *testing.T) {
	t.Parallel()
	uuid := "{aabbccdd-1234-5678-abcd-000000000001}"
	webURL := "https://bitbucket.org/myworkspace/my-service/addon/pipelines/home#!/results/42"
	fake := &testhelpers.FakeClient{
		T: t,
		GetPipelineFn: func(ns, slug, u string) (backend.Pipeline, error) {
			return backend.Pipeline{UUID: uuid, BuildNumber: 42, State: "SUCCESSFUL", WebURL: webURL}, nil
		},
	}
	browser := &testhelpers.FakeBrowserLauncher{}
	f, _, _ := factorytest.New(t, factorytest.Opts{InitialConfig: cmdtest.Config})
	factorytest.UseBackend(f, fake)
	f.GitRunner = func() run.Runner { return cmdtest.NewRunner() }
	f.Browser = browser
	cmd := view.NewCmdView(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service", uuid, "--web"})
	require.NoError(t, cmd.Execute())
	require.Len(t, browser.URLs, 1)
	assert.Equal(t, webURL, browser.URLs[0])
}

func TestView_APIError_PropagatesError(t *testing.T) {
	t.Parallel()
	uuid := "{aabbccdd-1234-5678-abcd-000000000001}"
	apiErr := errors.New("404 not found")
	fake := &testhelpers.FakeClient{
		T: t,
		GetPipelineFn: func(ns, slug, u string) (backend.Pipeline, error) {
			return backend.Pipeline{}, apiErr
		},
	}
	f, _, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := view.NewCmdView(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service", uuid})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}

func TestView_JSON_EmitsObject(t *testing.T) {
	t.Parallel()
	uuid := "{aabbccdd-1234-5678-abcd-000000000001}"
	fake := &testhelpers.FakeClient{
		T: t,
		GetPipelineFn: func(ns, slug, u string) (backend.Pipeline, error) {
			return backend.Pipeline{UUID: uuid, BuildNumber: 42, State: "SUCCESSFUL"}, nil
		},
	}
	f, out, _ := cmdtest.NewFactory(t, fake, cmdtest.NewRunner())
	cmd := view.NewCmdView(f, nil)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"myworkspace/my-service", uuid, "--json"})
	require.NoError(t, cmd.Execute())
	got := out.String()
	assert.Contains(t, got, `"buildNumber":42`)
	assert.Contains(t, got, `"state":"SUCCESSFUL"`)
}
