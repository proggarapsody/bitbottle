package ipallowlist

import (
	"fmt"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// resolveWorkspace returns the workspace slug from the explicit arg, or falls
// back to the pinned repo's namespace (Project field). An error is returned
// when neither is available.
func resolveWorkspace(f *factory.Factory, explicit string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}
	ref, err := f.BaseRepo()
	if err == nil && ref.Project != "" {
		return ref.Project, nil
	}
	return "", fmt.Errorf("workspace required: pass a workspace slug as an argument or run from inside a Cloud checkout")
}
