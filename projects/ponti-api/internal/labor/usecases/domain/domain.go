package domain

import "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/base"

type Labor struct {
	ID             int64
	Name           string
	ContractorName string
	Price          float64
	ProjectId      int64
	CategoryId     int64

	base.BaseModel
}

type ListedLabor struct {
	ID             int64
	Name           string
	ContractorName string
	Price          float64
	projectId      int64
	CategoryId     int64

	base.BaseModel
}
