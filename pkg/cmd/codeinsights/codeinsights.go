// Package codeinsights is the root of the `code-insights` command group.
// Code Insights is a Bitbucket Server / Data Center feature only —
// invocations against Cloud surface a typed ErrUnsupportedOnHost via the
// backend.AsCodeInsightsClient accessor.
package codeinsights

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/codeinsights/annotation"
	"github.com/proggarapsody/bitbottle/pkg/cmd/codeinsights/mergecheck"
	"github.com/proggarapsody/bitbottle/pkg/cmd/codeinsights/report"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdCodeInsights builds the `code-insights` cobra command tree.
func NewCmdCodeInsights(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "code-insights",
		Short: "Manage Code Insights reports and annotations (Bitbucket Server / DC only)",
		Long: `Manage Code Insights reports, annotations, and merge checks on
Bitbucket Server / Data Center. Code Insights enables CI tools, scanners,
and quality gates to attach structured results to commits.

Bitbucket Cloud does not support this API — calling these subcommands
against Cloud returns a typed "unsupported on host" error.`,
		Annotations: map[string]string{
			"help:arguments": `A repository can be supplied as PROJECT/REPO. When omitted, the
repository is inferred from the "origin" git remote in the current
directory.`,
		},
	}
	factory.EnableRepoOverride(cmd, f)
	cmd.AddCommand(report.NewCmdReport(f))
	cmd.AddCommand(annotation.NewCmdAnnotation(f))
	cmd.AddCommand(mergecheck.NewCmdMergeCheck(f))
	return cmd
}
