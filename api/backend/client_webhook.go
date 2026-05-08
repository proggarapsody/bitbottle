package backend

// WebhookLister lists webhooks for a repository.
type WebhookLister interface {
	ListWebhooks(ns, slug string) ([]Webhook, error)
}

// WebhookReader reads a single webhook by ID.
type WebhookReader interface {
	GetWebhook(ns, slug, id string) (Webhook, error)
}

// WebhookCreator creates a webhook.
type WebhookCreator interface {
	CreateWebhook(ns, slug string, in CreateWebhookInput) (Webhook, error)
}

// WebhookDeleter deletes a webhook.
type WebhookDeleter interface {
	DeleteWebhook(ns, slug, id string) error
}
