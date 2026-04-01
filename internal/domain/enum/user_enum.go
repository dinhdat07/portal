package enum

type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
)

func (r UserRole) IsValid() bool {
	switch r {
	case RoleUser, RoleAdmin:
		return true
	default:
		return false
	}
}

type UserStatus string

const (
	StatusActive  UserStatus = "active"
	StatusPending UserStatus = "pending_verification"
	StatusDeleted UserStatus = "deleted"
)

func (r UserStatus) IsValid() bool {
	switch r {
	case StatusActive, StatusDeleted, StatusPending:
		return true
	default:
		return false
	}
}
