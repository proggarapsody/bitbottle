package transfer_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/repo/transfer"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestNewCmdTransfer_RequiresOneArg(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{})
	cmd := transfer.NewCmdTransfer(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"--to", "NEWPROJ"})
	err := cmd.Execute()
	require.Error(t, err)
}

func TestNewCmdTransfer_RequiresToFlag(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{})
	cmd := transfer.NewCmdTransfer(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"MYPROJ/my-repo"})
	err := cmd.Execute()
	require.Error(t, err)
}

func TestRepoTransfer_CallsBackendWithTarget(t *testing.T) {
	t.Parallel()
	var gotNS, gotSlug, gotTarget string
	fake := &testhelpers.FakeClient{
		T: t,
		TransferRepoFn: func(ns, slug, target string) (backend.Repository, error) {
			gotNS, gotSlug, gotTarget = ns, slug, target
			return backend.Repository{Slug: slug, Name: slug, Namespace: target}, nil
		},
	}
	f, _, _ := newRepoFactory(t, fake)
	cmd := transfer.NewCmdTransfer(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"MYPROJ/my-repo", "--to", "NEWPROJ"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "MYPROJ", gotNS)
	assert.Equal(t, "my-repo", gotSlug)
	assert.Equal(t, "NEWPROJ", gotTarget)
}

func TestRepoTransfer_PrintsTransferredCoordinate(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		TransferRepoFn: func(ns, slug, target string) (backend.Repository, error) {
			return backend.Repository{Slug: slug, Name: slug, Namespace: target}, nil
		},
	}
	f, out, _ := newRepoFactory(t, fake)
	cmd := transfer.NewCmdTransfer(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"MYPROJ/my-repo", "--to", "NEWPROJ"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "NEWPROJ/my-repo")
}

func TestRepoTransfer_JSONOutput(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		TransferRepoFn: func(ns, slug, target string) (backend.Repository, error) {
			return backend.Repository{Slug: slug, Name: slug, Namespace: target}, nil
		},
	}
	f, out, _ := newRepoFactory(t, fake)
	cmd := transfer.NewCmdTransfer(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"MYPROJ/my-repo", "--to", "NEWPROJ", "--json"})
	require.NoError(t, cmd.Execute())
	got := out.String()
	assert.Contains(t, got, `"slug":"my-repo"`)
	assert.Contains(t, got, `"namespace":"NEWPROJ"`)
}

func TestRepoTransfer_APIError_Propagates(t *testing.T) {
	t.Parallel()
	apiErr := errors.New("403 forbidden")
	fake := &testhelpers.FakeClient{
		T: t,
		TransferRepoFn: func(ns, slug, target string) (backend.Repository, error) {
			return backend.Repository{}, apiErr
		},
	}
	f, _, _ := newRepoFactory(t, fake)
	cmd := transfer.NewCmdTransfer(f)
	format.RegisterOutputFlags(cmd)
	cmd.SetArgs([]string{"MYPROJ/my-repo", "--to", "NEWPROJ"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "403")
}
