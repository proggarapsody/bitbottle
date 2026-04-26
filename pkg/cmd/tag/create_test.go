package tag_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/tag"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestTagCreate_RequiresStartAt(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig:   tagConfig,
		BackendOverride: fake,
	})
	cmd := tag.NewCmdTagCreate(f)
	cmd.SetArgs([]string{"myworkspace/my-service", "v1.0.0"}) // missing --start-at
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "start-at")
}

func TestTagCreate_PrintsWebURL(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		CreateTagFn: func(ns, slug string, in backend.CreateTagInput) (backend.Tag, error) {
			return backend.Tag{
				Name:   in.Name,
				Hash:   "abc1234",
				WebURL: "https://example.com/v1.0.0",
			}, nil
		},
	}

	f, out, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig:   tagConfig,
		BackendOverride: fake,
	})
	cmd := tag.NewCmdTagCreate(f)
	cmd.SetArgs([]string{"myworkspace/my-service", "v1.0.0", "--start-at", "main"})
	require.NoError(t, cmd.Execute())

	assert.Contains(t, out.String(), "https://example.com/v1.0.0")
}

func TestTagCreate_PassesMessageToAPI(t *testing.T) {
	t.Parallel()

	var gotIn backend.CreateTagInput
	fake := &testhelpers.FakeClient{
		T: t,
		CreateTagFn: func(ns, slug string, in backend.CreateTagInput) (backend.Tag, error) {
			gotIn = in
			return backend.Tag{Name: in.Name}, nil
		},
	}

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig:   tagConfig,
		BackendOverride: fake,
	})
	cmd := tag.NewCmdTagCreate(f)
	cmd.SetArgs([]string{"myworkspace/my-service", "v1.0.0", "--start-at", "main", "--message", "Release notes"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "v1.0.0", gotIn.Name)
	assert.Equal(t, "main", gotIn.StartAt)
	assert.Equal(t, "Release notes", gotIn.Message)
}

func TestTagCreate_APIError_PropagatesError(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		CreateTagFn: func(ns, slug string, in backend.CreateTagInput) (backend.Tag, error) {
			return backend.Tag{}, errors.New("tag already exists")
		},
	}

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig:   tagConfig,
		BackendOverride: fake,
	})
	cmd := tag.NewCmdTagCreate(f)
	cmd.SetArgs([]string{"myworkspace/my-service", "v1.0.0", "--start-at", "main"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tag already exists")
}
