package mappers

import (
	"errors"
	"portal-system/internal/services"

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
	case errors.Is(err, services.ErrInvalidInput),
		errors.Is(err, services.ErrInvalidToken),
		errors.Is(err, services.ErrInvalidUserID),
		errors.Is(err, services.ErrInvalidAction),
		errors.Is(err, services.ErrInvalidTimeRange),
		errors.Is(err, services.ErrIncorrectPassword),
		errors.Is(err, services.ErrPasswordConfirmationMismatch),
		errors.Is(err, services.ErrNewPasswordMustBeDifferent),
		errors.Is(err, services.ErrPasswordAlreadySet),
		errors.Is(err, services.ErrInvalidReplacementRole):
		return gstatus.Error(codes.InvalidArgument, err.Error())

	case errors.Is(err, services.ErrUnauthorized),
		errors.Is(err, services.ErrInvalidCredentials),
		errors.Is(err, services.ErrInvalidRefreshToken):
		return gstatus.Error(codes.Unauthenticated, err.Error())

	case errors.Is(err, services.ErrForbidden):
		return gstatus.Error(codes.PermissionDenied, err.Error())

	case errors.Is(err, services.ErrUserNotFound):
		return gstatus.Error(codes.NotFound, err.Error())

	case errors.Is(err, services.ErrRoleNotFound),
		errors.Is(err, services.ErrPermissionNotFound):
		return gstatus.Error(codes.NotFound, err.Error())

	case errors.Is(err, services.ErrEmailExists),
		errors.Is(err, services.ErrUsernameExists),
		errors.Is(err, services.ErrRoleCodeExists):
		return gstatus.Error(codes.AlreadyExists, err.Error())

	case errors.Is(err, services.ErrAccountNotVerified),
		errors.Is(err, services.ErrAccountDeleted),
		errors.Is(err, services.ErrUserAlreadyDeleted),
		errors.Is(err, services.ErrUserNotDeleted),
		errors.Is(err, services.ErrRoleInUse),
		errors.Is(err, services.ErrCannotModifySystemRole):
		return gstatus.Error(codes.FailedPrecondition, err.Error())

	default:
		return gstatus.Error(codes.Internal, err.Error())
	}
}
