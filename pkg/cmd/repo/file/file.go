package file

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/internal/bbrepo"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdFile is the parent of `repo file <action>`. It exists today
// only to host `get`, but the noun is established under `repo` so future
// file-scoped operations (compare, history, blame) attach naturally.
func NewCmdFile(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "file",
		Short: "Read repository files at a ref",
	}
	cmd.AddCommand(NewCmdFileGet(f))
	return cmd
}

// NewCmdFileGet builds `repo file get PROJECT/REPO PATH --ref REF`.
// Output is the raw file bytes — straight to stdout by default, or to
// the path given by --out. No --json: file content is the file's content,
// not metadata.
func NewCmdFileGet(f *factory.Factory) *cobra.Command {
	var ref string
	var outPath string
	var hostname string

	cmd := &cobra.Command{
		Use:   "get PROJECT/REPO PATH",
		Short: "Read a file's content at a ref",
		Long: "Fetch the raw bytes of a file at the given ref (branch, tag, or\n" +
			"commit hash). Useful for inspecting source without cloning.\n\n" +
			"Output is written verbatim — no encoding, no normalisation —\n" +
			"so binary files round-trip cleanly when used with --out.\n\n" +
			"Examples:\n" +
			"  bitbottle repo file get myws/my-svc README.md --ref main\n" +
			"  bitbottle repo file get myws/my-svc go.mod --ref v1.2.0\n" +
			"  bitbottle repo file get myws/my-svc logo.png --ref main --out logo.png",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if ref == "" {
				return fmt.Errorf("--ref is required")
			}
			repoArg, path := args[0], args[1]
			if path == "" {
				return fmt.Errorf("PATH is required")
			}
			refParsed, err := bbrepo.Parse(repoArg)
			if err != nil {
				return err
			}
			host, err := resolveHostname(f, hostname)
			if err != nil {
				return err
			}
			client, err := f.Backend(host)
			if err != nil {
				return err
			}
			body, err := client.GetFileContent(refParsed.Project, refParsed.Slug, ref, path)
			if err != nil {
				return err
			}
			if outPath != "" {
				return os.WriteFile(outPath, body, 0o644)
			}
			_, werr := f.IOStreams.Out.Write(body)
			return werr
		},
	}
	cmd.Flags().StringVar(&ref, "ref", "", "Branch, tag, or commit hash to read from (required)")
	cmd.Flags().StringVar(&outPath, "out", "", "Write content to the given file path instead of stdout")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

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
