package pr

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/bbrepo"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdPR(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pr",
		Short: "Manage pull requests",
		Annotations: map[string]string{
			"help:arguments": `A repository can be supplied as PROJECT/REPO. When omitted, the
repository is inferred from the "origin" git remote in the current
directory. Use --hostname to disambiguate when multiple Bitbucket
hosts are configured.`,
		},
	}
	factory.EnableRepoOverride(cmd, f)
	cmd.AddCommand(NewCmdPRList(f))
	cmd.AddCommand(NewCmdPRView(f))
	cmd.AddCommand(NewCmdPRCreate(f))
	cmd.AddCommand(NewCmdPRMerge(f))
	cmd.AddCommand(NewCmdPRApprove(f))
	cmd.AddCommand(NewCmdPRDiff(f))
	cmd.AddCommand(NewCmdPRCheckout(f))
	cmd.AddCommand(NewCmdPREdit(f))
	cmd.AddCommand(NewCmdPRDecline(f))
	cmd.AddCommand(NewCmdPRUnapprove(f))
	cmd.AddCommand(NewCmdPRReady(f))
	cmd.AddCommand(NewCmdPRRequestReview(f))
	cmd.AddCommand(NewCmdPRRequestChanges(f))
	cmd.AddCommand(NewCmdPRComment(f))
	return cmd
}

func parsePRID(arg string) (int, error) {
	id, err := strconv.Atoi(arg)
	if err != nil {
		return 0, fmt.Errorf("invalid PR ID %q: must be a positive integer", arg)
	}
	return id, nil
}

func resolvePRTarget(f *factory.Factory, args []string, hostnameFlag string) (bbrepo.RepoRef, int, backend.Client, error) {
	prID, err := parsePRID(args[0])
	if err != nil {
		return bbrepo.RepoRef{}, 0, nil, err
	}

	ref, err := resolveRepoRef(f, []string{}, hostnameFlag)
	if err != nil {
		return bbrepo.RepoRef{}, 0, nil, err
	}
	ref.Project = strings.ToUpper(ref.Project)

	client, err := f.Backend(ref.Host)
	if err != nil {
		return bbrepo.RepoRef{}, 0, nil, err
	}
	return ref, prID, client, nil
}
