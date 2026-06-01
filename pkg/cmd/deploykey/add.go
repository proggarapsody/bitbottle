package deploykey

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// mapPermission converts the CLI-facing permission value ("read" or "read-write")
// to the Bitbucket Cloud API wire value ("read" or "read_write").
// Empty string passes through unchanged (lets the API use its default).
func mapPermission(p string) (string, error) {
	switch p {
	case "read":
		return "read", nil
	case "read-write":
		return "read_write", nil
	case "":
		return "", nil
	default:
		return "", fmt.Errorf("invalid --permission %q: must be read or read-write", p)
	}
}

// NewCmdAdd builds the `deploy-key add` cobra command.
func NewCmdAdd(f *factory.Factory) *cobra.Command {
	var key, label, hostname, permission string
	cmd := &cobra.Command{
		Use:   "add [PROJECT/REPO]",
		Short: "Add a deploy key to a repository",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			wirePermission, err := mapPermission(permission)
			if err != nil {
				return err
			}
			ref, err := factory.ResolveTarget(f, args, hostname)
			if err != nil {
				return err
			}
			client, err := f.Backend(ref.Host)
			if err != nil {
				return err
			}
			dk, err := backend.AsDeployKeyClient(client, ref.Host)
			if err != nil {
				return err
			}
			added, err := dk.AddDeployKey(ref.Project, ref.Slug, backend.DeployKeyInput{
				Key:        key,
				Label:      label,
				Permission: wirePermission,
			})
			if err != nil {
				return err
			}
			p := deployKeyFields(f, format.ConfigFromCmd(cmd))
			p.AddItem(added)
			return p.Render()
		},
	}
	cmd.Flags().StringVar(&key, "key", "", "SSH public key (required)")
	cmd.Flags().StringVar(&label, "label", "", "Label for the key")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	cmd.Flags().StringVar(&permission, "permission", "read", "Key permission: read or read-write (Cloud only)")
	_ = cmd.MarkFlagRequired("key")
	return cmd
}
