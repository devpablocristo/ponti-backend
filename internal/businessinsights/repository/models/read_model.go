package models

import (
	"time"

	"github.com/google/uuid"
)

type ReadModel struct {
	InsightID uuid.UUID `gorm:"column:insight_id;type:uuid;primaryKey"`
	UserID    string    `gorm:"column:user_id;primaryKey"`
	ReadAt    time.Time `gorm:"column:read_at;not null"`
}

func (ReadModel) TableName() string {
	return "business_insight_reads"
}
