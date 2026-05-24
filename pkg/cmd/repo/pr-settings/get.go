package prsettings

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdGet builds `repo pr-settings get [PROJECT/REPO] [--json]`.
func NewCmdGet(f *factory.Factory) *cobra.Command {
	var hostname string

	cmd := &cobra.Command{
		Use:   "get [PROJECT/REPO]",
		Short: "Show pull request gate settings for a repository",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref, err := factory.ResolveTarget(f, args, hostname)
			if err != nil {
				return err
			}

			client, err := f.Backend(ref.Host)
			if err != nil {
				return err
			}

			prc, err := backend.AsRepoPRSettingsClient(client, ref.Host)
			if err != nil {
				return err
			}

			settings, err := prc.GetRepoPRSettings(ref.Project, ref.Slug)
			if err != nil {
				return err
			}

			p := prSettingsFields(f, format.ConfigFromCmd(cmd))
			p.SetSingleItem()
			p.AddItem(settings)
			return p.Render()
		},
	}

	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	format.RegisterOutputFlags(cmd)
	return cmd
}

func prSettingsFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.RepoPRSettings] {
	p := format.New[backend.RepoPRSettings](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
	p.AddField(format.Field[backend.RepoPRSettings]{
		Name:    "requiredApprovers",
		Header:  "APPROVERS",
		Extract: func(s backend.RepoPRSettings) any { return s.RequiredApprovers },
	})
	p.AddField(format.Field[backend.RepoPRSettings]{
		Name:    "requiredAllApprovers",
		Header:  "ALL_APPROVERS",
		Extract: func(s backend.RepoPRSettings) any { return s.RequiredAllApprovers },
	})
	p.AddField(format.Field[backend.RepoPRSettings]{
		Name:    "requiredAllTasksComplete",
		Header:  "ALL_TASKS",
		Extract: func(s backend.RepoPRSettings) any { return s.RequiredAllTasksComplete },
	})
	p.AddField(format.Field[backend.RepoPRSettings]{
		Name:    "requiredSuccessfulBuilds",
		Header:  "BUILDS",
		Extract: func(s backend.RepoPRSettings) any { return s.RequiredSuccessfulBuilds },
	})
	p.AddField(format.Field[backend.RepoPRSettings]{
		Name:    "mergeStrategy",
		Header:  "MERGE_STRATEGY",
		Extract: func(s backend.RepoPRSettings) any { return s.MergeStrategy },
	})
	p.AddField(format.Field[backend.RepoPRSettings]{
		Name:    "allowedStrategies",
		Header:  "ALLOWED_STRATEGIES",
		Extract: func(s backend.RepoPRSettings) any { return strings.Join(s.AllowedStrategies, ",") },
	})
	return p
}
