package mapper

import (
	"portal-system/internal/service"

	"google.golang.org/grpc/codes"
	gstatus "google.golang.org/grpc/status"
)

func MapError(err error) error {
	if err == nil {
		return nil
	}

	if _, ok := gstatus.FromError(err); ok {
		return err
	}

	appErr, ok := service.AsAppError(err)
	if !ok {
		return gstatus.Error(codes.Internal, "Something went wrong. Please try again later")
	}

	return gstatus.Error(mapAppErrorCode(appErr), appErr.Message)
}

func mapAppErrorCode(err *service.AppError) codes.Code {
	switch err.Code {
	case service.CodeInvalidInput,
		service.CodeInvalidToken,
		service.CodeInvalidUserID,
		service.CodeInvalidAction,
		service.CodeInvalidTimeRange,
		service.CodeIncorrectPassword,
		service.CodePasswordMismatch,
		service.CodePasswordNotChanged,
		service.CodePasswordAlreadySet,
		service.CodeInvalidReplacementRole:
		return codes.InvalidArgument

	case service.CodeUnauthorized,
		service.CodeInvalidCredentials,
		service.CodeInvalidRefreshToken:
		return codes.Unauthenticated

	case service.CodeForbidden:
		return codes.PermissionDenied

	case service.CodeUserNotFound,
		service.CodeRoleNotFound,
		service.CodePermissionNotFound:
		return codes.NotFound

	case service.CodeEmailExists,
		service.CodeUsernameExists,
		service.CodeRoleCodeExists:
		return codes.AlreadyExists

	case service.CodeAccountNotVerified,
		service.CodeAccountDeleted,
		service.CodeUserAlreadyDeleted,
		service.CodeUserNotDeleted,
		service.CodeRoleInUse,
		service.CodeCannotModifySystemRole:
		return codes.FailedPrecondition

	default:
		return fallbackCodeByGroup(err.Group)
	}
}

func fallbackCodeByGroup(group service.ErrorGroup) codes.Code {
	switch group {
	case service.ErrorGroupAuth,
		service.ErrorGroupToken:
		return codes.Unauthenticated

	case service.ErrorGroupValidation,
		service.ErrorGroupPassword:
		return codes.InvalidArgument

	case service.ErrorGroupPermission:
		return codes.PermissionDenied

	case service.ErrorGroupInternal,
		service.ErrorGroupAudit,
		service.ErrorGroupEmail:
		return codes.Internal

	default:
		return codes.Internal
	}
}
