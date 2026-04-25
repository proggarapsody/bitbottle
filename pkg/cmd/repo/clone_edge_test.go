package repo_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/repo"
	"github.com/proggarapsody/bitbottle/test/testhelpers"
)

// TestRepoClone_Cloud_HTTPS_UsesBitbucketOrg verifies that when the configured
// host is bitbucket.org with git_protocol: https (cloud + HTTPS) the clone URL
// is the canonical https://bitbucket.org/NS/SLUG.git form rather than a
// self-hosted /scm/ path.
func TestRepoClone_Cloud_HTTPS_UsesBitbucketOrg(t *testing.T) {
	t.Parallel()

	const cloudHTTPSConfig = "bitbucket.org:\n  oauth_token: tok\n  user: alice\n  git_protocol: https\n  backend_type: cloud\n"

	runner := testhelpers.NewFakeRunner(testhelpers.RunResponse{})
	f, _, _ := newRepoRunnerFactory(t, nil, cloudHTTPSConfig, runner)
	cmd := repo.NewCmdRepoClone(f)
	cmd.SetArgs([]string{"myworkspace/my-service"})
	require.NoError(t, cmd.Execute())

	require.NotEmpty(t, runner.Calls, "expected at least one git call")
	last := runner.Calls[len(runner.Calls)-1]
	require.GreaterOrEqual(t, len(last.Args), 2)
	url := last.Args[1]

	assert.True(t, strings.HasPrefix(url, "https://bitbucket.org/"),
		"cloud HTTPS clone URL must start with https://bitbucket.org/, got %q", url)
	assert.Contains(t, url, "myworkspace/my-service",
		"clone URL must contain workspace/slug")
	assert.NotContains(t, url, "/scm/",
		"cloud URL must not include the Data Center /scm/ prefix")
}
