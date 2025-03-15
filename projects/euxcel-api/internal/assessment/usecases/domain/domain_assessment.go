package domain

import "time"

// AssessmentStatus representa valores de estado posibles para una evaluación.
type AssessmentStatus string

// Constantes de AssessmentStatus.
const (
	StatusPending    AssessmentStatus = "pending"     // Pendiente
	StatusInProgress AssessmentStatus = "in_progress" // En progreso
	StatusCompleted  AssessmentStatus = "completed"   // Completado
)

// Assessment representa una entidad de evaluación.
type Assessment struct {
	ID          string           // Clave primaria
	HRID        string           // Clave foránea hacia HR
	CandidateID string           // Clave foránea hacia Candidate
	StartDate   time.Time        // Fecha de inicio de la evaluación
	EndDate     time.Time        // Fecha de finalización de la evaluación
	Status      AssessmentStatus // Estado: "pending", "in_progress", "completed"
	MaxDuration time.Duration    // Duración máxima (en minutos)
	Skills      []SkillConfig    // Lista de configuraciones de habilidades requeridas
	Problem     Problem          // Enunciado del problema para la evaluación
	UnitTests   []UnitTest       // Lista de pruebas unitarias
}

// SkillConfig representa la configuración de una habilidad requerida en la evaluación.
type SkillConfig struct {
	ID           string // Clave primaria
	AssessmentID string // Clave foránea hacia Assessment
	SkillName    string // Nombre de la habilidad (p. ej., "Golang", "Docker", etc.)
	SkillLevel   string // Nivel de dominio (p. ej., principiante, intermedio, experto)
}

// Problem almacena el enunciado del problema generado para la evaluación.
type Problem struct {
	ID           string // Clave primaria
	AssessmentID string // Clave foránea hacia Assessment
	Description  string // Descripción detallada del problema
}

// UnitTest representa una prueba unitaria para la evaluación.
type UnitTest struct {
	ID             string // Clave primaria
	AssessmentID   string // Clave foránea hacia Assessment
	TestName       string // Nombre de la prueba
	InputData      string // Datos de entrada para la prueba
	ExpectedOutput string // Salida esperada para la prueba
}
