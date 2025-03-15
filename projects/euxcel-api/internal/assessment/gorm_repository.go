package assessment

import (
	gorm "github.com/alphacodinggroup/euxcel-backend/pkg/databases/sql/gorm"
)

type repository struct {
	db gorm.Repository
}

func NewRepository(db gorm.Repository) Repository {
	return &repository{
		db: db,
	}
}
