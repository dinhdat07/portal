package mappers

import (
	commonv1 "portal-system/gen/go/common/v1"
	"portal-system/internal/domain"
	"portal-system/internal/models"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func UserModelToPB(u *models.User) *commonv1.User {
	if u == nil {
		return nil
	}

	out := &commonv1.User{
		Id:        u.ID.String(),
		Email:     u.Email,
		Username:  u.Username,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Role:      RoleSummaryModelToPB(&u.Role),
		Status:    UserStatusToPB(u.Status),
		CreatedAt: timestamppb.New(u.CreatedAt),
		UpdatedAt: timestamppb.New(u.UpdatedAt),
	}

	if u.DOB != nil {
		v := u.DOB.Format("2006-01-02")
		out.Dob = &v
	}
	if u.EmailVerifiedAt != nil {
		out.EmailVerifiedAt = timestamppb.New(*u.EmailVerifiedAt)
	}
	if u.LastLoginAt != nil {
		out.LastLoginAt = timestamppb.New(*u.LastLoginAt)
	}
	if u.DeletedAt.Valid {
		out.DeletedAt = timestamppb.New(u.DeletedAt.Time)
	}
	if u.DeletedBy != nil {
		v := u.DeletedBy.String()
		out.DeletedBy = &v
	}

	return out
}

func RoleSummaryModelToPB(role *models.Role) *commonv1.RoleSummary {
	if role == nil {
		return nil
	}

	return &commonv1.RoleSummary{
		Id:       role.ID.String(),
		Code:     string(role.Code),
		Name:     role.Name,
		IsSystem: role.IsSystem,
	}
}

func UserStatusToPB(status domain.UserStatus) commonv1.UserStatus {
	switch status {
	case domain.StatusPending:
		return commonv1.UserStatus_USER_STATUS_PENDING_VERIFICATION
	case domain.StatusActive:
		return commonv1.UserStatus_USER_STATUS_ACTIVE
	case domain.StatusDeleted:
		return commonv1.UserStatus_USER_STATUS_DELETED
	default:
		return commonv1.UserStatus_USER_STATUS_UNSPECIFIED
	}
}

func UserStatusFromPB(status commonv1.UserStatus) (domain.UserStatus, bool) {
	switch status {
	case commonv1.UserStatus_USER_STATUS_PENDING_VERIFICATION:
		return domain.StatusPending, true
	case commonv1.UserStatus_USER_STATUS_ACTIVE:
		return domain.StatusActive, true
	case commonv1.UserStatus_USER_STATUS_DELETED:
		return domain.StatusDeleted, true
	default:
		return "", false
	}
}
