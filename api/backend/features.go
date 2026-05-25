package backend

import "fmt"

// FeatureSpec describes one optional backend capability.
type FeatureSpec struct {
	// Name is the interface name, e.g. "PipelineClient".
	Name string
	// HumanLabel is a short lowercase phrase used in unsupported-host messages,
	// e.g. "pipelines", "code insights", "branch protection".
	HumanLabel string
	// Plural controls whether the unsupported-host message uses "are not supported"
	// (true) or "is not supported" (false). Set true when HumanLabel is a plural noun.
	Plural bool
	// Check reports whether c implements this capability via a type assertion.
	Check func(c Client) bool
	// Feature is the Feature constant associated with this spec.
	Feature Feature
	// CloudSupport declares whether the Cloud adapter implements this interface.
	CloudSupport bool
	// ServerSupport declares whether the Server adapter implements this interface.
	ServerSupport bool
}

// AllFeatureSpecs enumerates every optional capability in declaration order.
// Adding a new As<X>Client function requires adding an entry here so the
// capability contract test catches any support-declaration mismatch.
var AllFeatureSpecs = []FeatureSpec{
	{
		Name:          "AdminClient",
		HumanLabel:    "admin operations",
		Plural:        true,
		Check:         func(c Client) bool { _, ok := c.(AdminClient); return ok },
		Feature:       FeatureAdmin,
		CloudSupport:  false,
		ServerSupport: true,
	},
	{
		Name:          "BranchProtector",
		HumanLabel:    "branch protection",
		Check:         func(c Client) bool { _, ok := c.(BranchProtector); return ok },
		Feature:       FeatureBranchProtect,
		CloudSupport:  false,
		ServerSupport: true,
	},
	{
		Name:          "BranchModelClient",
		HumanLabel:    "branching model",
		Check:         func(c Client) bool { _, ok := c.(BranchModelClient); return ok },
		Feature:       FeatureBranchModel,
		CloudSupport:  true,
		ServerSupport: false,
	},
	{
		Name:          "BranchRuleClient",
		HumanLabel:    "branch restriction rules",
		Plural:        true,
		Check:         func(c Client) bool { _, ok := c.(BranchRuleClient); return ok },
		Feature:       FeatureBranchRules,
		CloudSupport:  true,
		ServerSupport: false,
	},
	{
		Name:          "CodeInsightsClient",
		HumanLabel:    "code insights",
		Check:         func(c Client) bool { _, ok := c.(CodeInsightsClient); return ok },
		Feature:       FeatureCodeInsights,
		CloudSupport:  false,
		ServerSupport: true,
	},
	{
		Name:          "CodeSearcher",
		HumanLabel:    "code search",
		Check:         func(c Client) bool { _, ok := c.(CodeSearcher); return ok },
		Feature:       FeatureCodeSearch,
		CloudSupport:  true,
		ServerSupport: false,
	},
	{
		Name:          "CloudProjectClient",
		HumanLabel:    "Cloud workspace projects",
		Plural:        true,
		Check:         func(c Client) bool { _, ok := c.(CloudProjectClient); return ok },
		Feature:       FeatureCloudProjects,
		CloudSupport:  true,
		ServerSupport: false,
	},
	{
		Name:          "CommentReactor",
		HumanLabel:    "pr comment reactions",
		Plural:        true,
		Check:         func(c Client) bool { _, ok := c.(CommentReactor); return ok },
		Feature:       FeatureCommentReactions,
		CloudSupport:  false,
		ServerSupport: true,
	},
	{
		Name:          "CommitCommentReactor",
		HumanLabel:    "commit comment reactions",
		Plural:        true,
		Check:         func(c Client) bool { _, ok := c.(CommitCommentReactor); return ok },
		Feature:       FeatureCommitCommentReactions,
		CloudSupport:  false,
		ServerSupport: true,
	},
	{
		Name:          "CommitFileClient",
		HumanLabel:    "listing commit files",
		Check:         func(c Client) bool { _, ok := c.(CommitFileClient); return ok },
		Feature:       FeatureCommitFiles,
		CloudSupport:  true,
		ServerSupport: true,
	},
	{
		Name:          "DefaultReviewerClient",
		HumanLabel:    "default reviewer management",
		Check:         func(c Client) bool { _, ok := c.(DefaultReviewerClient); return ok },
		Feature:       FeatureDefaultReviewerClient,
		CloudSupport:  true,
		ServerSupport: true,
	},
	{
		Name:          "DefaultReviewersResolver",
		HumanLabel:    "default reviewers lookup",
		Check:         func(c Client) bool { _, ok := c.(DefaultReviewersResolver); return ok },
		Feature:       FeatureDefaultReviewers,
		CloudSupport:  false,
		ServerSupport: true,
	},
	{
		Name:          "DeployKeyClient",
		HumanLabel:    "deploy keys",
		Plural:        true,
		Check:         func(c Client) bool { _, ok := c.(DeployKeyClient); return ok },
		Feature:       FeatureDeployKeys,
		CloudSupport:  true,
		ServerSupport: true,
	},
	{
		Name:          "DeploymentClient",
		HumanLabel:    "deployments",
		Plural:        true,
		Check:         func(c Client) bool { _, ok := c.(DeploymentClient); return ok },
		Feature:       FeatureDeployments,
		CloudSupport:  true,
		ServerSupport: false,
	},
	{
		Name:          "DiffClient",
		HumanLabel:    "diff",
		Check:         func(c Client) bool { _, ok := c.(DiffClient); return ok },
		Feature:       FeatureDiff,
		CloudSupport:  true,
		ServerSupport: true,
	},
	{
		Name:          "GroupClient",
		HumanLabel:    "group management",
		Check:         func(c Client) bool { _, ok := c.(GroupClient); return ok },
		Feature:       FeatureGroup,
		CloudSupport:  false,
		ServerSupport: true,
	},
	{
		Name:          "GroupMemberClient",
		HumanLabel:    "group member management",
		Check:         func(c Client) bool { _, ok := c.(GroupMemberClient); return ok },
		Feature:       FeatureGroupMember,
		CloudSupport:  false,
		ServerSupport: true,
	},
	{
		Name:          "IssueAttacher",
		HumanLabel:    "issue attachments",
		Plural:        true,
		Check:         func(c Client) bool { _, ok := c.(IssueAttacher); return ok },
		Feature:       FeatureIssueAttachments,
		CloudSupport:  true,
		ServerSupport: false,
	},
	{
		Name:          "IssueClient",
		HumanLabel:    "issues",
		Plural:        true,
		Check:         func(c Client) bool { _, ok := c.(IssueClient); return ok },
		Feature:       FeatureIssues,
		CloudSupport:  true,
		ServerSupport: false,
	},
	{
		Name:          "IssueVoter",
		HumanLabel:    "issue voting",
		Check:         func(c Client) bool { _, ok := c.(IssueVoter); return ok },
		Feature:       FeatureIssueVoting,
		CloudSupport:  true,
		ServerSupport: false,
	},
	{
		Name:          "IssueWatcher",
		HumanLabel:    "issue watching",
		Check:         func(c Client) bool { _, ok := c.(IssueWatcher); return ok },
		Feature:       FeatureIssueWatching,
		CloudSupport:  true,
		ServerSupport: false,
	},
	{
		Name:          "MirrorClient",
		HumanLabel:    "mirror servers",
		Plural:        true,
		Check:         func(c Client) bool { _, ok := c.(MirrorClient); return ok },
		Feature:       FeatureMirror,
		CloudSupport:  false,
		ServerSupport: true,
	},
	{
		Name:          "IssueVersionClient",
		HumanLabel:    "issue versions",
		Plural:        true,
		Check:         func(c Client) bool { _, ok := c.(IssueVersionClient); return ok },
		Feature:       FeatureIssueVersions,
		CloudSupport:  true,
		ServerSupport: false,
	},
	{
		Name:          "MilestoneClient",
		HumanLabel:    "issue milestones",
		Plural:        true,
		Check:         func(c Client) bool { _, ok := c.(MilestoneClient); return ok },
		Feature:       FeatureMilestones,
		CloudSupport:  true,
		ServerSupport: false,
	},
	{
		Name:          "PRCommitClient",
		HumanLabel:    "PR commits",
		Plural:        true,
		Check:         func(c Client) bool { _, ok := c.(PRCommitClient); return ok },
		Feature:       FeaturePRCommits,
		CloudSupport:  true,
		ServerSupport: true,
	},
	{
		Name:          "PRCommentResolver",
		HumanLabel:    "pr comment resolve",
		Check:         func(c Client) bool { _, ok := c.(PRCommentResolver); return ok },
		Feature:       FeaturePRCommentResolve,
		CloudSupport:  true,
		ServerSupport: false,
	},
	{
		Name:          "PRCommentStateSetter",
		HumanLabel:    "pr task resolve/reopen",
		Check:         func(c Client) bool { _, ok := c.(PRCommentStateSetter); return ok },
		Feature:       FeaturePRCommentStateSet,
		CloudSupport:  false,
		ServerSupport: true,
	},
	{
		Name:          "PRFileClient",
		HumanLabel:    "PR files",
		Plural:        true,
		Check:         func(c Client) bool { _, ok := c.(PRFileClient); return ok },
		Feature:       FeaturePRFiles,
		CloudSupport:  true,
		ServerSupport: true,
	},
	{
		Name:          "PRParticipantClient",
		HumanLabel:    "PR participants",
		Plural:        true,
		Check:         func(c Client) bool { _, ok := c.(PRParticipantClient); return ok },
		Feature:       FeaturePRParticipants,
		CloudSupport:  true,
		ServerSupport: true,
	},
	{
		Name:          "PRReopener",
		HumanLabel:    "pr reopen",
		Check:         func(c Client) bool { _, ok := c.(PRReopener); return ok },
		Feature:       FeaturePRReopen,
		CloudSupport:  false,
		ServerSupport: true,
	},
	{
		Name:          "PermissionsClient",
		HumanLabel:    "permission management",
		Check:         func(c Client) bool { _, ok := c.(PermissionsClient); return ok },
		Feature:       FeaturePermissions,
		CloudSupport:  false,
		ServerSupport: true,
	},
	{
		Name:          "PipelineCacheClient",
		HumanLabel:    "pipeline caches",
		Plural:        true,
		Check:         func(c Client) bool { _, ok := c.(PipelineCacheClient); return ok },
		Feature:       FeaturePipelineCache,
		CloudSupport:  true,
		ServerSupport: false,
	},
	{
		Name:          "PipelineClient",
		HumanLabel:    "pipelines",
		Plural:        true,
		Check:         func(c Client) bool { _, ok := c.(PipelineClient); return ok },
		Feature:       FeaturePipelines,
		CloudSupport:  true,
		ServerSupport: false,
	},
	{
		Name:          "PipelineScheduleClient",
		HumanLabel:    "pipeline schedules",
		Plural:        true,
		Check:         func(c Client) bool { _, ok := c.(PipelineScheduleClient); return ok },
		Feature:       FeaturePipelineSchedules,
		CloudSupport:  true,
		ServerSupport: false,
	},
	{
		Name:          "PipelineArtifactClient",
		HumanLabel:    "pipeline artifacts",
		Plural:        true,
		Check:         func(c Client) bool { _, ok := c.(PipelineArtifactClient); return ok },
		Feature:       FeaturePipelineArtifacts,
		CloudSupport:  true,
		ServerSupport: false,
	},
	{
		Name:          "PipelineTriggerClient",
		HumanLabel:    "pipeline trigger",
		Check:         func(c Client) bool { _, ok := c.(PipelineTriggerClient); return ok },
		Feature:       FeaturePipelineTrigger,
		CloudSupport:  true,
		ServerSupport: false,
	},
	{
		Name:          "PipelineTestReportClient",
		HumanLabel:    "pipeline test reports",
		Plural:        true,
		Check:         func(c Client) bool { _, ok := c.(PipelineTestReportClient); return ok },
		Feature:       FeaturePipelineTestReports,
		CloudSupport:  true,
		ServerSupport: false,
	},
	{
		Name:          "RefComparer",
		HumanLabel:    "branch compare",
		Check:         func(c Client) bool { _, ok := c.(RefComparer); return ok },
		Feature:       FeatureRefCompare,
		CloudSupport:  true,
		ServerSupport: true,
	},
	{
		Name:          "RepoDownloadClient",
		HumanLabel:    "repo downloads",
		Plural:        true,
		Check:         func(c Client) bool { _, ok := c.(RepoDownloadClient); return ok },
		Feature:       FeatureRepoDownloads,
		CloudSupport:  true,
		ServerSupport: false,
	},
	{
		Name:          "RepoEditor",
		HumanLabel:    "repo edit",
		Check:         func(c Client) bool { _, ok := c.(RepoEditor); return ok },
		Feature:       FeatureRepoEdit,
		CloudSupport:  true,
		ServerSupport: true,
	},
	{
		Name:          "RepoPRSettingsClient",
		HumanLabel:    "repo PR settings",
		Check:         func(c Client) bool { _, ok := c.(RepoPRSettingsClient); return ok },
		Feature:       FeatureRepoPRSettings,
		CloudSupport:  false,
		ServerSupport: true,
	},
	{
		Name:          "RepoForker",
		HumanLabel:    "repo fork",
		Check:         func(c Client) bool { _, ok := c.(RepoForker); return ok },
		Feature:       FeatureRepoFork,
		CloudSupport:  true,
		ServerSupport: false,
	},
	{
		Name:          "RepoForksLister",
		HumanLabel:    "repo fork list",
		Check:         func(c Client) bool { _, ok := c.(RepoForksLister); return ok },
		Feature:       FeatureRepoForks,
		CloudSupport:  true,
		ServerSupport: true,
	},
	{
		Name:          "RepoTransferClient",
		HumanLabel:    "repo transfer",
		Check:         func(c Client) bool { _, ok := c.(RepoTransferClient); return ok },
		Feature:       FeatureRepoTransfer,
		CloudSupport:  true,
		ServerSupport: true,
	},
	{
		Name:          "RepoLabelClient",
		HumanLabel:    "repository labels",
		Plural:        true,
		Check:         func(c Client) bool { _, ok := c.(RepoLabelClient); return ok },
		Feature:       FeatureRepoLabels,
		CloudSupport:  true,
		ServerSupport: true,
	},
	{
		Name:          "RepoWatcherClient",
		HumanLabel:    "repo watchers",
		Plural:        true,
		Check:         func(c Client) bool { _, ok := c.(RepoWatcherClient); return ok },
		Feature:       FeatureRepoWatchers,
		CloudSupport:  true,
		ServerSupport: true,
	},
	{
		Name:          "ReviewerGroupClient",
		HumanLabel:    "reviewer group management",
		Check:         func(c Client) bool { _, ok := c.(ReviewerGroupClient); return ok },
		Feature:       FeatureReviewerGroup,
		CloudSupport:  false,
		ServerSupport: true,
	},
	{
		Name:          "PATClient",
		HumanLabel:    "personal access token management",
		Check:         func(c Client) bool { _, ok := c.(PATClient); return ok },
		Feature:       FeaturePAT,
		CloudSupport:  false,
		ServerSupport: true,
	},
	{
		Name:          "ServerProjectClient",
		HumanLabel:    "server project management",
		Check:         func(c Client) bool { _, ok := c.(ServerProjectClient); return ok },
		Feature:       FeatureServerProject,
		CloudSupport:  false,
		ServerSupport: true,
	},
	{
		Name:          "SnippetClient",
		HumanLabel:    "snippets",
		Plural:        true,
		Check:         func(c Client) bool { _, ok := c.(SnippetClient); return ok },
		Feature:       FeatureSnippets,
		CloudSupport:  true,
		ServerSupport: false,
	},
	{
		Name:          "SSHKeyClient",
		HumanLabel:    "user SSH keys",
		Plural:        true,
		Check:         func(c Client) bool { _, ok := c.(SSHKeyClient); return ok },
		Feature:       FeatureSSHKeys,
		CloudSupport:  true,
		ServerSupport: true,
	},
	{
		Name:          "SuggestionApplier",
		HumanLabel:    "pr suggestion apply",
		Check:         func(c Client) bool { _, ok := c.(SuggestionApplier); return ok },
		Feature:       FeaturePRSuggestion,
		CloudSupport:  false,
		ServerSupport: true,
	},
	{
		Name:          "VersionedServer",
		HumanLabel:    "server version",
		Check:         func(c Client) bool { _, ok := c.(VersionedServer); return ok },
		Feature:       FeatureServerVersion,
		CloudSupport:  false,
		ServerSupport: true,
	},
	{
		Name:          "WorkspaceClient",
		HumanLabel:    "workspaces",
		Plural:        true,
		Check:         func(c Client) bool { _, ok := c.(WorkspaceClient); return ok },
		Feature:       FeatureWorkspaces,
		CloudSupport:  true,
		ServerSupport: false,
	},
	{
		Name:          "WorkspaceMemberClient",
		HumanLabel:    "workspace members",
		Plural:        true,
		Check:         func(c Client) bool { _, ok := c.(WorkspaceMemberClient); return ok },
		Feature:       FeatureWorkspaceMembers,
		CloudSupport:  true,
		ServerSupport: false,
	},
	{
		Name:          "WorkspacePermsClient",
		HumanLabel:    "workspace permissions",
		Plural:        true,
		Check:         func(c Client) bool { _, ok := c.(WorkspacePermsClient); return ok },
		Feature:       FeatureWorkspacePerms,
		CloudSupport:  true,
		ServerSupport: false,
	},
	{
		Name:          "WorkspaceVariableClient",
		HumanLabel:    "workspace variables",
		Plural:        true,
		Check:         func(c Client) bool { _, ok := c.(WorkspaceVariableClient); return ok },
		Feature:       FeatureWorkspaceVariables,
		CloudSupport:  true,
		ServerSupport: false,
	},
	{
		Name:          "WorkspacePipelineVariableClient",
		HumanLabel:    "workspace pipeline variables",
		Plural:        true,
		Check:         func(c Client) bool { _, ok := c.(WorkspacePipelineVariableClient); return ok },
		Feature:       FeatureWorkspacePipelineVars,
		CloudSupport:  true,
		ServerSupport: false,
	},
	{
		Name:          "WorkspaceWebhookClient",
		HumanLabel:    "workspace webhooks",
		Plural:        true,
		Check:         func(c Client) bool { _, ok := c.(WorkspaceWebhookClient); return ok },
		Feature:       FeatureWorkspaceWebhooks,
		CloudSupport:  true,
		ServerSupport: false,
	},
	{
		Name:          "RunnerClient",
		HumanLabel:    "pipeline runners",
		Plural:        true,
		Check:         func(c Client) bool { _, ok := c.(RunnerClient); return ok },
		Feature:       FeatureRunner,
		CloudSupport:  true,
		ServerSupport: false,
	},
	{
		Name:          "AuditClient",
		HumanLabel:    "audit log",
		Plural:        false,
		Check:         func(c Client) bool { _, ok := c.(AuditClient); return ok },
		Feature:       FeatureAudit,
		CloudSupport:  true,
		ServerSupport: false,
	},
	{
		Name:          "IPAllowlistClient",
		HumanLabel:    "IP allowlist",
		Plural:        false,
		Check:         func(c Client) bool { _, ok := c.(IPAllowlistClient); return ok },
		Feature:       FeatureIPAllowlist,
		CloudSupport:  true,
		ServerSupport: false,
	},
	{
		Name:          "SourceWriter",
		HumanLabel:    "writing files via the API",
		Check:         func(c Client) bool { _, ok := c.(SourceWriter); return ok },
		Feature:       FeatureSourceWrite,
		CloudSupport:  true,
		ServerSupport: true,
	},
	{
		Name:          "CommitCherryPicker",
		HumanLabel:    "cherry-pick",
		Check:         func(c Client) bool { _, ok := c.(CommitCherryPicker); return ok },
		Feature:       FeatureCherryPick,
		CloudSupport:  false,
		ServerSupport: true,
	},
	{
		Name:          "PipelineConfigClient",
		HumanLabel:    "pipeline config",
		Check:         func(c Client) bool { _, ok := c.(PipelineConfigClient); return ok },
		Feature:       FeaturePipelineConfig,
		CloudSupport:  true,
		ServerSupport: false,
	},
	{
		Name:          "PipelineSSHKeyPairClient",
		HumanLabel:    "pipeline SSH key pair",
		Check:         func(c Client) bool { _, ok := c.(PipelineSSHKeyPairClient); return ok },
		Feature:       FeaturePipelineSSHKeyPair,
		CloudSupport:  true,
		ServerSupport: false,
	},
	{
		Name:          "PipelineKnownHostsClient",
		HumanLabel:    "pipeline known hosts",
		Plural:        true,
		Check:         func(c Client) bool { _, ok := c.(PipelineKnownHostsClient); return ok },
		Feature:       FeaturePipelineKnownHosts,
		CloudSupport:  true,
		ServerSupport: false,
	},
}

// specFor returns the FeatureSpec for the given Feature constant.
// Panics if not found (programming error: spec must be registered in AllFeatureSpecs).
func specFor(f Feature) FeatureSpec {
	for _, s := range AllFeatureSpecs {
		if s.Feature == f {
			return s
		}
	}
	panic("backend: no FeatureSpec registered for Feature " + string(f))
}

// requireFeature asserts that c implements T based on spec.
// Returns *DomainError{ErrUnsupportedOnHost} if the assertion fails.
func requireFeature[T any](c Client, host string, spec FeatureSpec) (T, error) {
	if v, ok := c.(T); ok {
		return v, nil
	}
	var zero T
	var suffix string
	switch {
	case spec.CloudSupport && !spec.ServerSupport:
		suffix = " (Bitbucket Cloud only)"
	case !spec.CloudSupport && spec.ServerSupport:
		suffix = " (Bitbucket Server / Data Center only)"
	}
	verb := "is"
	if spec.Plural {
		verb = "are"
	}
	return zero, &DomainError{
		Kind:    ErrUnsupportedOnHost,
		Code:    CodeHostUnsupported,
		Feature: string(spec.Feature),
		Host:    host,
		Message: fmt.Sprintf("%s %s not supported on %s%s", spec.HumanLabel, verb, host, suffix),
	}
}
