package testhelpers

import (
	"context"
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
	ListReposFn    func(ns string, limit int) ([]backend.Repository, error)
	GetRepoFn      func(ns, slug string) (backend.Repository, error)
	CreateRepoFn   func(ns string, in backend.CreateRepoInput) (backend.Repository, error)
	DeleteRepoFn   func(ns, slug string) error
	RenameRepoFn   func(ns, slug, newName string) (backend.Repository, error)
	ForkRepoFn     func(ns, slug string, in backend.ForkRepoInput) (backend.Repository, error)
	TransferRepoFn func(ns, slug, target string) (backend.Repository, error)

	// PR methods
	ListPRsFn          func(ns, slug, state string, limit int) ([]backend.PullRequest, error)
	GetPRFn            func(ns, slug string, id int) (backend.PullRequest, error)
	CreatePRFn         func(ns, slug string, in backend.CreatePRInput) (backend.PullRequest, error)
	MergePRFn          func(ns, slug string, id int, in backend.MergePRInput) (backend.PullRequest, error)
	EnableAutoMergeFn  func(ns, slug string, id int, strategy string) error
	DisableAutoMergeFn func(ns, slug string, id int) error
	ApprovePRFn        func(ns, slug string, id int) error
	GetPRDiffFn        func(ns, slug string, id int) (string, error)

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
	SubmitReviewFn  func(ns, slug string, id int, in backend.SubmitReviewInput) error

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

	// Commit comment methods
	ListCommitCommentsFn  func(ns, slug, hash string, limit int) ([]backend.CommitComment, error)
	AddCommitCommentFn    func(ns, slug, hash string, in backend.AddCommitCommentInput) (backend.CommitComment, error)
	EditCommitCommentFn   func(ns, slug, hash string, commentID int, body string) (backend.CommitComment, error)
	DeleteCommitCommentFn func(ns, slug, hash string, commentID int) error

	// PR comment methods
	ListPRCommentsFn  func(ns, slug string, id int) ([]backend.PRComment, error)
	AddPRCommentFn    func(ns, slug string, id int, in backend.AddPRCommentInput) (backend.PRComment, error)
	EditPRCommentFn   func(ns, slug string, id, commentID int, body string) (backend.PRComment, error)
	DeletePRCommentFn func(ns, slug string, id, commentID int) error

	// PR comment reaction methods (Server-only; satisfies backend.CommentReactor when set)
	ListCommentReactionsFn  func(ns, slug string, prID, commentID int) ([]backend.CommentReaction, error)
	AddCommentReactionFn    func(ns, slug string, prID, commentID int, emoji string) error
	RemoveCommentReactionFn func(ns, slug string, prID, commentID int, emoji string) error

	// PR activity
	GetPRActivityFn func(ns, slug string, id int, limit int) ([]backend.PRActivityEvent, error)

	// Commit status methods
	ListCommitStatusesFn func(ns, slug, hash string) ([]backend.CommitStatus, error)
	ReportCommitStatusFn func(ns, slug, hash string, input backend.CommitStatusInput) (backend.CommitStatus, error)

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

	// GHP methods
	UpdatePRBranchFn func(ns, slug string, prID int) error
	ListMyPRsFn      func(ns, slug string) ([]backend.MyPREntry, error)

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

	// Deployment methods (Cloud-only; satisfies backend.DeploymentClient when set)
	ListDeploymentsFn   func(ns, slug string, limit int) ([]backend.Deployment, error)
	GetDeploymentFn     func(ns, slug, uuid string) (backend.Deployment, error)
	ListEnvironmentsFn  func(ns, slug string) ([]backend.Environment, error)
	CreateEnvironmentFn func(ns, slug string, in backend.CreateEnvironmentInput) (backend.Environment, error)
	DeleteEnvironmentFn func(ns, slug, uuid string) error
	ListEnvVariablesFn  func(ns, slug, envUUID string) ([]backend.EnvVariable, error)
	SetEnvVariableFn    func(ns, slug, envUUID string, in backend.EnvVariableInput) (backend.EnvVariable, error)
	DeleteEnvVariableFn func(ns, slug, envUUID, varUUID string) error

	// Workspace variable methods (Cloud-only; satisfies backend.WorkspaceVariableClient when set)
	ListWorkspaceVariablesFn  func(ns string) ([]backend.PipelineVariable, error)
	SetWorkspaceVariableFn    func(ns string, in backend.PipelineVariableInput) (backend.PipelineVariable, error)
	DeleteWorkspaceVariableFn func(ns, key string) error

	// Permissions methods (Server-only; satisfies backend.PermissionsClient when set)
	ListProjectPermissionsFn  func(ctx context.Context, project string) ([]backend.PermissionGrant, error)
	GrantProjectPermissionFn  func(ctx context.Context, project string, subject backend.PermissionSubject, perm string) error
	RevokeProjectPermissionFn func(ctx context.Context, project string, subject backend.PermissionSubject) error
	ListRepoPermissionsFn     func(ctx context.Context, project, slug string) ([]backend.PermissionGrant, error)
	GrantRepoPermissionFn     func(ctx context.Context, project, slug string, subject backend.PermissionSubject, perm string) error
	RevokeRepoPermissionFn    func(ctx context.Context, project, slug string, subject backend.PermissionSubject) error

	// Admin methods (Server-only; satisfies backend.AdminClient when set)
	RotateSecretsFn    func() error
	GetLoggingConfigFn func() (backend.LoggingConfig, error)
	SetLoggingConfigFn func(in backend.LoggingConfigInput) error

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

	// Branch rule methods (Cloud-only; satisfies backend.BranchRuleClient when set)
	ListBranchRulesFn  func(ns, slug string) ([]backend.BranchRule, error)
	AddBranchRuleFn    func(ns, slug string, input backend.BranchRuleInput) (backend.BranchRule, error)
	DeleteBranchRuleFn func(ns, slug string, id int) error

	// Deploy key methods (both backends; satisfies backend.DeployKeyClient when set)
	ListDeployKeysFn  func(ns, slug string) ([]backend.DeployKey, error)
	AddDeployKeyFn    func(ns, slug string, input backend.DeployKeyInput) (backend.DeployKey, error)
	DeleteDeployKeyFn func(ns, slug string, id int) error

	// Commit file methods (both backends; satisfies backend.CommitFileClient when set)
	ListCommitFilesFn func(ns, slug, hash string) ([]backend.DiffStatEntry, error)

	// SSH key methods (Cloud-only; satisfies backend.SSHKeyClient when set)
	ListSSHKeysFn  func() ([]backend.SSHKey, error)
	AddSSHKeyFn    func(input backend.SSHKeyInput) (backend.SSHKey, error)
	DeleteSSHKeyFn func(id int) error

	// Default reviewer CRUD methods (both backends; satisfies backend.DefaultReviewerClient when set)
	ListDefaultReviewersFn  func(ns, slug string) ([]backend.DefaultReviewer, error)
	AddDefaultReviewerFn    func(ns, slug, userSlug string) error
	RemoveDefaultReviewerFn func(ns, slug, userSlug string) error

	// Pipeline trigger methods (Cloud-only; satisfies backend.PipelineTriggerClient when set)
	TriggerPipelineFn func(ns, slug string, input backend.PipelineTriggerInput) (backend.PipelineTriggerResult, error)

	// Pipeline schedule methods (Cloud-only; satisfies backend.PipelineScheduleClient when set)
	ListPipelineSchedulesFn  func(ns, slug string) ([]backend.PipelineSchedule, error)
	CreatePipelineScheduleFn func(ns, slug string, input backend.PipelineScheduleInput) (backend.PipelineSchedule, error)
	DeletePipelineScheduleFn func(ns, slug, uuid string) error

	// Diff methods (both backends; satisfies backend.DiffClient when set)
	GetDiffFn     func(ns, slug, from, to string) (string, error)
	GetDiffStatFn func(ns, slug, from, to string) (backend.DiffStat, error)

	// PR commit methods (both backends; satisfies backend.PRCommitClient when set)
	ListPRCommitsFn func(ns, slug string, prID int) ([]backend.Commit, error)

	// PR file methods (both backends; satisfies backend.PRFileClient when set)
	ListPRFilesFn func(ns, slug string, prID int) ([]backend.DiffStatEntry, error)

	// Repo watcher methods (both backends; satisfies backend.RepoWatcherClient when set)
	ListRepoWatchersFn func(ns, slug string) ([]backend.User, error)
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

func (c *FakeClient) TransferRepo(ns, slug, target string) (backend.Repository, error) {
	if c.TransferRepoFn != nil {
		return c.TransferRepoFn(ns, slug, target)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.TransferRepo; set TransferRepoFn in your test")
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

func (c *FakeClient) EnableAutoMerge(ns, slug string, id int, strategy string) error {
	if c.EnableAutoMergeFn != nil {
		return c.EnableAutoMergeFn(ns, slug, id, strategy)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.EnableAutoMerge; set EnableAutoMergeFn in your test")
	}
	return nil
}

func (c *FakeClient) DisableAutoMerge(ns, slug string, id int) error {
	if c.DisableAutoMergeFn != nil {
		return c.DisableAutoMergeFn(ns, slug, id)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.DisableAutoMerge; set DisableAutoMergeFn in your test")
	}
	return nil
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

func (c *FakeClient) SubmitReview(ns, slug string, id int, in backend.SubmitReviewInput) error {
	if c.SubmitReviewFn != nil {
		return c.SubmitReviewFn(ns, slug, id, in)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.SubmitReview; set SubmitReviewFn in your test")
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

func (c *FakeClient) ListCommitComments(ns, slug, hash string, limit int) ([]backend.CommitComment, error) {
	if c.ListCommitCommentsFn != nil {
		return c.ListCommitCommentsFn(ns, slug, hash, limit)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListCommitComments; set ListCommitCommentsFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) AddCommitComment(ns, slug, hash string, in backend.AddCommitCommentInput) (backend.CommitComment, error) {
	if c.AddCommitCommentFn != nil {
		return c.AddCommitCommentFn(ns, slug, hash, in)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.AddCommitComment; set AddCommitCommentFn in your test")
	}
	return backend.CommitComment{}, nil
}

func (c *FakeClient) EditCommitComment(ns, slug, hash string, commentID int, body string) (backend.CommitComment, error) {
	if c.EditCommitCommentFn != nil {
		return c.EditCommitCommentFn(ns, slug, hash, commentID, body)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.EditCommitComment; set EditCommitCommentFn in your test")
	}
	return backend.CommitComment{}, nil
}

func (c *FakeClient) DeleteCommitComment(ns, slug, hash string, commentID int) error {
	if c.DeleteCommitCommentFn != nil {
		return c.DeleteCommitCommentFn(ns, slug, hash, commentID)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.DeleteCommitComment; set DeleteCommitCommentFn in your test")
	}
	return nil
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

func (c *FakeClient) EditPRComment(ns, slug string, id, commentID int, body string) (backend.PRComment, error) {
	if c.EditPRCommentFn != nil {
		return c.EditPRCommentFn(ns, slug, id, commentID, body)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.EditPRComment; set EditPRCommentFn in your test")
	}
	return backend.PRComment{}, nil
}

func (c *FakeClient) DeletePRComment(ns, slug string, id, commentID int) error {
	if c.DeletePRCommentFn != nil {
		return c.DeletePRCommentFn(ns, slug, id, commentID)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.DeletePRComment; set DeletePRCommentFn in your test")
	}
	return nil
}

func (c *FakeClient) GetPRActivity(ns, slug string, id int, limit int) ([]backend.PRActivityEvent, error) {
	if c.GetPRActivityFn != nil {
		return c.GetPRActivityFn(ns, slug, id, limit)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.GetPRActivity; set GetPRActivityFn in your test")
	}
	return nil, nil
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

func (c *FakeClient) ReportCommitStatus(ns, slug, hash string, input backend.CommitStatusInput) (backend.CommitStatus, error) {
	if c.ReportCommitStatusFn != nil {
		return c.ReportCommitStatusFn(ns, slug, hash, input)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ReportCommitStatus; set ReportCommitStatusFn in your test")
	}
	return backend.CommitStatus{}, nil
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

func (c *FakeClient) UpdatePRBranch(ns, slug string, prID int) error {
	if c.UpdatePRBranchFn != nil {
		return c.UpdatePRBranchFn(ns, slug, prID)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.UpdatePRBranch; set UpdatePRBranchFn in your test")
	}
	return nil
}

func (c *FakeClient) ListMyPRs(ns, slug string) ([]backend.MyPREntry, error) {
	if c.ListMyPRsFn != nil {
		return c.ListMyPRsFn(ns, slug)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListMyPRs; set ListMyPRsFn in your test")
	}
	return nil, nil
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

// ── WorkspaceVariableClient ──────────────────────────────────────────────────

func (c *FakeClient) ListWorkspaceVariables(ns string) ([]backend.PipelineVariable, error) {
	if c.ListWorkspaceVariablesFn != nil {
		return c.ListWorkspaceVariablesFn(ns)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListWorkspaceVariables; set ListWorkspaceVariablesFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) SetWorkspaceVariable(ns string, in backend.PipelineVariableInput) (backend.PipelineVariable, error) {
	if c.SetWorkspaceVariableFn != nil {
		return c.SetWorkspaceVariableFn(ns, in)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.SetWorkspaceVariable; set SetWorkspaceVariableFn in your test")
	}
	return backend.PipelineVariable{}, nil
}

func (c *FakeClient) DeleteWorkspaceVariable(ns, key string) error {
	if c.DeleteWorkspaceVariableFn != nil {
		return c.DeleteWorkspaceVariableFn(ns, key)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.DeleteWorkspaceVariable; set DeleteWorkspaceVariableFn in your test")
	}
	return nil
}

// ── DeploymentClient ─────────────────────────────────────────────────────────

func (c *FakeClient) ListDeployments(ns, slug string, limit int) ([]backend.Deployment, error) {
	if c.ListDeploymentsFn != nil {
		return c.ListDeploymentsFn(ns, slug, limit)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListDeployments; set ListDeploymentsFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) GetDeployment(ns, slug, uuid string) (backend.Deployment, error) {
	if c.GetDeploymentFn != nil {
		return c.GetDeploymentFn(ns, slug, uuid)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.GetDeployment; set GetDeploymentFn in your test")
	}
	return backend.Deployment{}, nil
}

func (c *FakeClient) ListEnvironments(ns, slug string) ([]backend.Environment, error) {
	if c.ListEnvironmentsFn != nil {
		return c.ListEnvironmentsFn(ns, slug)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListEnvironments; set ListEnvironmentsFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) CreateEnvironment(ns, slug string, in backend.CreateEnvironmentInput) (backend.Environment, error) {
	if c.CreateEnvironmentFn != nil {
		return c.CreateEnvironmentFn(ns, slug, in)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.CreateEnvironment; set CreateEnvironmentFn in your test")
	}
	return backend.Environment{}, nil
}

func (c *FakeClient) DeleteEnvironment(ns, slug, uuid string) error {
	if c.DeleteEnvironmentFn != nil {
		return c.DeleteEnvironmentFn(ns, slug, uuid)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.DeleteEnvironment; set DeleteEnvironmentFn in your test")
	}
	return nil
}

func (c *FakeClient) ListEnvVariables(ns, slug, envUUID string) ([]backend.EnvVariable, error) {
	if c.ListEnvVariablesFn != nil {
		return c.ListEnvVariablesFn(ns, slug, envUUID)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListEnvVariables; set ListEnvVariablesFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) SetEnvVariable(ns, slug, envUUID string, in backend.EnvVariableInput) (backend.EnvVariable, error) {
	if c.SetEnvVariableFn != nil {
		return c.SetEnvVariableFn(ns, slug, envUUID, in)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.SetEnvVariable; set SetEnvVariableFn in your test")
	}
	return backend.EnvVariable{}, nil
}

func (c *FakeClient) DeleteEnvVariable(ns, slug, envUUID, varUUID string) error {
	if c.DeleteEnvVariableFn != nil {
		return c.DeleteEnvVariableFn(ns, slug, envUUID, varUUID)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.DeleteEnvVariable; set DeleteEnvVariableFn in your test")
	}
	return nil
}

// ── AdminClient ──────────────────────────────────────────────────────────────

func (c *FakeClient) RotateSecrets() error {
	if c.RotateSecretsFn != nil {
		return c.RotateSecretsFn()
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.RotateSecrets; set RotateSecretsFn in your test")
	}
	return nil
}

func (c *FakeClient) GetLoggingConfig() (backend.LoggingConfig, error) {
	if c.GetLoggingConfigFn != nil {
		return c.GetLoggingConfigFn()
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.GetLoggingConfig; set GetLoggingConfigFn in your test")
	}
	return backend.LoggingConfig{}, nil
}

func (c *FakeClient) SetLoggingConfig(in backend.LoggingConfigInput) error {
	if c.SetLoggingConfigFn != nil {
		return c.SetLoggingConfigFn(in)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.SetLoggingConfig; set SetLoggingConfigFn in your test")
	}
	return nil
}

// ── PermissionsClient ────────────────────────────────────────────────────────

func (c *FakeClient) ListProjectPermissions(ctx context.Context, project string) ([]backend.PermissionGrant, error) {
	if c.ListProjectPermissionsFn != nil {
		return c.ListProjectPermissionsFn(ctx, project)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListProjectPermissions; set ListProjectPermissionsFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) GrantProjectPermission(ctx context.Context, project string, subject backend.PermissionSubject, perm string) error {
	if c.GrantProjectPermissionFn != nil {
		return c.GrantProjectPermissionFn(ctx, project, subject, perm)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.GrantProjectPermission; set GrantProjectPermissionFn in your test")
	}
	return nil
}

func (c *FakeClient) RevokeProjectPermission(ctx context.Context, project string, subject backend.PermissionSubject) error {
	if c.RevokeProjectPermissionFn != nil {
		return c.RevokeProjectPermissionFn(ctx, project, subject)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.RevokeProjectPermission; set RevokeProjectPermissionFn in your test")
	}
	return nil
}

func (c *FakeClient) ListRepoPermissions(ctx context.Context, project, slug string) ([]backend.PermissionGrant, error) {
	if c.ListRepoPermissionsFn != nil {
		return c.ListRepoPermissionsFn(ctx, project, slug)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListRepoPermissions; set ListRepoPermissionsFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) GrantRepoPermission(ctx context.Context, project, slug string, subject backend.PermissionSubject, perm string) error {
	if c.GrantRepoPermissionFn != nil {
		return c.GrantRepoPermissionFn(ctx, project, slug, subject, perm)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.GrantRepoPermission; set GrantRepoPermissionFn in your test")
	}
	return nil
}

func (c *FakeClient) RevokeRepoPermission(ctx context.Context, project, slug string, subject backend.PermissionSubject) error {
	if c.RevokeRepoPermissionFn != nil {
		return c.RevokeRepoPermissionFn(ctx, project, slug, subject)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.RevokeRepoPermission; set RevokeRepoPermissionFn in your test")
	}
	return nil
}

// ── BranchRuleClient ─────────────────────────────────────────────────────────

func (c *FakeClient) ListBranchRules(ns, slug string) ([]backend.BranchRule, error) {
	if c.ListBranchRulesFn != nil {
		return c.ListBranchRulesFn(ns, slug)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListBranchRules; set ListBranchRulesFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) AddBranchRule(ns, slug string, input backend.BranchRuleInput) (backend.BranchRule, error) {
	if c.AddBranchRuleFn != nil {
		return c.AddBranchRuleFn(ns, slug, input)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.AddBranchRule; set AddBranchRuleFn in your test")
	}
	return backend.BranchRule{}, nil
}

func (c *FakeClient) DeleteBranchRule(ns, slug string, id int) error {
	if c.DeleteBranchRuleFn != nil {
		return c.DeleteBranchRuleFn(ns, slug, id)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.DeleteBranchRule; set DeleteBranchRuleFn in your test")
	}
	return nil
}

func (c *FakeClient) ListDeployKeys(ns, slug string) ([]backend.DeployKey, error) {
	if c.ListDeployKeysFn != nil {
		return c.ListDeployKeysFn(ns, slug)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListDeployKeys; set ListDeployKeysFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) AddDeployKey(ns, slug string, input backend.DeployKeyInput) (backend.DeployKey, error) {
	if c.AddDeployKeyFn != nil {
		return c.AddDeployKeyFn(ns, slug, input)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.AddDeployKey; set AddDeployKeyFn in your test")
	}
	return backend.DeployKey{}, nil
}

func (c *FakeClient) DeleteDeployKey(ns, slug string, id int) error {
	if c.DeleteDeployKeyFn != nil {
		return c.DeleteDeployKeyFn(ns, slug, id)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.DeleteDeployKey; set DeleteDeployKeyFn in your test")
	}
	return nil
}

// ── PipelineTriggerClient ────────────────────────────────────────────────────

func (c *FakeClient) TriggerPipeline(ns, slug string, input backend.PipelineTriggerInput) (backend.PipelineTriggerResult, error) {
	if c.TriggerPipelineFn != nil {
		return c.TriggerPipelineFn(ns, slug, input)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.TriggerPipeline; set TriggerPipelineFn in your test")
	}
	return backend.PipelineTriggerResult{}, nil
}

// ── PipelineScheduleClient ───────────────────────────────────────────────────

func (c *FakeClient) ListPipelineSchedules(ns, slug string) ([]backend.PipelineSchedule, error) {
	if c.ListPipelineSchedulesFn != nil {
		return c.ListPipelineSchedulesFn(ns, slug)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListPipelineSchedules; set ListPipelineSchedulesFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) CreatePipelineSchedule(ns, slug string, input backend.PipelineScheduleInput) (backend.PipelineSchedule, error) {
	if c.CreatePipelineScheduleFn != nil {
		return c.CreatePipelineScheduleFn(ns, slug, input)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.CreatePipelineSchedule; set CreatePipelineScheduleFn in your test")
	}
	return backend.PipelineSchedule{}, nil
}

func (c *FakeClient) DeletePipelineSchedule(ns, slug, uuid string) error {
	if c.DeletePipelineScheduleFn != nil {
		return c.DeletePipelineScheduleFn(ns, slug, uuid)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.DeletePipelineSchedule; set DeletePipelineScheduleFn in your test")
	}
	return nil
}

// ── DiffClient ───────────────────────────────────────────────────────────────

func (c *FakeClient) GetDiff(ns, slug, from, to string) (string, error) {
	if c.GetDiffFn != nil {
		return c.GetDiffFn(ns, slug, from, to)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.GetDiff; set GetDiffFn in your test")
	}
	return "", nil
}

func (c *FakeClient) GetDiffStat(ns, slug, from, to string) (backend.DiffStat, error) {
	if c.GetDiffStatFn != nil {
		return c.GetDiffStatFn(ns, slug, from, to)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.GetDiffStat; set GetDiffStatFn in your test")
	}
	return backend.DiffStat{}, nil
}

func (c *FakeClient) ListDefaultReviewers(ns, slug string) ([]backend.DefaultReviewer, error) {
	if c.ListDefaultReviewersFn != nil {
		return c.ListDefaultReviewersFn(ns, slug)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListDefaultReviewers; set ListDefaultReviewersFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) AddDefaultReviewer(ns, slug, userSlug string) error {
	if c.AddDefaultReviewerFn != nil {
		return c.AddDefaultReviewerFn(ns, slug, userSlug)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.AddDefaultReviewer; set AddDefaultReviewerFn in your test")
	}
	return nil
}

func (c *FakeClient) RemoveDefaultReviewer(ns, slug, userSlug string) error {
	if c.RemoveDefaultReviewerFn != nil {
		return c.RemoveDefaultReviewerFn(ns, slug, userSlug)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.RemoveDefaultReviewer; set RemoveDefaultReviewerFn in your test")
	}
	return nil
}

// ── SSHKeyClient ─────────────────────────────────────────────────────────────

func (c *FakeClient) ListSSHKeys() ([]backend.SSHKey, error) {
	if c.ListSSHKeysFn != nil {
		return c.ListSSHKeysFn()
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListSSHKeys; set ListSSHKeysFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) AddSSHKey(input backend.SSHKeyInput) (backend.SSHKey, error) {
	if c.AddSSHKeyFn != nil {
		return c.AddSSHKeyFn(input)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.AddSSHKey; set AddSSHKeyFn in your test")
	}
	return backend.SSHKey{}, nil
}

func (c *FakeClient) DeleteSSHKey(id int) error {
	if c.DeleteSSHKeyFn != nil {
		return c.DeleteSSHKeyFn(id)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.DeleteSSHKey; set DeleteSSHKeyFn in your test")
	}
	return nil
}

// ── CommitFileClient ─────────────────────────────────────────────────────────

func (c *FakeClient) ListCommitFiles(ns, slug, hash string) ([]backend.DiffStatEntry, error) {
	if c.ListCommitFilesFn != nil {
		return c.ListCommitFilesFn(ns, slug, hash)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListCommitFiles; set ListCommitFilesFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) ListPRCommits(ns, slug string, prID int) ([]backend.Commit, error) {
	if c.ListPRCommitsFn != nil {
		return c.ListPRCommitsFn(ns, slug, prID)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListPRCommits; set ListPRCommitsFn in your test")
	}
	return nil, nil
}

// ── PRFileClient ─────────────────────────────────────────────────────────────

func (c *FakeClient) ListPRFiles(ns, slug string, prID int) ([]backend.DiffStatEntry, error) {
	if c.ListPRFilesFn != nil {
		return c.ListPRFilesFn(ns, slug, prID)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListPRFiles; set ListPRFilesFn in your test")
	}
	return nil, nil
}

// ── RepoWatcherClient ─────────────────────────────────────────────────────────

func (c *FakeClient) ListRepoWatchers(ns, slug string) ([]backend.User, error) {
	if c.ListRepoWatchersFn != nil {
		return c.ListRepoWatchersFn(ns, slug)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListRepoWatchers; set ListRepoWatchersFn in your test")
	}
	return nil, nil
}
