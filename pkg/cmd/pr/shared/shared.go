// Package shared holds helpers shared by pr sub-command subpackages.
package shared

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/bbrepo"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// ParsePRID parses a string argument as a positive integer PR ID.
func ParsePRID(arg string) (int, error) {
	id, err := strconv.Atoi(arg)
	if err != nil {
		return 0, fmt.Errorf("invalid PR ID %q: must be a positive integer", arg)
	}
	return id, nil
}

// ResolvePRTarget resolves the repository target and parses the PR ID from
// args[0]. hostnameFlag may be empty to use auto-detection.
func ResolvePRTarget(f *factory.Factory, args []string, hostnameFlag string) (bbrepo.RepoRef, int, backend.Client, error) {
	prID, err := ParsePRID(args[0])
	if err != nil {
		return bbrepo.RepoRef{}, 0, nil, err
	}

	ref, err := factory.ResolveTarget(f, nil, hostnameFlag)
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
