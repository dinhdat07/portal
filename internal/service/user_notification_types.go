package service

import (
	"portal-system/internal/model"
)

type UserNotificationListFilter struct {
	Page       int
	PageSize   int
	UnreadOnly bool
}

type UserNotificationListResult struct {
	Notifications []model.UserNotification
	Total         int64
	Page          int
	PageSize      int
}
