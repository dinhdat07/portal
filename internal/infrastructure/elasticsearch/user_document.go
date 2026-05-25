package elasticsearch

type UserDocument struct {
	ID              string  `json:"id"`
	Email           string  `json:"email"`
	Username        string  `json:"username"`
	FirstName       string  `json:"first_name"`
	LastName        string  `json:"last_name"`
	FullName        string  `json:"full_name"`
	Dob             *string `json:"dob,omitempty"`
	RoleID          string  `json:"role_id"`
	Status          string  `json:"status"`
	EmailVerifiedAt *string `json:"email_verified_at,omitempty"`
	LastLoginAt     *string `json:"last_login_at,omitempty"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
	DeletedAt       *string `json:"deleted_at,omitempty"`
	DeletedBy       *string `json:"deleted_by,omitempty"`
}
