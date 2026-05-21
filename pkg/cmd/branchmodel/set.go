package branchmodel

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdSet builds the `branch-model set` cobra command.
func NewCmdSet(f *factory.Factory) *cobra.Command {
	var (
		hostname        string
		devBranch       string
		prodBranch      string
		prodEnabled     bool
		branchTypePairs []string
	)
	cmd := &cobra.Command{
		Use:   "set [PROJECT/REPO]",
		Short: "Update branching model settings",
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
			bm, err := backend.AsBranchModelClient(client, ref.Host)
			if err != nil {
				return err
			}

			// Fetch current settings to merge branch type prefixes.
			current, err := bm.GetBranchModelSettings(ref.Project, ref.Slug)
			if err != nil {
				return err
			}

			in := backend.BranchModelSettingsInput{}

			if cmd.Flags().Changed("dev-branch") {
				in.Development = &backend.BranchModelSettingsBranch{
					Name:          devBranch,
					UseMainbranch: devBranch == "",
				}
			}

			if cmd.Flags().Changed("prod-branch") || cmd.Flags().Changed("prod-enabled") {
				// When --prod-branch is set but --prod-enabled is not explicitly
				// given, default IsValid to true — omitting it would silently
				// disable the production branch.
				isValid := prodEnabled
				if cmd.Flags().Changed("prod-branch") && !cmd.Flags().Changed("prod-enabled") {
					isValid = true
				}
				pb := backend.BranchModelSettingsBranch{
					Name:          prodBranch,
					UseMainbranch: prodBranch == "",
					IsValid:       isValid,
				}
				in.Production = &pb
			}

			if len(branchTypePairs) > 0 {
				overrides, parseErr := parseBranchTypePrefixes(branchTypePairs)
				if parseErr != nil {
					return parseErr
				}
				// Merge: start from current, apply overrides.
				merged := make([]backend.BranchTypeSettings, 0, len(current.BranchTypes))
				for _, bt := range current.BranchTypes {
					if prefix, ok := overrides[bt.Kind]; ok {
						bt.Prefix = prefix
						delete(overrides, bt.Kind)
					}
					merged = append(merged, bt)
				}
				// Any remaining overrides are new kinds.
				for kind, prefix := range overrides {
					merged = append(merged, backend.BranchTypeSettings{
						Enabled: true,
						Kind:    kind,
						Prefix:  prefix,
					})
				}
				in.BranchTypes = merged
			}

			updated, err := bm.UpdateBranchModelSettings(ref.Project, ref.Slug, in)
			if err != nil {
				return err
			}

			cfg := format.ConfigFromCmd(cmd)
			if cfg.Format != format.FormatTable {
				p := format.New[backend.BranchModelSettings](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
				p.SetSingleItem()
				p.AddField(format.Field[backend.BranchModelSettings]{
					Name:    "development",
					Header:  "DEVELOPMENT",
					Extract: func(s backend.BranchModelSettings) any { return s.Development.Name },
				})
				p.AddField(format.Field[backend.BranchModelSettings]{
					Name:    "production",
					Header:  "PRODUCTION",
					Extract: func(s backend.BranchModelSettings) any { return s.Production.Name },
				})
				p.AddField(format.Field[backend.BranchModelSettings]{
					Name:     "branch_types",
					Header:   "BRANCH_TYPES",
					JSONOnly: true,
					Extract:  func(s backend.BranchModelSettings) any { return s.BranchTypes },
				})
				p.AddItem(updated)
				return p.Render()
			}
			return printBranchModelSettings(f, updated)
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	cmd.Flags().StringVar(&devBranch, "dev-branch", "", "Development branch name")
	cmd.Flags().StringVar(&prodBranch, "prod-branch", "", "Production branch name")
	cmd.Flags().BoolVar(&prodEnabled, "prod-enabled", false, "Enable production branch")
	cmd.Flags().StringArrayVar(&branchTypePairs, "branch-type-prefix", nil, "Branch type prefix as kind=prefix (repeatable or comma-separated)")
	return cmd
}

// parseBranchTypePrefixes parses []string of "kind=prefix" or "k1=p1,k2=p2" into a map.
func parseBranchTypePrefixes(pairs []string) (map[string]string, error) {
	out := make(map[string]string)
	for _, raw := range pairs {
		// Each element may contain commas.
		parts := strings.Split(raw, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			idx := strings.IndexByte(part, '=')
			if idx <= 0 {
				return nil, fmt.Errorf("invalid --branch-type-prefix %q: expected kind=prefix", part)
			}
			kind := strings.TrimSpace(part[:idx])
			prefix := strings.TrimSpace(part[idx+1:])
			if kind == "" {
				return nil, fmt.Errorf("invalid --branch-type-prefix %q: kind is empty", part)
			}
			out[kind] = prefix
		}
	}
	return out, nil
}

func printBranchModelSettings(f *factory.Factory, s backend.BranchModelSettings) error {
	out := f.IOStreams.Out
	dev := s.Development
	fmt.Fprintf(out, "Development: %s", dev.Name)
	if dev.UseMainbranch {
		fmt.Fprintf(out, " (main branch)")
	}
	fmt.Fprintln(out)
	prod := s.Production
	fmt.Fprintf(out, "Production:  %s", prod.Name)
	if prod.UseMainbranch {
		fmt.Fprintf(out, " (main branch)")
	}
	fmt.Fprintln(out)
	if len(s.BranchTypes) > 0 {
		fmt.Fprintln(out, "Branch types:")
		for _, bt := range s.BranchTypes {
			enabled := "disabled"
			if bt.Enabled {
				enabled = "enabled"
			}
			fmt.Fprintf(out, "  %-12s %-10s %s\n", bt.Kind, enabled, bt.Prefix)
		}
	}
	return nil
}
