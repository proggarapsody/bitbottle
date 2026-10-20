// Package contract_test verifies cross-cutting contracts that cannot live in a
// single leaf package without creating import cycles.
package contract_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory/factorytest"
	"github.com/proggarapsody/bitbottle/pkg/cmd/root"
	"github.com/proggarapsody/bitbottle/pkg/errfmt"
)

var backtickSpan = regexp.MustCompile("`[^`]*`")
var flagToken = regexp.MustCompile(`--[\w-]+|-[a-zA-Z]`)

// Standalone flag spans promise inherited root flags. A full inline invocation
// supplies a command context, so command-specific flags must exist on that path.
func validateHintFlags(cmd *cobra.Command, hint string) error {
	for _, quoted := range backtickSpan.FindAllString(hint, -1) {
		span := strings.Trim(quoted, "`")
		flags := flagToken.FindAllString(span, -1)
		if len(flags) == 0 {
			continue
		}
		sets := []*pflag.FlagSet{cmd.PersistentFlags()}
		words := strings.Fields(span)
		if len(words) > 1 && words[0] == "bitbottle" {
			target, _, err := cmd.Find(words[1:])
			if err != nil {
				return fmt.Errorf("invalid command in %q: %w", span, err)
			}
			sets = []*pflag.FlagSet{target.Flags(), target.PersistentFlags(), target.InheritedFlags()}
		}
		for _, flag := range flags {
			found := false
			for _, set := range sets {
				if strings.HasPrefix(flag, "--") {
					found = found || set.Lookup(strings.TrimPrefix(flag, "--")) != nil
				} else {
					found = found || set.ShorthandLookup(strings.TrimPrefix(flag, "-")) != nil
				}
			}
			if !found {
				return fmt.Errorf("flag %q is not registered for %q", flag, span)
			}
		}
	}
	return nil
}

func TestHintFlagContract(t *testing.T) {
	f, _, _ := factorytest.New(t, factorytest.Opts{})
	cmd := root.NewCmdRoot(f)
	for code, hints := range errfmt.CatalogueHints() {
		for _, hint := range hints {
			if err := validateHintFlags(cmd, hint); err != nil {
				t.Errorf("code %s: hint %q: %v", code, hint, err)
			}
		}
	}
}

func TestHintFlagCommandContext(t *testing.T) {
	for _, tt := range []struct {
		name, hint string
		valid      bool
	}{
		{"merge strategy", "Use `bitbottle pr merge --merge`.", true},
		{"merge with ID", "Use `bitbottle pr merge 42 --squash`.", true},
		{"inherited flag", "Use `bitbottle pr merge --skip-tls-verify`.", true},
		{"standalone inherited flag", "Use `-k`.", true},
		{"unknown flag", "Use `bitbottle pr merge --imaginary`.", false},
		{"wrong command", "Use `bitbottle pr list --merge`.", false},
		{"standalone local flag", "Use `--merge`.", false},
		{"real debug flag", "Use `--debug`.", true},
		{"phantom debug flag", "Use `--debug-imaginary`.", false},
		{"unknown command", "Use `bitbottle invented --skip-tls-verify`.", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			f, _, _ := factorytest.New(t, factorytest.Opts{})
			err := validateHintFlags(root.NewCmdRoot(f), tt.hint)
			if tt.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
