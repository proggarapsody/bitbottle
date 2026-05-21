package backend

// BranchRuleClient is implemented only by Bitbucket Cloud clients.
// Bitbucket Server / Data Center has no equivalent branch restriction API,
// so the entire surface is gated behind AsBranchRuleClient.
type BranchRuleClient interface {
	ListBranchRules(ns, slug string) ([]BranchRule, error)
	AddBranchRule(ns, slug string, input BranchRuleInput) (BranchRule, error)
	DeleteBranchRule(ns, slug string, id int) error
}

// FeatureBranchRules names the branch-rules capability for typed-error reporting.
const FeatureBranchRules Feature = "branch_rules"

// AsBranchRuleClient returns the BranchRuleClient view of c, or a typed
// *DomainError (Kind=ErrUnsupportedOnHost) when called against a Server/DC backend.
func AsBranchRuleClient(c Client, host string) (BranchRuleClient, error) {
	return requireFeature[BranchRuleClient](c, host, specFor(FeatureBranchRules))
}
