package pipeline_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pipeline"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdPipelineView_HasFlags(t *testing.T) {
	t.Parallel()
	f, _, _ := newPipelineFactory(t, &testhelpers.FakeClient{T: t}, newPipelineRunner())
	cmd := pipeline.NewCmdPipelineView(f)
	assert.NotNil(t, cmd.Flag("web"))
	assert.NotNil(t, cmd.Flag("json"))
	assert.NotNil(t, cmd.Flag("jq"))
}

func TestNewCmdPipelineView_RequiresTwoArgs(t *testing.T) {
	t.Parallel()
	f, _, _ := newPipelineFactory(t, &testhelpers.FakeClient{T: t}, newPipelineRunner())
	cmd := pipeline.NewCmdPipelineView(f)
	cmd.SetArgs([]string{"myworkspace/my-service"}) // missing UUID
	err := cmd.Execute()
	require.Error(t, err)
}

func TestPipelineView_PrintsDetails(t *testing.T) {
	t.Parallel()

	uuid := "{aabbccdd-1234-5678-abcd-000000000001}"
	fake := &testhelpers.FakeClient{
		T: t,
		GetPipelineFn: func(ns, slug, u string) (backend.Pipeline, error) {
			return backend.Pipeline{
				UUID:        uuid,
				BuildNumber: 42,
				State:       "SUCCESSFUL",
				RefName:     "main",
				Duration:    120,
			}, nil
		},
	}

	f, out, _ := newPipelineFactory(t, fake, newPipelineRunner())
	cmd := pipeline.NewCmdPipelineView(f)
	cmd.SetArgs([]string{"myworkspace/my-service", uuid})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "42")
	assert.Contains(t, got, "SUCCESSFUL")
	assert.Contains(t, got, "main")
}

func TestPipelineView_WebFlag_OpensBrowser(t *testing.T) {
	t.Parallel()

	uuid := "{aabbccdd-1234-5678-abcd-000000000001}"
	webURL := "https://bitbucket.org/myworkspace/my-service/addon/pipelines/home#!/results/42"
	fake := &testhelpers.FakeClient{
		T: t,
		GetPipelineFn: func(ns, slug, u string) (backend.Pipeline, error) {
			return backend.Pipeline{
				UUID:        uuid,
				BuildNumber: 42,
				State:       "SUCCESSFUL",
				WebURL:      webURL,
			}, nil
		},
	}

	browser := &testhelpers.FakeBrowserLauncher{}
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig:   pipelineConfig,
		BackendOverride: fake,
		GitRunner:       newPipelineRunner(),
		Browser:         browser,
	})
	cmd := pipeline.NewCmdPipelineView(f)
	cmd.SetArgs([]string{"myworkspace/my-service", uuid, "--web"})
	require.NoError(t, cmd.Execute())

	require.Len(t, browser.URLs, 1)
	assert.Equal(t, webURL, browser.URLs[0])
}

func TestPipelineView_APIError_PropagatesError(t *testing.T) {
	t.Parallel()

	uuid := "{aabbccdd-1234-5678-abcd-000000000001}"
	apiErr := errors.New("404 not found")
	fake := &testhelpers.FakeClient{
		T: t,
		GetPipelineFn: func(ns, slug, u string) (backend.Pipeline, error) {
			return backend.Pipeline{}, apiErr
		},
	}

	f, _, _ := newPipelineFactory(t, fake, newPipelineRunner())
	cmd := pipeline.NewCmdPipelineView(f)
	cmd.SetArgs([]string{"myworkspace/my-service", uuid})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}

func TestPipelineView_JSON_EmitsObject(t *testing.T) {
	t.Parallel()

	uuid := "{aabbccdd-1234-5678-abcd-000000000001}"
	fake := &testhelpers.FakeClient{
		T: t,
		GetPipelineFn: func(ns, slug, u string) (backend.Pipeline, error) {
			return backend.Pipeline{
				UUID:        uuid,
				BuildNumber: 42,
				State:       "SUCCESSFUL",
			}, nil
		},
	}

	f, out, _ := newPipelineFactory(t, fake, newPipelineRunner())
	cmd := pipeline.NewCmdPipelineView(f)
	cmd.SetArgs([]string{"myworkspace/my-service", uuid, "--json", "buildNumber,state"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, `"buildNumber":42`)
	assert.Contains(t, got, `"state":"SUCCESSFUL"`)
}
