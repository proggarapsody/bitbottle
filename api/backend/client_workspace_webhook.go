package backend

// WorkspaceWebhookLister lists webhooks scoped to a workspace.
type WorkspaceWebhookLister interface {
	ListWorkspaceWebhooks(workspace string) ([]Webhook, error)
}

// WorkspaceWebhookCreator creates a workspace-level webhook.
type WorkspaceWebhookCreator interface {
	CreateWorkspaceWebhook(workspace string, in CreateWebhookInput) (Webhook, error)
}

// WorkspaceWebhookDeleter deletes a workspace-level webhook by UUID.
type WorkspaceWebhookDeleter interface {
	DeleteWorkspaceWebhook(workspace, uuid string) error
}

// WorkspaceWebhookClient bundles the three workspace-webhook operations.
type WorkspaceWebhookClient interface {
	WorkspaceWebhookLister
	WorkspaceWebhookCreator
	WorkspaceWebhookDeleter
}

// FeatureWorkspaceWebhooks names the workspace-webhook capability.
const FeatureWorkspaceWebhooks Feature = "workspace-webhooks"

// AsWorkspaceWebhookClient returns the WorkspaceWebhookClient view of c,
// or a typed *DomainError (ErrUnsupportedOnHost) if the backend is not Cloud.
func AsWorkspaceWebhookClient(c Client, host string) (WorkspaceWebhookClient, error) {
	return requireFeature[WorkspaceWebhookClient](c, host, specFor(FeatureWorkspaceWebhooks))
}
