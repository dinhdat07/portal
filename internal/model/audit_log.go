package model

import (
	"portal-system/internal/domain"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type AuditLog struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey"`

	Action domain.ActionName `gorm:"type:varchar(50);not null;index"`

	ActorUserID   *uuid.UUID       `gorm:"type:uuid;index"`
	ActorUsername *string          `gorm:"type:varchar(50)"`
	ActorEmail    *string          `gorm:"type:varchar(255)"`
	ActorRole     *domain.RoleCode `gorm:"type:varchar(20)"`

	TargetUserID   *uuid.UUID       `gorm:"type:uuid;index"`
	TargetUsername *string          `gorm:"type:varchar(50)"`
	TargetEmail    *string          `gorm:"type:varchar(255)"`
	TargetRole     *domain.RoleCode `gorm:"type:varchar(20)"`

	Metadata  *datatypes.JSON `gorm:"type:jsonb"`
	IPAddress *string         `gorm:"type:varchar(45)"`
	UserAgent *string         `gorm:"type:text"`

	CreatedAt time.Time `gorm:"not null;autoCreateTime;index:idx_action_logs_time,sort:desc"`
}

func (AuditLog) TableName() string {
	return "action_logs"
}

func (a *AuditLog) BeforeCreate(tx *gorm.DB) error {
	a.ID = uuid.New()
	return nil
}
