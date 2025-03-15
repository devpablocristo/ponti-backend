package dto

type CreateCandidate struct {
	Candidate
}

// Response
type CreateCandidateResponse struct {
	Message     string `json:"message"`
	CandidateID string `json:"candidate_id"`
}
