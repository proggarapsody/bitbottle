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

// flagsInHint returns the flag tokens (--flag / -f) referenced inside
// backtick-quoted spans of a hint string. Restricting to backtick spans
// avoids false positives from hyphenated words like "auto-merge".
func flagsInHint(hint string) []string {
	var tokens []string
	for _, span := range backtickSpan.FindAllString(hint, -1) {
		tokens = append(tokens, flagToken.FindAllString(span, -1)...)
	}
	return tokens
}

// TestHintFlagContract asserts that every --flag / -f token mentioned inside
// backtick spans of errfmt catalogue hints is registered as a root persistent
// flag. This catches hint drift where a flag is promised to the user but not
// wired on the command path (e.g. gh#387 `-k` missing, gh#394 phantom --debug).
func TestHintFlagContract(t *testing.T) {
	cmd := root.NewCmdRoot(factory.New())

	registered := map[string]bool{}
	cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		registered["--"+f.Name] = true
		if f.Shorthand != "" {
			registered["-"+f.Shorthand] = true
		}
	})

	for code, hints := range errfmt.CatalogueHints() {
		for _, hint := range hints {
			for _, tok := range flagsInHint(hint) {
				if !registered[tok] {
					t.Errorf("code %s: hint references %q which is not a root persistent flag", code, tok)
				}
			}
		}
	}
}
