package protect

import (
	"strings"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// fields is the shared printer for `branch protect list`. ID, type and the
// matcher form the always-on columns; users/groups become a single comma-
// joined column to keep the table readable on narrow terminals while
// remaining queryable via --json.
func fields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.BranchProtection] {
	p := format.New[backend.BranchProtection](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), cfg)
	p.AddField(format.Field[backend.BranchProtection]{Name: "id", Header: "ID", Extract: func(b backend.BranchProtection) any { return b.ID }})
	p.AddField(format.Field[backend.BranchProtection]{Name: "type", Header: "TYPE", Extract: func(b backend.BranchProtection) any { return b.Type }})
	p.AddField(format.Field[backend.BranchProtection]{Name: "matcher", Header: "MATCHER", Extract: func(b backend.BranchProtection) any { return b.MatcherID }})
	p.AddField(format.Field[backend.BranchProtection]{Name: "kind", Header: "KIND", Extract: func(b backend.BranchProtection) any { return b.MatcherKind }})
	p.AddField(format.Field[backend.BranchProtection]{Name: "exempt", Header: "EXEMPT", Extract: func(b backend.BranchProtection) any {
		all := append([]string{}, b.Users...)
		for _, g := range b.Groups {
			all = append(all, "@"+g)
		}
		return strings.Join(all, ",")
	}})
	return p
}
