package pr

import (
	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

func prFields(f *factory.Factory, jsonFields, jqExpr string) *format.Printer[backend.PullRequest] {
	p := format.New[backend.PullRequest](f.IOStreams.Out, f.IOStreams.IsStdoutTTY(), jsonFields, jqExpr)
	p.AddField(format.Field[backend.PullRequest]{Name: "id", Header: "ID", Extract: func(pr backend.PullRequest) any { return pr.ID }})
	p.AddField(format.Field[backend.PullRequest]{Name: "title", Header: "TITLE", Extract: func(pr backend.PullRequest) any { return pr.Title }})
	p.AddField(format.Field[backend.PullRequest]{Name: "state", Header: "STATE", Extract: func(pr backend.PullRequest) any { return pr.State }})
	p.AddField(format.Field[backend.PullRequest]{Name: "draft", Header: "DRAFT", Extract: func(pr backend.PullRequest) any { return pr.Draft }})
	p.AddField(format.Field[backend.PullRequest]{Name: "author", Header: "AUTHOR", Extract: func(pr backend.PullRequest) any { return pr.Author.Slug }})
	p.AddField(format.Field[backend.PullRequest]{Name: "fromBranch", Header: "FROM", Extract: func(pr backend.PullRequest) any { return pr.FromBranch }})
	p.AddField(format.Field[backend.PullRequest]{Name: "toBranch", Header: "TO", Extract: func(pr backend.PullRequest) any { return pr.ToBranch }})
	p.AddField(format.Field[backend.PullRequest]{Name: "webURL", Header: "URL", Aliases: []string{"url", "link"}, Extract: func(pr backend.PullRequest) any { return pr.WebURL }})
	return p
}

func prFieldsWithDescription(f *factory.Factory, jsonFields, jqExpr string) *format.Printer[backend.PullRequest] {
	p := prFields(f, jsonFields, jqExpr)
	p.AddField(format.Field[backend.PullRequest]{Name: "description", Header: "DESCRIPTION", Extract: func(pr backend.PullRequest) any { return pr.Description }})
	return p
}
