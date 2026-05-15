package pr

import "github.com/proggarapsody/bitbottle/pkg/cmdregistry"

func init() {
	cmdregistry.Register(NewCmdPR)
}
