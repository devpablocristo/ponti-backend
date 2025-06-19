package base

import "time"

type BaseModel struct {
	CreatedAt time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time  `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt *time.Time `gorm:"column:deleted_at"`
	CreatedBy *int64     `gorm:"column:created_by"`
	UpdatedBy *int64     `gorm:"column:updated_by"`
	DeletedBy *int64     `gorm:"column:deleted_by"`
}
