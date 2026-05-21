package visibility_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/repo/visibility"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestRepoVisibility_GetPrivate(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetRepoFn: func(ns, slug string) (backend.Repository, error) {
			return backend.Repository{
				Slug:      slug,
				Namespace: ns,
				IsPrivate: true,
			}, nil
		},
	}
	f, out, _ := newRepoFactory(t, fake)
	cmd := visibility.NewCmdVisibility(f)
	cmd.SetArgs([]string{"MYPROJ/my-service"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "private\n", out.String())
}

func TestRepoVisibility_GetPublic(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		GetRepoFn: func(ns, slug string) (backend.Repository, error) {
			return backend.Repository{
				Slug:      slug,
				Namespace: ns,
				IsPrivate: false,
			}, nil
		},
	}
	f, out, _ := newRepoFactory(t, fake)
	cmd := visibility.NewCmdVisibility(f)
	cmd.SetArgs([]string{"MYPROJ/my-service"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "public\n", out.String())
}

func TestRepoVisibility_SetPublic(t *testing.T) {
	t.Parallel()
	var gotNS, gotSlug string
	var gotPrivate bool
	fake := &testhelpers.FakeClient{
		T: t,
		SetRepoVisibilityFn: func(ns, slug string, isPrivate bool) error {
			gotNS, gotSlug, gotPrivate = ns, slug, isPrivate
			return nil
		},
	}
	f, out, _ := newRepoFactory(t, fake)
	cmd := visibility.NewCmdVisibility(f)
	cmd.SetArgs([]string{"MYPROJ/my-service", "public"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "MYPROJ", gotNS)
	assert.Equal(t, "my-service", gotSlug)
	assert.False(t, gotPrivate)
	assert.Contains(t, out.String(), "MYPROJ/my-service")
	assert.Contains(t, out.String(), "public")
}

func TestRepoVisibility_SetPrivate(t *testing.T) {
	t.Parallel()
	var gotPrivate bool
	fake := &testhelpers.FakeClient{
		T: t,
		SetRepoVisibilityFn: func(ns, slug string, isPrivate bool) error {
			gotPrivate = isPrivate
			return nil
		},
	}
	f, out, _ := newRepoFactory(t, fake)
	cmd := visibility.NewCmdVisibility(f)
	cmd.SetArgs([]string{"MYPROJ/my-service", "private"})
	require.NoError(t, cmd.Execute())
	assert.True(t, gotPrivate)
	assert.Contains(t, out.String(), "private")
}

func TestRepoVisibility_InvalidArg(t *testing.T) {
	t.Parallel()
	f, _, _ := newRepoFactory(t, nil)
	cmd := visibility.NewCmdVisibility(f)
	cmd.SetArgs([]string{"MYPROJ/my-service", "internal"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "invalid visibility") ||
		strings.Contains(err.Error(), "internal"),
		"error should mention invalid visibility: %v", err)
}
