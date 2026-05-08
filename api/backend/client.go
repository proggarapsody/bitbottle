package backend

import (
	"fmt"
	"io"
)

type RepoLister interface {
	// ListRepos lists repositories. ns is the workspace (Bitbucket Cloud) or
	// project key (Bitbucket Server); pass "" for Server to list all repos.
	ListRepos(ns string, limit int) ([]Repository, error)
}

type RepoReader interface {
	GetRepo(ns, slug string) (Repository, error)
}

type RepoWriter interface {
	CreateRepo(ns string, in CreateRepoInput) (Repository, error)
}

type RepoDeleter interface {
	DeleteRepo(ns, slug string) error
}

type RepoRenamer interface {
	RenameRepo(ns, slug, newName string) (Repository, error)
}

type PRLister interface {
	ListPRs(ns, slug, state string, limit int) ([]PullRequest, error)
}

type PRReader interface {
	GetPR(ns, slug string, id int) (PullRequest, error)
}

type PRCreator interface {
	CreatePR(ns, slug string, in CreatePRInput) (PullRequest, error)
}

type PRMerger interface {
	MergePR(ns, slug string, id int, in MergePRInput) (PullRequest, error)
}

type PRApprover interface {
	ApprovePR(ns, slug string, id int) error
}

type PRDiffer interface {
	GetPRDiff(ns, slug string, id int) (string, error)
}

type BranchLister interface {
	ListBranches(ns, slug string, limit int) ([]Branch, error)
}

type BranchDeleter interface {
	DeleteBranch(ns, slug, branch string) error
}

type UserGetter interface {
	GetCurrentUser() (User, error)
}

type BranchCreator interface {
	CreateBranch(ns, slug string, in CreateBranchInput) (Branch, error)
}

type TagLister interface {
	ListTags(ns, slug string, limit int) ([]Tag, error)
}

type TagCreator interface {
	CreateTag(ns, slug string, in CreateTagInput) (Tag, error)
}

type TagDeleter interface {
	DeleteTag(ns, slug, name string) error
}

type PREditor interface {
	UpdatePR(ns, slug string, id int, in UpdatePRInput) (PullRequest, error)
}

type PRDecliner interface {
	DeclinePR(ns, slug string, id int) error
}

// PRReopener reverses a decline. Implemented only by Bitbucket Server / Data
// Center (POST /pull-requests/{id}/reopen). Bitbucket Cloud has no reopen
// primitive — declined PRs are terminal there (BCLOUD-23807) — so callers
// route through AsPRReopener so the Cloud-only constraint surfaces as a
// typed ErrUnsupportedOnHost rather than a panic.
type PRReopener interface {
	ReopenPR(ns, slug string, id int) error
}

type PRUnapprover interface {
	UnapprovePR(ns, slug string, id int) error
}

type PRReadier interface {
	ReadyPR(ns, slug string, id int) error
}

type PRReviewRequester interface {
	RequestReview(ns, slug string, id int, users []string) error
}

type CommitLister interface {
	ListCommits(ns, slug, branch string, limit int) ([]Commit, error)
}

type CommitReader interface {
	GetCommit(ns, slug, hash string) (Commit, error)
}

// PRChangesRequester can request changes on a pull request (Cloud only).
// Access via type assertion — not embedded in Client.
type PRChangesRequester interface {
	RequestChangesPR(ns, slug string, id int) error
}

type PRCommentLister interface {
	ListPRComments(ns, slug string, id int) ([]PRComment, error)
}

type PRCommentAdder interface {
	AddPRComment(ns, slug string, id int, in AddPRCommentInput) (PRComment, error)
}

type CommitStatusLister interface {
	ListCommitStatuses(ns, slug, hash string) ([]CommitStatus, error)
}

type WebhookLister interface {
	ListWebhooks(ns, slug string) ([]Webhook, error)
}

type WebhookReader interface {
	GetWebhook(ns, slug, id string) (Webhook, error)
}

type WebhookCreator interface {
	CreateWebhook(ns, slug string, in CreateWebhookInput) (Webhook, error)
}

type WebhookDeleter interface {
	DeleteWebhook(ns, slug, id string) error
}

type Client interface {
	RepoLister
	RepoReader
	RepoWriter
	RepoDeleter
	RepoRenamer
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
	PRReviewRequester
	UserGetter
	CommitLister
	CommitReader
	PRCommentLister
	PRCommentAdder
	CommitStatusLister
	WebhookLister
	WebhookReader
	WebhookCreator
	WebhookDeleter
}

// ServerCapabilities is implemented only by Bitbucket Data Center clients.
type ServerCapabilities interface {
	GetApplicationProperties() (AppProperties, error)
}

// RepoForker is implemented only by Bitbucket Cloud clients — Bitbucket
// Server / Data Center has no fork primitive in its REST API. Access via
// AsRepoForker so the Cloud-only constraint surfaces as a typed
// ErrUnsupportedOnHost error rather than a panic.
type RepoForker interface {
	ForkRepo(ns, slug string, in ForkRepoInput) (Repository, error)
}

// PipelineClient is implemented only by Bitbucket Cloud clients.
type PipelineClient interface {
	ListPipelines(ns, slug string, limit int) ([]Pipeline, error)
	GetPipeline(ns, slug, uuid string) (Pipeline, error)
	RunPipeline(ns, slug string, in RunPipelineInput) (Pipeline, error)
	ListPipelineSteps(ns, slug, uuid string) ([]PipelineStep, error)
	GetPipelineStepLog(ns, slug, pipelineUUID, stepUUID string) (io.ReadCloser, error)
	ListPipelineVariables(ns, slug string) ([]PipelineVariable, error)
	SetPipelineVariable(ns, slug string, in PipelineVariableInput) (PipelineVariable, error)
	DeletePipelineVariable(ns, slug, key string) error
}

// Feature names a capability that some backends may not implement. The
// registry maps each Feature to the optional interface a Client must
// satisfy to expose that capability.
type Feature string

const (
	FeaturePipelines Feature = "pipelines"
	FeatureRepoFork  Feature = "repo-fork"
)

// AsPipelineClient returns the PipelineClient view of c, or a typed
// *DomainError (Kind=ErrUnsupportedOnHost) if the backend at host does not
// implement the Pipelines capability.
func AsPipelineClient(c Client, host string) (PipelineClient, error) {
	pc, ok := c.(PipelineClient)
	if !ok {
		return nil, &DomainError{
			Kind:    ErrUnsupportedOnHost,
			Code:    CodeHostUnsupported,
			Host:    host,
			Feature: string(FeaturePipelines),
			Message: fmt.Sprintf("pipelines are not supported on %s (Bitbucket Cloud only)", host),
		}
	}
	return pc, nil
}

// AsRepoForker returns the RepoForker view of c, or a typed *DomainError
// (Kind=ErrUnsupportedOnHost) if the backend at host has no fork primitive
// (Bitbucket Server / Data Center).
func AsRepoForker(c Client, host string) (RepoForker, error) {
	rf, ok := c.(RepoForker)
	if !ok {
		return nil, &DomainError{
			Kind:    ErrUnsupportedOnHost,
			Code:    CodeHostUnsupported,
			Host:    host,
			Feature: string(FeatureRepoFork),
			Message: fmt.Sprintf("repo fork is not supported on %s (Bitbucket Cloud only)", host),
		}
	}
	return rf, nil
}

// WorkspaceClient is implemented only by Bitbucket Cloud clients. Bitbucket
// Server / Data Center has no workspace concept — its projects live directly
// under the instance — so the workspace and project list operations are
// Cloud-only and accessed via the AsWorkspaceClient type assertion.
type WorkspaceClient interface {
	ListWorkspaces(limit int) ([]Workspace, error)
	ListProjects(workspace string, limit int) ([]Project, error)
}

// FeatureWorkspaces names the workspace/project listing capability for
// typed-error reporting.
const FeatureWorkspaces Feature = "workspaces"

// AsWorkspaceClient returns the WorkspaceClient view of c, or a typed
// *DomainError (Kind=ErrUnsupportedOnHost) if the backend at host does not
// support workspaces.
func AsWorkspaceClient(c Client, host string) (WorkspaceClient, error) {
	wc, ok := c.(WorkspaceClient)
	if !ok {
		return nil, &DomainError{
			Kind:    ErrUnsupportedOnHost,
			Code:    CodeHostUnsupported,
			Host:    host,
			Feature: string(FeatureWorkspaces),
			Message: fmt.Sprintf("workspaces are not supported on %s (Bitbucket Cloud only)", host),
		}
	}
	return wc, nil
}

// IssueClient is implemented only by Bitbucket Cloud clients. Bitbucket
// Server / Data Center has no built-in issue tracker, so the entire issue
// surface is gated behind AsIssueClient.
type IssueClient interface {
	ListIssues(ns, slug, state string, limit int) ([]Issue, error)
	GetIssue(ns, slug string, id int) (Issue, error)
	CreateIssue(ns, slug string, in CreateIssueInput) (Issue, error)
	UpdateIssue(ns, slug string, id int, in UpdateIssueInput) (Issue, error)
}

// FeatureIssues names the issues capability for typed-error reporting.
const FeatureIssues Feature = "issues"

// AsIssueClient returns the IssueClient view of c, or a typed *DomainError
// (Kind=ErrUnsupportedOnHost) when called against a Server/DC backend.
func AsIssueClient(c Client, host string) (IssueClient, error) {
	ic, ok := c.(IssueClient)
	if !ok {
		return nil, &DomainError{
			Kind:    ErrUnsupportedOnHost,
			Code:    CodeHostUnsupported,
			Host:    host,
			Feature: string(FeatureIssues),
			Message: fmt.Sprintf("issues are not supported on %s (Bitbucket Cloud only)", host),
		}
	}
	return ic, nil
}

// DefaultReviewersResolver looks up the configured "default reviewers" for a
// repository given a source/target ref pair. Only Bitbucket Server / Data
// Center exposes this — Cloud has a similar feature with a different,
// non-trivial wire shape that we don't model yet.
//
// Implementations may need additional context (e.g. numeric repo ID) which
// they obtain themselves; callers pass only the values they natively know.
type DefaultReviewersResolver interface {
	DefaultReviewers(ns, slug, fromBranch, toBranch string) ([]User, error)
}

// FeatureDefaultReviewers names the default-reviewers capability for
// typed-error reporting.
const FeatureDefaultReviewers Feature = "default-reviewers"

// BranchProtector exposes branch-restriction management on Bitbucket
// Server / Data Center. The Cloud backend has a different "branch
// restrictions" shape that's not modelled here — calls against a Cloud
// client surface ErrUnsupportedOnHost via AsBranchProtector.
type BranchProtector interface {
	ListBranchProtections(ns, slug string, limit int) ([]BranchProtection, error)
	CreateBranchProtection(ns, slug string, in CreateBranchProtectionInput) (BranchProtection, error)
	DeleteBranchProtection(ns, slug string, id int) error
}

// FeatureBranchProtect names the branch-protection capability for typed-
// error reporting.
const FeatureBranchProtect Feature = "branch-protect"

// FeaturePRReopen names the PR-reopen capability for typed-error reporting.
// Bitbucket Cloud has no reopen primitive (BCLOUD-23807), so callers gate
// the feature behind AsPRReopener.
const FeaturePRReopen Feature = "pr-reopen"

// AsPRReopener returns the PRReopener view of c, or a typed *DomainError
// (Kind=ErrUnsupportedOnHost) when the backend at host has no reopen
// primitive (currently Bitbucket Cloud).
func AsPRReopener(c Client, host string) (PRReopener, error) {
	r, ok := c.(PRReopener)
	if !ok {
		return nil, &DomainError{
			Kind:    ErrUnsupportedOnHost,
			Code:    CodeHostUnsupported,
			Host:    host,
			Feature: string(FeaturePRReopen),
			Message: fmt.Sprintf("pr reopen is not supported on %s (Bitbucket Server / Data Center only)", host),
		}
	}
	return r, nil
}

// AsBranchProtector returns the BranchProtector view of c, or a typed
// *DomainError (Kind=ErrUnsupportedOnHost) when called against a backend
// that doesn't model branch protections (currently Cloud).
func AsBranchProtector(c Client, host string) (BranchProtector, error) {
	bp, ok := c.(BranchProtector)
	if !ok {
		return nil, &DomainError{
			Kind:    ErrUnsupportedOnHost,
			Code:    CodeHostUnsupported,
			Host:    host,
			Feature: string(FeatureBranchProtect),
			Message: fmt.Sprintf("branch protection is not supported on %s (Bitbucket Server / Data Center only)", host),
		}
	}
	return bp, nil
}

// AsDefaultReviewersResolver returns the DefaultReviewersResolver view of c,
// or a typed *DomainError when the backend doesn't implement it (currently
// Cloud). Callers use the returned error to decide whether to skip the
// auto-apply step entirely.
func AsDefaultReviewersResolver(c Client, host string) (DefaultReviewersResolver, error) {
	r, ok := c.(DefaultReviewersResolver)
	if !ok {
		return nil, &DomainError{
			Kind:    ErrUnsupportedOnHost,
			Code:    CodeHostUnsupported,
			Host:    host,
			Feature: string(FeatureDefaultReviewers),
			Message: fmt.Sprintf("default reviewers lookup is not supported on %s (Bitbucket Server / Data Center only)", host),
		}
	}
	return r, nil
}

// CodeSearcher performs workspace-scoped code search on Bitbucket Cloud.
// Bitbucket Server / Data Center does not expose a first-class REST code-
// search endpoint (search there is provided by the separate Sourcegraph
// integration or third-party plugins), so the entire surface is gated
// behind AsCodeSearcher. Server adapters intentionally do NOT implement
// this interface — the type-assertion in AsCodeSearcher is what surfaces
// the typed host.unsupported error to callers.
type CodeSearcher interface {
	SearchCode(workspace, query string, limit int) ([]CodeSearchHit, error)
}

// FeatureCodeSearch names the code-search capability for typed-error
// reporting.
const FeatureCodeSearch Feature = "code-search"

// AsCodeSearcher returns the CodeSearcher view of c, or a typed *DomainError
// (Kind=ErrUnsupportedOnHost) when called against a backend that doesn't
// model code search (currently Server/DC).
func AsCodeSearcher(c Client, host string) (CodeSearcher, error) {
	cs, ok := c.(CodeSearcher)
	if !ok {
		return nil, &DomainError{
			Kind:    ErrUnsupportedOnHost,
			Code:    CodeHostUnsupported,
			Host:    host,
			Feature: string(FeatureCodeSearch),
			Message: fmt.Sprintf("code search is not supported on %s (Bitbucket Cloud only)", host),
		}
	}
	return cs, nil
}
