package repo_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/repo"
	"github.com/proggarapsody/bitbottle/pkg/iostreams"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

// TestRepoDelete_TTYConfirmAborted verifies that when the user is prompted
// interactively (TTY iostreams) and answers "n", the command:
//   - prints "Deletion aborted." to stdout
//   - does NOT call DeleteRepo
//   - exits with no error (decision belongs to the user)
func TestRepoDelete_TTYConfirmAborted(t *testing.T) {
	t.Parallel()

	// FakeClient with no DeleteRepoFn set will t.Fatalf on any unexpected call
	// — so this asserts non-invocation purely by the absence of test failure.
	fake := &testhelpers.FakeClient{T: t}

	stdout := &bytes.Buffer{}
	ios := iostreams.TestTTY()
	ios.In = io.NopCloser(strings.NewReader("n\n"))
	ios.Out = stdout

	f, _, _ := factory.NewTestFactory(t, factory.TestFactoryOpts{
		InitialConfig:   repoConfig,
		BackendOverride: fake,
		IOStreams:       ios,
	})

	cmd := repo.NewCmdRepoDelete(f)
	cmd.SetArgs([]string{"MYPROJ/my-service"})
	require.NoError(t, cmd.Execute())

	assert.Contains(t, stdout.String(), "Deletion aborted",
		"command must print abort message when user answers n")
}
