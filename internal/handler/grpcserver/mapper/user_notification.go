package mapper

import (
	commonv1 "portal-system/gen/go/common/v1"
	"portal-system/internal/model"
	"portal-system/internal/service"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func UserNotificationModelToPB(model *model.UserNotification) *commonv1.UserNotification {
	if model == nil {
		return nil
	}

	var readAt *timestamppb.Timestamp
	if model.ReadAt != nil {
		readAt = timestamppb.New(*model.ReadAt)
	}

	return &commonv1.UserNotification{
		Id:           model.ID.String(),
		UserId:       model.UserID.String(),
		Announcement: AnnouncementModelToPB(&model.Announcement),
		IsRead:       model.IsRead,
		ReadAt:       readAt,
		CreatedAt:    timestamppb.New(model.CreatedAt),
	}
}

func ListUserNotificationsResultToPB(result *service.UserNotificationListResult) []*commonv1.UserNotification {
	if result == nil {
		return nil
	}

	pbList := make([]*commonv1.UserNotification, 0, len(result.Notifications))
	for i := range result.Notifications {
		pbList = append(pbList, UserNotificationModelToPB(&result.Notifications[i]))
	}

	return pbList
}
