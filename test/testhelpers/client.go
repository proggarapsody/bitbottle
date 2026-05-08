package testhelpers

import (
	"io"
	"testing"

	"github.com/proggarapsody/bitbottle/api/backend"
)

// FakeClient is a test double for backend.Client.
// Set the Fn fields for the methods your test exercises.
// Unset methods call t.Fatalf so unexpected calls are loud failures.
type FakeClient struct {
	T *testing.T

	// Repo methods
	ListReposFn  func(ns string, limit int) ([]backend.Repository, error)
	GetRepoFn    func(ns, slug string) (backend.Repository, error)
	CreateRepoFn func(ns string, in backend.CreateRepoInput) (backend.Repository, error)
	DeleteRepoFn func(ns, slug string) error
	RenameRepoFn func(ns, slug, newName string) (backend.Repository, error)
	ForkRepoFn   func(ns, slug string, in backend.ForkRepoInput) (backend.Repository, error)

	// PR methods
	ListPRsFn   func(ns, slug, state string, limit int) ([]backend.PullRequest, error)
	GetPRFn     func(ns, slug string, id int) (backend.PullRequest, error)
	CreatePRFn  func(ns, slug string, in backend.CreatePRInput) (backend.PullRequest, error)
	MergePRFn   func(ns, slug string, id int, in backend.MergePRInput) (backend.PullRequest, error)
	ApprovePRFn func(ns, slug string, id int) error
	GetPRDiffFn func(ns, slug string, id int) (string, error)

	// Branch / user methods
	ListBranchesFn   func(ns, slug string, limit int) ([]backend.Branch, error)
	DeleteBranchFn   func(ns, slug, branch string) error
	GetCurrentUserFn func() (backend.User, error)

	// Branch create method
	CreateBranchFn func(ns, slug string, in backend.CreateBranchInput) (backend.Branch, error)

	// Tag methods
	ListTagsFn  func(ns, slug string, limit int) ([]backend.Tag, error)
	CreateTagFn func(ns, slug string, in backend.CreateTagInput) (backend.Tag, error)
	DeleteTagFn func(ns, slug, name string) error

	// PR lifecycle methods
	UpdatePRFn      func(ns, slug string, id int, in backend.UpdatePRInput) (backend.PullRequest, error)
	DeclinePRFn     func(ns, slug string, id int) error
	UnapprovePRFn   func(ns, slug string, id int) error
	ReadyPRFn       func(ns, slug string, id int) error
	RequestReviewFn func(ns, slug string, id int, users []string) error

	// Pipeline methods (Cloud-only; satisfies backend.PipelineClient when set)
	ListPipelinesFn          func(ns, slug string, limit int) ([]backend.Pipeline, error)
	GetPipelineFn            func(ns, slug, uuid string) (backend.Pipeline, error)
	RunPipelineFn            func(ns, slug string, in backend.RunPipelineInput) (backend.Pipeline, error)
	ListPipelineStepsFn      func(ns, slug, uuid string) ([]backend.PipelineStep, error)
	GetPipelineStepLogFn     func(ns, slug, pipelineUUID, stepUUID string) (io.ReadCloser, error)
	ListPipelineVariablesFn  func(ns, slug string) ([]backend.PipelineVariable, error)
	SetPipelineVariableFn    func(ns, slug string, in backend.PipelineVariableInput) (backend.PipelineVariable, error)
	DeletePipelineVariableFn func(ns, slug, key string) error

	// Commit methods
	ListCommitsFn func(ns, slug, branch string, limit int) ([]backend.Commit, error)
	GetCommitFn   func(ns, slug, hash string) (backend.Commit, error)

	// PR comment methods
	ListPRCommentsFn func(ns, slug string, id int) ([]backend.PRComment, error)
	AddPRCommentFn   func(ns, slug string, id int, in backend.AddPRCommentInput) (backend.PRComment, error)

	// Commit status methods
	ListCommitStatusesFn func(ns, slug, hash string) ([]backend.CommitStatus, error)

	// Webhook methods
	ListWebhooksFn  func(ns, slug string) ([]backend.Webhook, error)
	GetWebhookFn    func(ns, slug, id string) (backend.Webhook, error)
	CreateWebhookFn func(ns, slug string, in backend.CreateWebhookInput) (backend.Webhook, error)
	DeleteWebhookFn func(ns, slug, id string) error

	// Workspace methods (Cloud-only; satisfies backend.WorkspaceClient when set)
	ListWorkspacesFn func(limit int) ([]backend.Workspace, error)
	ListProjectsFn   func(workspace string, limit int) ([]backend.Project, error)

	// Issue methods (Cloud-only; satisfies backend.IssueClient when set)
	ListIssuesFn         func(ns, slug, state string, limit int) ([]backend.Issue, error)
	GetIssueFn           func(ns, slug string, id int) (backend.Issue, error)
	CreateIssueFn        func(ns, slug string, in backend.CreateIssueInput) (backend.Issue, error)
	UpdateIssueFn        func(ns, slug string, id int, in backend.UpdateIssueInput) (backend.Issue, error)
	ReopenIssueFn        func(ns, slug string, id int) error
	AssignIssueFn        func(ns, slug string, id int, assignee string) error
	ListIssueCommentsFn  func(ns, slug string, id int) ([]backend.IssueComment, error)
	AddIssueCommentFn    func(ns, slug string, id int, body string) (backend.IssueComment, error)
	EditIssueCommentFn   func(ns, slug string, id, commentID int, body string) (backend.IssueComment, error)
	DeleteIssueCommentFn func(ns, slug string, id, commentID int) error

	// Default reviewers (Server-only; satisfies backend.DefaultReviewersResolver when set)
	DefaultReviewersFn func(ns, slug, fromBranch, toBranch string) ([]backend.User, error)

	// Branch protection (Server-only; satisfies backend.BranchProtector when set)
	ListBranchProtectionsFn  func(ns, slug string, limit int) ([]backend.BranchProtection, error)
	CreateBranchProtectionFn func(ns, slug string, in backend.CreateBranchProtectionInput) (backend.BranchProtection, error)
	DeleteBranchProtectionFn func(ns, slug string, id int) error

	// Code search (Cloud-only; satisfies backend.CodeSearcher when set)
	SearchCodeFn func(workspace, query string, limit int) ([]backend.CodeSearchHit, error)

	// Source primitives (both backends; satisfies backend.SourceReader when set)
	GetFileContentFn func(ns, slug, ref, path string) ([]byte, error)
	ListTreeFn       func(ns, slug, ref, path string) ([]backend.TreeEntry, error)

	// Code Insights (Server-only; satisfies backend.CodeInsightsClient when set)
	ListReportsFn       func(project, slug, hash string) ([]backend.CodeInsightsReport, error)
	GetReportFn         func(project, slug, hash, key string) (backend.CodeInsightsReport, error)
	SetReportFn         func(project, slug, hash, key string, in backend.CodeInsightsReportInput) (backend.CodeInsightsReport, error)
	DeleteReportFn      func(project, slug, hash, key string) error
	ListAnnotationsFn   func(project, slug, hash, key string) ([]backend.CodeInsightsAnnotation, error)
	AddAnnotationsFn    func(project, slug, hash, key string, in []backend.CodeInsightsAnnotationInput) error
	DeleteAnnotationsFn func(project, slug, hash, key string) error
	SetMergeCheckFn     func(project, slug, key string, in backend.MergeCheckInput) error
	GetMergeCheckFn     func(project, slug, key string) (backend.MergeCheck, error)
	DeleteMergeCheckFn  func(project, slug, key string) error
}

// Compile-time interface check.
var _ backend.Client = (*FakeClient)(nil)

func (c *FakeClient) ListRepos(ns string, limit int) ([]backend.Repository, error) {
	if c.ListReposFn != nil {
		return c.ListReposFn(ns, limit)
	}
	return nil, nil
}

func (c *FakeClient) GetRepo(ns, slug string) (backend.Repository, error) {
	if c.GetRepoFn != nil {
		return c.GetRepoFn(ns, slug)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.GetRepo; set GetRepoFn in your test")
	}
	return backend.Repository{}, nil
}

func (c *FakeClient) CreateRepo(ns string, in backend.CreateRepoInput) (backend.Repository, error) {
	if c.CreateRepoFn != nil {
		return c.CreateRepoFn(ns, in)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.CreateRepo; set CreateRepoFn in your test")
	}
	return backend.Repository{}, nil
}

func (c *FakeClient) DeleteRepo(ns, slug string) error {
	if c.DeleteRepoFn != nil {
		return c.DeleteRepoFn(ns, slug)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.DeleteRepo; set DeleteRepoFn in your test")
	}
	return nil
}

func (c *FakeClient) RenameRepo(ns, slug, newName string) (backend.Repository, error) {
	if c.RenameRepoFn != nil {
		return c.RenameRepoFn(ns, slug, newName)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.RenameRepo; set RenameRepoFn in your test")
	}
	return backend.Repository{}, nil
}

func (c *FakeClient) ForkRepo(ns, slug string, in backend.ForkRepoInput) (backend.Repository, error) {
	if c.ForkRepoFn != nil {
		return c.ForkRepoFn(ns, slug, in)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ForkRepo; set ForkRepoFn in your test")
	}
	return backend.Repository{}, nil
}

func (c *FakeClient) ListPRs(ns, slug, state string, limit int) ([]backend.PullRequest, error) {
	if c.ListPRsFn != nil {
		return c.ListPRsFn(ns, slug, state, limit)
	}
	return nil, nil
}

func (c *FakeClient) GetPR(ns, slug string, id int) (backend.PullRequest, error) {
	if c.GetPRFn != nil {
		return c.GetPRFn(ns, slug, id)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.GetPR; set GetPRFn in your test")
	}
	return backend.PullRequest{}, nil
}

func (c *FakeClient) CreatePR(ns, slug string, in backend.CreatePRInput) (backend.PullRequest, error) {
	if c.CreatePRFn != nil {
		return c.CreatePRFn(ns, slug, in)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.CreatePR; set CreatePRFn in your test")
	}
	return backend.PullRequest{}, nil
}

func (c *FakeClient) MergePR(ns, slug string, id int, in backend.MergePRInput) (backend.PullRequest, error) {
	if c.MergePRFn != nil {
		return c.MergePRFn(ns, slug, id, in)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.MergePR; set MergePRFn in your test")
	}
	return backend.PullRequest{}, nil
}

func (c *FakeClient) ApprovePR(ns, slug string, id int) error {
	if c.ApprovePRFn != nil {
		return c.ApprovePRFn(ns, slug, id)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ApprovePR; set ApprovePRFn in your test")
	}
	return nil
}

func (c *FakeClient) GetPRDiff(ns, slug string, id int) (string, error) {
	if c.GetPRDiffFn != nil {
		return c.GetPRDiffFn(ns, slug, id)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.GetPRDiff; set GetPRDiffFn in your test")
	}
	return "", nil
}

func (c *FakeClient) ListBranches(ns, slug string, limit int) ([]backend.Branch, error) {
	if c.ListBranchesFn != nil {
		return c.ListBranchesFn(ns, slug, limit)
	}
	return nil, nil
}

func (c *FakeClient) DeleteBranch(ns, slug, branch string) error {
	if c.DeleteBranchFn != nil {
		return c.DeleteBranchFn(ns, slug, branch)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.DeleteBranch; set DeleteBranchFn in your test")
	}
	return nil
}

func (c *FakeClient) GetCurrentUser() (backend.User, error) {
	if c.GetCurrentUserFn != nil {
		return c.GetCurrentUserFn()
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.GetCurrentUser; set GetCurrentUserFn in your test")
	}
	return backend.User{}, nil
}

func (c *FakeClient) CreateBranch(ns, slug string, in backend.CreateBranchInput) (backend.Branch, error) {
	if c.CreateBranchFn != nil {
		return c.CreateBranchFn(ns, slug, in)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.CreateBranch; set CreateBranchFn in your test")
	}
	return backend.Branch{}, nil
}

func (c *FakeClient) ListTags(ns, slug string, limit int) ([]backend.Tag, error) {
	if c.ListTagsFn != nil {
		return c.ListTagsFn(ns, slug, limit)
	}
	return nil, nil
}

func (c *FakeClient) CreateTag(ns, slug string, in backend.CreateTagInput) (backend.Tag, error) {
	if c.CreateTagFn != nil {
		return c.CreateTagFn(ns, slug, in)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.CreateTag; set CreateTagFn in your test")
	}
	return backend.Tag{}, nil
}

func (c *FakeClient) DeleteTag(ns, slug, name string) error {
	if c.DeleteTagFn != nil {
		return c.DeleteTagFn(ns, slug, name)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.DeleteTag; set DeleteTagFn in your test")
	}
	return nil
}

func (c *FakeClient) UpdatePR(ns, slug string, id int, in backend.UpdatePRInput) (backend.PullRequest, error) {
	if c.UpdatePRFn != nil {
		return c.UpdatePRFn(ns, slug, id, in)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.UpdatePR; set UpdatePRFn in your test")
	}
	return backend.PullRequest{}, nil
}

func (c *FakeClient) DeclinePR(ns, slug string, id int) error {
	if c.DeclinePRFn != nil {
		return c.DeclinePRFn(ns, slug, id)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.DeclinePR; set DeclinePRFn in your test")
	}
	return nil
}

func (c *FakeClient) UnapprovePR(ns, slug string, id int) error {
	if c.UnapprovePRFn != nil {
		return c.UnapprovePRFn(ns, slug, id)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.UnapprovePR; set UnapprovePRFn in your test")
	}
	return nil
}

func (c *FakeClient) ReadyPR(ns, slug string, id int) error {
	if c.ReadyPRFn != nil {
		return c.ReadyPRFn(ns, slug, id)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ReadyPR; set ReadyPRFn in your test")
	}
	return nil
}

func (c *FakeClient) RequestReview(ns, slug string, id int, users []string) error {
	if c.RequestReviewFn != nil {
		return c.RequestReviewFn(ns, slug, id, users)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.RequestReview; set RequestReviewFn in your test")
	}
	return nil
}

func (c *FakeClient) ListPipelines(ns, slug string, limit int) ([]backend.Pipeline, error) {
	if c.ListPipelinesFn != nil {
		return c.ListPipelinesFn(ns, slug, limit)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListPipelines; set ListPipelinesFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) GetPipeline(ns, slug, uuid string) (backend.Pipeline, error) {
	if c.GetPipelineFn != nil {
		return c.GetPipelineFn(ns, slug, uuid)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.GetPipeline; set GetPipelineFn in your test")
	}
	return backend.Pipeline{}, nil
}

func (c *FakeClient) RunPipeline(ns, slug string, in backend.RunPipelineInput) (backend.Pipeline, error) {
	if c.RunPipelineFn != nil {
		return c.RunPipelineFn(ns, slug, in)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.RunPipeline; set RunPipelineFn in your test")
	}
	return backend.Pipeline{}, nil
}

func (c *FakeClient) ListPipelineSteps(ns, slug, uuid string) ([]backend.PipelineStep, error) {
	if c.ListPipelineStepsFn != nil {
		return c.ListPipelineStepsFn(ns, slug, uuid)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListPipelineSteps; set ListPipelineStepsFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) GetPipelineStepLog(ns, slug, pipelineUUID, stepUUID string) (io.ReadCloser, error) {
	if c.GetPipelineStepLogFn != nil {
		return c.GetPipelineStepLogFn(ns, slug, pipelineUUID, stepUUID)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.GetPipelineStepLog; set GetPipelineStepLogFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) ListPipelineVariables(ns, slug string) ([]backend.PipelineVariable, error) {
	if c.ListPipelineVariablesFn != nil {
		return c.ListPipelineVariablesFn(ns, slug)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListPipelineVariables; set ListPipelineVariablesFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) SetPipelineVariable(ns, slug string, in backend.PipelineVariableInput) (backend.PipelineVariable, error) {
	if c.SetPipelineVariableFn != nil {
		return c.SetPipelineVariableFn(ns, slug, in)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.SetPipelineVariable; set SetPipelineVariableFn in your test")
	}
	return backend.PipelineVariable{}, nil
}

func (c *FakeClient) DeletePipelineVariable(ns, slug, key string) error {
	if c.DeletePipelineVariableFn != nil {
		return c.DeletePipelineVariableFn(ns, slug, key)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.DeletePipelineVariable; set DeletePipelineVariableFn in your test")
	}
	return nil
}

func (c *FakeClient) ListCommits(ns, slug, branch string, limit int) ([]backend.Commit, error) {
	if c.ListCommitsFn != nil {
		return c.ListCommitsFn(ns, slug, branch, limit)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListCommits; set ListCommitsFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) GetCommit(ns, slug, hash string) (backend.Commit, error) {
	if c.GetCommitFn != nil {
		return c.GetCommitFn(ns, slug, hash)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.GetCommit; set GetCommitFn in your test")
	}
	return backend.Commit{}, nil
}

func (c *FakeClient) ListPRComments(ns, slug string, id int) ([]backend.PRComment, error) {
	if c.ListPRCommentsFn != nil {
		return c.ListPRCommentsFn(ns, slug, id)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListPRComments; set ListPRCommentsFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) AddPRComment(ns, slug string, id int, in backend.AddPRCommentInput) (backend.PRComment, error) {
	if c.AddPRCommentFn != nil {
		return c.AddPRCommentFn(ns, slug, id, in)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.AddPRComment; set AddPRCommentFn in your test")
	}
	return backend.PRComment{}, nil
}

func (c *FakeClient) ListCommitStatuses(ns, slug, hash string) ([]backend.CommitStatus, error) {
	if c.ListCommitStatusesFn != nil {
		return c.ListCommitStatusesFn(ns, slug, hash)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListCommitStatuses; set ListCommitStatusesFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) ListWebhooks(ns, slug string) ([]backend.Webhook, error) {
	if c.ListWebhooksFn != nil {
		return c.ListWebhooksFn(ns, slug)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListWebhooks; set ListWebhooksFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) GetWebhook(ns, slug, id string) (backend.Webhook, error) {
	if c.GetWebhookFn != nil {
		return c.GetWebhookFn(ns, slug, id)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.GetWebhook; set GetWebhookFn in your test")
	}
	return backend.Webhook{}, nil
}

func (c *FakeClient) CreateWebhook(ns, slug string, in backend.CreateWebhookInput) (backend.Webhook, error) {
	if c.CreateWebhookFn != nil {
		return c.CreateWebhookFn(ns, slug, in)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.CreateWebhook; set CreateWebhookFn in your test")
	}
	return backend.Webhook{}, nil
}

func (c *FakeClient) DeleteWebhook(ns, slug, id string) error {
	if c.DeleteWebhookFn != nil {
		return c.DeleteWebhookFn(ns, slug, id)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.DeleteWebhook; set DeleteWebhookFn in your test")
	}
	return nil
}

func (c *FakeClient) ListWorkspaces(limit int) ([]backend.Workspace, error) {
	if c.ListWorkspacesFn != nil {
		return c.ListWorkspacesFn(limit)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListWorkspaces; set ListWorkspacesFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) ListProjects(workspace string, limit int) ([]backend.Project, error) {
	if c.ListProjectsFn != nil {
		return c.ListProjectsFn(workspace, limit)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListProjects; set ListProjectsFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) ListIssues(ns, slug, state string, limit int) ([]backend.Issue, error) {
	if c.ListIssuesFn != nil {
		return c.ListIssuesFn(ns, slug, state, limit)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListIssues; set ListIssuesFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) GetIssue(ns, slug string, id int) (backend.Issue, error) {
	if c.GetIssueFn != nil {
		return c.GetIssueFn(ns, slug, id)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.GetIssue; set GetIssueFn in your test")
	}
	return backend.Issue{}, nil
}

func (c *FakeClient) CreateIssue(ns, slug string, in backend.CreateIssueInput) (backend.Issue, error) {
	if c.CreateIssueFn != nil {
		return c.CreateIssueFn(ns, slug, in)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.CreateIssue; set CreateIssueFn in your test")
	}
	return backend.Issue{}, nil
}

func (c *FakeClient) UpdateIssue(ns, slug string, id int, in backend.UpdateIssueInput) (backend.Issue, error) {
	if c.UpdateIssueFn != nil {
		return c.UpdateIssueFn(ns, slug, id, in)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.UpdateIssue; set UpdateIssueFn in your test")
	}
	return backend.Issue{}, nil
}

func (c *FakeClient) ReopenIssue(ns, slug string, id int) error {
	if c.ReopenIssueFn != nil {
		return c.ReopenIssueFn(ns, slug, id)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ReopenIssue; set ReopenIssueFn in your test")
	}
	return nil
}

func (c *FakeClient) AssignIssue(ns, slug string, id int, assignee string) error {
	if c.AssignIssueFn != nil {
		return c.AssignIssueFn(ns, slug, id, assignee)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.AssignIssue; set AssignIssueFn in your test")
	}
	return nil
}

func (c *FakeClient) ListIssueComments(ns, slug string, id int) ([]backend.IssueComment, error) {
	if c.ListIssueCommentsFn != nil {
		return c.ListIssueCommentsFn(ns, slug, id)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListIssueComments; set ListIssueCommentsFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) AddIssueComment(ns, slug string, id int, body string) (backend.IssueComment, error) {
	if c.AddIssueCommentFn != nil {
		return c.AddIssueCommentFn(ns, slug, id, body)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.AddIssueComment; set AddIssueCommentFn in your test")
	}
	return backend.IssueComment{}, nil
}

func (c *FakeClient) EditIssueComment(ns, slug string, id, commentID int, body string) (backend.IssueComment, error) {
	if c.EditIssueCommentFn != nil {
		return c.EditIssueCommentFn(ns, slug, id, commentID, body)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.EditIssueComment; set EditIssueCommentFn in your test")
	}
	return backend.IssueComment{}, nil
}

func (c *FakeClient) DeleteIssueComment(ns, slug string, id, commentID int) error {
	if c.DeleteIssueCommentFn != nil {
		return c.DeleteIssueCommentFn(ns, slug, id, commentID)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.DeleteIssueComment; set DeleteIssueCommentFn in your test")
	}
	return nil
}

// DefaultReviewers defaults to "no defaults configured" (nil, nil) when the
// Fn is unset — unlike most FakeClient methods which Fatalf on unset Fn.
// Rationale: default-reviewers is a passive lookup automatically triggered
// by `pr create`, not an action a test deliberately exercises. Most pr-create
// tests don't care about reviewers and shouldn't be forced to set a stub
// just to silence the auto-fetch path. Tests that DO assert on the lookup
// (see create_default_reviewers_test.go) wire DefaultReviewersFn explicitly.
func (c *FakeClient) DefaultReviewers(ns, slug, fromBranch, toBranch string) ([]backend.User, error) {
	if c.DefaultReviewersFn != nil {
		return c.DefaultReviewersFn(ns, slug, fromBranch, toBranch)
	}
	return nil, nil
}

func (c *FakeClient) ListBranchProtections(ns, slug string, limit int) ([]backend.BranchProtection, error) {
	if c.ListBranchProtectionsFn != nil {
		return c.ListBranchProtectionsFn(ns, slug, limit)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListBranchProtections; set ListBranchProtectionsFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) CreateBranchProtection(ns, slug string, in backend.CreateBranchProtectionInput) (backend.BranchProtection, error) {
	if c.CreateBranchProtectionFn != nil {
		return c.CreateBranchProtectionFn(ns, slug, in)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.CreateBranchProtection; set CreateBranchProtectionFn in your test")
	}
	return backend.BranchProtection{}, nil
}

func (c *FakeClient) DeleteBranchProtection(ns, slug string, id int) error {
	if c.DeleteBranchProtectionFn != nil {
		return c.DeleteBranchProtectionFn(ns, slug, id)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.DeleteBranchProtection; set DeleteBranchProtectionFn in your test")
	}
	return nil
}

func (c *FakeClient) SearchCode(workspace, query string, limit int) ([]backend.CodeSearchHit, error) {
	if c.SearchCodeFn != nil {
		return c.SearchCodeFn(workspace, query, limit)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.SearchCode; set SearchCodeFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) GetFileContent(ns, slug, ref, path string) ([]byte, error) {
	if c.GetFileContentFn != nil {
		return c.GetFileContentFn(ns, slug, ref, path)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.GetFileContent; set GetFileContentFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) ListTree(ns, slug, ref, path string) ([]backend.TreeEntry, error) {
	if c.ListTreeFn != nil {
		return c.ListTreeFn(ns, slug, ref, path)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListTree; set ListTreeFn in your test")
	}
	return nil, nil
}

// ── CodeInsightsClient ───────────────────────────────────────────────────────

func (c *FakeClient) ListReports(project, slug, hash string) ([]backend.CodeInsightsReport, error) {
	if c.ListReportsFn != nil {
		return c.ListReportsFn(project, slug, hash)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListReports; set ListReportsFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) GetReport(project, slug, hash, key string) (backend.CodeInsightsReport, error) {
	if c.GetReportFn != nil {
		return c.GetReportFn(project, slug, hash, key)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.GetReport; set GetReportFn in your test")
	}
	return backend.CodeInsightsReport{}, nil
}

func (c *FakeClient) SetReport(project, slug, hash, key string, in backend.CodeInsightsReportInput) (backend.CodeInsightsReport, error) {
	if c.SetReportFn != nil {
		return c.SetReportFn(project, slug, hash, key, in)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.SetReport; set SetReportFn in your test")
	}
	return backend.CodeInsightsReport{}, nil
}

func (c *FakeClient) DeleteReport(project, slug, hash, key string) error {
	if c.DeleteReportFn != nil {
		return c.DeleteReportFn(project, slug, hash, key)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.DeleteReport; set DeleteReportFn in your test")
	}
	return nil
}

func (c *FakeClient) ListAnnotations(project, slug, hash, key string) ([]backend.CodeInsightsAnnotation, error) {
	if c.ListAnnotationsFn != nil {
		return c.ListAnnotationsFn(project, slug, hash, key)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListAnnotations; set ListAnnotationsFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) AddAnnotations(project, slug, hash, key string, in []backend.CodeInsightsAnnotationInput) error {
	if c.AddAnnotationsFn != nil {
		return c.AddAnnotationsFn(project, slug, hash, key, in)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.AddAnnotations; set AddAnnotationsFn in your test")
	}
	return nil
}

func (c *FakeClient) DeleteAnnotations(project, slug, hash, key string) error {
	if c.DeleteAnnotationsFn != nil {
		return c.DeleteAnnotationsFn(project, slug, hash, key)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.DeleteAnnotations; set DeleteAnnotationsFn in your test")
	}
	return nil
}

func (c *FakeClient) SetMergeCheck(project, slug, key string, in backend.MergeCheckInput) error {
	if c.SetMergeCheckFn != nil {
		return c.SetMergeCheckFn(project, slug, key, in)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.SetMergeCheck; set SetMergeCheckFn in your test")
	}
	return nil
}

func (c *FakeClient) GetMergeCheck(project, slug, key string) (backend.MergeCheck, error) {
	if c.GetMergeCheckFn != nil {
		return c.GetMergeCheckFn(project, slug, key)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.GetMergeCheck; set GetMergeCheckFn in your test")
	}
	return backend.MergeCheck{}, nil
}

func (c *FakeClient) DeleteMergeCheck(project, slug, key string) error {
	if c.DeleteMergeCheckFn != nil {
		return c.DeleteMergeCheckFn(project, slug, key)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.DeleteMergeCheck; set DeleteMergeCheckFn in your test")
	}
	return nil
}
