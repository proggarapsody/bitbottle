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
			Host:    host,
			Feature: string(FeatureWorkspaces),
			Message: fmt.Sprintf("workspaces are not supported on %s (Bitbucket Cloud only)", host),
		}
	}
	return wc, nil
}
