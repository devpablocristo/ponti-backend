package dto

import "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project/usecases/domain"

type CreateProject struct {
	Project
}

type CreateProjectResponse struct {
	Message string          `json:"message"`
	Project *domain.Project `json:"project"`
}
