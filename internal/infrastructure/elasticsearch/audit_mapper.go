package elasticsearch

func AuditLogDocumentFromDebeziumAfter(after map[string]any) AuditLogDocument {
	return AuditLogDocument{
		ID:             stringFromMap(after, "id"),
		Action:         stringFromMap(after, "action"),
		ActorUserID:    stringPtrFromMap(after, "actor_user_id"),
		ActorUsername:  stringPtrFromMap(after, "actor_username"),
		ActorEmail:     stringPtrFromMap(after, "actor_email"),
		ActorRole:      stringPtrFromMap(after, "actor_role"),
		TargetUserID:   stringPtrFromMap(after, "target_user_id"),
		TargetUsername: stringPtrFromMap(after, "target_username"),
		TargetEmail:    stringPtrFromMap(after, "target_email"),
		TargetRole:     stringPtrFromMap(after, "target_role"),
		Metadata:       mapFromMap(after, "metadata"),
		IPAddress:      stringPtrFromMap(after, "ip_address"),
		UserAgent:      stringPtrFromMap(after, "user_agent"),
		CreatedAt:      stringFromMap(after, "created_at"),
	}
}
