package completion

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func NewCmdCompletion(f *factory.Factory) *cobra.Command {
	var shell string
	cmd := &cobra.Command{
		Use:   "completion",
		Short: "Generate shell completion scripts",
		RunE: func(cmd *cobra.Command, args []string) error {
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
	_ = cmd.MarkFlagRequired("shell")
	return cmd
}
