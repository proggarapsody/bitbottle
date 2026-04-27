// Adapted from cli/cli (MIT) — pkg/cmd/root/help.go.
// Renders sectioned --help output with a custom ARGUMENTS section sourced
// from Annotations["help:arguments"] on the command or any ancestor.

package root

import (
	"bytes"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

// argumentsAnnotationKey is the Cobra command annotation read by helpFunc to
// render the ARGUMENTS section. Set it on a parent command to share argument
// documentation across all leaves.
const argumentsAnnotationKey = "help:arguments"

// SetHelpFunc installs the bitbottle help renderer on cmd and every descendant.
func SetHelpFunc(cmd *cobra.Command) {
	cmd.SetHelpFunc(helpFunc)
}

func helpFunc(cmd *cobra.Command, _ []string) {
	out := cmd.OutOrStdout()

	if cmd.Long != "" {
		fmt.Fprintln(out, cmd.Long)
	} else if cmd.Short != "" {
		fmt.Fprintln(out, cmd.Short)
	}
	fmt.Fprintln(out)

	type section struct {
		title string
		body  string
	}
	var sections []section

	sections = append(sections, section{"USAGE", "  " + cmd.UseLine()})

	if len(cmd.Aliases) > 0 {
		sections = append(sections, section{"ALIASES", "  " + strings.Join(cmd.Aliases, ", ")})
	}

	if cmd.HasAvailableSubCommands() {
		sections = append(sections, section{"COMMANDS", subcommandList(cmd)})
	}

	if cmd.HasAvailableLocalFlags() {
		sections = append(sections, section{"FLAGS", strings.TrimRight(cmd.LocalFlags().FlagUsages(), "\n")})
	}
	if cmd.HasAvailableInheritedFlags() {
		sections = append(sections, section{"INHERITED FLAGS", strings.TrimRight(cmd.InheritedFlags().FlagUsages(), "\n")})
	}

	if a := lookupAncestorAnnotation(cmd, argumentsAnnotationKey); a != "" {
		sections = append(sections, section{"ARGUMENTS", a})
	}

	if cmd.Example != "" {
		sections = append(sections, section{"EXAMPLES", cmd.Example})
	}

	for _, s := range sections {
		fmt.Fprintf(out, "%s\n", s.title)
		body := strings.TrimRight(s.body, "\n")
		// Pre-formatted multi-column blocks (FLAGS, COMMANDS) carry their own
		// internal alignment from tabwriter/pflag — only indent if every line
		// already starts with whitespace, otherwise add a uniform indent.
		fmt.Fprintln(out, ensureIndent(body, "  "))
		fmt.Fprintln(out)
	}
}

// ensureIndent applies prefix to every non-empty line that doesn't already
// start with whitespace.
func ensureIndent(s, prefix string) string {
	var buf bytes.Buffer
	for _, line := range strings.Split(s, "\n") {
		if line == "" {
			fmt.Fprintln(&buf)
			continue
		}
		if line[0] == ' ' || line[0] == '\t' {
			fmt.Fprintln(&buf, line)
		} else {
			fmt.Fprintln(&buf, prefix+line)
		}
	}
	return strings.TrimRight(buf.String(), "\n")
}

// lookupAncestorAnnotation walks up the command chain returning the first
// non-empty value of the given annotation key.
func lookupAncestorAnnotation(cmd *cobra.Command, key string) string {
	for c := cmd; c != nil; c = c.Parent() {
		if v, ok := c.Annotations[key]; ok && v != "" {
			return v
		}
	}
	return ""
}

func subcommandList(cmd *cobra.Command) string {
	var buf bytes.Buffer
	tw := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	for _, sub := range cmd.Commands() {
		if !sub.IsAvailableCommand() {
			continue
		}
		fmt.Fprintf(tw, "%s\t%s\n", sub.Name(), sub.Short)
	}
	_ = tw.Flush()
	return buf.String()
}
