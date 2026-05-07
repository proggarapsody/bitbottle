package commit_test

import (
	"testing"

	"github.com/proggarapsody/bitbottle/pkg/cmd/commit"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmdutil"
)

// TestCommitView_HasPagerAnnotation pins commit view as a $PAGER opt-in.
// Commit messages with multi-paragraph bodies plus parent/diff metadata
// can exceed a terminal page, mirroring `git show` behaviour.
func TestCommitView_HasPagerAnnotation(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{})
	cmd := commit.NewCmdCommitView(f)
	if got := cmd.Annotations[cmdutil.PagerAnnotation]; got != "true" {
		t.Errorf("commit view: %s annotation = %q, want %q", cmdutil.PagerAnnotation, got, "true")
	}
}
