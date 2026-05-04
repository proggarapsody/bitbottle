package cmdutil

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/iostreams"
)

// PagerAnnotation is the cobra annotation key that marks a command as
// long-output and pager-eligible. Commands set
//
//	Annotations: map[string]string{cmdutil.PagerAnnotation: "true"}
//
// instead of calling StartPager/StopPager inline. This keeps pager
// lifecycle out of command bodies — adding pager support to a new
// command is one annotation, not a 5-line copy-paste with an easily-
// missed defer.
const PagerAnnotation = "pager"

// EnablePagerForAnnotated walks the command tree rooted at root and wraps
// every annotated command's RunE so the pager is started before the
// command runs and stopped after. The pager is a no-op outside a TTY,
// so unmarked commands and piped invocations are unaffected.
//
// Walking happens at registration time (call once after building the
// tree). The wrap is invisible to the command body — a command can
// still call f.IOStreams.Out directly, and gets the pager pipe for free.
func EnablePagerForAnnotated(root *cobra.Command, ios *iostreams.IOStreams) {
	wrapIfAnnotated(root, ios)
	for _, cmd := range root.Commands() {
		EnablePagerForAnnotated(cmd, ios) // recurse into subtrees
	}
}

func wrapIfAnnotated(cmd *cobra.Command, ios *iostreams.IOStreams) {
	if cmd.Annotations[PagerAnnotation] != "true" || cmd.RunE == nil {
		return
	}
	original := cmd.RunE
	cmd.RunE = func(c *cobra.Command, args []string) error {
		if err := ios.StartPager(); err != nil {
			return err
		}
		defer ios.StopPager()
		return original(c, args)
	}
}
