package workspace

import "github.com/proggarapsody/bitbottle/pkg/cmdregistry"

func init() {
	cmdregistry.Register(NewCmdWorkspace)
}
