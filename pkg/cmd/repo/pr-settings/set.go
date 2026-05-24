package prsettings

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

const errMsgNoFlags = "at least one flag is required (--required-approvers, --required-all-approvers, --required-all-tasks-complete, --required-successful-builds, --merge-strategy, --allowed-strategies)"

// NewCmdSet builds `repo pr-settings set [PROJECT/REPO] [flags] [--json]`.
func NewCmdSet(f *factory.Factory) *cobra.Command {
	var (
		hostname                 string
		requiredApprovers        int
		requiredAllApprovers     bool
		requiredAllTasksComplete bool
		requiredSuccessfulBuilds int
		mergeStrategy            string
		allowedStrategies        string
	)

	cmd := &cobra.Command{
		Use:   "set [PROJECT/REPO]",
		Short: "Update pull request gate settings for a repository",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref, err := factory.ResolveTarget(f, args, hostname)
			if err != nil {
				return err
			}

			// Build input — require at least one flag.
			in := backend.RepoPRSettingsInput{}
			noop := true

			if cmd.Flags().Changed("required-approvers") {
				in.RequiredApprovers = &requiredApprovers
				noop = false
			}
			if cmd.Flags().Changed("required-all-approvers") {
				in.RequiredAllApprovers = &requiredAllApprovers
				noop = false
			}
			if cmd.Flags().Changed("required-all-tasks-complete") {
				in.RequiredAllTasksComplete = &requiredAllTasksComplete
				noop = false
			}
			if cmd.Flags().Changed("required-successful-builds") {
				in.RequiredSuccessfulBuilds = &requiredSuccessfulBuilds
				noop = false
			}
			if cmd.Flags().Changed("merge-strategy") {
				in.MergeStrategy = &mergeStrategy
				noop = false
			}
			if cmd.Flags().Changed("allowed-strategies") {
				parts := strings.Split(allowedStrategies, ",")
				cleaned := make([]string, 0, len(parts))
				for _, p := range parts {
					p = strings.TrimSpace(p)
					if p != "" {
						cleaned = append(cleaned, p)
					}
				}
				in.AllowedStrategies = &cleaned
				noop = false
			}

			if noop {
				return &backend.DomainError{
					Kind:    backend.ErrInvalidRequest,
					Code:    backend.CodeInvalidRequest,
					Message: errMsgNoFlags,
				}
			}

			client, err := f.Backend(ref.Host)
			if err != nil {
				return err
			}

			prc, err := backend.AsRepoPRSettingsClient(client, ref.Host)
			if err != nil {
				return err
			}

			settings, err := prc.UpdateRepoPRSettings(ref.Project, ref.Slug, in)
			if err != nil {
				return err
			}

			cfg := format.ConfigFromCmd(cmd)
			if cfg.Format != format.FormatTable {
				p := prSettingsFields(f, cfg)
				p.SetSingleItem()
				p.AddItem(settings)
				return p.Render()
			}

			fmt.Fprintf(f.IOStreams.Out, "✓ PR settings for %s/%s updated.\n", ref.Project, ref.Slug)
			return nil
		},
	}

	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	cmd.Flags().IntVar(&requiredApprovers, "required-approvers", 0, "Minimum number of approvals required")
	cmd.Flags().BoolVar(&requiredAllApprovers, "required-all-approvers", false, "Require all reviewers to approve")
	cmd.Flags().BoolVar(&requiredAllTasksComplete, "required-all-tasks-complete", false, "Require all tasks to be resolved")
	cmd.Flags().IntVar(&requiredSuccessfulBuilds, "required-successful-builds", 0, "Minimum number of successful builds required")
	cmd.Flags().StringVar(&mergeStrategy, "merge-strategy", "", "Default merge strategy (e.g. no-ff, squash, ff, ff-only, rebase)")
	cmd.Flags().StringVar(&allowedStrategies, "allowed-strategies", "", "Comma-separated list of allowed merge strategies")
	format.RegisterOutputFlags(cmd)
	return cmd
}
