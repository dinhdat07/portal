package mapper

import (
	commonv1 "portal-system/gen/go/common/v1"
	"portal-system/internal/model"
	"portal-system/internal/service"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func AnnouncementModelToPB(model *model.Announcement) *commonv1.Announcement {
	if model == nil {
		return nil
	}

	targetRoles := make([]string, len(model.TargetRoles))
	for i, v := range model.TargetRoles {
		targetRoles[i] = string(v)
	}

	return &commonv1.Announcement{
		Id:             model.ID.String(),
		Title:          model.Title,
		Content:        model.Content,
		Type:           string(model.Type),
		TargetRoles:    targetRoles,
		DispatchStatus: string(model.DispatchStatus),
		CreatedBy:      model.CreatedBy.String(),
		CreatedAt:      timestamppb.New(model.CreatedAt),
		UpdatedAt:      timestamppb.New(model.UpdatedAt),
	}
}

func ListAnnouncementsResultToPB(result *service.AnnouncementListResult) []*commonv1.Announcement {
	if result == nil {
		return nil
	}

	pbList := make([]*commonv1.Announcement, 0, len(result.Announcements))
	for i := range result.Announcements {
		pbList = append(pbList, AnnouncementModelToPB(&result.Announcements[i]))
	}

	return pbList
}
