package group

import (
	"fmt"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// resolveHostname returns the hostname to use for group commands. If flag is
// non-empty it takes precedence. Otherwise the single configured host is used.
func resolveHostname(f *factory.Factory, flag string) (string, error) {
	if flag != "" {
		return flag, nil
	}
	cfg, err := f.Config()
	if err != nil {
		return "", err
	}
	hosts := cfg.Hosts()
	switch len(hosts) {
	case 0:
		return "", fmt.Errorf("not authenticated; run `bitbottle auth login` first")
	case 1:
		return hosts[0], nil
	default:
		return "", fmt.Errorf("multiple hosts configured; use --hostname to specify one")
	}
}
