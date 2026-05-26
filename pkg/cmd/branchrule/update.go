package branchrule

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdUpdate builds the `branch-rule update` cobra command.
func NewCmdUpdate(f *factory.Factory) *cobra.Command {
	var pattern, users, groups, hostname string
	var value int
	var valueSet bool
	cmd := &cobra.Command{
		Use:   "update [PROJECT/REPO] ID",
		Short: "Update a branch restriction rule in a repository",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var repoArgs []string
			var idArg string
			if len(args) == 2 {
				repoArgs = args[:1]
				idArg = args[1]
			} else {
				// single arg must be the ID; repo inferred from git remote
				repoArgs = nil
				idArg = args[0]
			}
			id, err := strconv.Atoi(idArg)
			if err != nil || id <= 0 {
				return fmt.Errorf("ID must be a positive integer, got %q", idArg)
			}

			// At least one flag must be set
			patternChanged := cmd.Flags().Changed("pattern")
			usersChanged := cmd.Flags().Changed("users")
			groupsChanged := cmd.Flags().Changed("groups")
			valueSet = cmd.Flags().Changed("value")
			if !patternChanged && !usersChanged && !groupsChanged && !valueSet {
				return fmt.Errorf("at least one of --pattern, --users, --groups, --value must be provided")
			}

			ref, err := factory.ResolveTarget(f, repoArgs, hostname)
			if err != nil {
				return err
			}
			client, err := f.Backend(ref.Host)
			if err != nil {
				return err
			}
			br, err := backend.AsBranchRuleClient(client, ref.Host)
			if err != nil {
				return err
			}

			var in backend.UpdateBranchRuleInput
			if patternChanged {
				in.Pattern = &pattern
			}
			if valueSet {
				in.Value = &value
			}
			if usersChanged {
				var us []string
				if users != "" {
					us = strings.Split(users, ",")
				}
				in.Users = &us
			}
			if groupsChanged {
				var gs []string
				if groups != "" {
					gs = strings.Split(groups, ",")
				}
				in.Groups = &gs
			}

			updated, err := br.UpdateBranchRule(ref.Project, ref.Slug, id, in)
			if err != nil {
				return err
			}
			p := branchRuleFields(f, format.ConfigFromCmd(cmd))
			p.AddItem(updated)
			return p.Render()
		},
	}
	cmd.Flags().StringVar(&pattern, "pattern", "", "New branch pattern")
	cmd.Flags().StringVar(&users, "users", "", "Comma-separated user slugs (replaces existing)")
	cmd.Flags().StringVar(&groups, "groups", "", "Comma-separated group slugs (replaces existing)")
	cmd.Flags().IntVar(&value, "value", 0, "Numeric value (e.g. required approvers count)")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	return cmd
}
