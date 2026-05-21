package backend

// FeatureSpec describes one optional backend capability.
type FeatureSpec struct {
	// Name is the interface name, e.g. "PipelineClient".
	Name string
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
		Check:         func(c Client) bool { _, ok := c.(AdminClient); return ok },
		Feature:       FeatureAdmin,
		CloudSupport:  false,
		ServerSupport: true,
	},
	{
		Name:          "BranchProtector",
		Check:         func(c Client) bool { _, ok := c.(BranchProtector); return ok },
		Feature:       FeatureBranchProtect,
		CloudSupport:  false,
		ServerSupport: true,
	},
	{
		Name:          "BranchRuleClient",
		Check:         func(c Client) bool { _, ok := c.(BranchRuleClient); return ok },
		Feature:       FeatureBranchRules,
		CloudSupport:  true,
		ServerSupport: false,
	},
	{
		Name:          "CodeInsightsClient",
		Check:         func(c Client) bool { _, ok := c.(CodeInsightsClient); return ok },
		Feature:       FeatureCodeInsights,
		CloudSupport:  false,
		ServerSupport: true,
	},
	{
		Name:          "CodeSearcher",
		Check:         func(c Client) bool { _, ok := c.(CodeSearcher); return ok },
		Feature:       FeatureCodeSearch,
		CloudSupport:  true,
		ServerSupport: false,
	},
	{
		Name:          "CommentReactor",
		Check:         func(c Client) bool { _, ok := c.(CommentReactor); return ok },
		Feature:       FeatureCommentReactions,
		CloudSupport:  false,
		ServerSupport: true,
	},
	{
		Name:          "CommitCommentReactor",
		Check:         func(c Client) bool { _, ok := c.(CommitCommentReactor); return ok },
		Feature:       FeatureCommitCommentReactions,
		CloudSupport:  false,
		ServerSupport: true,
	},
	{
		Name:          "CommitFileClient",
		Check:         func(c Client) bool { _, ok := c.(CommitFileClient); return ok },
		Feature:       FeatureCommitFiles,
		CloudSupport:  true,
		ServerSupport: true,
	},
	{
		Name:          "DefaultReviewerClient",
		Check:         func(c Client) bool { _, ok := c.(DefaultReviewerClient); return ok },
		Feature:       FeatureDefaultReviewerClient,
		CloudSupport:  true,
		ServerSupport: true,
	},
	{
		Name:          "DefaultReviewersResolver",
		Check:         func(c Client) bool { _, ok := c.(DefaultReviewersResolver); return ok },
		Feature:       FeatureDefaultReviewers,
		CloudSupport:  false,
		ServerSupport: true,
	},
	{
		Name:          "DeployKeyClient",
		Check:         func(c Client) bool { _, ok := c.(DeployKeyClient); return ok },
		Feature:       FeatureDeployKeys,
		CloudSupport:  true,
		ServerSupport: true,
	},
	{
		Name:          "DeploymentClient",
		Check:         func(c Client) bool { _, ok := c.(DeploymentClient); return ok },
		Feature:       FeatureDeployments,
		CloudSupport:  true,
		ServerSupport: false,
	},
	{
		Name:          "DiffClient",
		Check:         func(c Client) bool { _, ok := c.(DiffClient); return ok },
		Feature:       FeatureDiff,
		CloudSupport:  true,
		ServerSupport: true,
	},
	{
		Name:          "IssueClient",
		Check:         func(c Client) bool { _, ok := c.(IssueClient); return ok },
		Feature:       FeatureIssues,
		CloudSupport:  true,
		ServerSupport: false,
	},
	{
		Name:          "PRCommitClient",
		Check:         func(c Client) bool { _, ok := c.(PRCommitClient); return ok },
		Feature:       FeaturePRCommits,
		CloudSupport:  true,
		ServerSupport: true,
	},
	{
		Name:          "PRCommentResolver",
		Check:         func(c Client) bool { _, ok := c.(PRCommentResolver); return ok },
		Feature:       FeaturePRCommentResolve,
		CloudSupport:  true,
		ServerSupport: false,
	},
	{
		Name:          "PRCommentStateSetter",
		Check:         func(c Client) bool { _, ok := c.(PRCommentStateSetter); return ok },
		Feature:       FeaturePRCommentStateSet,
		CloudSupport:  false,
		ServerSupport: true,
	},
	{
		Name:          "PRFileClient",
		Check:         func(c Client) bool { _, ok := c.(PRFileClient); return ok },
		Feature:       FeaturePRFiles,
		CloudSupport:  true,
		ServerSupport: true,
	},
	{
		Name:          "PRParticipantClient",
		Check:         func(c Client) bool { _, ok := c.(PRParticipantClient); return ok },
		Feature:       FeaturePRParticipants,
		CloudSupport:  true,
		ServerSupport: true,
	},
	{
		Name:          "PRReopener",
		Check:         func(c Client) bool { _, ok := c.(PRReopener); return ok },
		Feature:       FeaturePRReopen,
		CloudSupport:  false,
		ServerSupport: true,
	},
	{
		Name:          "PermissionsClient",
		Check:         func(c Client) bool { _, ok := c.(PermissionsClient); return ok },
		Feature:       FeaturePermissions,
		CloudSupport:  false,
		ServerSupport: true,
	},
	{
		Name:          "PipelineCacheClient",
		Check:         func(c Client) bool { _, ok := c.(PipelineCacheClient); return ok },
		Feature:       FeaturePipelineCache,
		CloudSupport:  true,
		ServerSupport: false,
	},
	{
		Name:          "PipelineClient",
		Check:         func(c Client) bool { _, ok := c.(PipelineClient); return ok },
		Feature:       FeaturePipelines,
		CloudSupport:  true,
		ServerSupport: false,
	},
	{
		Name:          "PipelineScheduleClient",
		Check:         func(c Client) bool { _, ok := c.(PipelineScheduleClient); return ok },
		Feature:       FeaturePipelineSchedules,
		CloudSupport:  true,
		ServerSupport: false,
	},
	{
		Name:          "PipelineTriggerClient",
		Check:         func(c Client) bool { _, ok := c.(PipelineTriggerClient); return ok },
		Feature:       FeaturePipelineTrigger,
		CloudSupport:  true,
		ServerSupport: false,
	},
	{
		Name:          "RepoEditor",
		Check:         func(c Client) bool { _, ok := c.(RepoEditor); return ok },
		Feature:       FeatureRepoEdit,
		CloudSupport:  true,
		ServerSupport: true,
	},
	{
		Name:          "RepoForker",
		Check:         func(c Client) bool { _, ok := c.(RepoForker); return ok },
		Feature:       FeatureRepoFork,
		CloudSupport:  true,
		ServerSupport: false,
	},
	{
		Name:          "RepoForksLister",
		Check:         func(c Client) bool { _, ok := c.(RepoForksLister); return ok },
		Feature:       FeatureRepoForks,
		CloudSupport:  true,
		ServerSupport: true,
	},
	{
		Name:          "RepoTransferClient",
		Check:         func(c Client) bool { _, ok := c.(RepoTransferClient); return ok },
		Feature:       FeatureRepoTransfer,
		CloudSupport:  true,
		ServerSupport: true,
	},
	{
		Name:          "RepoWatcherClient",
		Check:         func(c Client) bool { _, ok := c.(RepoWatcherClient); return ok },
		Feature:       FeatureRepoWatchers,
		CloudSupport:  true,
		ServerSupport: true,
	},
	{
		Name:          "ReviewerGroupClient",
		Check:         func(c Client) bool { _, ok := c.(ReviewerGroupClient); return ok },
		Feature:       FeatureReviewerGroup,
		CloudSupport:  false,
		ServerSupport: true,
	},
	{
		Name:          "SnippetClient",
		Check:         func(c Client) bool { _, ok := c.(SnippetClient); return ok },
		Feature:       FeatureSnippets,
		CloudSupport:  true,
		ServerSupport: false,
	},
	{
		Name:          "SSHKeyClient",
		Check:         func(c Client) bool { _, ok := c.(SSHKeyClient); return ok },
		Feature:       FeatureSSHKeys,
		CloudSupport:  true,
		ServerSupport: false,
	},
	{
		Name:          "SuggestionApplier",
		Check:         func(c Client) bool { _, ok := c.(SuggestionApplier); return ok },
		Feature:       FeaturePRSuggestion,
		CloudSupport:  false,
		ServerSupport: true,
	},
	{
		Name:          "VersionedServer",
		Check:         func(c Client) bool { _, ok := c.(VersionedServer); return ok },
		Feature:       FeatureServerVersion,
		CloudSupport:  false,
		ServerSupport: true,
	},
	{
		Name:          "WorkspaceClient",
		Check:         func(c Client) bool { _, ok := c.(WorkspaceClient); return ok },
		Feature:       FeatureWorkspaces,
		CloudSupport:  true,
		ServerSupport: false,
	},
	{
		Name:          "WorkspaceMemberClient",
		Check:         func(c Client) bool { _, ok := c.(WorkspaceMemberClient); return ok },
		Feature:       FeatureWorkspaceMembers,
		CloudSupport:  true,
		ServerSupport: false,
	},
	{
		Name:          "WorkspaceVariableClient",
		Check:         func(c Client) bool { _, ok := c.(WorkspaceVariableClient); return ok },
		Feature:       FeatureWorkspaceVariables,
		CloudSupport:  true,
		ServerSupport: false,
	},
	{
		Name:          "WorkspaceWebhookClient",
		Check:         func(c Client) bool { _, ok := c.(WorkspaceWebhookClient); return ok },
		Feature:       FeatureWorkspaceWebhooks,
		CloudSupport:  true,
		ServerSupport: false,
	},
}
