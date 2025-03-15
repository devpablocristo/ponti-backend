package models

import (
	"time"

	"github.com/devpablocristo/monorepo/projects/qh/internal/assessment/usecases/domain"
)

// Assessment representa la evaluación en la capa GORM.
type Assessment struct {
	ID          string     `gorm:"primaryKey"`
	HRID        string     `gorm:"index"`                     // Foreign Key a HR
	CandidateID string     `gorm:"index"`                     // Foreign Key a Candidate
	StartDate   time.Time  `gorm:"not null"`                  // Fecha de inicio
	EndDate     *time.Time `gorm:""`                          // Fecha de fin (nullable)
	Status      string     `gorm:"type:varchar(50);not null"` // "pending", "in_progress", "completed"
	// Se asume que en la BD se almacena el valor en entero (minutos)
	MaxDuration int64         `gorm:"not null"`                                            // Duración máxima en minutos
	Skills      []SkillConfig `gorm:"foreignKey:AssessmentID;constraint:OnDelete:CASCADE"` // Configuraciones de skills requeridas
	Problem     *Problem      `gorm:"foreignKey:AssessmentID;constraint:OnDelete:CASCADE"` // Problema asociado a la evaluación
	UnitTests   []UnitTest    `gorm:"foreignKey:AssessmentID;constraint:OnDelete:CASCADE"` // Pruebas unitarias
	CreatedAt   time.Time     `gorm:"autoCreateTime"`
	UpdatedAt   time.Time     `gorm:"autoUpdateTime"`
}

// SkillConfig representa la configuración de una skill requerida.
type SkillConfig struct {
	ID           string    `gorm:"primaryKey"`
	AssessmentID string    `gorm:"index;not null"`             // Foreign Key a Assessment
	SkillName    string    `gorm:"type:varchar(100);not null"` // Nombre de la skill
	SkillLevel   string    `gorm:"type:varchar(50);not null"`  // Nivel de habilidad
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}

// Problem representa el enunciado del problema para la evaluación.
type Problem struct {
	ID           string    `gorm:"primaryKey"`
	AssessmentID string    `gorm:"index;not null"`     // Foreign Key a Assessment
	Description  string    `gorm:"type:text;not null"` // Descripción del problema
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}

// UnitTest representa una prueba unitaria asociada a la evaluación.
type UnitTest struct {
	ID             string    `gorm:"primaryKey"`
	AssessmentID   string    `gorm:"index;not null"`             // Foreign Key a Assessment
	TestName       string    `gorm:"type:varchar(100);not null"` // Nombre del test
	InputData      string    `gorm:"type:text;not null"`         // Datos de entrada
	ExpectedOutput string    `gorm:"type:text;not null"`         // Salida esperada
	CreatedAt      time.Time `gorm:"autoCreateTime"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime"`
}

// ToDomain convierte este objeto GORM (DTO) a su equivalente en la capa de dominio.
func (dto Assessment) ToDomain() *domain.Assessment {
	return &domain.Assessment{
		ID:          dto.ID,
		HRID:        dto.HRID,
		CandidateID: dto.CandidateID,
		StartDate:   dto.StartDate,
		EndDate: func() time.Time {
			if dto.EndDate != nil {
				return *dto.EndDate
			}
			return time.Time{}
		}(),
		Status: domain.AssessmentStatus(dto.Status),
		// Convertimos el valor almacenado (en minutos) a una duración:
		MaxDuration: time.Duration(dto.MaxDuration) * time.Minute,
		Skills:      SkillConfigToDomain(dto.Skills),
		Problem: func() domain.Problem {
			p := ProblemToDomain(dto.Problem)
			if p == nil {
				return domain.Problem{}
			}
			return *p
		}(),
		UnitTests: UnitTestToDomain(dto.UnitTests),
	}
}

// FromDomainAssessment convierte un objeto del dominio a su equivalente GORM (DTO).
func FromDomainAssessment(assessment *domain.Assessment) *Assessment {
	var endDatePtr *time.Time
	if !assessment.EndDate.IsZero() {
		endDatePtr = &assessment.EndDate
	}

	return &Assessment{
		ID:          assessment.ID,
		HRID:        assessment.HRID,
		CandidateID: assessment.CandidateID,
		StartDate:   assessment.StartDate,
		EndDate:     endDatePtr,
		Status:      string(assessment.Status),
		// Convertimos la duración (en minutos) a un entero:
		MaxDuration: int64(assessment.MaxDuration.Minutes()),
		Skills:      SkillConfigFromDomainAssessment(assessment.Skills),
		Problem:     ProblemFromDomainAssessment(&assessment.Problem),
		UnitTests:   UnitTestFromDomainAssessment(assessment.UnitTests),
	}
}

//
// Funciones de conversión (sin prefijo "map")
//

// Convierte un slice de SkillConfig del modelo a SkillConfig del dominio.
func SkillConfigToDomain(gs []SkillConfig) []domain.SkillConfig {
	var ds []domain.SkillConfig
	for _, g := range gs {
		ds = append(ds, domain.SkillConfig{
			ID:           g.ID,
			AssessmentID: g.AssessmentID,
			SkillName:    g.SkillName,
			SkillLevel:   g.SkillLevel,
		})
	}
	return ds
}

// Convierte un slice de SkillConfig del dominio al modelo.
func SkillConfigFromDomainAssessment(ds []domain.SkillConfig) []SkillConfig {
	var gs []SkillConfig
	for _, d := range ds {
		gs = append(gs, SkillConfig{
			ID:           d.ID,
			AssessmentID: d.AssessmentID,
			SkillName:    d.SkillName,
			SkillLevel:   d.SkillLevel,
		})
	}
	return gs
}

// Convierte un objeto Problem del modelo a Problem del dominio.
func ProblemToDomain(g *Problem) *domain.Problem {
	if g == nil {
		return nil
	}
	return &domain.Problem{
		ID:           g.ID,
		AssessmentID: g.AssessmentID,
		Description:  g.Description,
	}
}

// Convierte un objeto Problem del dominio al modelo.
func ProblemFromDomainAssessment(d *domain.Problem) *Problem {
	if d == nil {
		return nil
	}
	return &Problem{
		ID:           d.ID,
		AssessmentID: d.AssessmentID,
		Description:  d.Description,
	}
}

// Convierte un slice de UnitTest del modelo a UnitTest del dominio.
func UnitTestToDomain(gu []UnitTest) []domain.UnitTest {
	var du []domain.UnitTest
	for _, g := range gu {
		du = append(du, domain.UnitTest{
			ID:             g.ID,
			AssessmentID:   g.AssessmentID,
			TestName:       g.TestName,
			InputData:      g.InputData,
			ExpectedOutput: g.ExpectedOutput,
		})
	}
	return du
}

// Convierte un slice de UnitTest del dominio al modelo.
func UnitTestFromDomainAssessment(du []domain.UnitTest) []UnitTest {
	var gu []UnitTest
	for _, d := range du {
		gu = append(gu, UnitTest{
			ID:             d.ID,
			AssessmentID:   d.AssessmentID,
			TestName:       d.TestName,
			InputData:      d.InputData,
			ExpectedOutput: d.ExpectedOutput,
		})
	}
	return gu
}
