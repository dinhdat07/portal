package domain

type PermissionCode string

const (
	// profile
	PermProfileReadSelf       PermissionCode = "profile:read_self"
	PermProfileUpdateSelf     PermissionCode = "profile:update_self"
	PermProfileChangePassword PermissionCode = "profile:change_password"

	// user management
	PermUserList       PermissionCode = "users:list"
	PermUserReadDetail PermissionCode = "users:read_detail"
	PermUserCreate     PermissionCode = "users:create"
	PermUserUpdate     PermissionCode = "users:update"
	PermUserDelete     PermissionCode = "users:delete"
	PermUserRestore    PermissionCode = "users:restore"
	PermUserRoleUpdate PermissionCode = "users:update_role"

	// role management
	PermRoleList             PermissionCode = "roles:list"
	PermRoleCreate           PermissionCode = "roles:create"
	PermRoleDelete           PermissionCode = "roles:delete"
	PermRoleAssignPermission PermissionCode = "roles:assign_permission"
	PermRoleRemovePermission PermissionCode = "roles:remove_permission"

	// permission management
	PermPermissionList PermissionCode = "permissions:list"
)

var AllPermissions = []PermissionCode{
	PermUserList,
	PermUserReadDetail,
	PermUserCreate,
	PermUserUpdate,
	PermUserDelete,
	PermUserRestore,
	PermUserRoleUpdate,
	PermRoleList,
	PermRoleCreate,
	PermRoleDelete,
	PermRoleAssignPermission,
	PermRoleRemovePermission,
	PermPermissionList,
	PermProfileReadSelf,
	PermProfileUpdateSelf,
	PermProfileChangePassword,
}
