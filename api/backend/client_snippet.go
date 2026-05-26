package backend

import "time"

// SnippetClient is implemented only by Bitbucket Cloud clients. Bitbucket
// Server / Data Center has no snippets API, so the entire snippet surface is
// gated behind AsSnippetClient.
type SnippetClient interface {
	ListSnippets(workspace string, limit int) ([]Snippet, error)
	GetSnippet(workspace, id string) (Snippet, error)
	CreateSnippet(workspace string, in CreateSnippetInput) (Snippet, error)
	DeleteSnippet(workspace, id string) error
	ListSnippetComments(workspace, snippetID string, limit int) ([]SnippetComment, error)
	AddSnippetComment(workspace, snippetID, body string) (SnippetComment, error)
	DeleteSnippetComment(workspace, snippetID string, commentID int) error
}

// FeatureSnippets names the snippets capability for typed-error reporting.
const FeatureSnippets Feature = "snippets"

// AsSnippetClient returns the SnippetClient view of c, or a typed *DomainError
// (Kind=ErrUnsupportedOnHost) when called against a Server/DC backend.
func AsSnippetClient(c Client, host string) (SnippetClient, error) {
	return requireFeature[SnippetClient](c, host, specFor(FeatureSnippets))
}

// Snippet is the domain representation of a Bitbucket Cloud snippet.
type Snippet struct {
	ID        string        `json:"id"`
	Title     string        `json:"title"`
	Owner     string        `json:"owner"`
	IsPrivate bool          `json:"is_private"`
	CreatedOn time.Time     `json:"created_on"`
	Files     []SnippetFile `json:"files"`
	WebURL    string        `json:"web_url"`
}

// SnippetFile is a file attached to a snippet.
type SnippetFile struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

// CreateSnippetInput holds the parameters for creating a new snippet.
type CreateSnippetInput struct {
	Title     string
	IsPrivate bool
	Files     []SnippetFile
}

// SnippetComment is the domain representation of a comment on a Bitbucket Cloud snippet.
type SnippetComment struct {
	ID        int    `json:"id"`
	Body      string `json:"body"`
	CreatedOn string `json:"created_on"`
	UpdatedOn string `json:"updated_on"`
	Author    string `json:"author"` // display name
	WebURL    string `json:"web_url"`
}
