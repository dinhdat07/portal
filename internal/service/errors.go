package service

import (
	"errors"
	"fmt"
)

type ErrorGroup string

const (
	ErrorGroupAuth       ErrorGroup = "AUTH"
	ErrorGroupUser       ErrorGroup = "USER"
	ErrorGroupRole       ErrorGroup = "ROLE"
	ErrorGroupPermission ErrorGroup = "PERMISSION"
	ErrorGroupEmail      ErrorGroup = "EMAIL"
	ErrorGroupPassword   ErrorGroup = "PASSWORD"
	ErrorGroupToken      ErrorGroup = "TOKEN"
	ErrorGroupValidation ErrorGroup = "VALIDATION"
	ErrorGroupAudit      ErrorGroup = "AUDIT"
	ErrorGroupInternal   ErrorGroup = "INTERNAL"
)

type ErrorCode string

const (
	CodeRoleNotFound           ErrorCode = "ROLE_NOT_FOUND"
	CodeRoleCodeExists         ErrorCode = "ROLE_CODE_EXISTS"
	CodeRoleInUse              ErrorCode = "ROLE_IN_USE"
	CodeInvalidReplacementRole ErrorCode = "ROLE_INVALID_REPLACEMENT"
	CodeCannotModifySystemRole ErrorCode = "ROLE_CANNOT_MODIFY_SYSTEM"

	CodePermissionNotFound ErrorCode = "PERMISSION_NOT_FOUND"

	CodeInvalidCredentials ErrorCode = "AUTH_INVALID_CREDENTIALS"
	CodeUnauthorized       ErrorCode = "AUTH_UNAUTHORIZED"
	CodeForbidden          ErrorCode = "AUTH_FORBIDDEN"

	CodeAccountNotVerified  ErrorCode = "AUTH_ACCOUNT_NOT_VERIFIED"
	CodeAccountDeleted      ErrorCode = "AUTH_ACCOUNT_DELETED"
	CodeUserInactive        ErrorCode = "AUTH_USER_INACTIVE"
	CodeInvalidRefreshToken ErrorCode = "TOKEN_INVALID_REFRESH"

	CodeUserNotFound       ErrorCode = "USER_NOT_FOUND"
	CodeUserAlreadyDeleted ErrorCode = "USER_ALREADY_DELETED"
	CodeInvalidUserID      ErrorCode = "USER_INVALID_ID"
	CodeUserNotDeleted     ErrorCode = "USER_NOT_DELETED"

	CodeEmailExists      ErrorCode = "EMAIL_ALREADY_EXISTS"
	CodeUsernameExists   ErrorCode = "USER_USERNAME_ALREADY_EXISTS"
	CodeEmailBlacklisted ErrorCode = "EMAIL_BLACKLISTED"

	CodeIncorrectPassword  ErrorCode = "PASSWORD_INCORRECT"
	CodePasswordMismatch   ErrorCode = "PASSWORD_CONFIRMATION_MISMATCH"
	CodePasswordNotChanged ErrorCode = "PASSWORD_NOT_CHANGED"
	CodePasswordAlreadySet ErrorCode = "PASSWORD_ALREADY_SET"

	CodeInvalidInput     ErrorCode = "VALIDATION_INVALID_INPUT"
	CodeInvalidAction    ErrorCode = "VALIDATION_INVALID_ACTION"
	CodeInvalidTimeRange ErrorCode = "VALIDATION_INVALID_TIME_RANGE"

	CodeInternalError               ErrorCode = "INTERNAL_ERROR"
	CodeInvalidToken                ErrorCode = "TOKEN_INVALID"
	CodeAuditLogFailed              ErrorCode = "AUDIT_LOG_FAILED"
	CodeSendVerificationEmailFailed ErrorCode = "EMAIL_SEND_VERIFICATION_FAILED"
	CodeSendResetPasswordFailed     ErrorCode = "EMAIL_SEND_RESET_PASSWORD_FAILED"
	CodeSendSetPasswordFailed       ErrorCode = "EMAIL_SEND_SET_PASSWORD_FAILED"
)

type AppError struct {
	Group   ErrorGroup `json:"group"`
	Code    ErrorCode  `json:"code"`
	Message string     `json:"message"`
	Err     error      `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}

	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NewAppError(group ErrorGroup, code ErrorCode, message string) *AppError {
	return &AppError{
		Group:   group,
		Code:    code,
		Message: message,
	}
}

func WrapAppError(group ErrorGroup, code ErrorCode, message string, err error) *AppError {
	return &AppError{
		Group:   group,
		Code:    code,
		Message: message,
		Err:     err,
	}
}

func AsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}

	return nil, false
}

var (
	ErrRoleNotFound = NewAppError(
		ErrorGroupRole,
		CodeRoleNotFound,
		"Role not found",
	)

	ErrRoleCodeExists = NewAppError(
		ErrorGroupRole,
		CodeRoleCodeExists,
		"Role code already exists",
	)

	ErrRoleInUse = NewAppError(
		ErrorGroupRole,
		CodeRoleInUse,
		"Role is assigned to users",
	)

	ErrInvalidReplacementRole = NewAppError(
		ErrorGroupRole,
		CodeInvalidReplacementRole,
		"Invalid replacement role",
	)

	ErrCannotModifySystemRole = NewAppError(
		ErrorGroupRole,
		CodeCannotModifySystemRole,
		"System role cannot be modified",
	)

	ErrPermissionNotFound = NewAppError(
		ErrorGroupPermission,
		CodePermissionNotFound,
		"Permission not found",
	)
)

var (
	ErrInvalidCredentials = NewAppError(
		ErrorGroupAuth,
		CodeInvalidCredentials,
		"Email/username or password is incorrect",
	)

	ErrUnauthorized = NewAppError(
		ErrorGroupAuth,
		CodeUnauthorized,
		"You are not authenticated",
	)

	ErrForbidden = NewAppError(
		ErrorGroupAuth,
		CodeForbidden,
		"You do not have permission to perform this action",
	)
)

var (
	ErrAccountNotVerified = NewAppError(
		ErrorGroupAuth,
		CodeAccountNotVerified,
		"Your account is not verified. Please check your email",
	)

	ErrAccountDeleted = NewAppError(
		ErrorGroupAuth,
		CodeAccountDeleted,
		"This account has been deleted",
	)

	ErrUserInactive = NewAppError(
		ErrorGroupAuth,
		CodeUserInactive,
		"Your account is inactive",
	)

	ErrInvalidRefreshToken = NewAppError(
		ErrorGroupToken,
		CodeInvalidRefreshToken,
		"Your session is expired",
	)
)

var (
	ErrUserNotFound = NewAppError(
		ErrorGroupUser,
		CodeUserNotFound,
		"User not found",
	)

	ErrUserAlreadyDeleted = NewAppError(
		ErrorGroupUser,
		CodeUserAlreadyDeleted,
		"User is already deleted",
	)

	ErrInvalidUserID = NewAppError(
		ErrorGroupUser,
		CodeInvalidUserID,
		"Invalid user ID",
	)

	ErrUserNotDeleted = NewAppError(
		ErrorGroupUser,
		CodeUserNotDeleted,
		"User is not deleted",
	)
)

var (
	ErrEmailExists = NewAppError(
		ErrorGroupEmail,
		CodeEmailExists,
		"Email is already in use",
	)

	ErrUsernameExists = NewAppError(
		ErrorGroupUser,
		CodeUsernameExists,
		"Username is already taken",
	)

	ErrEmailBlacklisted = NewAppError(
		ErrorGroupEmail,
		CodeEmailBlacklisted,
		"This email cannot be used",
	)
)

var (
	ErrIncorrectPassword = NewAppError(
		ErrorGroupPassword,
		CodeIncorrectPassword,
		"Current password is incorrect",
	)

	ErrPasswordConfirmationMismatch = NewAppError(
		ErrorGroupPassword,
		CodePasswordMismatch,
		"Password confirmation does not match",
	)

	ErrNewPasswordMustBeDifferent = NewAppError(
		ErrorGroupPassword,
		CodePasswordNotChanged,
		"New password must be different from the current one",
	)

	ErrPasswordAlreadySet = NewAppError(
		ErrorGroupPassword,
		CodePasswordAlreadySet,
		"Password has already been set",
	)
)

var (
	ErrInvalidInput = NewAppError(
		ErrorGroupValidation,
		CodeInvalidInput,
		"Invalid input data",
	)

	ErrInvalidAction = NewAppError(
		ErrorGroupValidation,
		CodeInvalidAction,
		"Invalid action",
	)

	ErrInvalidTimeRange = NewAppError(
		ErrorGroupValidation,
		CodeInvalidTimeRange,
		"Invalid time range",
	)
)

var (
	ErrInternalServer = NewAppError(
		ErrorGroupInternal,
		CodeInternalError,
		"Something went wrong. Please try again later",
	)

	ErrInvalidToken = NewAppError(
		ErrorGroupToken,
		CodeInvalidToken,
		"Invalid or expired token",
	)

	ErrAuditLogger = NewAppError(
		ErrorGroupAudit,
		CodeAuditLogFailed,
		"Failed to record activity",
	)

	ErrSendVerificationEmail = NewAppError(
		ErrorGroupEmail,
		CodeSendVerificationEmailFailed,
		"Failed to send verification email",
	)

	ErrSendResetPasswordEmail = NewAppError(
		ErrorGroupEmail,
		CodeSendResetPasswordFailed,
		"Failed to send reset password email",
	)

	ErrSendSetPasswordEmail = NewAppError(
		ErrorGroupEmail,
		CodeSendSetPasswordFailed,
		"Failed to send set password email",
	)
)
