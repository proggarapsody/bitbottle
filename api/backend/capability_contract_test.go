package backend_test

import (
	"testing"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/api/cloud"
	"github.com/proggarapsody/bitbottle/api/server"
)

// cloudClient is a nil *cloud.Client cast to backend.Client.
// A nil pointer satisfies interface type assertions — we only check type
// (dynamic type is *cloud.Client), never call any method.
var cloudClient backend.Client = (*cloud.Client)(nil)
var serverClient backend.Client = (*server.Client)(nil)

func TestAllFeatureSpecs_CloudSupport(t *testing.T) {
	t.Parallel()
	for _, spec := range backend.AllFeatureSpecs {
		spec := spec
		t.Run(spec.Name, func(t *testing.T) {
			t.Parallel()
			got := spec.Check(cloudClient)
			if got != spec.CloudSupport {
				t.Errorf("Cloud.%s: Check()=%v but CloudSupport=%v — update AllFeatureSpecs or the adapter",
					spec.Name, got, spec.CloudSupport)
			}
		})
	}
}

func TestAllFeatureSpecs_ServerSupport(t *testing.T) {
	t.Parallel()
	for _, spec := range backend.AllFeatureSpecs {
		spec := spec
		t.Run(spec.Name, func(t *testing.T) {
			t.Parallel()
			got := spec.Check(serverClient)
			if got != spec.ServerSupport {
				t.Errorf("Server.%s: Check()=%v but ServerSupport=%v — update AllFeatureSpecs or the adapter",
					spec.Name, got, spec.ServerSupport)
			}
		})
	}
}

func TestAllFeatureSpecs_Coverage(t *testing.T) {
	t.Parallel()
	// All 55 expected names must be present in AllFeatureSpecs.
	expected := []string{
		"AdminClient", "AuditClient", "BranchModelClient", "BranchProtector", "BranchRuleClient", "CodeInsightsClient",
		"CodeSearcher", "CommentReactor", "CommitCherryPicker", "CommitCommentReactor", "CommitFileClient",
		"DefaultReviewerClient", "DefaultReviewersResolver", "DeployKeyClient", "DeploymentClient",
		"DiffClient", "GroupClient", "GroupMemberClient",
		"IPAllowlistClient",
		"IssueAttacher", "IssueClient", "IssueVoter", "IssueWatcher",
		"PRCommitClient", "PRCommentResolver",
		"PRCommentStateSetter", "PRFileClient", "PRParticipantClient", "PRReopener",
		"PATClient", "PermissionsClient", "PipelineArtifactClient", "PipelineCacheClient", "PipelineClient",
		"PipelineScheduleClient", "PipelineTriggerClient", "RepoEditor", "RepoForker", "RepoForksLister",
		"RepoPRSettingsClient", "RepoLabelClient", "RepoTransferClient", "RepoWatcherClient", "ReviewerGroupClient",
		"RunnerClient",
		"ServerProjectClient", "SnippetClient", "SSHKeyClient",
		"SourceWriter",
		"SuggestionApplier", "VersionedServer", "WorkspaceClient", "WorkspaceMemberClient",
		"WorkspaceVariableClient", "WorkspaceWebhookClient",
	}
	names := make(map[string]bool, len(backend.AllFeatureSpecs))
	for _, spec := range backend.AllFeatureSpecs {
		names[spec.Name] = true
	}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("AllFeatureSpecs is missing %q", name)
		}
	}
	if len(backend.AllFeatureSpecs) != len(expected) {
		t.Errorf("AllFeatureSpecs has %d entries, expected %d — add/remove from expected slice", len(backend.AllFeatureSpecs), len(expected))
	}
}
