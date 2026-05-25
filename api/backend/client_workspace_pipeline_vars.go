package backend

// WorkspacePipelineVariableClient is implemented by Bitbucket Cloud only.
type WorkspacePipelineVariableClient interface {
	ListWorkspacePipelineVariables(workspace string) ([]PipelineVariable, error)
	GetWorkspacePipelineVariable(workspace, uuid string) (PipelineVariable, error)
	SetWorkspacePipelineVariable(workspace string, in PipelineVariableInput) (PipelineVariable, error)
	DeleteWorkspacePipelineVariable(workspace, uuid string) error
}

const FeatureWorkspacePipelineVars Feature = "workspace-pipeline-vars"

// AsWorkspacePipelineVariableClient returns the WorkspacePipelineVariableClient
// view of c, or a typed *DomainError (Kind=ErrUnsupportedOnHost) if not
// available.
func AsWorkspacePipelineVariableClient(c Client, host string) (WorkspacePipelineVariableClient, error) {
	return requireFeature[WorkspacePipelineVariableClient](c, host, specFor(FeatureWorkspacePipelineVars))
}
