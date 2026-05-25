// Package knownhosts implements `bitbottle pipeline ssh known-hosts` subcommands.
package knownhosts

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmdutil"
)

// NewCmdKnownHosts builds the `pipeline ssh known-hosts` parent command.
func NewCmdKnownHosts(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "known-hosts",
		Short: "Manage pipeline known hosts (Cloud only)",
		Annotations: map[string]string{
			"help:arguments": `A repository can be supplied as WORKSPACE/REPO. When omitted, the
repository is inferred from the "origin" git remote in the current
directory.`,
		},
	}
	factory.EnableRepoOverride(cmd, f)
	cmd.AddCommand(NewCmdList(f, nil))
	cmd.AddCommand(NewCmdKHView(f, nil))
	cmd.AddCommand(NewCmdAdd(f, nil))
	cmd.AddCommand(NewCmdDelete(f, nil))
	return cmd
}

// ── list ──────────────────────────────────────────────────────────────────────

// ListOptions holds parsed flags for `pipeline ssh known-hosts list`.
type ListOptions struct {
	Hostname string
	Args     []string
}

// NewCmdList builds the `pipeline ssh known-hosts list` cobra command.
func NewCmdList(f *factory.Factory, runF func(*ListOptions) error) *cobra.Command {
	opts := &ListOptions{}
	cmd := &cobra.Command{
		Use:   "list [PROJECT/REPO]",
		Short: "List pipeline known hosts",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args
			if runF != nil {
				return runF(opts)
			}
			return runList(f, cmd, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	format.RegisterOutputFlags(cmd)
	return cmd
}

func runList(f *factory.Factory, cmd *cobra.Command, opts *ListOptions) error {
	ref, err := factory.ResolveTarget(f, opts.Args, opts.Hostname)
	if err != nil {
		return err
	}
	client, err := f.Backend(ref.Host)
	if err != nil {
		return err
	}
	kc, err := backend.AsPipelineKnownHostsClient(client, ref.Host)
	if err != nil {
		return err
	}
	hosts, listErr := kc.ListPipelineKnownHosts(ref.Project, ref.Slug)
	if listErr != nil && len(hosts) == 0 {
		return listErr
	}

	cfg := format.ConfigFromCmd(cmd)
	if cfg.Format != format.FormatTable {
		p := format.New[backend.PipelineKnownHost](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
		p.AddField(format.Field[backend.PipelineKnownHost]{Name: "uuid", Header: "UUID", Extract: func(h backend.PipelineKnownHost) any { return h.UUID }})
		p.AddField(format.Field[backend.PipelineKnownHost]{Name: "hostname", Header: "HOSTNAME", Extract: func(h backend.PipelineKnownHost) any { return h.Hostname }})
		p.AddField(format.Field[backend.PipelineKnownHost]{Name: "key_type", Header: "KEY_TYPE", Extract: func(h backend.PipelineKnownHost) any { return h.PublicKey.KeyType }})
		p.AddField(format.Field[backend.PipelineKnownHost]{Name: "md5_fingerprint", Header: "MD5", Extract: func(h backend.PipelineKnownHost) any { return h.PublicKey.MD5 }})
		for _, h := range hosts {
			p.AddItem(h)
		}
		if err := p.Render(); err != nil {
			return err
		}
		cmdutil.PartialWarn(f.IOStreams.ErrOut, len(hosts), listErr)
		return listErr
	}

	out := f.IOStreams.Out
	if len(hosts) == 0 {
		fmt.Fprintln(out, "No pipeline known hosts found.")
		return nil
	}
	fmt.Fprintf(out, "%-38s  %-30s  %-10s  %s\n", "UUID", "HOSTNAME", "KEY_TYPE", "MD5")
	for _, h := range hosts {
		fmt.Fprintf(out, "%-38s  %-30s  %-10s  %s\n", h.UUID, h.Hostname, h.PublicKey.KeyType, h.PublicKey.MD5)
	}
	cmdutil.PartialWarn(f.IOStreams.ErrOut, len(hosts), listErr)
	return listErr
}

// ── view ──────────────────────────────────────────────────────────────────────

// KHViewOptions holds parsed flags for `pipeline ssh known-hosts view`.
type KHViewOptions struct {
	Output   format.OutputConfig
	Hostname string
	UUID     string
	Args     []string
}

// NewCmdKHView builds the `pipeline ssh known-hosts view` cobra command.
func NewCmdKHView(f *factory.Factory, runF func(*KHViewOptions) error) *cobra.Command {
	opts := &KHViewOptions{}
	cmd := &cobra.Command{
		Use:   "view UUID [PROJECT/REPO]",
		Short: "View a pipeline known host",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.UUID = args[0]
			opts.Args = args[1:]
			opts.Output = format.ConfigFromCmd(cmd)
			if runF != nil {
				return runF(opts)
			}
			return runKHView(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	format.RegisterOutputFlags(cmd)
	return cmd
}

func runKHView(f *factory.Factory, opts *KHViewOptions) error {
	ref, err := factory.ResolveTarget(f, opts.Args, opts.Hostname)
	if err != nil {
		return err
	}
	client, err := f.Backend(ref.Host)
	if err != nil {
		return err
	}
	kc, err := backend.AsPipelineKnownHostsClient(client, ref.Host)
	if err != nil {
		return err
	}
	host, err := kc.GetPipelineKnownHost(ref.Project, ref.Slug, opts.UUID)
	if err != nil {
		return err
	}

	if opts.Output.Format != format.FormatTable {
		p := format.New[backend.PipelineKnownHost](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), opts.Output)
		p.AddField(format.Field[backend.PipelineKnownHost]{Name: "uuid", Header: "UUID", Extract: func(h backend.PipelineKnownHost) any { return h.UUID }})
		p.AddField(format.Field[backend.PipelineKnownHost]{Name: "hostname", Header: "HOSTNAME", Extract: func(h backend.PipelineKnownHost) any { return h.Hostname }})
		p.AddField(format.Field[backend.PipelineKnownHost]{Name: "key_type", Header: "KEY_TYPE", Extract: func(h backend.PipelineKnownHost) any { return h.PublicKey.KeyType }})
		p.AddField(format.Field[backend.PipelineKnownHost]{Name: "key", Header: "KEY", Extract: func(h backend.PipelineKnownHost) any { return h.PublicKey.Key }})
		p.AddField(format.Field[backend.PipelineKnownHost]{Name: "md5_fingerprint", Header: "MD5", Extract: func(h backend.PipelineKnownHost) any { return h.PublicKey.MD5 }})
		p.AddField(format.Field[backend.PipelineKnownHost]{Name: "sha256_fingerprint", Header: "SHA256", Extract: func(h backend.PipelineKnownHost) any { return h.PublicKey.SHA256 }})
		p.SetSingleItem()
		p.AddItem(host)
		return p.Render()
	}

	out := f.IOStreams.Out
	fmt.Fprintf(out, "UUID:     %s\n", host.UUID)
	fmt.Fprintf(out, "Hostname: %s\n", host.Hostname)
	fmt.Fprintf(out, "Key type: %s\n", host.PublicKey.KeyType)
	fmt.Fprintf(out, "MD5:      %s\n", host.PublicKey.MD5)
	fmt.Fprintf(out, "SHA256:   %s\n", host.PublicKey.SHA256)
	return nil
}

// ── add ───────────────────────────────────────────────────────────────────────

// AddOptions holds parsed flags for `pipeline ssh known-hosts add`.
type AddOptions struct {
	Output   format.OutputConfig
	Hostname string
	KeyData  string
	KeyType  string
	Args     []string // Args[0] = HOSTNAME_ARG; Args[1] = optional PROJECT/REPO
}

// NewCmdAdd builds the `pipeline ssh known-hosts add` cobra command.
func NewCmdAdd(f *factory.Factory, runF func(*AddOptions) error) *cobra.Command {
	opts := &AddOptions{}
	cmd := &cobra.Command{
		Use:   "add HOSTNAME [PROJECT/REPO]",
		Short: "Add a known host to pipeline SSH config",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args
			opts.Output = format.ConfigFromCmd(cmd)
			if runF != nil {
				return runF(opts)
			}
			return runAdd(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	cmd.Flags().StringVar(&opts.KeyData, "key", "", "Base64-encoded public key material")
	cmd.Flags().StringVar(&opts.KeyType, "key-type", "RSA", "Key type: rsa, ecdsa, or ed25519")
	format.RegisterOutputFlags(cmd)
	return cmd
}

func runAdd(f *factory.Factory, opts *AddOptions) error {
	hostnameArg := opts.Args[0]
	repoArgs := opts.Args[1:]

	ref, err := factory.ResolveTarget(f, repoArgs, opts.Hostname)
	if err != nil {
		return err
	}
	client, err := f.Backend(ref.Host)
	if err != nil {
		return err
	}
	kc, err := backend.AsPipelineKnownHostsClient(client, ref.Host)
	if err != nil {
		return err
	}
	in := backend.PipelineKnownHostInput{
		Hostname: hostnameArg,
		PublicKey: backend.PipelineSSHPublicKey{
			KeyType: opts.KeyType,
			Key:     opts.KeyData,
		},
	}
	host, err := kc.AddPipelineKnownHost(ref.Project, ref.Slug, in)
	if err != nil {
		return err
	}

	if opts.Output.Format != format.FormatTable {
		p := format.New[backend.PipelineKnownHost](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), opts.Output)
		p.AddField(format.Field[backend.PipelineKnownHost]{Name: "uuid", Header: "UUID", Extract: func(h backend.PipelineKnownHost) any { return h.UUID }})
		p.AddField(format.Field[backend.PipelineKnownHost]{Name: "hostname", Header: "HOSTNAME", Extract: func(h backend.PipelineKnownHost) any { return h.Hostname }})
		p.AddField(format.Field[backend.PipelineKnownHost]{Name: "key_type", Header: "KEY_TYPE", Extract: func(h backend.PipelineKnownHost) any { return h.PublicKey.KeyType }})
		p.SetSingleItem()
		p.AddItem(host)
		return p.Render()
	}

	fmt.Fprintf(f.IOStreams.Out, "Added known host %s (UUID: %s)\n", host.Hostname, host.UUID)
	return nil
}

// ── delete ────────────────────────────────────────────────────────────────────

// DeleteOptions holds parsed flags for `pipeline ssh known-hosts delete`.
type DeleteOptions struct {
	Hostname string
	Confirm  bool
	UUID     string
	Args     []string
}

// NewCmdDelete builds the `pipeline ssh known-hosts delete` cobra command.
func NewCmdDelete(f *factory.Factory, runF func(*DeleteOptions) error) *cobra.Command {
	opts := &DeleteOptions{}
	cmd := &cobra.Command{
		Use:   "delete UUID [PROJECT/REPO]",
		Short: "Delete a pipeline known host",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.UUID = args[0]
			opts.Args = args[1:]
			if runF != nil {
				return runF(opts)
			}
			return runDelete(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	cmd.Flags().BoolVar(&opts.Confirm, "confirm", false, "Confirm deletion without a TTY prompt")
	return cmd
}

func runDelete(f *factory.Factory, opts *DeleteOptions) error {
	if !f.IOStreams.IsStdoutTTY() && !opts.Confirm {
		return fmt.Errorf("pass --confirm to delete a pipeline known host")
	}

	ref, err := factory.ResolveTarget(f, opts.Args, opts.Hostname)
	if err != nil {
		return err
	}
	client, err := f.Backend(ref.Host)
	if err != nil {
		return err
	}
	kc, err := backend.AsPipelineKnownHostsClient(client, ref.Host)
	if err != nil {
		return err
	}
	if err := kc.DeletePipelineKnownHost(ref.Project, ref.Slug, opts.UUID); err != nil {
		return err
	}
	fmt.Fprintf(f.IOStreams.Out, "Deleted known host %s\n", opts.UUID)
	return nil
}
