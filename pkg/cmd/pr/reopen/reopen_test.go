package reopen_test

import (
	"errors"
	"testing"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/internal/cmdtest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pr/reopen"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeReopenClient embeds FakeClient and additionally implements
// backend.PRReopener, for use in reopen tests.
type fakeReopenClient struct {
	*testhelpers.FakeClient
	ReopenPRFn func(ns, slug string, id int) error
}

func (f *fakeReopenClient) ReopenPR(ns, slug string, id int) error {
	if f.ReopenPRFn != nil {
		return f.ReopenPRFn(ns, slug, id)
	}
	if f.T != nil {
		f.T.Fatalf("unexpected call to fakeReopenClient.ReopenPR")
	}
	return nil
}

// Compile-time assertion: fakeReopenClient implements PRReopener.
var _ backend.PRReopener = (*fakeReopenClient)(nil)

func TestPRReopen_PrintsConfirmation(t *testing.T) {
	t.Parallel()
	fake := &fakeReopenClient{
		FakeClient: &testhelpers.FakeClient{T: t},
		ReopenPRFn: func(ns, slug string, id int) error {
			return nil
		},
	}
	f, out, _ := cmdtest.NewPRFactory(t, fake, cmdtest.NewPRRunner())
	cmd := reopen.NewCmdReopen(f)
	cmd.SetArgs([]string{"42"})
	require.NoError(t, cmd.Execute())
	assert.Contains(t, out.String(), "Reopened pull request #42")
}

func TestPRReopen_PassesProjectSlugAndID(t *testing.T) {
	t.Parallel()
	var gotNS, gotSlug string
	var gotID int
	fake := &fakeReopenClient{
		FakeClient: &testhelpers.FakeClient{T: t},
		ReopenPRFn: func(ns, slug string, id int) error {
			gotNS, gotSlug, gotID = ns, slug, id
			return nil
		},
	}
	f, _, _ := cmdtest.NewPRFactory(t, fake, cmdtest.NewPRRunner())
	cmd := reopen.NewCmdReopen(f)
	cmd.SetArgs([]string{"7"})
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "MYPROJ", gotNS, "project must be uppercased by ResolvePRTarget")
	assert.Equal(t, "my-service", gotSlug)
	assert.Equal(t, 7, gotID)
}

func TestPRReopen_UnsupportedOnCloud(t *testing.T) {
	t.Parallel()
	// Plain FakeClient does NOT implement PRReopener — simulates a Cloud
	// backend, where reopen is unsupported.
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := cmdtest.NewPRFactory(t, fake, cmdtest.NewPRRunner())
	cmd := reopen.NewCmdReopen(f)
	cmd.SetArgs([]string{"42"})
	err := cmd.Execute()
	require.Error(t, err)
	// AsPRReopener returns a typed ErrUnsupportedOnHost.
	assert.ErrorIs(t, err, backend.ErrUnsupportedOnHost)
}

func TestPRReopen_APIError_PropagatesError(t *testing.T) {
	t.Parallel()
	fake := &fakeReopenClient{
		FakeClient: &testhelpers.FakeClient{T: t},
		ReopenPRFn: func(ns, slug string, id int) error {
			return errors.New("409 conflict")
		},
	}
	f, _, _ := cmdtest.NewPRFactory(t, fake, cmdtest.NewPRRunner())
	cmd := reopen.NewCmdReopen(f)
	cmd.SetArgs([]string{"42"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "409")
}

func TestPRReopen_RejectsNonNumericPRID(t *testing.T) {
	t.Parallel()
	fake := &testhelpers.FakeClient{T: t}
	f, _, _ := cmdtest.NewPRFactory(t, fake, cmdtest.NewPRRunner())
	cmd := reopen.NewCmdReopen(f)
	cmd.SetArgs([]string{"abc"})
	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid PR ID")
}
