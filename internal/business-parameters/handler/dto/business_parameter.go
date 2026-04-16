package dto

import (
	domain "github.com/devpablocristo/ponti-backend/internal/business-parameters/usecases/domain"
)

type BusinessParameterResponse struct {
	ID          int64  `json:"id"`
	Key         string `json:"key"`
	Value       string `json:"value"`
	Type        string `json:"type"`
	Category    string `json:"category"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type CreateBusinessParameterRequest struct {
	Key         string `json:"key" binding:"required"`
	Value       string `json:"value" binding:"required"`
	Type        string `json:"type" binding:"required,oneof=decimal integer string boolean"`
	Category    string `json:"category" binding:"required,oneof=units calculations business_rules"`
	Description string `json:"description"`
}

type UpdateBusinessParameterRequest struct {
	Key         string `json:"key" binding:"required"`
	Value       string `json:"value" binding:"required"`
	Type        string `json:"type" binding:"required,oneof=decimal integer string boolean"`
	Category    string `json:"category" binding:"required,oneof=units calculations business_rules"`
	Description string `json:"description"`
}

func FromDomain(param *domain.BusinessParameter) BusinessParameterResponse {
	return BusinessParameterResponse{
		ID:          param.ID,
		Key:         param.Key,
		Value:       param.Value,
		Type:        param.Type,
		Category:    param.Category,
		Description: param.Description,
		CreatedAt:   param.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   param.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func (req *CreateBusinessParameterRequest) ToDomain() *domain.BusinessParameter {
	return &domain.BusinessParameter{
		Key:         req.Key,
		Value:       req.Value,
		Type:        req.Type,
		Category:    req.Category,
		Description: req.Description,
	}
}

func (req *UpdateBusinessParameterRequest) ToDomain(id int64) *domain.BusinessParameter {
	return &domain.BusinessParameter{
		ID:          id,
		Key:         req.Key,
		Value:       req.Value,
		Type:        req.Type,
		Category:    req.Category,
		Description: req.Description,
	}
}
