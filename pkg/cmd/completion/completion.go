package completion

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdCompletion(f *factory.Factory) *cobra.Command {
	var shell string
	cmd := &cobra.Command{
		Use:       "completion [bash|zsh|fish|powershell]",
		Short:     "Generate shell completion scripts",
		ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
		Args:      cobra.MatchAll(cobra.MaximumNArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				shell = args[0]
			}
			if shell == "" {
				return fmt.Errorf("shell is required: bash, zsh, fish, or powershell")
			}
			switch shell {
			case "bash":
				return cmd.Root().GenBashCompletion(f.IOStreams.Out)
			case "zsh":
				return cmd.Root().GenZshCompletion(f.IOStreams.Out)
			case "fish":
				return cmd.Root().GenFishCompletion(f.IOStreams.Out, true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletionWithDesc(f.IOStreams.Out)
			default:
				return fmt.Errorf("unsupported shell %q: must be bash, zsh, fish, or powershell", shell)
			}
		},
	}
	cmd.Flags().StringVarP(&shell, "shell", "s", "", "Shell type: bash, zsh, fish, powershell")
	return cmd
}
