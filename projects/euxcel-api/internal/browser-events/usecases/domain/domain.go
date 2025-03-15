package domain

import "time"

type BrowserEvent struct {
	ID            string
	CandidateID   string
	AssessmentIDs []string // Cambiado de string a []string
	EventType     string
	Timestamp     time.Time
	TargetID      string
	Payload       map[string]any
}
