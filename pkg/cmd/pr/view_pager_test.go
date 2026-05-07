package pr_test

import (
	"testing"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pr"
	"github.com/proggarapsody/bitbottle/pkg/cmdutil"
)

// TestPRView_HasPagerAnnotation pins that pr view opts in to $PAGER on a
// TTY. PR descriptions and reviewer/build-status sections routinely
// overflow a terminal page, and the annotation handler at the root
// (cmdutil.EnablePagerForAnnotated) wraps RunE only when this annotation
// is set. Without it, users see scrollback-only output.
func TestPRView_HasPagerAnnotation(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{})
	cmd := pr.NewCmdPRView(f)
	if got := cmd.Annotations[cmdutil.PagerAnnotation]; got != "true" {
		t.Errorf("pr view: %s annotation = %q, want %q", cmdutil.PagerAnnotation, got, "true")
	}
}
