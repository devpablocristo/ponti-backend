package domain

import shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"

type Labor struct {
	ID             int64
	Name           string
	ContractorName string
	Price          float64
	ProjectId      int64
	CategoryId     int64

	shareddomain.Base
}

type ListedLabor struct {
	ID             int64
	Name           string
	ContractorName string
	Price          float64
	ProjectId      int64
	CategoryId     int64

	shareddomain.Base
}
