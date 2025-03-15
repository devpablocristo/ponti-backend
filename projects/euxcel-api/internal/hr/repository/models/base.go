package models

type HR struct {
	ID     string `gorm:"primaryKey"`
	UserID string `gorm:"index"`
}
