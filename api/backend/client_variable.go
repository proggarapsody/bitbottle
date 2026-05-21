package backend

// WorkspaceVariableClient is implemented by Bitbucket Cloud only.
type WorkspaceVariableClient interface {
	ListWorkspaceVariables(ns string) ([]PipelineVariable, error)
	SetWorkspaceVariable(ns string, in PipelineVariableInput) (PipelineVariable, error)
	DeleteWorkspaceVariable(ns, key string) error
}

const FeatureWorkspaceVariables Feature = "workspace-variables"

// AsWorkspaceVariableClient returns the WorkspaceVariableClient view of c,
// or a typed *DomainError (Kind=ErrUnsupportedOnHost) if not available.
func AsWorkspaceVariableClient(c Client, host string) (WorkspaceVariableClient, error) {
	return requireFeature[WorkspaceVariableClient](c, host, specFor(FeatureWorkspaceVariables))
}
