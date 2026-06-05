package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserNotification struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID         uuid.UUID `gorm:"type:uuid;not null;index:idx_user_notification_user_announcement,unique"`
	AnnouncementID uuid.UUID `gorm:"type:uuid;not null;index:idx_user_notification_user_announcement,unique;index"`
	IsRead         bool      `gorm:"default:false;index"`
	ReadAt         *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time

	User         User         `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Announcement Announcement `gorm:"foreignKey:AnnouncementID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (u *UserNotification) BeforeCreate(tx *gorm.DB) error {
	u.ID = uuid.New()
	return nil
}
