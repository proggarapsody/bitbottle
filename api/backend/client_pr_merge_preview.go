package backend

// PRMergePreviewClient performs a dry-run merge check on a pull request.
// Both Bitbucket Cloud and Bitbucket Server / Data Center expose dedicated
// dry-run endpoints that report conflicts without actually merging.
type PRMergePreviewClient interface {
	DryRunMergePR(ns, slug string, prID int, strategy string) (MergeDryRunResult, error)
}

// FeaturePRMergePreview names the PR merge-preview capability.
const FeaturePRMergePreview Feature = "pr_merge_preview"

// AsPRMergePreviewClient returns the PRMergePreviewClient view of c, or a
// typed *DomainError (Kind=ErrUnsupportedOnHost) when the backend does not
// implement the dry-run capability.
func AsPRMergePreviewClient(c Client, host string) (PRMergePreviewClient, error) {
	return requireFeature[PRMergePreviewClient](c, host, specFor(FeaturePRMergePreview))
}
