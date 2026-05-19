package backend

// Client is the composite capability interface that every backend adapter must
// satisfy. Capability interfaces that only some backends implement (e.g.
// PipelineClient, IssueClient, BranchProtector) are NOT embedded here; they
// are accessed via their respective AsXxx helper which returns a typed
// ErrUnsupportedOnHost when the backend doesn't implement the capability.
//
// Per-feature interface declarations live in client_<feature>.go files in this
// package so that new features can add interfaces without touching this file.
// The composite Client embedding block is the single intended conflict point
// for parallel development — per-feature optional interfaces never touch it.
type Client interface {
	RepoLister
	RepoReader
	RepoWriter
	RepoDeleter
	RepoRenamer
	RepoVisibilitySetter
	RepoDefaultBranchSetter
	SourceReader
	PRLister
	PRReader
	PRCreator
	PRMerger
	PRApprover
	PRDiffer
	BranchLister
	BranchCreator
	BranchDeleter
	TagLister
	TagCreator
	TagDeleter
	PREditor
	PRDecliner
	PRUnapprover
	PRReadier
	PRUnreadier
	PRReviewRequester
	PRReviewer
	UserGetter
	CommitLister
	CommitReader
	PRCommentLister
	PRCommentAdder
	PRCommentEditor
	PRCommentDeleter
	PRActivityReader
	CommitCommenter
	CommitStatusLister
	CommitStatusReporter
	WebhookLister
	WebhookReader
	WebhookCreator
	WebhookDeleter
}

// Feature names a capability that some backends may not implement. The
// As<X> helpers map each Feature to the optional interface a Client must
// satisfy to expose that capability.
type Feature string
