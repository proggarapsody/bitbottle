package pr_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pr"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

// fakeChangesClient embeds FakeClient and additionally implements
// backend.PRChangesRequester, for use in request-changes tests.
type fakeChangesClient struct {
	*testhelpers.FakeClient
	RequestChangesPRFn func(ns, slug string, id int) error
}

func (f *fakeChangesClient) RequestChangesPR(ns, slug string, id int) error {
	if f.RequestChangesPRFn != nil {
		return f.RequestChangesPRFn(ns, slug, id)
	}
	if f.T != nil {
		f.T.Fatalf("unexpected call to fakeChangesClient.RequestChangesPR")
	}
	return nil
}

// Compile-time assertion: fakeChangesClient implements PRChangesRequester.
var _ backend.PRChangesRequester = (*fakeChangesClient)(nil)

func TestPRRequestChanges_PrintsConfirmation(t *testing.T) {
	t.Parallel()
	fake := &fakeChangesClient{
		FakeClient: &testhelpers.FakeClient{T: t},
		RequestChangesPRFn: func(ns, slug string, id int) error {
			return nil
		},
	}
	f, out, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRRequestChanges(f)
	cmd.SetArgs([]string{"42"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "Requested changes on pull request #42")
}

func TestPRRequestChanges_UnsupportedOnServer(t *testing.T) {
	t.Parallel()
	// Plain FakeClient does NOT implement PRChangesRequester.
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRRequestChanges(f)
	cmd.SetArgs([]string{"42"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not supported")
}

func TestPRRequestChanges_APIError_PropagatesError(t *testing.T) {
	t.Parallel()
	fake := &fakeChangesClient{
		FakeClient: &testhelpers.FakeClient{T: t},
		RequestChangesPRFn: func(ns, slug string, id int) error {
			return errors.New("403 forbidden")
		},
	}
	f, _, _ := newPRFactory(t, fake, newPRRunner())
	cmd := pr.NewCmdPRRequestChanges(f)
	cmd.SetArgs([]string{"42"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "403")
}
