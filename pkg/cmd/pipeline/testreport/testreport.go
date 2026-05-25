// Package testreport implements `bitbottle pipeline test-report` and
// `bitbottle pipeline test-case` subcommand groups.
package testreport

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdTestReport builds the `pipeline test-report` parent command.
func NewCmdTestReport(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test-report",
		Short: "View pipeline test reports (Cloud only)",
		Annotations: map[string]string{
			"help:arguments": `A repository can be supplied as WORKSPACE/REPO. When omitted, the
repository is inferred from the "origin" git remote in the current
directory.`,
		},
	}
	factory.EnableRepoOverride(cmd, f)
	cmd.AddCommand(NewCmdView(f, nil))
	return cmd
}

// NewCmdTestCase builds the `pipeline test-case` parent command.
func NewCmdTestCase(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test-case",
		Short: "List pipeline test cases (Cloud only)",
		Annotations: map[string]string{
			"help:arguments": `A repository can be supplied as WORKSPACE/REPO. When omitted, the
repository is inferred from the "origin" git remote in the current
directory.`,
		},
	}
	factory.EnableRepoOverride(cmd, f)
	cmd.AddCommand(NewCmdList(f, nil))
	return cmd
}
