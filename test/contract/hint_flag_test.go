// Package contract_test verifies cross-cutting contracts that cannot live in a
// single leaf package without creating import cycles.
package contract_test

import (
	"regexp"
	"testing"

	"github.com/spf13/pflag"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/cmd/root"
	"github.com/proggarapsody/bitbottle/pkg/errfmt"
)

// backtickSpan matches the content between a pair of backticks.
var backtickSpan = regexp.MustCompile("`[^`]*`")

// flagToken matches --long-flag and -f short-flag tokens.
var flagToken = regexp.MustCompile(`--[\w-]+|-[a-zA-Z]`)

// flagsInHint returns flag tokens (--flag / -f) that appear inside backtick
// spans of a hint string. Restricting to backtick spans avoids false positives
// from hyphenated words like "auto-merge" or "self-signed".
func flagsInHint(hint string) []string {
	var flags []string
	for _, span := range backtickSpan.FindAllString(hint, -1) {
		flags = append(flags, flagToken.FindAllString(span, -1)...)
	}
	return flags
}

// TestHintFlagContract asserts that every --flag / -f token mentioned inside
// backtick spans of errfmt catalogue hints is registered as a root persistent
// flag. This catches hint drift where a flag is promised to the user but not
// wired on the command path (e.g. #387 `-k` missing, #394 phantom --debug).
//
// Scope note: this only checks root persistent flags (inherited by every
// subcommand). Per-command flag coverage — where a hint references a flag that
// is only present on a specific subcommand — is not yet covered here.
func TestHintFlagContract(t *testing.T) {
	cmd := root.NewCmdRoot(factory.New())

	registered := map[string]struct{}{}
	cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		registered["--"+f.Name] = struct{}{}
		if f.Shorthand != "" {
			registered["-"+f.Shorthand] = struct{}{}
		}
	})

	for code, hints := range errfmt.CatalogueHints() {
		for _, hint := range hints {
			for _, flag := range flagsInHint(hint) {
				if _, ok := registered[flag]; !ok {
					t.Errorf("code %s: hint %q references flag %q which is not a root persistent flag", code, hint, flag)
				}
			}
		}
	}
}
