package script_test

import (
	"os"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"

	"github.com/proggarapsody/bitbottle/internal/app"
)

func TestMain(m *testing.M) {
	testscript.Main(m, map[string]func(){
		"bitbottle": func() { os.Exit(app.Run()) },
	})
}
