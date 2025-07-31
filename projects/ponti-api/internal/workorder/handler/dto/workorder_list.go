package dto

import (
	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
)

// Request/Response para listado

type WorkorderListResponse struct {
	PageInfo types.PageInfo    `json:"page_info"`
	Items    []WorkorderDetail `json:"items"`
}
