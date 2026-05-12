// Package set implements the `variable set` subcommand.
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

// Options holds parsed flags for `variable set`.
type Options struct {
	Hostname string
	Body     string // takes precedence over positional VALUE; "-" reads stdin
	Secured  bool
	Scope    string // "repository" (default) | "workspace" | "deployment"
	EnvUUID  string // required when scope=deployment

	// Args[0] = PROJECT/REPO, Args[1] = KEY, Args[2] (optional) = VALUE
	Args []string

	// Stdin is overridable in tests; defaults to os.Stdin when nil.
	Stdin io.Reader
}

// NewCmdSet builds the `variable set` cobra command.
func NewCmdSet(f *factory.Factory, runF func(*Options) error) *cobra.Command {
	opts := &Options{}
	cmd := &cobra.Command{
		Use:   "set PROJECT/REPO KEY [VALUE]",
		Short: "Create or update a variable (upsert by KEY)",
		Long: `Set a variable. The variable is upserted by KEY: created if absent, updated
if present.

Use --scope to target repository (default), workspace, or deployment variables.
For --scope deployment, --env ENV-UUID is required.

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
	cmd.Flags().StringVar(&opts.Scope, "scope", "repository", "Variable scope: repository, workspace, or deployment")
	cmd.Flags().StringVar(&opts.EnvUUID, "env", "", "Environment UUID (required for --scope deployment)")
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
	key := opts.Args[1]

	switch opts.Scope {
	case "repository":
		pc, err := backend.AsPipelineClient(client, ref.Host)
		if err != nil {
			return err
		}
		v, err := pc.SetPipelineVariable(ref.Project, ref.Slug, backend.PipelineVariableInput{
			Key:     key,
			Value:   value,
			Secured: opts.Secured,
		})
		if err != nil {
			return err
		}
		return printSet(f, v.Secured, v.Key)

	case "workspace":
		wc, err := backend.AsWorkspaceVariableClient(client, ref.Host)
		if err != nil {
			return err
		}
		v, err := wc.SetWorkspaceVariable(ref.Project, backend.PipelineVariableInput{
			Key:     key,
			Value:   value,
			Secured: opts.Secured,
		})
		if err != nil {
			return err
		}
		return printSet(f, v.Secured, v.Key)

	case "deployment":
		if opts.EnvUUID == "" {
			return fmt.Errorf("--env ENV-UUID is required for --scope deployment")
		}
		dc, err := backend.AsDeploymentClient(client, ref.Host)
		if err != nil {
			return err
		}
		v, err := dc.SetEnvVariable(ref.Project, ref.Slug, opts.EnvUUID, backend.EnvVariableInput{
			Key:     key,
			Value:   value,
			Secured: opts.Secured,
		})
		if err != nil {
			return err
		}
		return printSet(f, v.Secured, v.Key)

	default:
		return fmt.Errorf("unknown scope %q; valid: repository, workspace, deployment", opts.Scope)
	}
}

func printSet(f *factory.Factory, secured bool, key string) error {
	if secured {
		fmt.Fprintf(f.IOStreams.Out, "Set secured variable %s\n", key)
	} else {
		fmt.Fprintf(f.IOStreams.Out, "Set variable %s\n", key)
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
