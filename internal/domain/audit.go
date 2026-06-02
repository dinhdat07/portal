package domain

type ActionName string

const (
	ActionRegister           ActionName = "REGISTER"
	ActionLogin              ActionName = "LOGIN"
	ActionVerifyEmail        ActionName = "VERIFY_EMAIL"
	ActionResendVerification ActionName = "RESEND_VERIFICATION"
	ActionSetPassword        ActionName = "SET_PASSWORD"
	ActionResetPassword      ActionName = "RESET_PASSWORD"
	ActionForgotPassword     ActionName = "FORGOT_PASSWORD"

	ActionLogout    ActionName = "LOGOUT"
	ActionLogoutAll ActionName = "LOGOUT_ALL"

	ActionRefreshTokenReuseDetected ActionName = "REFRESH_TOKEN_REUSE_DETECTED"

	ActionUpdateProfile  ActionName = "UPDATE_PROFILE"
	ActionChangePassword ActionName = "CHANGE_PASSWORD"

	ActionAdminSearchUser  ActionName = "ADMIN_SEARCH_USER"
	ActionAdminViewUser    ActionName = "ADMIN_VIEW_USER"
	ActionAdminCreateUser  ActionName = "ADMIN_CREATE_USER"
	ActionAdminUpdateUser  ActionName = "ADMIN_UPDATE_USER"
	ActionAdminDeleteUser  ActionName = "ADMIN_DELETE_USER"
	ActionAdminRestoreUser ActionName = "ADMIN_RESTORE_USER"
	ActionAdminAssignRole  ActionName = "ADMIN_ASSIGN_ROLE"

	ActionCreateRole       ActionName = "ADMIN_CREATE_ROLE"
	ActionUpdateRole       ActionName = "ADMIN_UPDATE_ROLE"
	ActionDeleteRole       ActionName = "ADMIN_DELETE_ROLE"
	ActionListRoles        ActionName = "ADMIN_LIST_ROLES"
	ActionAssignPermission ActionName = "ADMIN_ASSIGN_PERMISSION_TO_ROLE"
	ActionRemovePermission ActionName = "ADMIN_REMOVE_PERMISSION_FROM_ROLE"

	ActionListPermissions ActionName = "ADMIN_LIST_PERMISSIONS"

	ActionAdminCreateAnnouncement ActionName = "ADMIN_CREATE_ANNOUNCEMENT"
	ActionAdminSearchAnnouncement ActionName = "ADMIN_SEARCH_ANNOUNCEMENT"
	ActionAdminViewAnnouncement   ActionName = "ADMIN_VIEW_ANNOUNCEMENT"
)

func (a ActionName) IsValid() bool {
	switch a {
	case ActionRegister,
		ActionVerifyEmail,
		ActionLogin,
		ActionLogout,
		ActionLogoutAll,
		ActionResendVerification,
		ActionSetPassword,
		ActionResetPassword,
		ActionForgotPassword,
		ActionRefreshTokenReuseDetected,

		ActionUpdateProfile,
		ActionChangePassword,
		ActionAdminSearchUser,
		ActionAdminViewUser,
		ActionAdminCreateUser,
		ActionAdminUpdateUser,
		ActionAdminDeleteUser,
		ActionAdminRestoreUser,
		ActionAdminAssignRole,
		ActionAdminCreateAnnouncement,
		ActionAdminSearchAnnouncement,
		ActionAdminViewAnnouncement:
		return true
	default:
		return false
	}
}
