package factory

import (
	"fmt"

	"github.com/proggarapsody/bitbottle/internal/bbrepo"
)

// ResolveRef parses a PROJECT/REPO argument and resolves the Bitbucket host.
// hostnameFlag takes precedence when set; otherwise the host is inferred from
// the argument itself or from the single configured host.
func (f *Factory) ResolveRef(arg, hostnameFlag string) (bbrepo.RepoRef, error) {
	ref, err := bbrepo.Parse(arg)
	if err != nil {
		return bbrepo.RepoRef{}, err
	}
	if hostnameFlag != "" {
		ref.Host = hostnameFlag
		return ref, nil
	}
	if ref.Host == "" {
		cfg, err := f.Config()
		if err != nil {
			return bbrepo.RepoRef{}, err
		}
		hosts := cfg.Hosts()
		switch len(hosts) {
		case 0:
			return bbrepo.RepoRef{}, fmt.Errorf("not authenticated; run `bitbottle auth login` first")
		case 1:
			ref.Host = hosts[0]
		default:
			return bbrepo.RepoRef{}, fmt.Errorf("multiple hosts configured; use --hostname to specify one")
		}
	}
	return ref, nil
}
