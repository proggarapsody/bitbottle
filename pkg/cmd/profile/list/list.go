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
	JSONFields string
	JQExpr     string
}

// NewCmdList builds the `profile list` cobra command.
func NewCmdList(f *factory.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List named credential profiles",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if runF != nil {
				return runF(opts)
			}
			return listRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.JSONFields, "json", "", "Output JSON with specified fields (comma-separated)")
	cmd.Flags().StringVar(&opts.JQExpr, "jq", "", "Filter JSON output with a jq expression")
	return cmd
}

func listRun(f *factory.Factory, opts *Options) error {
	store, err := f.Profiles()
	if err != nil {
		return err
	}
	all := store.All()

	p := profileListFields(f, opts.JSONFields, opts.JQExpr)
	for _, np := range all {
		p.AddItem(np)
	}
	return p.Render()
}

func profileListFields(f *factory.Factory, jsonFields, jqExpr string) *format.Printer[profiles.NamedProfile] {
	isTTY := f.IOStreams.IsStdoutTTY()
	p := format.New[profiles.NamedProfile](f.IOStreams.Out, isTTY, jsonFields, jqExpr)

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
