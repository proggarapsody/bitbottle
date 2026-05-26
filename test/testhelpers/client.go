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
	ListReposFn            func(ns string, limit int) ([]backend.Repository, error)
	GetRepoFn              func(ns, slug string) (backend.Repository, error)
	CreateRepoFn           func(ns string, in backend.CreateRepoInput) (backend.Repository, error)
	DeleteRepoFn           func(ns, slug string) error
	RenameRepoFn           func(ns, slug, newName string) (backend.Repository, error)
	ForkRepoFn             func(ns, slug string, in backend.ForkRepoInput) (backend.Repository, error)
	TransferRepoFn         func(ns, slug, target string) (backend.Repository, error)
	SetRepoVisibilityFn    func(ns, slug string, isPrivate bool) error
	SetRepoDefaultBranchFn func(ns, slug, branch string) error

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
	UpdatePRFn        func(ns, slug string, id int, in backend.UpdatePRInput) (backend.PullRequest, error)
	DeclinePRFn       func(ns, slug string, id int) error
	UnapprovePRFn     func(ns, slug string, id int) error
	ReadyPRFn         func(ns, slug string, id int) error
	UnreadyPRFn       func(ns, slug string, id int) error
	RequestReviewFn   func(ns, slug string, id int, users []string) error
	RemoveReviewersFn func(ns, slug string, id int, users []string) error
	SubmitReviewFn    func(ns, slug string, id int, in backend.SubmitReviewInput) error

	// Pipeline methods (Cloud-only; satisfies backend.PipelineClient when set)
	ListPipelinesFn          func(ns, slug string, limit int) ([]backend.Pipeline, error)
	GetPipelineFn            func(ns, slug, uuid string) (backend.Pipeline, error)
	RunPipelineFn            func(ns, slug string, in backend.RunPipelineInput) (backend.Pipeline, error)
	StopPipelineFn           func(ws, slug, pipelineUUID string) error
	RerunPipelineFn          func(ns, slug, sourceUUID string) (backend.Pipeline, error)
	ListPipelineStepsFn      func(ns, slug, uuid string) ([]backend.PipelineStep, error)
	GetPipelineStepLogFn     func(ns, slug, pipelineUUID, stepUUID string) (io.ReadCloser, error)
	ListPipelineVariablesFn  func(ns, slug string) ([]backend.PipelineVariable, error)
	SetPipelineVariableFn    func(ns, slug string, in backend.PipelineVariableInput) (backend.PipelineVariable, error)
	DeletePipelineVariableFn func(ns, slug, key string) error

	// Pipeline artifact methods (Cloud-only; satisfies backend.PipelineArtifactClient when set)
	ListPipelineArtifactsFn    func(ws, slug, pipelineUUID, stepUUID string, limit int) ([]backend.PipelineArtifact, error)
	DownloadPipelineArtifactFn func(ws, slug, pipelineUUID, stepUUID, name string, out io.Writer) error

	// Commit methods
	ListCommitsFn      func(ns, slug, branch string, limit int) ([]backend.Commit, error)
	GetCommitFn        func(ns, slug, hash string) (backend.Commit, error)
	CherryPickCommitFn func(ns, slug string, in backend.CherryPickInput) (backend.Commit, error)

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

	// Workspace member methods (Cloud-only; satisfies backend.WorkspaceMemberClient when set)
	ListWorkspaceMembersFn func(workspace string, limit int) ([]backend.WorkspaceMember, error)

	// Workspace webhook methods (Cloud-only; satisfies backend.WorkspaceWebhookClient when set)
	ListWorkspaceWebhooksFn  func(workspace string) ([]backend.Webhook, error)
	CreateWorkspaceWebhookFn func(workspace string, in backend.CreateWebhookInput) (backend.Webhook, error)
	DeleteWorkspaceWebhookFn func(workspace, uuid string) error

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

	// Workspace pipeline variable methods (Cloud-only; satisfies backend.WorkspacePipelineVariableClient when set)
	ListWorkspacePipelineVariablesFn  func(workspace string) ([]backend.PipelineVariable, error)
	GetWorkspacePipelineVariableFn    func(workspace, uuid string) (backend.PipelineVariable, error)
	SetWorkspacePipelineVariableFn    func(workspace string, in backend.PipelineVariableInput) (backend.PipelineVariable, error)
	DeleteWorkspacePipelineVariableFn func(workspace, uuid string) error

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

	// Admin user methods (Server-only)
	ListAdminUsersFn func(filter string, limit int) ([]backend.AdminUser, error)
	RenameUserFn     func(slug, newSlug string) error
	ActivateUserFn   func(slug string) error
	DeactivateUserFn func(slug string) error

	// Admin system methods (Server-only)
	GetLicenseFn      func() (backend.AdminLicense, error)
	GetClusterNodesFn func() ([]backend.ClusterNode, error)

	// Admin mail server methods (Server-only)
	GetMailServerConfigFn func() (backend.MailServerConfig, error)
	SetMailServerConfigFn func(in backend.MailServerConfig) error

	// Admin banner methods (Server-only)
	GetBannerFn   func() (backend.BannerConfig, error)
	SetBannerFn   func(in backend.BannerConfig) error
	ClearBannerFn func() error

	// Source write methods (both backends; satisfies backend.SourceWriter when set)
	PutFileFn func(ns, slug, path string, in backend.PutFileInput) error

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

	// Branch model methods (Cloud-only; satisfies backend.BranchModelClient when set)
	GetBranchModelFn            func(ws, slug string) (backend.BranchModel, error)
	GetBranchModelSettingsFn    func(ws, slug string) (backend.BranchModelSettings, error)
	UpdateBranchModelSettingsFn func(ws, slug string, in backend.BranchModelSettingsInput) (backend.BranchModelSettings, error)

	// Branch rule methods (Cloud-only; satisfies backend.BranchRuleClient when set)
	ListBranchRulesFn  func(ns, slug string) ([]backend.BranchRule, error)
	AddBranchRuleFn    func(ns, slug string, input backend.BranchRuleInput) (backend.BranchRule, error)
	DeleteBranchRuleFn func(ns, slug string, id int) error
	UpdateBranchRuleFn func(ns, slug string, id int, in backend.UpdateBranchRuleInput) (backend.BranchRule, error)

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

	// Reviewer group methods (Server-only; satisfies backend.ReviewerGroupClient when set)
	ListReviewerGroupsFn  func(ns, slug string) ([]backend.ReviewerGroup, error)
	CreateReviewerGroupFn func(ns, slug string, in backend.CreateReviewerGroupInput) (backend.ReviewerGroup, error)
	DeleteReviewerGroupFn func(ns, slug string, id int) error

	// Pipeline trigger methods (Cloud-only; satisfies backend.PipelineTriggerClient when set)
	TriggerPipelineFn func(ns, slug string, input backend.PipelineTriggerInput) (backend.PipelineTriggerResult, error)

	// Pipeline schedule methods (Cloud-only; satisfies backend.PipelineScheduleClient when set)
	ListPipelineSchedulesFn  func(ns, slug string) ([]backend.PipelineSchedule, error)
	CreatePipelineScheduleFn func(ns, slug string, input backend.PipelineScheduleInput) (backend.PipelineSchedule, error)
	DeletePipelineScheduleFn func(ns, slug, uuid string) error

	// Pipeline cache methods (Cloud-only; satisfies backend.PipelineCacheClient when set)
	ListPipelineCachesFn  func(ns, slug string) ([]backend.PipelineCache, error)
	DeletePipelineCacheFn func(ns, slug, uuid string) error

	// Runner methods (Cloud-only; satisfies backend.RunnerClient when set)
	ListRunnersFn  func(workspace string) ([]backend.Runner, error)
	CreateRunnerFn func(workspace string, in backend.CreateRunnerInput) (backend.Runner, error)
	DeleteRunnerFn func(workspace, runnerUUID string) error

	// Audit methods (Cloud-only; satisfies backend.AuditClient when set)
	ListAuditLogFn func(workspace string, opts backend.AuditLogOpts) ([]backend.AuditEvent, error)

	// IPAllowlist methods (Cloud-only; satisfies backend.IPAllowlistClient when set)
	ListIPAllowlistsFn  func(workspace string) ([]backend.IPAllowlist, error)
	CreateIPAllowlistFn func(workspace string, in backend.CreateIPAllowlistInput) (backend.IPAllowlist, error)
	DeleteIPAllowlistFn func(workspace, uuid string) error

	// Diff methods (both backends; satisfies backend.DiffClient when set)
	GetDiffFn     func(ns, slug, from, to string) (string, error)
	GetDiffStatFn func(ns, slug, from, to string) (backend.DiffStat, error)

	// PR commit methods (both backends; satisfies backend.PRCommitClient when set)
	ListPRCommitsFn func(ns, slug string, prID int) ([]backend.Commit, error)

	// PR file methods (both backends; satisfies backend.PRFileClient when set)
	ListPRFilesFn func(ns, slug string, prID int) ([]backend.DiffStatEntry, error)

	// PR participant methods (both backends; satisfies backend.PRParticipantClient when set)
	ListPRParticipantsFn func(ns, slug string, prID int) ([]backend.PRParticipant, error)

	// PR participant update methods (Cloud-only; satisfies backend.PRParticipantUpdater when set)
	UpdatePRParticipantFn func(ns, slug string, prID int, accountID, state string) (backend.PRParticipant, error)

	// Repo watcher methods (both backends; satisfies backend.RepoWatcherClient when set)
	ListRepoWatchersFn func(ns, slug string) ([]backend.User, error)

	// Repo forks methods (both backends; satisfies backend.RepoForksLister when set)
	ListRepoForksFn func(ns, slug string, limit int) ([]backend.Repository, error)

	// Repo edit methods (both backends; satisfies backend.RepoEditor when set)
	EditRepoFn func(ns, slug string, in backend.EditRepoInput) (backend.Repository, error)

	// Repo PR settings methods (Server-only; satisfies backend.RepoPRSettingsClient when set)
	GetRepoPRSettingsFn    func(ns, slug string) (backend.RepoPRSettings, error)
	UpdateRepoPRSettingsFn func(ns, slug string, in backend.RepoPRSettingsInput) (backend.RepoPRSettings, error)

	// Snippet methods (Cloud-only; satisfies backend.SnippetClient when set)
	ListSnippetsFn         func(workspace string, limit int) ([]backend.Snippet, error)
	GetSnippetFn           func(workspace, id string) (backend.Snippet, error)
	CreateSnippetFn        func(workspace string, in backend.CreateSnippetInput) (backend.Snippet, error)
	DeleteSnippetFn        func(workspace, id string) error
	ListSnippetCommentsFn  func(workspace, snippetID string, limit int) ([]backend.SnippetComment, error)
	AddSnippetCommentFn    func(workspace, snippetID, body string) (backend.SnippetComment, error)
	DeleteSnippetCommentFn func(workspace, snippetID string, commentID int) error

	// Group methods (Server-only; satisfies backend.GroupClient when set)
	ListGroupsFn  func(filter string, limit int) ([]backend.Group, error)
	CreateGroupFn func(name string) (backend.Group, error)
	DeleteGroupFn func(name string) error

	// Group member methods (Server-only; satisfies backend.GroupMemberClient when set)
	ListGroupMembersFn  func(groupName string, limit int) ([]backend.GroupMember, error)
	AddGroupMemberFn    func(groupName, user string) error
	RemoveGroupMemberFn func(groupName, user string) error

	// Server project methods (Server-only; satisfies backend.ServerProjectClient when set)
	ListServerProjectsFn  func(filter string, limit int) ([]backend.ServerProject, error)
	GetServerProjectFn    func(key string) (backend.ServerProject, error)
	CreateServerProjectFn func(in backend.CreateServerProjectInput) (backend.ServerProject, error)
	UpdateServerProjectFn func(key string, in backend.UpdateServerProjectInput) (backend.ServerProject, error)
	DeleteServerProjectFn func(key string) error

	// PAT methods (Server-only; satisfies backend.PATClient when set)
	ListPATsFn  func(userSlug string, limit int) ([]backend.PAT, error)
	CreatePATFn func(userSlug string, in backend.CreatePATInput) (backend.PATWithSecret, error)
	RevokePATFn func(userSlug, tokenID string) error

	// IssueAttacher methods (Cloud-only; satisfies backend.IssueAttacher when set)
	ListIssueAttachmentsFn  func(ns, slug string, id int) ([]backend.IssueAttachment, error)
	DeleteIssueAttachmentFn func(ns, slug string, id int, filename string) error

	// IssueVoter methods (Cloud-only; satisfies backend.IssueVoter when set)
	VoteIssueFn   func(ns, slug string, id int) error
	UnvoteIssueFn func(ns, slug string, id int) error

	// IssueWatcher methods (Cloud-only; satisfies backend.IssueWatcher when set)
	WatchIssueFn   func(ns, slug string, id int) error
	UnwatchIssueFn func(ns, slug string, id int) error

	// IssueActivityClient methods (Cloud-only; satisfies backend.IssueActivityClient when set)
	ListIssueActivityFn func(ns, slug string, issueID int, limit int) ([]backend.IssueChange, error)

	// WorkspaceSearcher methods (Cloud-only; satisfies backend.WorkspaceSearcher when set)
	SearchWorkspacesFn func(opts backend.WorkspaceSearchOpts) ([]backend.Workspace, error)

	// WorkspaceProjectPermsClient methods (Cloud-only; satisfies backend.WorkspaceProjectPermsClient when set)
	ListWorkspaceProjectPermsFn  func(workspace, projectKey string, limit int) ([]backend.WorkspaceProjectPerm, error)
	GrantWorkspaceProjectPermFn  func(workspace, projectKey string, in backend.WorkspaceProjectPermInput) error
	RevokeWorkspaceProjectPermFn func(workspace, projectKey, subjectSlug string, isGroup bool) error

	// WorkspaceProjectDefaultReviewerClient methods (Cloud-only; satisfies backend.WorkspaceProjectDefaultReviewerClient when set)
	ListProjectDefaultReviewersFn  func(workspace, projectKey string, limit int) ([]backend.ProjectDefaultReviewer, error)
	AddProjectDefaultReviewerFn    func(workspace, projectKey, accountID string) error
	RemoveProjectDefaultReviewerFn func(workspace, projectKey, accountID string) error

	// RepoLabelClient methods (both backends; satisfies backend.RepoLabelClient when set)
	ListRepoLabelsFn  func(ns, slug string) ([]backend.RepoLabel, error)
	CreateRepoLabelFn func(ns, slug string, in backend.CreateRepoLabelInput) (backend.RepoLabel, error)
	UpdateRepoLabelFn func(ns, slug string, id int, in backend.UpdateRepoLabelInput) (backend.RepoLabel, error)
	DeleteRepoLabelFn func(ns, slug string, id int) error

	// PipelineConfigClient methods (Cloud-only; satisfies backend.PipelineConfigClient when set)
	GetPipelinesConfigFn    func(ws, slug string) (backend.PipelineConfig, error)
	UpdatePipelinesConfigFn func(ws, slug string, in backend.PipelineConfig) (backend.PipelineConfig, error)

	// PipelineTestReportClient methods (Cloud-only)
	GetPipelineTestReportFn func(ws, slug, pipelineUUID, stepUUID string) (backend.PipelineTestReport, error)
	ListPipelineTestCasesFn func(ws, slug, pipelineUUID, stepUUID string, filter backend.TestCaseFilter) ([]backend.PipelineTestCase, error)

	// RefComparer methods (Cloud + Server/DC)
	CompareRefsFn func(ns, slug, base, head string, limit int) (backend.RefComparison, error)

	// RepoDownloadClient methods (Cloud-only; satisfies backend.RepoDownloadClient when set)
	ListRepoDownloadsFn    func(ns, slug string, limit int) ([]backend.RepoDownload, error)
	UploadRepoDownloadFn   func(ns, slug, name string, body io.Reader) (backend.RepoDownload, error)
	DownloadRepoDownloadFn func(ns, slug, name string, out io.Writer) error
	DeleteRepoDownloadFn   func(ns, slug, name string) error

	// MilestoneClient methods (Cloud-only; satisfies backend.MilestoneClient when set)
	ListMilestonesFn func(ns, slug string, limit int) ([]backend.Milestone, error)
	GetMilestoneFn   func(ns, slug string, id int) (backend.Milestone, error)

	// IssueVersionClient methods (Cloud-only; satisfies backend.IssueVersionClient when set)
	ListIssueVersionsFn  func(ns, slug string, limit int) ([]backend.IssueVersion, error)
	GetIssueVersionFn    func(ns, slug string, id int) (backend.IssueVersion, error)
	CreateIssueVersionFn func(ns, slug, name string) (backend.IssueVersion, error)
	DeleteIssueVersionFn func(ns, slug string, id int) error

	// CloudProjectClient methods (Cloud-only; satisfies backend.CloudProjectClient when set)
	CreateWorkspaceProjectFn func(ws string, input backend.CreateWorkspaceProjectInput) (backend.WorkspaceProject, error)
	GetWorkspaceProjectFn    func(ws, key string) (backend.WorkspaceProject, error)
	UpdateWorkspaceProjectFn func(ws, key string, input backend.UpdateWorkspaceProjectInput) (backend.WorkspaceProject, error)
	DeleteWorkspaceProjectFn func(ws, key string) error

	// MirrorClient methods (Server-only; satisfies backend.MirrorClient when set)
	ListMirrorServersFn func(limit int) ([]backend.MirrorServer, error)
	GetMirrorServerFn   func(id string) (backend.MirrorServer, error)
	ListMirroredReposFn func(mirrorID string, limit int) ([]backend.MirroredRepo, error)

	// WorkspacePermsClient methods (Cloud-only; satisfies backend.WorkspacePermsClient when set)
	ListWorkspaceMemberPermsFn func(ws string, limit int) ([]backend.WorkspaceMemberPerm, error)
	ListWorkspaceRepoPermsFn   func(ws string, limit int) ([]backend.WorkspaceRepoPerm, error)
	GrantWorkspacePermFn       func(ws, user, permission string) error
	RevokeWorkspacePermFn      func(ws, user string) error

	// PipelineSSHKeyPairClient methods (Cloud-only; satisfies backend.PipelineSSHKeyPairClient when set)
	GetPipelineSSHKeyPairFn        func(ns, slug string) (backend.PipelineSSHKeyPair, error)
	RegeneratePipelineSSHKeyPairFn func(ns, slug string, bits int) (backend.PipelineSSHKeyPair, error)

	// PipelineKnownHostsClient methods (Cloud-only; satisfies backend.PipelineKnownHostsClient when set)
	ListPipelineKnownHostsFn  func(ns, slug string) ([]backend.PipelineKnownHost, error)
	GetPipelineKnownHostFn    func(ns, slug, uuid string) (backend.PipelineKnownHost, error)
	AddPipelineKnownHostFn    func(ns, slug string, in backend.PipelineKnownHostInput) (backend.PipelineKnownHost, error)
	DeletePipelineKnownHostFn func(ns, slug, uuid string) error
}

// ── Compile-time interface assertions ─────────────────────────────────────────
// If FakeClient is missing a method required by any of these interfaces, this
// file fails to compile, catching the gap before test-time.
var (
	// Base composite interface — all backends must satisfy this.
	_ backend.Client = (*FakeClient)(nil)

	// Optional interfaces — FakeClient implements all of these.
	_ backend.AdminClient                           = (*FakeClient)(nil)
	_ backend.BranchModelClient                     = (*FakeClient)(nil)
	_ backend.BranchProtector                       = (*FakeClient)(nil)
	_ backend.BranchRuleClient                      = (*FakeClient)(nil)
	_ backend.CodeInsightsClient                    = (*FakeClient)(nil)
	_ backend.CodeSearcher                          = (*FakeClient)(nil)
	_ backend.CommitFileClient                      = (*FakeClient)(nil)
	_ backend.DefaultReviewerClient                 = (*FakeClient)(nil)
	_ backend.DefaultReviewersResolver              = (*FakeClient)(nil)
	_ backend.DeployKeyClient                       = (*FakeClient)(nil)
	_ backend.DeploymentClient                      = (*FakeClient)(nil)
	_ backend.DiffClient                            = (*FakeClient)(nil)
	_ backend.IssueClient                           = (*FakeClient)(nil)
	_ backend.PermissionsClient                     = (*FakeClient)(nil)
	_ backend.PipelineCacheClient                   = (*FakeClient)(nil)
	_ backend.PipelineClient                        = (*FakeClient)(nil)
	_ backend.PipelineScheduleClient                = (*FakeClient)(nil)
	_ backend.PipelineTriggerClient                 = (*FakeClient)(nil)
	_ backend.PRBranchUpdater                       = (*FakeClient)(nil)
	_ backend.PRCommitClient                        = (*FakeClient)(nil)
	_ backend.PRFileClient                          = (*FakeClient)(nil)
	_ backend.PRParticipantClient                   = (*FakeClient)(nil)
	_ backend.PRParticipantUpdater                  = (*FakeClient)(nil)
	_ backend.PRStatusLister                        = (*FakeClient)(nil)
	_ backend.RepoEditor                            = (*FakeClient)(nil)
	_ backend.RepoPRSettingsClient                  = (*FakeClient)(nil)
	_ backend.RepoForker                            = (*FakeClient)(nil)
	_ backend.RepoForksLister                       = (*FakeClient)(nil)
	_ backend.RepoTransferClient                    = (*FakeClient)(nil)
	_ backend.RepoWatcherClient                     = (*FakeClient)(nil)
	_ backend.ReviewerGroupClient                   = (*FakeClient)(nil)
	_ backend.SnippetClient                         = (*FakeClient)(nil)
	_ backend.SSHKeyClient                          = (*FakeClient)(nil)
	_ backend.WorkspaceClient                       = (*FakeClient)(nil)
	_ backend.WorkspaceMemberClient                 = (*FakeClient)(nil)
	_ backend.WorkspaceVariableClient               = (*FakeClient)(nil)
	_ backend.WorkspacePipelineVariableClient       = (*FakeClient)(nil)
	_ backend.PRReviewerRemover                     = (*FakeClient)(nil)
	_ backend.WorkspaceWebhookClient                = (*FakeClient)(nil)
	_ backend.GroupClient                           = (*FakeClient)(nil)
	_ backend.GroupMemberClient                     = (*FakeClient)(nil)
	_ backend.ServerProjectClient                   = (*FakeClient)(nil)
	_ backend.PATClient                             = (*FakeClient)(nil)
	_ backend.IssueAttacher                         = (*FakeClient)(nil)
	_ backend.IssueVoter                            = (*FakeClient)(nil)
	_ backend.IssueWatcher                          = (*FakeClient)(nil)
	_ backend.RepoLabelClient                       = (*FakeClient)(nil)
	_ backend.RunnerClient                          = (*FakeClient)(nil)
	_ backend.SourceWriter                          = (*FakeClient)(nil)
	_ backend.AuditClient                           = (*FakeClient)(nil)
	_ backend.CommitCherryPicker                    = (*FakeClient)(nil)
	_ backend.PipelineConfigClient                  = (*FakeClient)(nil)
	_ backend.PipelineTestReportClient              = (*FakeClient)(nil)
	_ backend.PipelineSSHKeyPairClient              = (*FakeClient)(nil)
	_ backend.PipelineKnownHostsClient              = (*FakeClient)(nil)
	_ backend.RefComparer                           = (*FakeClient)(nil)
	_ backend.RepoDownloadClient                    = (*FakeClient)(nil)
	_ backend.MilestoneClient                       = (*FakeClient)(nil)
	_ backend.IssueVersionClient                    = (*FakeClient)(nil)
	_ backend.CloudProjectClient                    = (*FakeClient)(nil)
	_ backend.MirrorClient                          = (*FakeClient)(nil)
	_ backend.WorkspacePermsClient                  = (*FakeClient)(nil)
	_ backend.IssueActivityClient                   = (*FakeClient)(nil)
	_ backend.WorkspaceSearcher                     = (*FakeClient)(nil)
	_ backend.WorkspaceProjectPermsClient           = (*FakeClient)(nil)
	_ backend.WorkspaceProjectDefaultReviewerClient = (*FakeClient)(nil)
)

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

func (c *FakeClient) SetRepoVisibility(ns, slug string, isPrivate bool) error {
	if c.SetRepoVisibilityFn != nil {
		return c.SetRepoVisibilityFn(ns, slug, isPrivate)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.SetRepoVisibility; set SetRepoVisibilityFn in your test")
	}
	return nil
}

func (c *FakeClient) SetRepoDefaultBranch(ns, slug, branch string) error {
	if c.SetRepoDefaultBranchFn != nil {
		return c.SetRepoDefaultBranchFn(ns, slug, branch)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.SetRepoDefaultBranch; set SetRepoDefaultBranchFn in your test")
	}
	return nil
}

func (c *FakeClient) EditRepo(ns, slug string, in backend.EditRepoInput) (backend.Repository, error) {
	if c.EditRepoFn != nil {
		return c.EditRepoFn(ns, slug, in)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.EditRepo; set EditRepoFn in your test")
	}
	return backend.Repository{}, nil
}

// ── RepoPRSettingsClient ──────────────────────────────────────────────────────

func (c *FakeClient) GetRepoPRSettings(ns, slug string) (backend.RepoPRSettings, error) {
	if c.GetRepoPRSettingsFn != nil {
		return c.GetRepoPRSettingsFn(ns, slug)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.GetRepoPRSettings; set GetRepoPRSettingsFn in your test")
	}
	return backend.RepoPRSettings{}, nil
}

func (c *FakeClient) UpdateRepoPRSettings(ns, slug string, in backend.RepoPRSettingsInput) (backend.RepoPRSettings, error) {
	if c.UpdateRepoPRSettingsFn != nil {
		return c.UpdateRepoPRSettingsFn(ns, slug, in)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.UpdateRepoPRSettings; set UpdateRepoPRSettingsFn in your test")
	}
	return backend.RepoPRSettings{}, nil
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

func (c *FakeClient) UnreadyPR(ns, slug string, id int) error {
	if c.UnreadyPRFn != nil {
		return c.UnreadyPRFn(ns, slug, id)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.UnreadyPR; set UnreadyPRFn in your test")
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

func (c *FakeClient) RemoveReviewers(ns, slug string, id int, users []string) error {
	if c.RemoveReviewersFn != nil {
		return c.RemoveReviewersFn(ns, slug, id, users)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.RemoveReviewers; set RemoveReviewersFn in your test")
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

func (c *FakeClient) StopPipeline(ws, slug, pipelineUUID string) error {
	if c.StopPipelineFn != nil {
		return c.StopPipelineFn(ws, slug, pipelineUUID)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.StopPipeline; set StopPipelineFn in your test")
	}
	return nil
}

func (c *FakeClient) RerunPipeline(ns, slug, sourceUUID string) (backend.Pipeline, error) {
	if c.RerunPipelineFn != nil {
		return c.RerunPipelineFn(ns, slug, sourceUUID)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.RerunPipeline; set RerunPipelineFn in your test")
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

// ── PipelineArtifactClient ───────────────────────────────────────────────────

func (c *FakeClient) ListPipelineArtifacts(ws, slug, pipelineUUID, stepUUID string, limit int) ([]backend.PipelineArtifact, error) {
	if c.ListPipelineArtifactsFn != nil {
		return c.ListPipelineArtifactsFn(ws, slug, pipelineUUID, stepUUID, limit)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListPipelineArtifacts; set ListPipelineArtifactsFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) DownloadPipelineArtifact(ws, slug, pipelineUUID, stepUUID, name string, out io.Writer) error {
	if c.DownloadPipelineArtifactFn != nil {
		return c.DownloadPipelineArtifactFn(ws, slug, pipelineUUID, stepUUID, name, out)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.DownloadPipelineArtifact; set DownloadPipelineArtifactFn in your test")
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

func (c *FakeClient) CherryPickCommit(ns, slug string, in backend.CherryPickInput) (backend.Commit, error) {
	if c.CherryPickCommitFn != nil {
		return c.CherryPickCommitFn(ns, slug, in)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.CherryPickCommit; set CherryPickCommitFn in your test")
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

// ── WorkspacePipelineVariableClient ──────────────────────────────────────────

func (c *FakeClient) ListWorkspacePipelineVariables(workspace string) ([]backend.PipelineVariable, error) {
	if c.ListWorkspacePipelineVariablesFn != nil {
		return c.ListWorkspacePipelineVariablesFn(workspace)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListWorkspacePipelineVariables; set ListWorkspacePipelineVariablesFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) GetWorkspacePipelineVariable(workspace, uuid string) (backend.PipelineVariable, error) {
	if c.GetWorkspacePipelineVariableFn != nil {
		return c.GetWorkspacePipelineVariableFn(workspace, uuid)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.GetWorkspacePipelineVariable; set GetWorkspacePipelineVariableFn in your test")
	}
	return backend.PipelineVariable{}, nil
}

func (c *FakeClient) SetWorkspacePipelineVariable(workspace string, in backend.PipelineVariableInput) (backend.PipelineVariable, error) {
	if c.SetWorkspacePipelineVariableFn != nil {
		return c.SetWorkspacePipelineVariableFn(workspace, in)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.SetWorkspacePipelineVariable; set SetWorkspacePipelineVariableFn in your test")
	}
	return backend.PipelineVariable{}, nil
}

func (c *FakeClient) DeleteWorkspacePipelineVariable(workspace, uuid string) error {
	if c.DeleteWorkspacePipelineVariableFn != nil {
		return c.DeleteWorkspacePipelineVariableFn(workspace, uuid)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.DeleteWorkspacePipelineVariable; set DeleteWorkspacePipelineVariableFn in your test")
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

func (c *FakeClient) ListAdminUsers(filter string, limit int) ([]backend.AdminUser, error) {
	if c.ListAdminUsersFn != nil {
		return c.ListAdminUsersFn(filter, limit)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListAdminUsers; set ListAdminUsersFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) RenameUser(slug, newSlug string) error {
	if c.RenameUserFn != nil {
		return c.RenameUserFn(slug, newSlug)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.RenameUser; set RenameUserFn in your test")
	}
	return nil
}

func (c *FakeClient) ActivateUser(slug string) error {
	if c.ActivateUserFn != nil {
		return c.ActivateUserFn(slug)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ActivateUser; set ActivateUserFn in your test")
	}
	return nil
}

func (c *FakeClient) DeactivateUser(slug string) error {
	if c.DeactivateUserFn != nil {
		return c.DeactivateUserFn(slug)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.DeactivateUser; set DeactivateUserFn in your test")
	}
	return nil
}

func (c *FakeClient) GetLicense() (backend.AdminLicense, error) {
	if c.GetLicenseFn != nil {
		return c.GetLicenseFn()
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.GetLicense; set GetLicenseFn in your test")
	}
	return backend.AdminLicense{}, nil
}

func (c *FakeClient) GetClusterNodes() ([]backend.ClusterNode, error) {
	if c.GetClusterNodesFn != nil {
		return c.GetClusterNodesFn()
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.GetClusterNodes; set GetClusterNodesFn in your test")
	}
	return nil, nil
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

func (c *FakeClient) GetBranchModel(ws, slug string) (backend.BranchModel, error) {
	if c.GetBranchModelFn != nil {
		return c.GetBranchModelFn(ws, slug)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.GetBranchModel; set GetBranchModelFn in your test")
	}
	return backend.BranchModel{}, nil
}

func (c *FakeClient) GetBranchModelSettings(ws, slug string) (backend.BranchModelSettings, error) {
	if c.GetBranchModelSettingsFn != nil {
		return c.GetBranchModelSettingsFn(ws, slug)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.GetBranchModelSettings; set GetBranchModelSettingsFn in your test")
	}
	return backend.BranchModelSettings{}, nil
}

func (c *FakeClient) UpdateBranchModelSettings(ws, slug string, in backend.BranchModelSettingsInput) (backend.BranchModelSettings, error) {
	if c.UpdateBranchModelSettingsFn != nil {
		return c.UpdateBranchModelSettingsFn(ws, slug, in)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.UpdateBranchModelSettings; set UpdateBranchModelSettingsFn in your test")
	}
	return backend.BranchModelSettings{}, nil
}

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

func (c *FakeClient) UpdateBranchRule(ns, slug string, id int, in backend.UpdateBranchRuleInput) (backend.BranchRule, error) {
	if c.UpdateBranchRuleFn != nil {
		return c.UpdateBranchRuleFn(ns, slug, id, in)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.UpdateBranchRule; set UpdateBranchRuleFn in your test")
	}
	return backend.BranchRule{}, nil
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

// ── PipelineCacheClient ──────────────────────────────────────────────────────

func (c *FakeClient) ListPipelineCaches(ns, slug string) ([]backend.PipelineCache, error) {
	if c.ListPipelineCachesFn != nil {
		return c.ListPipelineCachesFn(ns, slug)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListPipelineCaches; set ListPipelineCachesFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) DeletePipelineCache(ns, slug, uuid string) error {
	if c.DeletePipelineCacheFn != nil {
		return c.DeletePipelineCacheFn(ns, slug, uuid)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.DeletePipelineCache; set DeletePipelineCacheFn in your test")
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

// ── PRParticipantClient ───────────────────────────────────────────────────────

func (c *FakeClient) ListPRParticipants(ns, slug string, prID int) ([]backend.PRParticipant, error) {
	if c.ListPRParticipantsFn != nil {
		return c.ListPRParticipantsFn(ns, slug, prID)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListPRParticipants; set ListPRParticipantsFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) UpdatePRParticipant(ns, slug string, prID int, accountID, state string) (backend.PRParticipant, error) {
	if c.UpdatePRParticipantFn != nil {
		return c.UpdatePRParticipantFn(ns, slug, prID, accountID, state)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.UpdatePRParticipant; set UpdatePRParticipantFn in your test")
	}
	return backend.PRParticipant{}, nil
}

// ── WorkspaceMemberClient ─────────────────────────────────────────────────────

func (c *FakeClient) ListWorkspaceMembers(workspace string, limit int) ([]backend.WorkspaceMember, error) {
	if c.ListWorkspaceMembersFn != nil {
		return c.ListWorkspaceMembersFn(workspace, limit)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListWorkspaceMembers; set ListWorkspaceMembersFn in your test")
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

// ── RepoForksLister ────────────────────────────────────────────────────────────

func (c *FakeClient) ListRepoForks(ns, slug string, limit int) ([]backend.Repository, error) {
	if c.ListRepoForksFn != nil {
		return c.ListRepoForksFn(ns, slug, limit)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListRepoForks; set ListRepoForksFn in your test")
	}
	return nil, nil
}

// ── ReviewerGroupClient ───────────────────────────────────────────────────────

func (c *FakeClient) ListReviewerGroups(ns, slug string) ([]backend.ReviewerGroup, error) {
	if c.ListReviewerGroupsFn != nil {
		return c.ListReviewerGroupsFn(ns, slug)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListReviewerGroups; set ListReviewerGroupsFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) CreateReviewerGroup(ns, slug string, in backend.CreateReviewerGroupInput) (backend.ReviewerGroup, error) {
	if c.CreateReviewerGroupFn != nil {
		return c.CreateReviewerGroupFn(ns, slug, in)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.CreateReviewerGroup; set CreateReviewerGroupFn in your test")
	}
	return backend.ReviewerGroup{}, nil
}

func (c *FakeClient) DeleteReviewerGroup(ns, slug string, id int) error {
	if c.DeleteReviewerGroupFn != nil {
		return c.DeleteReviewerGroupFn(ns, slug, id)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.DeleteReviewerGroup; set DeleteReviewerGroupFn in your test")
	}
	return nil
}

// ── WorkspaceWebhookClient ────────────────────────────────────────────────────

func (c *FakeClient) ListWorkspaceWebhooks(workspace string) ([]backend.Webhook, error) {
	if c.ListWorkspaceWebhooksFn != nil {
		return c.ListWorkspaceWebhooksFn(workspace)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListWorkspaceWebhooks; set ListWorkspaceWebhooksFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) CreateWorkspaceWebhook(workspace string, in backend.CreateWebhookInput) (backend.Webhook, error) {
	if c.CreateWorkspaceWebhookFn != nil {
		return c.CreateWorkspaceWebhookFn(workspace, in)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.CreateWorkspaceWebhook; set CreateWorkspaceWebhookFn in your test")
	}
	return backend.Webhook{}, nil
}

func (c *FakeClient) DeleteWorkspaceWebhook(workspace, uuid string) error {
	if c.DeleteWorkspaceWebhookFn != nil {
		return c.DeleteWorkspaceWebhookFn(workspace, uuid)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.DeleteWorkspaceWebhook; set DeleteWorkspaceWebhookFn in your test")
	}
	return nil
}

func (c *FakeClient) ListSnippets(workspace string, limit int) ([]backend.Snippet, error) {
	if c.ListSnippetsFn != nil {
		return c.ListSnippetsFn(workspace, limit)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListSnippets; set ListSnippetsFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) GetSnippet(workspace, id string) (backend.Snippet, error) {
	if c.GetSnippetFn != nil {
		return c.GetSnippetFn(workspace, id)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.GetSnippet; set GetSnippetFn in your test")
	}
	return backend.Snippet{}, nil
}

func (c *FakeClient) CreateSnippet(workspace string, in backend.CreateSnippetInput) (backend.Snippet, error) {
	if c.CreateSnippetFn != nil {
		return c.CreateSnippetFn(workspace, in)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.CreateSnippet; set CreateSnippetFn in your test")
	}
	return backend.Snippet{}, nil
}

func (c *FakeClient) DeleteSnippet(workspace, id string) error {
	if c.DeleteSnippetFn != nil {
		return c.DeleteSnippetFn(workspace, id)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.DeleteSnippet; set DeleteSnippetFn in your test")
	}
	return nil
}

func (c *FakeClient) ListSnippetComments(workspace, snippetID string, limit int) ([]backend.SnippetComment, error) {
	if c.ListSnippetCommentsFn != nil {
		return c.ListSnippetCommentsFn(workspace, snippetID, limit)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListSnippetComments; set ListSnippetCommentsFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) AddSnippetComment(workspace, snippetID, body string) (backend.SnippetComment, error) {
	if c.AddSnippetCommentFn != nil {
		return c.AddSnippetCommentFn(workspace, snippetID, body)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.AddSnippetComment; set AddSnippetCommentFn in your test")
	}
	return backend.SnippetComment{}, nil
}

func (c *FakeClient) DeleteSnippetComment(workspace, snippetID string, commentID int) error {
	if c.DeleteSnippetCommentFn != nil {
		return c.DeleteSnippetCommentFn(workspace, snippetID, commentID)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.DeleteSnippetComment; set DeleteSnippetCommentFn in your test")
	}
	return nil
}

// ── GroupClient ───────────────────────────────────────────────────────────────

func (c *FakeClient) ListGroups(filter string, limit int) ([]backend.Group, error) {
	if c.ListGroupsFn != nil {
		return c.ListGroupsFn(filter, limit)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListGroups; set ListGroupsFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) CreateGroup(name string) (backend.Group, error) {
	if c.CreateGroupFn != nil {
		return c.CreateGroupFn(name)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.CreateGroup; set CreateGroupFn in your test")
	}
	return backend.Group{}, nil
}

func (c *FakeClient) DeleteGroup(name string) error {
	if c.DeleteGroupFn != nil {
		return c.DeleteGroupFn(name)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.DeleteGroup; set DeleteGroupFn in your test")
	}
	return nil
}

// ── GroupMemberClient ─────────────────────────────────────────────────────────

func (c *FakeClient) ListGroupMembers(groupName string, limit int) ([]backend.GroupMember, error) {
	if c.ListGroupMembersFn != nil {
		return c.ListGroupMembersFn(groupName, limit)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListGroupMembers; set ListGroupMembersFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) AddGroupMember(groupName, user string) error {
	if c.AddGroupMemberFn != nil {
		return c.AddGroupMemberFn(groupName, user)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.AddGroupMember; set AddGroupMemberFn in your test")
	}
	return nil
}

func (c *FakeClient) RemoveGroupMember(groupName, user string) error {
	if c.RemoveGroupMemberFn != nil {
		return c.RemoveGroupMemberFn(groupName, user)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.RemoveGroupMember; set RemoveGroupMemberFn in your test")
	}
	return nil
}

// ── ServerProjectClient ───────────────────────────────────────────────────────

func (c *FakeClient) ListServerProjects(filter string, limit int) ([]backend.ServerProject, error) {
	if c.ListServerProjectsFn != nil {
		return c.ListServerProjectsFn(filter, limit)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListServerProjects; set ListServerProjectsFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) GetServerProject(key string) (backend.ServerProject, error) {
	if c.GetServerProjectFn != nil {
		return c.GetServerProjectFn(key)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.GetServerProject; set GetServerProjectFn in your test")
	}
	return backend.ServerProject{}, nil
}

func (c *FakeClient) CreateServerProject(in backend.CreateServerProjectInput) (backend.ServerProject, error) {
	if c.CreateServerProjectFn != nil {
		return c.CreateServerProjectFn(in)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.CreateServerProject; set CreateServerProjectFn in your test")
	}
	return backend.ServerProject{}, nil
}

func (c *FakeClient) UpdateServerProject(key string, in backend.UpdateServerProjectInput) (backend.ServerProject, error) {
	if c.UpdateServerProjectFn != nil {
		return c.UpdateServerProjectFn(key, in)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.UpdateServerProject; set UpdateServerProjectFn in your test")
	}
	return backend.ServerProject{}, nil
}

func (c *FakeClient) DeleteServerProject(key string) error {
	if c.DeleteServerProjectFn != nil {
		return c.DeleteServerProjectFn(key)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.DeleteServerProject; set DeleteServerProjectFn in your test")
	}
	return nil
}

// ── PATClient ─────────────────────────────────────────────────────────────────

func (c *FakeClient) ListPATs(userSlug string, limit int) ([]backend.PAT, error) {
	if c.ListPATsFn != nil {
		return c.ListPATsFn(userSlug, limit)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListPATs; set ListPATsFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) CreatePAT(userSlug string, in backend.CreatePATInput) (backend.PATWithSecret, error) {
	if c.CreatePATFn != nil {
		return c.CreatePATFn(userSlug, in)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.CreatePAT; set CreatePATFn in your test")
	}
	return backend.PATWithSecret{}, nil
}

func (c *FakeClient) RevokePAT(userSlug, tokenID string) error {
	if c.RevokePATFn != nil {
		return c.RevokePATFn(userSlug, tokenID)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.RevokePAT; set RevokePATFn in your test")
	}
	return nil
}

// ── IssueAttacher ─────────────────────────────────────────────────────────────

func (c *FakeClient) ListIssueAttachments(ns, slug string, id int) ([]backend.IssueAttachment, error) {
	if c.ListIssueAttachmentsFn != nil {
		return c.ListIssueAttachmentsFn(ns, slug, id)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListIssueAttachments; set ListIssueAttachmentsFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) DeleteIssueAttachment(ns, slug string, id int, filename string) error {
	if c.DeleteIssueAttachmentFn != nil {
		return c.DeleteIssueAttachmentFn(ns, slug, id, filename)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.DeleteIssueAttachment; set DeleteIssueAttachmentFn in your test")
	}
	return nil
}

// ── IssueVoter ────────────────────────────────────────────────────────────────

func (c *FakeClient) VoteIssue(ns, slug string, id int) error {
	if c.VoteIssueFn != nil {
		return c.VoteIssueFn(ns, slug, id)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.VoteIssue; set VoteIssueFn in your test")
	}
	return nil
}

func (c *FakeClient) UnvoteIssue(ns, slug string, id int) error {
	if c.UnvoteIssueFn != nil {
		return c.UnvoteIssueFn(ns, slug, id)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.UnvoteIssue; set UnvoteIssueFn in your test")
	}
	return nil
}

// ── IssueWatcher ─────────────────────────────────────────────────────────────

func (c *FakeClient) WatchIssue(ns, slug string, id int) error {
	if c.WatchIssueFn != nil {
		return c.WatchIssueFn(ns, slug, id)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.WatchIssue; set WatchIssueFn in your test")
	}
	return nil
}

func (c *FakeClient) UnwatchIssue(ns, slug string, id int) error {
	if c.UnwatchIssueFn != nil {
		return c.UnwatchIssueFn(ns, slug, id)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.UnwatchIssue; set UnwatchIssueFn in your test")
	}
	return nil
}

// ── IssueActivityClient ───────────────────────────────────────────────────────

func (c *FakeClient) ListIssueActivity(ns, slug string, issueID int, limit int) ([]backend.IssueChange, error) {
	if c.ListIssueActivityFn != nil {
		return c.ListIssueActivityFn(ns, slug, issueID, limit)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListIssueActivity; set ListIssueActivityFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) SearchWorkspaces(opts backend.WorkspaceSearchOpts) ([]backend.Workspace, error) {
	if c.SearchWorkspacesFn != nil {
		return c.SearchWorkspacesFn(opts)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.SearchWorkspaces; set SearchWorkspacesFn in your test")
	}
	return nil, nil
}

// ── RepoLabelClient ───────────────────────────────────────────────────────────

func (c *FakeClient) ListRepoLabels(ns, slug string) ([]backend.RepoLabel, error) {
	if c.ListRepoLabelsFn != nil {
		return c.ListRepoLabelsFn(ns, slug)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListRepoLabels; set ListRepoLabelsFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) CreateRepoLabel(ns, slug string, in backend.CreateRepoLabelInput) (backend.RepoLabel, error) {
	if c.CreateRepoLabelFn != nil {
		return c.CreateRepoLabelFn(ns, slug, in)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.CreateRepoLabel; set CreateRepoLabelFn in your test")
	}
	return backend.RepoLabel{}, nil
}

func (c *FakeClient) UpdateRepoLabel(ns, slug string, id int, in backend.UpdateRepoLabelInput) (backend.RepoLabel, error) {
	if c.UpdateRepoLabelFn != nil {
		return c.UpdateRepoLabelFn(ns, slug, id, in)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.UpdateRepoLabel; set UpdateRepoLabelFn in your test")
	}
	return backend.RepoLabel{}, nil
}

func (c *FakeClient) DeleteRepoLabel(ns, slug string, id int) error {
	if c.DeleteRepoLabelFn != nil {
		return c.DeleteRepoLabelFn(ns, slug, id)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.DeleteRepoLabel; set DeleteRepoLabelFn in your test")
	}
	return nil
}

// ── SourceWriter ──────────────────────────────────────────────────────────────

func (c *FakeClient) PutFile(ns, slug, path string, in backend.PutFileInput) error {
	if c.PutFileFn != nil {
		return c.PutFileFn(ns, slug, path, in)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.PutFile; set PutFileFn in your test")
	}
	return nil
}

// ── Admin mail server ─────────────────────────────────────────────────────────

func (c *FakeClient) GetMailServerConfig() (backend.MailServerConfig, error) {
	if c.GetMailServerConfigFn != nil {
		return c.GetMailServerConfigFn()
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.GetMailServerConfig; set GetMailServerConfigFn in your test")
	}
	return backend.MailServerConfig{}, nil
}

func (c *FakeClient) SetMailServerConfig(in backend.MailServerConfig) error {
	if c.SetMailServerConfigFn != nil {
		return c.SetMailServerConfigFn(in)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.SetMailServerConfig; set SetMailServerConfigFn in your test")
	}
	return nil
}

// ── Admin banner ──────────────────────────────────────────────────────────────

func (c *FakeClient) GetBanner() (backend.BannerConfig, error) {
	if c.GetBannerFn != nil {
		return c.GetBannerFn()
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.GetBanner; set GetBannerFn in your test")
	}
	return backend.BannerConfig{}, nil
}

func (c *FakeClient) SetBanner(in backend.BannerConfig) error {
	if c.SetBannerFn != nil {
		return c.SetBannerFn(in)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.SetBanner; set SetBannerFn in your test")
	}
	return nil
}

func (c *FakeClient) ClearBanner() error {
	if c.ClearBannerFn != nil {
		return c.ClearBannerFn()
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ClearBanner; set ClearBannerFn in your test")
	}
	return nil
}

// ── RunnerClient ──────────────────────────────────────────────────────────────

func (c *FakeClient) ListRunners(workspace string) ([]backend.Runner, error) {
	if c.ListRunnersFn != nil {
		return c.ListRunnersFn(workspace)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListRunners; set ListRunnersFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) CreateRunner(workspace string, in backend.CreateRunnerInput) (backend.Runner, error) {
	if c.CreateRunnerFn != nil {
		return c.CreateRunnerFn(workspace, in)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.CreateRunner; set CreateRunnerFn in your test")
	}
	return backend.Runner{}, nil
}

func (c *FakeClient) DeleteRunner(workspace, runnerUUID string) error {
	if c.DeleteRunnerFn != nil {
		return c.DeleteRunnerFn(workspace, runnerUUID)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.DeleteRunner; set DeleteRunnerFn in your test")
	}
	return nil
}

// ── AuditClient ───────────────────────────────────────────────────────────────

func (c *FakeClient) ListAuditLog(workspace string, opts backend.AuditLogOpts) ([]backend.AuditEvent, error) {
	if c.ListAuditLogFn != nil {
		return c.ListAuditLogFn(workspace, opts)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListAuditLog; set ListAuditLogFn in your test")
	}
	return nil, nil
}

// ── IPAllowlistClient ─────────────────────────────────────────────────────────

var _ backend.IPAllowlistClient = (*FakeClient)(nil)

func (c *FakeClient) ListIPAllowlists(workspace string) ([]backend.IPAllowlist, error) {
	if c.ListIPAllowlistsFn != nil {
		return c.ListIPAllowlistsFn(workspace)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListIPAllowlists; set ListIPAllowlistsFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) CreateIPAllowlist(workspace string, in backend.CreateIPAllowlistInput) (backend.IPAllowlist, error) {
	if c.CreateIPAllowlistFn != nil {
		return c.CreateIPAllowlistFn(workspace, in)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.CreateIPAllowlist; set CreateIPAllowlistFn in your test")
	}
	return backend.IPAllowlist{}, nil
}

func (c *FakeClient) DeleteIPAllowlist(workspace, uuid string) error {
	if c.DeleteIPAllowlistFn != nil {
		return c.DeleteIPAllowlistFn(workspace, uuid)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.DeleteIPAllowlist; set DeleteIPAllowlistFn in your test")
	}
	return nil
}

// ── PipelineConfigClient ─────────────────────────────────────────────────────

func (c *FakeClient) GetPipelinesConfig(ws, slug string) (backend.PipelineConfig, error) {
	if c.GetPipelinesConfigFn != nil {
		return c.GetPipelinesConfigFn(ws, slug)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.GetPipelinesConfig; set GetPipelinesConfigFn in your test")
	}
	return backend.PipelineConfig{}, nil
}

func (c *FakeClient) UpdatePipelinesConfig(ws, slug string, in backend.PipelineConfig) (backend.PipelineConfig, error) {
	if c.UpdatePipelinesConfigFn != nil {
		return c.UpdatePipelinesConfigFn(ws, slug, in)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.UpdatePipelinesConfig; set UpdatePipelinesConfigFn in your test")
	}
	return backend.PipelineConfig{}, nil
}

// ── PipelineTestReportClient ──────────────────────────────────────────────────

var _ backend.PipelineTestReportClient = (*FakeClient)(nil)

func (c *FakeClient) GetPipelineTestReport(ws, slug, pipelineUUID, stepUUID string) (backend.PipelineTestReport, error) {
	if c.GetPipelineTestReportFn != nil {
		return c.GetPipelineTestReportFn(ws, slug, pipelineUUID, stepUUID)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.GetPipelineTestReport; set GetPipelineTestReportFn in your test")
	}
	return backend.PipelineTestReport{}, nil
}

func (c *FakeClient) ListPipelineTestCases(ws, slug, pipelineUUID, stepUUID string, filter backend.TestCaseFilter) ([]backend.PipelineTestCase, error) {
	if c.ListPipelineTestCasesFn != nil {
		return c.ListPipelineTestCasesFn(ws, slug, pipelineUUID, stepUUID, filter)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListPipelineTestCases; set ListPipelineTestCasesFn in your test")
	}
	return nil, nil
}

// ── RefComparer ───────────────────────────────────────────────────────────────

var _ backend.RefComparer = (*FakeClient)(nil)

func (c *FakeClient) CompareRefs(ns, slug, base, head string, limit int) (backend.RefComparison, error) {
	if c.CompareRefsFn != nil {
		return c.CompareRefsFn(ns, slug, base, head, limit)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.CompareRefs; set CompareRefsFn in your test")
	}
	return backend.RefComparison{}, nil
}

func (c *FakeClient) ListRepoDownloads(ns, slug string, limit int) ([]backend.RepoDownload, error) {
	if c.ListRepoDownloadsFn != nil {
		return c.ListRepoDownloadsFn(ns, slug, limit)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListRepoDownloads; set ListRepoDownloadsFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) UploadRepoDownload(ns, slug, name string, body io.Reader) (backend.RepoDownload, error) {
	if c.UploadRepoDownloadFn != nil {
		return c.UploadRepoDownloadFn(ns, slug, name, body)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.UploadRepoDownload; set UploadRepoDownloadFn in your test")
	}
	return backend.RepoDownload{}, nil
}

func (c *FakeClient) DownloadRepoDownload(ns, slug, name string, out io.Writer) error {
	if c.DownloadRepoDownloadFn != nil {
		return c.DownloadRepoDownloadFn(ns, slug, name, out)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.DownloadRepoDownload; set DownloadRepoDownloadFn in your test")
	}
	return nil
}

func (c *FakeClient) DeleteRepoDownload(ns, slug, name string) error {
	if c.DeleteRepoDownloadFn != nil {
		return c.DeleteRepoDownloadFn(ns, slug, name)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.DeleteRepoDownload; set DeleteRepoDownloadFn in your test")
	}
	return nil
}

func (c *FakeClient) ListMilestones(ns, slug string, limit int) ([]backend.Milestone, error) {
	if c.ListMilestonesFn != nil {
		return c.ListMilestonesFn(ns, slug, limit)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListMilestones; set ListMilestonesFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) GetMilestone(ns, slug string, id int) (backend.Milestone, error) {
	if c.GetMilestoneFn != nil {
		return c.GetMilestoneFn(ns, slug, id)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.GetMilestone; set GetMilestoneFn in your test")
	}
	return backend.Milestone{}, nil
}

func (c *FakeClient) ListIssueVersions(ns, slug string, limit int) ([]backend.IssueVersion, error) {
	if c.ListIssueVersionsFn != nil {
		return c.ListIssueVersionsFn(ns, slug, limit)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListIssueVersions; set ListIssueVersionsFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) GetIssueVersion(ns, slug string, id int) (backend.IssueVersion, error) {
	if c.GetIssueVersionFn != nil {
		return c.GetIssueVersionFn(ns, slug, id)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.GetIssueVersion; set GetIssueVersionFn in your test")
	}
	return backend.IssueVersion{}, nil
}

func (c *FakeClient) CreateIssueVersion(ns, slug, name string) (backend.IssueVersion, error) {
	if c.CreateIssueVersionFn != nil {
		return c.CreateIssueVersionFn(ns, slug, name)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.CreateIssueVersion; set CreateIssueVersionFn in your test")
	}
	return backend.IssueVersion{}, nil
}

func (c *FakeClient) DeleteIssueVersion(ns, slug string, id int) error {
	if c.DeleteIssueVersionFn != nil {
		return c.DeleteIssueVersionFn(ns, slug, id)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.DeleteIssueVersion; set DeleteIssueVersionFn in your test")
	}
	return nil
}

func (c *FakeClient) CreateWorkspaceProject(ws string, input backend.CreateWorkspaceProjectInput) (backend.WorkspaceProject, error) {
	if c.CreateWorkspaceProjectFn != nil {
		return c.CreateWorkspaceProjectFn(ws, input)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.CreateWorkspaceProject; set CreateWorkspaceProjectFn in your test")
	}
	return backend.WorkspaceProject{}, nil
}

func (c *FakeClient) GetWorkspaceProject(ws, key string) (backend.WorkspaceProject, error) {
	if c.GetWorkspaceProjectFn != nil {
		return c.GetWorkspaceProjectFn(ws, key)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.GetWorkspaceProject; set GetWorkspaceProjectFn in your test")
	}
	return backend.WorkspaceProject{}, nil
}

func (c *FakeClient) UpdateWorkspaceProject(ws, key string, input backend.UpdateWorkspaceProjectInput) (backend.WorkspaceProject, error) {
	if c.UpdateWorkspaceProjectFn != nil {
		return c.UpdateWorkspaceProjectFn(ws, key, input)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.UpdateWorkspaceProject; set UpdateWorkspaceProjectFn in your test")
	}
	return backend.WorkspaceProject{}, nil
}

func (c *FakeClient) DeleteWorkspaceProject(ws, key string) error {
	if c.DeleteWorkspaceProjectFn != nil {
		return c.DeleteWorkspaceProjectFn(ws, key)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.DeleteWorkspaceProject; set DeleteWorkspaceProjectFn in your test")
	}
	return nil
}

func (c *FakeClient) ListMirrorServers(limit int) ([]backend.MirrorServer, error) {
	if c.ListMirrorServersFn != nil {
		return c.ListMirrorServersFn(limit)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListMirrorServers; set ListMirrorServersFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) GetMirrorServer(id string) (backend.MirrorServer, error) {
	if c.GetMirrorServerFn != nil {
		return c.GetMirrorServerFn(id)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.GetMirrorServer; set GetMirrorServerFn in your test")
	}
	return backend.MirrorServer{}, nil
}

func (c *FakeClient) ListMirroredRepos(mirrorID string, limit int) ([]backend.MirroredRepo, error) {
	if c.ListMirroredReposFn != nil {
		return c.ListMirroredReposFn(mirrorID, limit)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListMirroredRepos; set ListMirroredReposFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) ListWorkspaceMemberPerms(ws string, limit int) ([]backend.WorkspaceMemberPerm, error) {
	if c.ListWorkspaceMemberPermsFn != nil {
		return c.ListWorkspaceMemberPermsFn(ws, limit)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListWorkspaceMemberPerms; set ListWorkspaceMemberPermsFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) ListWorkspaceRepoPerms(ws string, limit int) ([]backend.WorkspaceRepoPerm, error) {
	if c.ListWorkspaceRepoPermsFn != nil {
		return c.ListWorkspaceRepoPermsFn(ws, limit)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListWorkspaceRepoPerms; set ListWorkspaceRepoPermsFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) GrantWorkspacePerm(ws, user, permission string) error {
	if c.GrantWorkspacePermFn != nil {
		return c.GrantWorkspacePermFn(ws, user, permission)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.GrantWorkspacePerm; set GrantWorkspacePermFn in your test")
	}
	return nil
}

func (c *FakeClient) RevokeWorkspacePerm(ws, user string) error {
	if c.RevokeWorkspacePermFn != nil {
		return c.RevokeWorkspacePermFn(ws, user)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.RevokeWorkspacePerm; set RevokeWorkspacePermFn in your test")
	}
	return nil
}

// ── PipelineSSHKeyPairClient ──────────────────────────────────────────────────

func (c *FakeClient) GetPipelineSSHKeyPair(ns, slug string) (backend.PipelineSSHKeyPair, error) {
	if c.GetPipelineSSHKeyPairFn != nil {
		return c.GetPipelineSSHKeyPairFn(ns, slug)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.GetPipelineSSHKeyPair; set GetPipelineSSHKeyPairFn in your test")
	}
	return backend.PipelineSSHKeyPair{}, nil
}

func (c *FakeClient) RegeneratePipelineSSHKeyPair(ns, slug string, bits int) (backend.PipelineSSHKeyPair, error) {
	if c.RegeneratePipelineSSHKeyPairFn != nil {
		return c.RegeneratePipelineSSHKeyPairFn(ns, slug, bits)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.RegeneratePipelineSSHKeyPair; set RegeneratePipelineSSHKeyPairFn in your test")
	}
	return backend.PipelineSSHKeyPair{}, nil
}

// ── PipelineKnownHostsClient ──────────────────────────────────────────────────

func (c *FakeClient) ListPipelineKnownHosts(ns, slug string) ([]backend.PipelineKnownHost, error) {
	if c.ListPipelineKnownHostsFn != nil {
		return c.ListPipelineKnownHostsFn(ns, slug)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListPipelineKnownHosts; set ListPipelineKnownHostsFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) GetPipelineKnownHost(ns, slug, uuid string) (backend.PipelineKnownHost, error) {
	if c.GetPipelineKnownHostFn != nil {
		return c.GetPipelineKnownHostFn(ns, slug, uuid)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.GetPipelineKnownHost; set GetPipelineKnownHostFn in your test")
	}
	return backend.PipelineKnownHost{}, nil
}

func (c *FakeClient) AddPipelineKnownHost(ns, slug string, in backend.PipelineKnownHostInput) (backend.PipelineKnownHost, error) {
	if c.AddPipelineKnownHostFn != nil {
		return c.AddPipelineKnownHostFn(ns, slug, in)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.AddPipelineKnownHost; set AddPipelineKnownHostFn in your test")
	}
	return backend.PipelineKnownHost{}, nil
}

func (c *FakeClient) DeletePipelineKnownHost(ns, slug, uuid string) error {
	if c.DeletePipelineKnownHostFn != nil {
		return c.DeletePipelineKnownHostFn(ns, slug, uuid)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.DeletePipelineKnownHost; set DeletePipelineKnownHostFn in your test")
	}
	return nil
}

// ── WorkspaceProjectPermsClient ───────────────────────────────────────────────

func (c *FakeClient) ListWorkspaceProjectPerms(workspace, projectKey string, limit int) ([]backend.WorkspaceProjectPerm, error) {
	if c.ListWorkspaceProjectPermsFn != nil {
		return c.ListWorkspaceProjectPermsFn(workspace, projectKey, limit)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListWorkspaceProjectPerms; set ListWorkspaceProjectPermsFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) GrantWorkspaceProjectPerm(workspace, projectKey string, in backend.WorkspaceProjectPermInput) error {
	if c.GrantWorkspaceProjectPermFn != nil {
		return c.GrantWorkspaceProjectPermFn(workspace, projectKey, in)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.GrantWorkspaceProjectPerm; set GrantWorkspaceProjectPermFn in your test")
	}
	return nil
}

func (c *FakeClient) RevokeWorkspaceProjectPerm(workspace, projectKey, subjectSlug string, isGroup bool) error {
	if c.RevokeWorkspaceProjectPermFn != nil {
		return c.RevokeWorkspaceProjectPermFn(workspace, projectKey, subjectSlug, isGroup)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.RevokeWorkspaceProjectPerm; set RevokeWorkspaceProjectPermFn in your test")
	}
	return nil
}

// ── WorkspaceProjectDefaultReviewerClient ────────────────────────────────────

func (c *FakeClient) ListProjectDefaultReviewers(workspace, projectKey string, limit int) ([]backend.ProjectDefaultReviewer, error) {
	if c.ListProjectDefaultReviewersFn != nil {
		return c.ListProjectDefaultReviewersFn(workspace, projectKey, limit)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.ListProjectDefaultReviewers; set ListProjectDefaultReviewersFn in your test")
	}
	return nil, nil
}

func (c *FakeClient) AddProjectDefaultReviewer(workspace, projectKey, accountID string) error {
	if c.AddProjectDefaultReviewerFn != nil {
		return c.AddProjectDefaultReviewerFn(workspace, projectKey, accountID)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.AddProjectDefaultReviewer; set AddProjectDefaultReviewerFn in your test")
	}
	return nil
}

func (c *FakeClient) RemoveProjectDefaultReviewer(workspace, projectKey, accountID string) error {
	if c.RemoveProjectDefaultReviewerFn != nil {
		return c.RemoveProjectDefaultReviewerFn(workspace, projectKey, accountID)
	}
	if c.T != nil {
		c.T.Fatalf("unexpected call to FakeClient.RemoveProjectDefaultReviewer; set RemoveProjectDefaultReviewerFn in your test")
	}
	return nil
}
