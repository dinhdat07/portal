package elasticsearch

import (
	"strings"
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
