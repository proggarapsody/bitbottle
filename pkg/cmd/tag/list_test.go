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

const tagConfig = "bitbucket.org:\n  oauth_token: tok\n  user: alice\n  git_protocol: https\n"

func TestTagList_PrintsTable(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		ListTagsFn: func(ns, slug string, limit int) ([]backend.Tag, error) {
			return []backend.Tag{
				{Name: "v1.0.0", Hash: "abc1234def567890", Message: "Release v1.0.0", WebURL: "https://example.com/v1.0.0"},
				{Name: "v0.9.0", Hash: "deadbeef01234567", Message: "", WebURL: ""},
			}, nil
		},
	}

	f, out, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig:   tagConfig,
		BackendOverride: fake,
	})
	cmd := tag.NewCmdTagList(f)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	require.NoError(t, cmd.Execute())

	got := out.String()
	assert.Contains(t, got, "v1.0.0")
	assert.Contains(t, got, "v0.9.0")
}

func TestTagList_PassesLimitToAPI(t *testing.T) {
	t.Parallel()

	var gotLimit int
	fake := &testhelpers.FakeClient{
		T: t,
		ListTagsFn: func(ns, slug string, limit int) ([]backend.Tag, error) {
			gotLimit = limit
			return nil, nil
		},
	}

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig:   tagConfig,
		BackendOverride: fake,
	})
	cmd := tag.NewCmdTagList(f)
	cmd.SetArgs([]string{"myworkspace/my-service", "--limit", "50"})
	require.NoError(t, cmd.Execute())

	assert.Equal(t, 50, gotLimit)
}

func TestTagList_APIError_PropagatesError(t *testing.T) {
	t.Parallel()

	fake := &testhelpers.FakeClient{
		T: t,
		ListTagsFn: func(ns, slug string, limit int) ([]backend.Tag, error) {
			return nil, errors.New("tags not found")
		},
	}

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig:   tagConfig,
		BackendOverride: fake,
	})
	cmd := tag.NewCmdTagList(f)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tags not found")
}
