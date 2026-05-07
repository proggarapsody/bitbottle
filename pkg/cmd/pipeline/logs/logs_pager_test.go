package logs_test

import (
	"testing"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/pipeline/logs"
	"github.com/proggarapsody/bitbottle/pkg/cmdutil"
)

// TestPipelineLogs_HasPagerAnnotation pins pipeline logs as a $PAGER
// opt-in. Build logs are routinely thousands of lines; without the
// annotation users hit unpaginated output on a TTY and lose the
// scroll/search affordances of less.
func TestPipelineLogs_HasPagerAnnotation(t *testing.T) {
	t.Parallel()
	f, _, _ := factorytest.New(t, factorytest.Opts{})
	cmd := logs.NewCmdLogs(f, nil)
	if got := cmd.Annotations[cmdutil.PagerAnnotation]; got != "true" {
		t.Errorf("pipeline logs: %s annotation = %q, want %q", cmdutil.PagerAnnotation, got, "true")
	}
}
