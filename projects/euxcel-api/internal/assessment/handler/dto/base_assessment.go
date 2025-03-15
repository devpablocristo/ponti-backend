package dto

import (
	"errors"
	"fmt"
	"time"

	"github.com/devpablocristo/monorepo/projects/qh/internal/assessment/usecases/domain"
)

// Assessment es el DTO que se utiliza para exponer/recibir datos vía JSON.
type Assessment struct {
	HRID        string        `json:"hr_id"`
	CandidateID string        `json:"candidate_id"`
	StartDate   time.Time     `json:"start_date"`
	EndDate     time.Time     `json:"end_date"`
	Status      string        `json:"status"` // tipo string en vez de domain.AssessmentStatus
	MaxDuration string        `json:"max_duration"`
	Skills      []SkillConfig `json:"skills"`
	Problem     Problem       `json:"problem"`
	UnitTests   []UnitTest    `json:"unit_tests"`
}

func (a Assessment) ToDomain() (*domain.Assessment, error) {
	duration, err := time.ParseDuration(a.MaxDuration) // Convierte string a time.Duration
	if err != nil {
		return nil, fmt.Errorf("invalid duration format: %w", err)
	}

	return &domain.Assessment{
		HRID:        a.HRID,
		CandidateID: a.CandidateID,
		StartDate:   a.StartDate,
		EndDate:     a.EndDate,
		Status:      domain.AssessmentStatus(a.Status),
		MaxDuration: duration,
		Skills:      ToDomainSkillConfigs(a.Skills),
		Problem:     a.Problem.ToDomain(),
		UnitTests:   ToDomainUnitTests(a.UnitTests),
	}, nil
}

// FromDomain convierte una entidad de dominio en el DTO correspondiente.
func FromDomain(a *domain.Assessment) (*Assessment, error) {
	if a == nil {
		return nil, errors.New("assessment cannot be nil")
	}
	return &Assessment{
		HRID:        a.HRID,
		CandidateID: a.CandidateID,
		StartDate:   a.StartDate,
		EndDate:     a.EndDate,
		Status:      string(a.Status), // Cast de domain.AssessmentStatus a string
		MaxDuration: a.MaxDuration.String(),
		Skills:      FromDomainToSkillConfigs(a.Skills),
		Problem:     FromDomainToProblem(&a.Problem),
		UnitTests:   FromDomainToUnitTests(a.UnitTests),
	}, nil
}

// SkillConfig es el DTO para la configuración de habilidades.
type SkillConfig struct {
	ID           string `json:"id"`
	AssessmentID string `json:"assessment_id,omitempty"`
	SkillName    string `json:"skill_name"`
	SkillLevel   string `json:"skill_level"`
}

// ToDomain convierte un DTO de SkillConfig a domain.SkillConfig.
func (s SkillConfig) ToDomain() domain.SkillConfig {
	return domain.SkillConfig{
		ID:           s.ID,
		AssessmentID: s.AssessmentID,
		SkillName:    s.SkillName,
		SkillLevel:   s.SkillLevel,
	}
}

// FromDomain convierte una entidad de dominio a SkillConfig (DTO).
func FromDomainSkillConfig(s domain.SkillConfig) SkillConfig {
	return SkillConfig{
		ID:           s.ID,
		AssessmentID: s.AssessmentID,
		SkillName:    s.SkillName,
		SkillLevel:   s.SkillLevel,
	}
}

// Problem es el DTO para el enunciado de un problema.
type Problem struct {
	ID           string `json:"id"`
	AssessmentID string `json:"assessment_id,omitempty"`
	Description  string `json:"description"`
}

// ToDomain convierte un DTO de Problem a domain.Problem.
func (p Problem) ToDomain() domain.Problem {
	return domain.Problem{
		ID:           p.ID,
		AssessmentID: p.AssessmentID,
		Description:  p.Description,
	}
}

// FromDomain convierte una entidad de dominio a Problem (DTO).
func FromDomainToProblem(p *domain.Problem) Problem {
	return Problem{
		ID:           p.ID,
		AssessmentID: p.AssessmentID,
		Description:  p.Description,
	}
}

// UnitTest es el DTO para una prueba unitaria.
type UnitTest struct {
	ID             string `json:"id"`
	AssessmentID   string `json:"assessment_id,omitempty"`
	TestName       string `json:"test_name"`
	InputData      string `json:"input_data"`
	ExpectedOutput string `json:"expected_output"`
}

// ToDomain convierte un DTO de UnitTest a domain.UnitTest.
func (u UnitTest) ToDomain() domain.UnitTest {
	return domain.UnitTest{
		ID:             u.ID,
		AssessmentID:   u.AssessmentID,
		TestName:       u.TestName,
		InputData:      u.InputData,
		ExpectedOutput: u.ExpectedOutput,
	}
}

// FromDomain convierte una entidad de dominio a UnitTest (DTO).
func FromDomainToUnitTest(u domain.UnitTest) UnitTest {
	return UnitTest{
		ID:             u.ID,
		AssessmentID:   u.AssessmentID,
		TestName:       u.TestName,
		InputData:      u.InputData,
		ExpectedOutput: u.ExpectedOutput,
	}
}

// ToDomainSkillConfigs mapea una slice de SkillConfig (DTO) a []domain.SkillConfig.
func ToDomainSkillConfigs(dtoSkills []SkillConfig) []domain.SkillConfig {
	domainSkills := make([]domain.SkillConfig, 0, len(dtoSkills))
	for _, dtoSkill := range dtoSkills {
		domainSkills = append(domainSkills, dtoSkill.ToDomain())
	}
	return domainSkills
}

// fromDomainSkillConfigs mapea una slice de domain.SkillConfig a []SkillConfig (DTO).
func FromDomainToSkillConfigs(domainSkills []domain.SkillConfig) []SkillConfig {
	dtoSkills := make([]SkillConfig, 0, len(domainSkills))
	for _, domainSkill := range domainSkills {
		dtoSkills = append(dtoSkills, FromDomainSkillConfig(domainSkill))
	}
	return dtoSkills
}

// ToDomainUnitTests mapea una slice de UnitTest (DTO) a []domain.UnitTest.
func ToDomainUnitTests(dtoUnitTests []UnitTest) []domain.UnitTest {
	domainUnitTests := make([]domain.UnitTest, 0, len(dtoUnitTests))
	for _, dtoUnitTest := range dtoUnitTests {
		domainUnitTests = append(domainUnitTests, dtoUnitTest.ToDomain())
	}
	return domainUnitTests
}

// fromDomainUnitTests mapea una slice de domain.UnitTest a []UnitTest (DTO).
func FromDomainToUnitTests(domainUnitTests []domain.UnitTest) []UnitTest {
	dtoUnitTests := make([]UnitTest, 0, len(domainUnitTests))
	for _, domainUnitTest := range domainUnitTests {
		dtoUnitTests = append(dtoUnitTests, FromDomainToUnitTest(domainUnitTest))
	}
	return dtoUnitTests
}
