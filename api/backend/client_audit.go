package backend

// AuditClient lists workspace audit log events.
type AuditClient interface {
	ListAuditLog(workspace string, opts AuditLogOpts) ([]AuditEvent, error)
}

// AuditLogOpts filters the audit log query.
type AuditLogOpts struct {
	Action string // filter by action (e.g. "workspace.member.create")
	From   string // ISO 8601 date-time; returns events at or after this time
	Limit  int    // 0 = no cap (use paging default)
}

// AuditEvent is one entry in the workspace audit log.
type AuditEvent struct {
	Actor     AuditActor  `json:"actor"`
	Action    string      `json:"action"`
	Object    AuditObject `json:"object"`
	CreatedAt string      `json:"created_on"`
}

// AuditActor is the user who performed the action.
type AuditActor struct {
	AccountID   string `json:"account_id"`
	DisplayName string `json:"display_name"`
	NickName    string `json:"nickname"`
}

// AuditObject is the resource affected by the action.
type AuditObject struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

// FeatureAudit is the feature key for AuditClient.
const FeatureAudit Feature = "Audit"

// AsAuditClient asserts AuditClient on c. Returns a typed
// DomainError{ErrUnsupportedOnHost} when the backend does not support it.
func AsAuditClient(c Client, host string) (AuditClient, error) {
	return requireFeature[AuditClient](c, host, specFor(FeatureAudit))
}
