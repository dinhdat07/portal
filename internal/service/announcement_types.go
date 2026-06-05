package service

import (
	"portal-system/internal/domain"
	"portal-system/internal/model"
)

type CreateAnnouncementInput struct {
	Title       string             `json:"title"`
	Content     string             `json:"content"`
	Type        string             `json:"type"`
	TargetRoles []domain.RoleCode  `json:"target_roles"`
}

type AnnouncementListFilter struct {
	Page             int
	PageSize         int
	AnnouncementType string
}

type AnnouncementListResult struct {
	Announcements []model.Announcement
	Total         int64
	Page          int
	PageSize      int
}
