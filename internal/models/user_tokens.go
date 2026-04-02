package models

import (
	"portal-system/internal/domain/enum"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserToken struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	TokenType enum.TokenType `gorm:"type:varchar(30);not null;index" json:"token_type"`
	TokenHash string         `gorm:"type:varchar(255);not null;uniqueIndex" json:"-"`
	ExpiresAt time.Time      `gorm:"not null;index" json:"expires_at"`
	UsedAt    *time.Time     `gorm:"default:null" json:"used_at,omitempty"`
	RevokedAt *time.Time     `gorm:"default:null" json:"revoked_at,omitempty"`
	CreatedAt time.Time      `gorm:"not null;autoCreateTime" json:"created_at"`

	User User `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
}

func (u *UserToken) BeforeCreate(tx *gorm.DB) error {
	u.ID = uuid.New()
	return nil
}
