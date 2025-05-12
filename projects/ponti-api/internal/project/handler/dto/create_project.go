package dto

type CreateProject struct {
	Project
}

type CreateProjectResponse struct {
	Message string   `json:"message"`
	Project *Project `json:"project"`
}
