package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"portal-system/internal/domain"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AnnouncementType string

const (
	AnnouncementTypeInfo    AnnouncementType = "INFO"
	AnnouncementTypeWarning AnnouncementType = "WARNING"
	AnnouncementTypeAlert   AnnouncementType = "ALERT"
)

type AnnouncementStatus string

const (
	AnnouncementStatusPending    AnnouncementStatus = "PENDING"
	AnnouncementStatusProcessing AnnouncementStatus = "PROCESSING"
	AnnouncementStatusCompleted  AnnouncementStatus = "COMPLETED"
)

// RoleCodeArray is a custom type to handle PostgreSQL string array or JSON array for RoleCode
type RoleCodeArray []domain.RoleCode

// Value implements the driver.Valuer interface
func (a RoleCodeArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return "[]", nil
	}
	return json.Marshal(a)
}

// Scan implements the sql.Scanner interface
func (a *RoleCodeArray) Scan(value interface{}) error {
	if value == nil {
		*a = []domain.RoleCode{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSONB value: %v", value)
	}

	return json.Unmarshal(bytes, a)
}

type Announcement struct {
	ID          uuid.UUID          `gorm:"type:uuid;primaryKey"`
	Title       string             `gorm:"size:255;not null"`
	Content     string             `gorm:"type:text;not null"`
	Type        AnnouncementType   `gorm:"size:20;not null;index"`
	TargetRoles RoleCodeArray      `gorm:"type:jsonb"`
	Status      AnnouncementStatus `gorm:"size:20;not null;index;default:'PENDING'"`
	CreatedBy   uuid.UUID          `gorm:"type:uuid;not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time

	Creator User `gorm:"foreignKey:CreatedBy;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
}

func (a *Announcement) BeforeCreate(tx *gorm.DB) error {
	a.ID = uuid.New()
	if a.Status == "" {
		a.Status = AnnouncementStatusPending
	}
	return nil
}
