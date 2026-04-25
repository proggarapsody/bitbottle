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
	}
	cmd.AddCommand(NewCmdPRList(f))
	cmd.AddCommand(NewCmdPRView(f))
	cmd.AddCommand(NewCmdPRCreate(f))
	cmd.AddCommand(NewCmdPRMerge(f))
	cmd.AddCommand(NewCmdPRApprove(f))
	cmd.AddCommand(NewCmdPRDiff(f))
	cmd.AddCommand(NewCmdPRCheckout(f))
	return cmd
}

// parsePRID parses a positional PR ID argument and returns a friendly error
// when the value is not a positive integer.
func parsePRID(arg string) (int, error) {
	id, err := strconv.Atoi(arg)
	if err != nil {
		return 0, fmt.Errorf("invalid PR ID %q: must be a positive integer", arg)
	}
	return id, nil
}

// resolvePRTarget is the common entry point for sub-commands that operate on a
// single existing pull request (`view`, `diff`, `approve`, `checkout`,
// `merge`). It parses the PR ID, resolves the repository reference from the
// current git remote (uppercasing the Bitbucket Data Center project key), and
// returns a backend client bound to the resolved host.
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
