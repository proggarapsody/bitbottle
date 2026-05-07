package issue

import (
	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
	"github.com/proggarapsody/bitbottle/pkg/iostreams"
)

// IssueStateColor maps Bitbucket Cloud issue states to colours. Active
// (open / new) renders green; closed states (resolved / closed) magenta;
// "won't fix" / "invalid" / "duplicate" red — terminal and negative;
// "on hold" yellow because it's neither active nor done.
//
// Exported so a future `issue view` formatter (or external tooling) can
// share the mapping. Unknown states pass through uncoloured.
func IssueStateColor(ios *iostreams.IOStreams) func(string) string {
	return func(state string) string {
		switch state {
		case "open", "new":
			return ios.ColorGreen(state)
		case "resolved", "closed":
			return ios.ColorMagenta(state)
		case "wontfix", "invalid", "duplicate":
			return ios.ColorRed(state)
		case "on hold":
			return ios.ColorYellow(state)
		default:
			return state
		}
	}
}

// reporterSlug returns the reporter's user slug, never panicking on the
// zero User.
func reporterSlug(i backend.Issue) any { return i.Reporter.Slug }

// assigneeSlug returns the assignee's slug or "" when unassigned. Pointer
// dereference is the chokepoint that protects every consumer.
func assigneeSlug(i backend.Issue) any {
	if i.Assignee == nil {
		return ""
	}
	return i.Assignee.Slug
}

func issueListFields(f *factory.Factory, jsonFields, jqExpr string) *format.Printer[backend.Issue] {
	p := format.New[backend.Issue](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), jsonFields, jqExpr)
	p.AddField(format.Field[backend.Issue]{Name: "id", Header: "ID", Extract: func(i backend.Issue) any { return i.ID }})
	p.AddField(format.Field[backend.Issue]{Name: "title", Header: "TITLE", Extract: func(i backend.Issue) any { return i.Title }})
	p.AddField(format.Field[backend.Issue]{
		Name: "state", Header: "STATE",
		Extract:   func(i backend.Issue) any { return i.State },
		ColorFunc: IssueStateColor(f.IOStreams),
	})
	p.AddField(format.Field[backend.Issue]{Name: "kind", Header: "KIND", Extract: func(i backend.Issue) any { return i.Kind }})
	p.AddField(format.Field[backend.Issue]{Name: "priority", Header: "PRIORITY", Extract: func(i backend.Issue) any { return i.Priority }})
	p.AddField(format.Field[backend.Issue]{Name: "reporter", Header: "REPORTER", Extract: reporterSlug})
	p.AddField(format.Field[backend.Issue]{Name: "assignee", Header: "ASSIGNEE", Extract: assigneeSlug})
	p.AddField(format.Field[backend.Issue]{Name: "webURL", Header: "URL", Aliases: []string{"url", "link"}, Extract: func(i backend.Issue) any { return i.WebURL }})
	return p
}

// issueViewFields adds the body content — useful on `issue view` but noisy
// in the list.
func issueViewFields(f *factory.Factory, jsonFields, jqExpr string) *format.Printer[backend.Issue] {
	p := issueListFields(f, jsonFields, jqExpr)
	p.AddField(format.Field[backend.Issue]{Name: "content", Header: "CONTENT", Extract: func(i backend.Issue) any { return i.Content }})
	return p
}
