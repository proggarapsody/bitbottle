package view_test

import (
	"testing"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/repo/view"
	"github.com/proggarapsody/bitbottle/pkg/cmdutil"
)

// TestRepoView_HasPagerAnnotation pins repo view as a $PAGER opt-in.
// Repository descriptions plus the metadata block can exceed a terminal
// page; same rationale as pr view and commit log.
func TestRepoView_HasPagerAnnotation(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{})
	cmd := view.NewCmdView(f)
	if got := cmd.Annotations[cmdutil.PagerAnnotation]; got != "true" {
		t.Errorf("repo view: %s annotation = %q, want %q", cmdutil.PagerAnnotation, got, "true")
	}
}
