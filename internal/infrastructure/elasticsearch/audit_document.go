package elasticsearch

type AuditLogDocument struct {
	ID             string         `json:"id"`
	Action         string         `json:"action"`
	ActorUserID    *string        `json:"actor_user_id,omitempty"`
	ActorUsername  *string        `json:"actor_username,omitempty"`
	ActorEmail     *string        `json:"actor_email,omitempty"`
	ActorRole      *string        `json:"actor_role,omitempty"`
	TargetUserID   *string        `json:"target_user_id,omitempty"`
	TargetUsername *string        `json:"target_username,omitempty"`
	TargetEmail    *string        `json:"target_email,omitempty"`
	TargetRole     *string        `json:"target_role,omitempty"`
	Metadata       map[string]any `json:"metadata,omitempty"`
	IPAddress      *string        `json:"ip_address,omitempty"`
	UserAgent      *string        `json:"user_agent,omitempty"`
	CreatedAt      string         `json:"created_at"`
}
