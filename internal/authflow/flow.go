package authflow

import (
	"github.com/proggarapsody/bitbottle/api"
	"github.com/proggarapsody/bitbottle/internal/config"
	"github.com/proggarapsody/bitbottle/internal/keyring"
	"github.com/proggarapsody/bitbottle/pkg/iostreams"
)

// Flow orchestrates the auth login sequence.
type Flow struct {
	ios     *iostreams.IOStreams
	cfg     *config.Config
	keyring keyring.Keyring
	client  func(hostname string, skipTLS bool) *api.Client
}

func New(ios *iostreams.IOStreams, cfg *config.Config, kr keyring.Keyring, client func(string, bool) *api.Client) *Flow {
	return &Flow{ios: ios, cfg: cfg, keyring: kr, client: client}
}

// Login runs the interactive login flow for hostname.
func (f *Flow) Login(hostname, token, gitProtocol string, skipTLS bool) error {
	panic("not implemented")
}
