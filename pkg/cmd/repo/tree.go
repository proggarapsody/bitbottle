package repo

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/bbrepo"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdRepoTree builds `repo tree PROJECT/REPO [PATH] --ref REF`.
// PATH defaults to "" (the repository root). Output is a TTY-aware table
// or JSON via --json/--jq.
func NewCmdRepoTree(f *factory.Factory) *cobra.Command {
	var ref string
	var hostname string

	cmd := &cobra.Command{
		Use:   "tree PROJECT/REPO [PATH]",
		Short: "List files at a ref",
		Long: "List the immediate children of a directory at the given ref.\n" +
			"With no PATH, lists the repository root.\n\n" +
			"Type is normalised to 'file' or 'dir' across both backends —\n" +
			"submodules surface as 'dir' so renderers can recurse into them\n" +
			"and the submodule pointer is exposed via the hash field.\n\n" +
			"Examples:\n" +
			"  bitbottle repo tree myws/my-svc --ref main\n" +
			"  bitbottle repo tree myws/my-svc cmd --ref main\n" +
			"  bitbottle repo tree myws/my-svc --ref v1.0.0 --json path,type",
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if ref == "" {
				return fmt.Errorf("--ref is required")
			}
			refParsed, err := bbrepo.Parse(args[0])
			if err != nil {
				return err
			}
			path := ""
			if len(args) == 2 {
				path = args[1]
			}
			host, err := resolveHostname(f, hostname)
			if err != nil {
				return err
			}
			client, err := f.Backend(host)
			if err != nil {
				return err
			}
			entries, err := client.ListTree(refParsed.Project, refParsed.Slug, ref, path)
			if err != nil {
				return err
			}
			p := treeFields(f, format.ConfigFromCmd(cmd))
			for _, e := range entries {
				p.AddItem(e)
			}
			return p.Render()
		},
	}
	cmd.Flags().StringVar(&ref, "ref", "", "Branch, tag, or commit hash to read from (required)")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func treeFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.TreeEntry] {
	p := format.New[backend.TreeEntry](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
	p.AddField(format.Field[backend.TreeEntry]{Name: "type", Header: "TYPE", Extract: func(e backend.TreeEntry) any { return e.Type }})
	p.AddField(format.Field[backend.TreeEntry]{Name: "path", Header: "PATH", Extract: func(e backend.TreeEntry) any { return e.Path }})
	p.AddField(format.Field[backend.TreeEntry]{Name: "size", Header: "SIZE", Extract: func(e backend.TreeEntry) any { return e.Size }})
	p.AddField(format.Field[backend.TreeEntry]{Name: "hash", Header: "HASH", Extract: func(e backend.TreeEntry) any { return e.Hash }})
	return p
}
