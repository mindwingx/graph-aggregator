package models

import (
	"gorm.io/gorm"
	"time"
)

type TransactionalOutboxMessages struct {
	gorm.Model
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt time.Time `json:"deleted_at,omitempty" gorm:"default:null"`
	ID        uint64    `json:"id" gorm:"primaryKey"`
	Uuid      string    `json:"uuid"` // this will be used for uniqueness and tracing
	Message   string    `json:"message"`
}
