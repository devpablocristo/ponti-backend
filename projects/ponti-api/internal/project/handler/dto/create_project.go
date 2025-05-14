package dto

type CreateProject struct {
	Project
}

type CreateProjectResponse struct {
	Message string `json:"message"`
	Project int64  `json:"project"`
}
