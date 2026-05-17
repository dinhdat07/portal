package service

import (
	"portal-system/internal/domain"
	"portal-system/internal/model"
	"time"
)

type UsersFilter struct {
	Page     int
	PageSize int
	Username string
	Email    string
	FullName string
	Dob      *time.Time

	// input API
	RoleCode *domain.RoleCode

	Status         domain.UserStatus
	IncludeDeleted bool
}

type ListUsersResult struct {
	Users    []model.User
	Total    int64
	Page     int
	PageSize int
}

type CreateUserInput struct {
	Email     string
	Username  string
	FirstName string
	LastName  string
	DOB       *time.Time
	RoleCode  domain.RoleCode
}

type UpdateUserInput struct {
	Username  *string
	FirstName *string
	LastName  *string
	DOB       *time.Time
}

type CreateRoleInput struct {
	Code domain.RoleCode
	Name string
}
