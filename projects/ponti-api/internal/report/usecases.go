// Package report proporciona funcionalidades para generar reportes financieros y operativos
package report

import (
	"context"
	"fmt"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/report/repository/models"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/report/usecases/domain"
)

// ReportRepositoryPort define la interfaz del repositorio
type ReportRepositoryPort interface {
	GetFieldCropMetrics(filters domain.ReportFilter) ([]domain.FieldCropMetric, error)
	GetProjectInfo(projectID int64) (*domain.ProjectInfo, error)
	BuildFieldCrop(filters domain.ReportFilter) (*domain.FieldCrop, error)
	GetInvestorContributionReport(ctx context.Context, filter domain.ReportFilter) (*domain.InvestorContributionReport, error)
	GetSummaryResults(filters domain.SummaryResultsFilter) ([]domain.SummaryResults, error)
}

// ReportUseCasePort define la interfaz del caso de uso
type ReportUseCasePort interface {
	GetFieldCropReport(filters domain.ReportFilter) (*domain.FieldCrop, error)
	GetInvestorContributionReport(ctx context.Context, filter domain.ReportFilter) (*domain.InvestorContributionReport, error)
	GetSummaryResultsReport(filters domain.SummaryResultsFilter) (*domain.SummaryResultsResponse, error)
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

// validateAtLeastOneFilter valida que al menos un filtro esté presente
func (uc *ReportUseCase) validateAtLeastOneFilter(projectID, customerID, campaignID, fieldID *int64) error {
	if projectID == nil && customerID == nil && campaignID == nil && fieldID == nil {
		return fmt.Errorf("at least one filter must be specified (project_id, customer_id, campaign_id, or field_id)")
	}
	return nil
}

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

// ===== REPORTE DE RESUMEN DE RESULTADOS =====

// GetSummaryResultsReport obtiene el reporte de resumen de resultados
func (uc *ReportUseCase) GetSummaryResultsReport(filters domain.SummaryResultsFilter) (*domain.SummaryResultsResponse, error) {
	// Validar que al menos un filtro esté presente
	if err := uc.validateAtLeastOneFilter(filters.ProjectID, filters.CustomerID, filters.CampaignID, filters.FieldID); err != nil {
		return nil, err
	}

	// Obtener datos del repositorio
	results, err := uc.repository.GetSummaryResults(filters)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo resumen de resultados: %w", err)
	}

	if len(results) == 0 {
		return &domain.SummaryResultsResponse{
			Crops:  []domain.SummaryResults{},
			Totals: domain.ProjectTotals{},
		}, nil
	}

	// Obtener información del proyecto (usar el primer resultado)
	projectID := results[0].ProjectID
	projectInfo, err := uc.repository.GetProjectInfo(projectID)
	if err != nil {
		return nil, fmt.Errorf("error getting project information: %w", err)
	}

	// Calcular totales del proyecto
	resultsPtr := make([]*domain.SummaryResults, len(results))
	for i := range results {
		resultsPtr[i] = &results[i]
	}
	totales := models.CalculateProjectTotals(resultsPtr)

	// Construir respuesta
	response := &domain.SummaryResultsResponse{
		ProjectID:    projectInfo.ProjectID,
		ProjectName:  projectInfo.ProjectName,
		CustomerID:   projectInfo.CustomerID,
		CustomerName: projectInfo.CustomerName,
		CampaignID:   projectInfo.CampaignID,
		CampaignName: projectInfo.CampaignName,
		Crops:        results,
		Totals:       *totales,
	}

	return response, nil
}
