package elasticsearch

import (
	"fmt"
	"strings"
	"time"
)

func UserDocumentFromDebeziumAfter(after map[string]any) (UserDocument, error) {
	dob, err := dateFromDebeziumDays(after["dob"])
	if err != nil {
		return UserDocument{}, err
	}

	firstName := stringFromMap(after, "first_name")
	lastName := stringFromMap(after, "last_name")

	return UserDocument{
		ID:              stringFromMap(after, "id"),
		Email:           stringFromMap(after, "email"),
		Username:        stringFromMap(after, "username"),
		FirstName:       firstName,
		LastName:        lastName,
		FullName:        strings.TrimSpace(firstName + " " + lastName),
		Dob:             dob,
		RoleID:          stringFromMap(after, "role_id"),
		Status:          stringFromMap(after, "status"),
		EmailVerifiedAt: stringPtrFromMap(after, "email_verified_at"),
		LastLoginAt:     stringPtrFromMap(after, "last_login_at"),
		CreatedAt:       stringFromMap(after, "created_at"),
		UpdatedAt:       stringFromMap(after, "updated_at"),
		DeletedAt:       stringPtrFromMap(after, "deleted_at"),
		DeletedBy:       stringPtrFromMap(after, "deleted_by"),
	}, nil
}

func stringFromMap(m map[string]any, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}

	s, ok := v.(string)
	if !ok {
		return ""
	}

	return s
}

func stringPtrFromMap(m map[string]any, key string) *string {
	v, ok := m[key]
	if !ok || v == nil {
		return nil
	}

	s, ok := v.(string)
	if !ok || s == "" {
		return nil
	}

	return &s
}

func dateFromDebeziumDays(v any) (*string, error) {
	if v == nil {
		return nil, nil
	}

	var days int64

	switch n := v.(type) {
	case float64:
		days = int64(n)
	case int:
		days = int64(n)
	case int64:
		days = n
	default:
		return nil, fmt.Errorf("unsupported debezium date type: %T", v)
	}

	t := time.Unix(0, 0).UTC().AddDate(0, 0, int(days))
	s := t.Format("2006-01-02")

	return &s, nil
}
