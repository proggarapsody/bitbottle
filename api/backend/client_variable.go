package backend

import "fmt"

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
	wc, ok := c.(WorkspaceVariableClient)
	if !ok {
		return nil, &DomainError{
			Kind:    ErrUnsupportedOnHost,
			Code:    CodeHostUnsupported,
			Host:    host,
			Feature: string(FeatureWorkspaceVariables),
			Message: fmt.Sprintf("workspace variables are not supported on %s (Bitbucket Cloud only)", host),
		}
	}
	return wc, nil
}
