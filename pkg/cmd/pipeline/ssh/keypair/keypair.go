// Package keypair implements `bitbottle pipeline ssh key-pair` subcommands.
package keypair

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdKeyPair builds the `pipeline ssh key-pair` parent command.
func NewCmdKeyPair(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "key-pair",
		Short: "Manage the pipeline SSH key pair (Cloud only)",
		Annotations: map[string]string{
			"help:arguments": `A repository can be supplied as WORKSPACE/REPO. When omitted, the
repository is inferred from the "origin" git remote in the current
directory.`,
		},
	}
	factory.EnableRepoOverride(cmd, f)
	cmd.AddCommand(NewCmdView(f, nil))
	cmd.AddCommand(NewCmdRegenerate(f, nil))
	return cmd
}

// ── view ──────────────────────────────────────────────────────────────────────

// ViewOptions holds parsed flags for `pipeline ssh key-pair view`.
type ViewOptions struct {
	Output   format.OutputConfig
	Hostname string
	Args     []string
}

// NewCmdView builds the `pipeline ssh key-pair view` cobra command.
func NewCmdView(f *factory.Factory, runF func(*ViewOptions) error) *cobra.Command {
	opts := &ViewOptions{}
	cmd := &cobra.Command{
		Use:   "view [PROJECT/REPO]",
		Short: "View the pipeline SSH key pair",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args
			opts.Output = format.ConfigFromCmd(cmd)
			if runF != nil {
				return runF(opts)
			}
			return runView(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	format.RegisterOutputFlags(cmd)
	return cmd
}

func runView(f *factory.Factory, opts *ViewOptions) error {
	ref, err := factory.ResolveTarget(f, opts.Args, opts.Hostname)
	if err != nil {
		return err
	}
	client, err := f.Backend(ref.Host)
	if err != nil {
		return err
	}
	sc, err := backend.AsPipelineSSHKeyPairClient(client, ref.Host)
	if err != nil {
		return err
	}
	kp, err := sc.GetPipelineSSHKeyPair(ref.Project, ref.Slug)
	if err != nil {
		return err
	}

	if opts.Output.Format != format.FormatTable {
		p := format.New[backend.PipelineSSHKeyPair](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), opts.Output)
		p.AddField(format.Field[backend.PipelineSSHKeyPair]{Name: "public_key", Header: "PUBLIC_KEY", Extract: func(k backend.PipelineSSHKeyPair) any { return k.PublicKey }})
		p.AddField(format.Field[backend.PipelineSSHKeyPair]{Name: "key_type", Header: "KEY_TYPE", Extract: func(k backend.PipelineSSHKeyPair) any { return k.KeyTypeLabel }})
		p.AddField(format.Field[backend.PipelineSSHKeyPair]{Name: "created", Header: "CREATED", Extract: func(k backend.PipelineSSHKeyPair) any { return k.Created.Format(time.RFC3339) }})
		p.SetSingleItem()
		p.AddItem(kp)
		return p.Render()
	}

	out := f.IOStreams.Out
	fmt.Fprintf(out, "Key type: %s\n", kp.KeyTypeLabel)
	fmt.Fprintf(out, "Created:  %s\n", kp.Created.Format(time.RFC3339))
	fmt.Fprintf(out, "Public key:\n%s\n", kp.PublicKey)
	return nil
}

// ── regenerate ────────────────────────────────────────────────────────────────

// RegenerateOptions holds parsed flags for `pipeline ssh key-pair regenerate`.
type RegenerateOptions struct {
	Output   format.OutputConfig
	Hostname string
	Bits     int
	Confirm  bool
	Args     []string
}

// NewCmdRegenerate builds the `pipeline ssh key-pair regenerate` cobra command.
func NewCmdRegenerate(f *factory.Factory, runF func(*RegenerateOptions) error) *cobra.Command {
	opts := &RegenerateOptions{}
	cmd := &cobra.Command{
		Use:   "regenerate [PROJECT/REPO]",
		Short: "Regenerate the pipeline SSH key pair",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args
			opts.Output = format.ConfigFromCmd(cmd)
			if runF != nil {
				return runF(opts)
			}
			return runRegenerate(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	cmd.Flags().IntVar(&opts.Bits, "bits", 0, "Key size in bits (2048 or 4096; default 2048)")
	cmd.Flags().BoolVar(&opts.Confirm, "confirm", false, "Confirm regeneration without a TTY prompt")
	format.RegisterOutputFlags(cmd)
	return cmd
}

func runRegenerate(f *factory.Factory, opts *RegenerateOptions) error {
	if !f.IOStreams.IsStdoutTTY() && !opts.Confirm {
		return fmt.Errorf("pass --confirm to regenerate the pipeline SSH key pair")
	}

	ref, err := factory.ResolveTarget(f, opts.Args, opts.Hostname)
	if err != nil {
		return err
	}
	client, err := f.Backend(ref.Host)
	if err != nil {
		return err
	}
	sc, err := backend.AsPipelineSSHKeyPairClient(client, ref.Host)
	if err != nil {
		return err
	}
	kp, err := sc.RegeneratePipelineSSHKeyPair(ref.Project, ref.Slug, opts.Bits)
	if err != nil {
		return err
	}

	if opts.Output.Format != format.FormatTable {
		p := format.New[backend.PipelineSSHKeyPair](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), opts.Output)
		p.AddField(format.Field[backend.PipelineSSHKeyPair]{Name: "public_key", Header: "PUBLIC_KEY", Extract: func(k backend.PipelineSSHKeyPair) any { return k.PublicKey }})
		p.AddField(format.Field[backend.PipelineSSHKeyPair]{Name: "key_type", Header: "KEY_TYPE", Extract: func(k backend.PipelineSSHKeyPair) any { return k.KeyTypeLabel }})
		p.AddField(format.Field[backend.PipelineSSHKeyPair]{Name: "created", Header: "CREATED", Extract: func(k backend.PipelineSSHKeyPair) any { return k.Created.Format(time.RFC3339) }})
		p.SetSingleItem()
		p.AddItem(kp)
		return p.Render()
	}

	out := f.IOStreams.Out
	fmt.Fprintf(out, "Regenerated SSH key pair (%s)\n", kp.KeyTypeLabel)
	fmt.Fprintf(out, "Public key:\n%s\n", kp.PublicKey)
	return nil
}
