// Package compare implements `bitbottle branch compare BASE..HEAD`.
package compare

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmdutil"
)

// Options holds parsed flags for `branch compare`.
type Options struct {
	Output   format.OutputConfig
	Hostname string
	Limit    int
	Base     string
	Head     string
	Args     []string // remaining args after BASE..HEAD
}

// NewCmdCompare builds the `branch compare` cobra command.
func NewCmdCompare(f *factory.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "compare BASE..HEAD [PROJECT/REPO]",
		Short: "Compare two branches or commits",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.ValidatePositiveLimit(opts.Limit); err != nil {
				return err
			}
			parts := strings.SplitN(args[0], "..", 2)
			if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
				return &backend.DomainError{
					Kind:    backend.ErrInvalidRequest,
					Code:    backend.CodeInvalidRequest,
					Message: "argument must be in BASE..HEAD form",
				}
			}
			opts.Base = parts[0]
			opts.Head = parts[1]
			opts.Args = args[1:]
			opts.Output = format.ConfigFromCmd(cmd)
			if runF != nil {
				return runF(opts)
			}
			return compareRun(f, opts)
		},
	}
	cmd.Flags().IntVar(&opts.Limit, "limit", 30, "Maximum number of commits to show per side")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	format.RegisterOutputFlags(cmd)
	return cmd
}

func compareRun(f *factory.Factory, opts *Options) error {
	ref, err := factory.ResolveTarget(f, opts.Args, opts.Hostname)
	if err != nil {
		return err
	}
	client, err := f.Backend(ref.Host)
	if err != nil {
		return err
	}
	rc, err := backend.AsRefComparer(client, ref.Host)
	if err != nil {
		return err
	}
	cmp, err := rc.CompareRefs(ref.Project, ref.Slug, opts.Base, opts.Head, opts.Limit)
	if err != nil {
		return err
	}

	if opts.Output.Format != format.FormatTable {
		p := format.New[backend.RefComparison](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), opts.Output)
		p.AddField(format.Field[backend.RefComparison]{Name: "base", Header: "BASE", Extract: func(c backend.RefComparison) any { return c.Base }})
		p.AddField(format.Field[backend.RefComparison]{Name: "head", Header: "HEAD", Extract: func(c backend.RefComparison) any { return c.Head }})
		p.AddField(format.Field[backend.RefComparison]{Name: "ahead_by", Header: "AHEAD", Extract: func(c backend.RefComparison) any { return c.AheadBy }})
		p.AddField(format.Field[backend.RefComparison]{Name: "behind_by", Header: "BEHIND", Extract: func(c backend.RefComparison) any { return c.BehindBy }})
		p.SetSingleItem()
		p.AddItem(cmp)
		return p.Render()
	}

	out := f.IOStreams.Out
	fmt.Fprintf(out, "%s..%s\n", cmp.Base, cmp.Head)
	fmt.Fprintf(out, "Ahead by:  %d commit(s)\n", cmp.AheadBy)
	fmt.Fprintf(out, "Behind by: %d commit(s)\n", cmp.BehindBy)
	if len(cmp.CommitsAhead) > 0 {
		fmt.Fprintf(out, "\nCommits ahead:\n")
		for _, c := range cmp.CommitsAhead {
			fmt.Fprintf(out, "  %s %s\n", c.Hash[:min(len(c.Hash), 7)], c.Message)
		}
	}
	if len(cmp.CommitsBehind) > 0 {
		fmt.Fprintf(out, "\nCommits behind:\n")
		for _, c := range cmp.CommitsBehind {
			fmt.Fprintf(out, "  %s %s\n", c.Hash[:min(len(c.Hash), 7)], c.Message)
		}
	}
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
