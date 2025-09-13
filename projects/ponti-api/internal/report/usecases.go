// Package report proporciona funcionalidades para generar reportes financieros y operativos
package report

import (
	"context"
	"fmt"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/report/usecases/domain"
)

// ReportRepositoryPort define la interfaz del repositorio
type ReportRepositoryPort interface {
	GetFieldCropMetrics(filters domain.ReportFilter) ([]domain.FieldCropMetric, error)
	GetProjectInfo(projectID int64) (*domain.ProjectInfo, error)
	BuildFieldCrop(filters domain.ReportFilter) (*domain.FieldCrop, error)
	GetInvestorContributionReport(ctx context.Context, filter domain.ReportFilter) (*domain.InvestorContributionReport, error)
}

// ReportUseCasePort define la interfaz del caso de uso
type ReportUseCasePort interface {
	GetFieldCropReport(filters domain.ReportFilter) (*domain.FieldCrop, error)
	GetInvestorContributionReport(ctx context.Context, filter domain.ReportFilter) (*domain.InvestorContributionReport, error)
}

// ReportUseCase implementa la lógica de negocio para reportes
type ReportUseCase struct {
	repository ReportRepositoryPort
}

// NewReportUseCase crea una nueva instancia del caso de uso
func NewReportUseCase(repository ReportRepositoryPort) *ReportUseCase {
	return &ReportUseCase{
		repository: repository,
	}
}

// ===== REPORTE POR CAMPO/CULTIVO =====

// GetFieldCropReport obtiene el reporte por campo/cultivo
func (uc *ReportUseCase) GetFieldCropReport(filters domain.ReportFilter) (*domain.FieldCrop, error) {

	// Obtener reporte del repositorio
	report, err := uc.repository.BuildFieldCrop(filters)
	if err != nil {
		return nil, fmt.Errorf("error al obtener reporte de campo/cultivo: %w", err)
	}

	return report, nil
}

// ===== VALIDACIONES =====

// GetInvestorContributionReport obtiene el reporte de aportes de inversores
func (uc *ReportUseCase) GetInvestorContributionReport(ctx context.Context, filter domain.ReportFilter) (*domain.InvestorContributionReport, error) {
	// Todos los filtros son opcionales - no hay validaciones requeridas

	// Obtener datos desde el repository (que consulta la vista de la DB)
	report, err := uc.repository.GetInvestorContributionReport(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo reporte de aportes de inversores: %w", err)
	}

	return report, nil
}
