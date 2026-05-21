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
		Name:          "IssueClient",
		HumanLabel:    "issues",
		Plural:        true,
		Check:         func(c Client) bool { _, ok := c.(IssueClient); return ok },
		Feature:       FeatureIssues,
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
		Name:          "PipelineTriggerClient",
		HumanLabel:    "pipeline trigger",
		Check:         func(c Client) bool { _, ok := c.(PipelineTriggerClient); return ok },
		Feature:       FeaturePipelineTrigger,
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
		ServerSupport: false,
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
		Name:          "WorkspaceVariableClient",
		HumanLabel:    "workspace variables",
		Plural:        true,
		Check:         func(c Client) bool { _, ok := c.(WorkspaceVariableClient); return ok },
		Feature:       FeatureWorkspaceVariables,
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
