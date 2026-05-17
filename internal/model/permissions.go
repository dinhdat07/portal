package model

import "github.com/google/uuid"

type Permission struct {
	ID   uuid.UUID `gorm:"type:uuid;primaryKey"`
	Code string    `gorm:"size:100;uniqueIndex;not null"`
	Name string    `gorm:"size:100;not null"`
}
