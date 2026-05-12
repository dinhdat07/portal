package mapper

import (
	"errors"
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

	switch {
	case errors.Is(err, service.ErrInvalidInput),
		errors.Is(err, service.ErrInvalidToken),
		errors.Is(err, service.ErrInvalidUserID),
		errors.Is(err, service.ErrInvalidAction),
		errors.Is(err, service.ErrInvalidTimeRange),
		errors.Is(err, service.ErrIncorrectPassword),
		errors.Is(err, service.ErrPasswordConfirmationMismatch),
		errors.Is(err, service.ErrNewPasswordMustBeDifferent),
		errors.Is(err, service.ErrPasswordAlreadySet),
		errors.Is(err, service.ErrInvalidReplacementRole):
		return gstatus.Error(codes.InvalidArgument, err.Error())

	case errors.Is(err, service.ErrUnauthorized),
		errors.Is(err, service.ErrInvalidCredentials),
		errors.Is(err, service.ErrInvalidRefreshToken):
		return gstatus.Error(codes.Unauthenticated, err.Error())

	case errors.Is(err, service.ErrForbidden):
		return gstatus.Error(codes.PermissionDenied, err.Error())

	case errors.Is(err, service.ErrUserNotFound):
		return gstatus.Error(codes.NotFound, err.Error())

	case errors.Is(err, service.ErrRoleNotFound),
		errors.Is(err, service.ErrPermissionNotFound):
		return gstatus.Error(codes.NotFound, err.Error())

	case errors.Is(err, service.ErrEmailExists),
		errors.Is(err, service.ErrUsernameExists),
		errors.Is(err, service.ErrRoleCodeExists):
		return gstatus.Error(codes.AlreadyExists, err.Error())

	case errors.Is(err, service.ErrAccountNotVerified),
		errors.Is(err, service.ErrAccountDeleted),
		errors.Is(err, service.ErrUserAlreadyDeleted),
		errors.Is(err, service.ErrUserNotDeleted),
		errors.Is(err, service.ErrRoleInUse),
		errors.Is(err, service.ErrCannotModifySystemRole):
		return gstatus.Error(codes.FailedPrecondition, err.Error())

	default:
		return gstatus.Error(codes.Internal, err.Error())
	}
}
