package dto

// CreateProject is the DTO for the create request of a project.
// It embeds the base Project DTO.
type CreateProject struct {
	Project
}

type CreateProjectResponse struct {
	Message    string   `json:"message"`
	NewProject *Project `json:"new_project"`
}
