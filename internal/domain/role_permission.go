package domain

var RolePermissions = map[RoleCode][]PermissionCode{
	RoleCodeAdmin: {
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
	},
	RoleCodeUser: {
		PermProfileReadSelf,
		PermProfileUpdateSelf,
		PermProfileChangePassword,
	},
}
