package pipeline

import (
	"fmt"

	"github.com/proggarapsody/bitbottle/internal/bbrepo"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func resolvePipelineRef(f *factory.Factory, arg, hostnameFlag string) (bbrepo.RepoRef, error) {
	ref, err := bbrepo.Parse(arg)
	if err != nil {
		return bbrepo.RepoRef{}, err
	}
	if hostnameFlag != "" {
		ref.Host = hostnameFlag
		return ref, nil
	}
	if ref.Host == "" {
		cfg, cerr := f.Config()
		if cerr != nil {
			return bbrepo.RepoRef{}, cerr
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
