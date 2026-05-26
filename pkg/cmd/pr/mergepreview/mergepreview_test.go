package mergepreview_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pr/mergepreview"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

func TestMergePreview_CanMerge_PrintsSuccess(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		DryRunMergePRFn: func(ns, slug string, prID int, strategy string) (backend.MergeDryRunResult, error) {
			return backend.MergeDryRunResult{CanMerge: true, Message: "All checks pass"}, nil
		},
	}
	f, out, _ := cmdtest.NewPRFactory(t, fake, cmdtest.NewPRRunner())
	cmd := mergepreview.NewCmdMergePreview(f)
	cmd.SetArgs([]string{"42"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "Can merge cleanly")
	assert.Contains(t, out.String(), "#42")
}

func TestMergePreview_CannotMerge_PrintsFailure(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		DryRunMergePRFn: func(ns, slug string, prID int, strategy string) (backend.MergeDryRunResult, error) {
			return backend.MergeDryRunResult{
				CanMerge: false,
				Message:  "Merge conflicts found",
				Vetoes: []backend.MergeVeto{
					{SummaryMessage: "Not all required approvals given", DetailMessage: "2 more approvals needed"},
				},
			}, nil
		},
	}
	f, out, _ := cmdtest.NewPRFactory(t, fake, cmdtest.NewPRRunner())
	cmd := mergepreview.NewCmdMergePreview(f)
	cmd.SetArgs([]string{"42"})
	require.NoError(t, cmd.Execute())
	o := out.String()
	assert.Contains(t, o, "Cannot merge")
	assert.Contains(t, o, "Not all required approvals given")
}

func TestMergePreview_JSONFlag_OutputsJSON(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		DryRunMergePRFn: func(ns, slug string, prID int, strategy string) (backend.MergeDryRunResult, error) {
			return backend.MergeDryRunResult{CanMerge: true}, nil
		},
	}
	f, out, _ := cmdtest.NewPRFactory(t, fake, cmdtest.NewPRRunner())
	cmd := mergepreview.NewCmdMergePreview(f)
	cmd.SetArgs([]string{"42", "--json"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), `"can_merge"`)
}

func TestMergePreview_InvalidStrategy_ReturnsError(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := cmdtest.NewPRFactory(t, fake, cmdtest.NewPRRunner())
	cmd := mergepreview.NewCmdMergePreview(f)
	cmd.SetArgs([]string{"42", "--strategy", "invalid"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid --strategy")
}

func TestMergePreview_StrategyPassedToClient(t *testing.T) {
	t.Parallel()
	var gotStrategy string
	fake := &testhelpers.FakeClient{
		T: t,
		DryRunMergePRFn: func(ns, slug string, prID int, strategy string) (backend.MergeDryRunResult, error) {
			gotStrategy = strategy
			return backend.MergeDryRunResult{CanMerge: true}, nil
		},
	}
	f, _, _ := cmdtest.NewPRFactory(t, fake, cmdtest.NewPRRunner())
	cmd := mergepreview.NewCmdMergePreview(f)
	cmd.SetArgs([]string{"42", "--strategy", "squash"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "squash", gotStrategy)
}

func TestMergePreview_APIError_PropagatesError(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{
		T: t,
		DryRunMergePRFn: func(ns, slug string, prID int, strategy string) (backend.MergeDryRunResult, error) {
			return backend.MergeDryRunResult{}, errors.New("network error")
		},
	}
	f, _, _ := cmdtest.NewPRFactory(t, fake, cmdtest.NewPRRunner())
	cmd := mergepreview.NewCmdMergePreview(f)
	cmd.SetArgs([]string{"42"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "network error")
}

func TestMergePreview_UnsupportedHost_ReturnsTypedError(t *testing.T) {
	t.Parallel()
	// A client that does NOT implement PRMergePreviewClient
	type noPreviewClient struct {
		backend.Client
	}
	wrapper := &noPreviewClient{Client: &testhelpers.FakeClient{T: t}}
	f, _, _ := cmdtest.NewPRFactory(t, wrapper, cmdtest.NewPRRunner())
	cmd := mergepreview.NewCmdMergePreview(f)
	cmd.SetArgs([]string{"42"})
	err := cmd.Execute()
	require.Error(t, err)
	var de *backend.DomainError
	require.ErrorAs(t, err, &de)
	assert.Equal(t, backend.ErrUnsupportedOnHost, de.Kind)
}
