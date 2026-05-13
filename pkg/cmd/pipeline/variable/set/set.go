package set

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// Options holds parsed flags for `pipeline variable set`.
type Options struct {
	Hostname string
	Body     string // takes precedence over positional VALUE; "-" reads stdin
	Secured  bool

	// Args[0] = PROJECT/REPO, Args[1] = KEY, Args[2] (optional) = VALUE
	Args []string

	// Stdin is overridable in tests; defaults to os.Stdin when nil.
	Stdin io.Reader
}

// NewCmdSet builds the `pipeline variable set` cobra command.
//
// Deprecated: use `bitbottle variable set --scope repository` instead.
func NewCmdSet(f *factory.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:        "set PROJECT/REPO KEY [VALUE]",
		Short:      "Create or update a pipeline variable (upsert by KEY)",
		Deprecated: "use `bitbottle variable set --scope repository` instead",
		Long: `Set a repository-level pipeline variable. The variable is upserted by KEY:
created if absent, updated if present.

DEPRECATED: use ` + "`bitbottle variable set --scope repository`" + ` instead.

The value can be supplied as the third positional argument, via --body, or by
passing --body=- to read the value from standard input. Reading from stdin is
the recommended path for secured variables — it keeps the value out of shell
history.`,
		Args: cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Args = args
			if runF != nil {
				return runF(opts)
			}
			return setRun(f, opts)
		},
	}
	cmd.Flags().StringVar(&opts.Body, "body", "", "Value (use \"-\" to read from stdin)")
	cmd.Flags().BoolVar(&opts.Secured, "secured", false, "Mark variable as secured (value will be redacted on read)")
	cmd.Flags().StringVar(&opts.Hostname, "hostname", "", "Bitbucket hostname (overrides auto-detection)")
	return cmd
}

func setRun(f *factory.Factory, opts *Options) error {
	ref, err := factory.ResolveTarget(f, opts.Args, opts.Hostname)
	if err != nil {
		return err
	}
	value, err := resolveValue(opts)
	if err != nil {
		return err
	}
	client, err := f.Backend(ref.Host)
	if err != nil {
		return err
	}
	pc, err := backend.AsPipelineClient(client, ref.Host)
	if err != nil {
		return err
	}
	v, err := pc.SetPipelineVariable(ref.Project, ref.Slug, backend.PipelineVariableInput{
		Key:     opts.Args[1],
		Value:   value,
		Secured: opts.Secured,
	})
	if err != nil {
		return err
	}
	out := f.IOStreams.Out
	if v.Secured {
		fmt.Fprintf(out, "Set secured variable %s\n", v.Key)
	} else {
		fmt.Fprintf(out, "Set variable %s\n", v.Key)
	}
	return nil
}

func resolveValue(opts *Options) (string, error) {
	switch {
	case opts.Body == "-":
		stdin := opts.Stdin
		if stdin == nil {
			stdin = os.Stdin
		}
		b, err := io.ReadAll(stdin)
		if err != nil {
			return "", fmt.Errorf("read value from stdin: %w", err)
		}
		return strings.TrimRight(string(b), "\n"), nil
	case opts.Body != "":
		return opts.Body, nil
	case len(opts.Args) >= 3:
		return opts.Args[2], nil
	default:
		return "", fmt.Errorf("value required: pass as the third argument, via --body, or via --body=- on stdin")
	}
}
