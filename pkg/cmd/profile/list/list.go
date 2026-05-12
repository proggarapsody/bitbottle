// Package list implements `profile list`.
package list

import (
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/internal/profiles"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// Options holds parsed flags for `profile list`.
type Options struct {
	Output format.OutputConfig
}

// NewCmdList builds the `profile list` cobra command.
func NewCmdList(f *factory.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List named credential profiles",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Output = format.ConfigFromCmd(cmd)
			if runF != nil {
				return runF(opts)
			}
			return listRun(f, opts)
		},
	}
	return cmd
}

func listRun(f *factory.Factory, opts *Options) error {
	store, err := f.Profiles()
	if err != nil {
		return err
	}
	all := store.All()

	p := profileListFields(f, opts.Output)
	for _, np := range all {
		p.AddItem(np)
	}
	return p.Render()
}

func profileListFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[profiles.NamedProfile] {
	isTTY := f.IOStreams.IsStdoutTTY()
	p := format.New[profiles.NamedProfile](f.IOStreams.Out, isTTY, cfg)

	p.AddField(format.Field[profiles.NamedProfile]{
		Name:    "name",
		Header:  "NAME",
		Extract: func(np profiles.NamedProfile) any { return np.Name },
	})

	p.AddField(format.Field[profiles.NamedProfile]{
		Name:    "hostname",
		Header:  "HOSTNAME",
		Extract: func(np profiles.NamedProfile) any { return np.Hostname },
	})

	p.AddField(format.Field[profiles.NamedProfile]{
		Name:    "user",
		Header:  "USER",
		Extract: func(np profiles.NamedProfile) any { return np.User },
	})

	p.AddField(format.Field[profiles.NamedProfile]{
		Name:     "backend_type",
		Header:   "BACKEND",
		JSONOnly: true,
		Extract:  func(np profiles.NamedProfile) any { return np.BackendType },
	})

	p.AddField(format.Field[profiles.NamedProfile]{
		Name:     "skip_tls_verify",
		Header:   "SKIP_TLS",
		JSONOnly: true,
		Extract:  func(np profiles.NamedProfile) any { return np.SkipTLSVerify },
	})

	return p
}
