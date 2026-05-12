package mappers

import (
	adminv1 "portal-system/gen/go/admin/v1"
	commonv1 "portal-system/gen/go/common/v1"
	"portal-system/internal/models"
	"portal-system/internal/services"
)

func ListUsersResultToPB(result *services.ListUsersResult) *adminv1.ListUsersResponse {
	if result == nil {
		return nil
	}

	data := make([]*commonv1.User, 0, len(result.Users))
	for i := range result.Users {
		data = append(data, UserModelToPB(&result.Users[i]))
	}

	return &adminv1.ListUsersResponse{
		Data: data,
		Meta: &commonv1.PaginationMeta{
			Page:     int32(result.Page),
			PageSize: int32(result.PageSize),
			Total:    result.Total,
		},
	}
}

func RolesToPB(roles []models.Role) *adminv1.ListRolesResponse {
	data := make([]*commonv1.Role, 0, len(roles))
	for i := range roles {
		data = append(data, RoleModelToPB(&roles[i]))
	}

	return &adminv1.ListRolesResponse{Data: data}
}

func RoleModelToPB(role *models.Role) *commonv1.Role {
	if role == nil {
		return nil
	}

	perms := make([]*commonv1.Permission, 0, len(role.Permissions))
	for i := range role.Permissions {
		perms = append(perms, PermissionModelToPB(&role.Permissions[i]))
	}

	return &commonv1.Role{
		Id:          role.ID.String(),
		Code:        string(role.Code),
		Name:        role.Name,
		IsSystem:    role.IsSystem,
		Permissions: perms,
	}
}

func PermissionsToPB(perms []models.Permission) *adminv1.ListPermissionsResponse {
	data := make([]*commonv1.Permission, 0, len(perms))
	for i := range perms {
		data = append(data, PermissionModelToPB(&perms[i]))
	}

	return &adminv1.ListPermissionsResponse{Data: data}
}

func PermissionModelToPB(perm *models.Permission) *commonv1.Permission {
	if perm == nil {
		return nil
	}

	return &commonv1.Permission{
		Id:   perm.ID.String(),
		Code: perm.Code,
		Name: perm.Name,
	}
}
