package completion

import "github.com/proggarapsody/bitbottle/pkg/cmdregistry"

func init() {
	cmdregistry.Register(NewCmdCompletion)
}
