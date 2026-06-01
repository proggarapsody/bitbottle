// Package codeinsights is the root of the `code-insights` command group.
// Code Insights is supported on both Bitbucket Server / Data Center and
// Bitbucket Cloud. Server/DC uses the CodeInsightsClient; Cloud uses the
// CloudCodeInsightsClient. Commands resolve the correct adapter automatically.
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
		Short: "Manage Code Insights reports, annotations, and merge checks",
		Long: `Manage Code Insights reports and annotations on Bitbucket Server / Data
Center and Bitbucket Cloud. Code Insights enables CI tools, scanners, and
quality gates to attach structured results to commits.

Merge checks (experimental) are Bitbucket Server / DC only.`,
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
