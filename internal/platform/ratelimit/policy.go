package ratelimit

import (
	"time"

	adminv1 "portal-system/gen/go/admin/v1"
	authv1 "portal-system/gen/go/auth/v1"
	userv1 "portal-system/gen/go/user/v1"
)

var DefaultPolicy = Policy{
	Name:   "global",
	Limit:  300,
	Burst:  300,
	Window: time.Minute,
	Phase:  PhasePostAuth,
	Scopes: []KeyScope{
		ScopeUser,
	},
}

var MethodPolicies = map[string]Policy{
	authv1.AuthService_Login_FullMethodName: {
		Name:   "login",
		Limit:  5,
		Burst:  5,
		Window: time.Minute,
		Phase:  PhasePreAuth,
		Scopes: []KeyScope{
			ScopeIP,
			ScopeIdentifier,
		},
	},
	authv1.AuthService_Register_FullMethodName: {
		Name:   "register",
		Limit:  5,
		Burst:  5,
		Window: time.Hour,
		Phase:  PhasePreAuth,
		Scopes: []KeyScope{
			ScopeIP,
			ScopeEmail,
		},
	},
	authv1.AuthService_ForgotPassword_FullMethodName: {
		Name:   "forgot_password",
		Limit:  3,
		Burst:  3,
		Window: 15 * time.Minute,
		Phase:  PhasePreAuth,
		Scopes: []KeyScope{
			ScopeIP,
			ScopeEmail,
		},
	},
	authv1.AuthService_ResendVerification_FullMethodName: {
		Name:   "resend_verification",
		Limit:  3,
		Burst:  3,
		Window: 15 * time.Minute,
		Phase:  PhasePreAuth,
		Scopes: []KeyScope{
			ScopeIP,
			ScopeEmail,
		},
	},

	// RefreshToken is public in your current buildGRPCPublicMethods(),
	// but it should usually be limited by IP pre-auth unless you parse token/user before auth.
	authv1.AuthService_RefreshToken_FullMethodName: {
		Name:   "refresh_token",
		Limit:  30,
		Burst:  10,
		Window: time.Minute,
		Phase:  PhasePreAuth,
		Scopes: []KeyScope{
			ScopeIP,
		},
	},

	userv1.UserService_GetMyProfile_FullMethodName: {
		Name:   "user_read",
		Limit:  60,
		Burst:  120,
		Window: time.Minute,
		Phase:  PhasePostAuth,
		Scopes: []KeyScope{
			ScopeUser,
		},
	},
	userv1.UserService_UpdateMyProfile_FullMethodName: {
		Name:   "user_write",
		Limit:  30,
		Burst:  60,
		Window: time.Minute,
		Phase:  PhasePostAuth,
		Scopes: []KeyScope{
			ScopeUser,
		},
	},
	userv1.UserService_ChangeMyPassword_FullMethodName: {
		Name:   "change_password",
		Limit:  5,
		Burst:  5,
		Window: time.Hour,
		Phase:  PhasePostAuth,
		Scopes: []KeyScope{
			ScopeUser,
		},
	},

	adminv1.AdminService_ListUsers_FullMethodName: {
		Name:   "admin_read",
		Limit:  120,
		Burst:  200,
		Window: time.Minute,
		Phase:  PhasePostAuth,
		Scopes: []KeyScope{
			ScopeUser,
		},
	},
	adminv1.AdminService_GetUserDetail_FullMethodName: {
		Name:   "admin_read",
		Limit:  120,
		Burst:  200,
		Window: time.Minute,
		Phase:  PhasePostAuth,
		Scopes: []KeyScope{
			ScopeUser,
		},
	},
	adminv1.AdminService_CreateUser_FullMethodName: {
		Name:   "admin_write",
		Limit:  60,
		Burst:  80,
		Window: time.Minute,
		Phase:  PhasePostAuth,
		Scopes: []KeyScope{
			ScopeUser,
		},
	},
	adminv1.AdminService_UpdateUser_FullMethodName: {
		Name:   "admin_write",
		Limit:  60,
		Burst:  80,
		Window: time.Minute,
		Phase:  PhasePostAuth,
		Scopes: []KeyScope{
			ScopeUser,
		},
	},
	adminv1.AdminService_DeleteUser_FullMethodName: {
		Name:   "admin_dangerous",
		Limit:  20,
		Burst:  20,
		Window: time.Minute,
		Phase:  PhasePostAuth,
		Scopes: []KeyScope{
			ScopeUser,
		},
	},
	adminv1.AdminService_RestoreUser_FullMethodName: {
		Name:   "admin_dangerous",
		Limit:  20,
		Burst:  20,
		Window: time.Minute,
		Phase:  PhasePostAuth,
		Scopes: []KeyScope{
			ScopeUser,
		},
	},
	adminv1.AdminService_UpdateUserRole_FullMethodName: {
		Name:   "admin_dangerous",
		Limit:  20,
		Burst:  20,
		Window: time.Minute,
		Phase:  PhasePostAuth,
		Scopes: []KeyScope{
			ScopeUser,
		},
	},
}

func PolicyForMethod(method string) Policy {
	if policy, ok := MethodPolicies[method]; ok {
		return normalizePolicy(policy)
	}

	return DefaultPolicy
}

func normalizePolicy(policy Policy) Policy {
	if policy.Burst <= 0 {
		policy.Burst = policy.Limit
	}

	if policy.Phase == "" {
		policy.Phase = PhasePostAuth
	}

	if len(policy.Scopes) == 0 {
		policy.Scopes = []KeyScope{ScopeUser}
	}

	return policy
}
