package domain

type ExperienceLevel string

const (
	Trainee    ExperienceLevel = "trainee"
	Junior     ExperienceLevel = "junior"
	Mid        ExperienceLevel = "mid"
	SemiSenior ExperienceLevel = "semi-senior"
	Senior     ExperienceLevel = "senior"
)

type Experience struct {
	Level ExperienceLevel
	Rank  int
}

var (
	ExperienceTrainee    = Experience{Level: Trainee, Rank: 1}
	ExperienceJunior     = Experience{Level: Junior, Rank: 2}
	ExperienceMid        = Experience{Level: Mid, Rank: 3}
	ExperienceSemiSenior = Experience{Level: SemiSenior, Rank: 4}
	ExperienceSenior     = Experience{Level: Senior, Rank: 5}
)

type Candidate struct {
	ID             string
	PersonID       string
	Email          string
	Experience     Experience
	AssessmentsIDs []AssessmentID
}

type AssessmentID string
