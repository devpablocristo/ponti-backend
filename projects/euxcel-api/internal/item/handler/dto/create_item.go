package dto

type CreateAssessment struct {
	Item
}

// Response
type CreateAssessmentResponse struct {
	Message      string `json:"message"`
	AssessmentID string `json:"assessment_id"`
}
