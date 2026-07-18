// Package report proporciona funcionalidades para generar reportes financieros y operativos.
package report

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/devpablocristo/ponti-backend/internal/report/repository/models"
	"github.com/devpablocristo/ponti-backend/internal/report/usecases"
	"github.com/devpablocristo/ponti-backend/internal/report/usecases/domain"
	"github.com/devpablocristo/ponti-backend/internal/report/usecases/mappers"
	sharedmodels "github.com/devpablocristo/ponti-backend/internal/shared/models"
)

// ===== PORTS (Hexagonal Architecture) =====

// ReportRepositoryPort define la interfaz del repositorio (Puerto de salida).
type ReportRepositoryPort interface {
	GetFieldCropMetrics(domain.ReportFilter) ([]domain.FieldCropMetric, error)
	GetProjectInfo(int64) (*domain.ProjectInfo, error)
	BuildFieldCrop(domain.ReportFilter) (*domain.FieldCrop, error)
	GetInvestorContributionReport(context.Context, domain.ReportFilter) (*domain.InvestorContributionReport, error)
	GetSummaryResults(domain.SummaryResultsFilter) ([]domain.SummaryResults, error)
}

// ReportUseCasePort define la interfaz del caso de uso (Puerto de entrada).
type ReportUseCasePort interface {
	GetFieldCropReport(domain.ReportFilter) (*domain.FieldCrop, error)
	GetInvestorContributionReport(context.Context, domain.ReportFilter) (*domain.InvestorContributionReport, error)
	GetSummaryResultsReport(context.Context, domain.SummaryResultsFilter) (*domain.SummaryResultsResponse, error)
}

type BusinessInsightsNotifier interface {
	NotifyOperatingResultNegative(ctx context.Context, tenantID uuid.UUID, actor string, issue OperatingResultNegativeInput) error
	MaybeResolveOperatingResultNegative(ctx context.Context, tenantID uuid.UUID, projectID string) error
}

type OperatingResultNegativeInput struct {
	ProjectID               string
	CustomerID              string
	CampaignID              string
	TotalOperatingResultUSD string
	ProjectReturnPct        string
	TotalInvestedProjectUSD string
	NegativeCrops           []OperatingResultNegativeCrop
}

type OperatingResultNegativeCrop struct {
	CropID             string
	CropName           string
	OperatingResultUSD string
	SurfaceHa          string
	ReturnPct          string
}

// ===== USE CASE IMPLEMENTATION =====

// ReportUseCase implementa la lógica de negocio para reportes.
type ReportUseCase struct {
	repository    ReportRepositoryPort
	validator     *usecases.ReportFilterValidator
	summaryMapper *mappers.SummaryResponseMapper
	notifier      BusinessInsightsNotifier
}

// NewReportUseCase crea una nueva instancia del caso de uso.
func NewReportUseCase(repository ReportRepositoryPort) *ReportUseCase {
	return &ReportUseCase{
		repository:    repository,
		validator:     usecases.NewReportFilterValidator(),
		summaryMapper: mappers.NewSummaryResponseMapper(),
	}
}

func (uc *ReportUseCase) SetBusinessInsightsNotifier(n BusinessInsightsNotifier) {
	uc.notifier = n
}

// ===== REPORTE POR CAMPO/CULTIVO =====

// GetFieldCropReport obtiene el reporte por campo/cultivo.
func (uc *ReportUseCase) GetFieldCropReport(filters domain.ReportFilter) (*domain.FieldCrop, error) {

	// Obtener reporte del repositorio
	report, err := uc.repository.BuildFieldCrop(filters)
	if err != nil {
		return nil, fmt.Errorf("error al obtener reporte de campo/cultivo: %w", err)
	}

	return report, nil
}

// GetInvestorContributionReport obtiene el reporte de aportes de inversores.
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

// GetSummaryResultsReport obtiene el reporte de resumen de resultados.
func (uc *ReportUseCase) GetSummaryResultsReport(ctx context.Context, filters domain.SummaryResultsFilter) (*domain.SummaryResultsResponse, error) {
	// Validar workspace completo: customer_id + project_id + campaign_id
	if err := uc.validator.ValidateRequiredWorkspaceFilter(filters); err != nil {
		return nil, err
	}

	// Obtener datos del repositorio
	results, err := uc.repository.GetSummaryResults(filters)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo resumen de resultados: %w", err)
	}

	// Retornar respuesta vacía si no hay resultados
	if len(results) == 0 {
		return uc.buildEmptySummaryResponse(), nil
	}

	// Construir respuesta con datos
	report, err := uc.buildSummaryResponse(results)
	if err != nil {
		return nil, err
	}
	uc.evaluateOperatingResultInsight(ctx, report)
	return report, nil
}

// ===== FUNCIONES PRIVADAS (DRY) =====

// buildEmptySummaryResponse construye una respuesta vacía usando el mapper.
func (uc *ReportUseCase) buildEmptySummaryResponse() *domain.SummaryResultsResponse {
	return uc.summaryMapper.BuildEmptyResponse()
}

// buildSummaryResponse construye la respuesta completa con datos usando el mapper.
func (uc *ReportUseCase) buildSummaryResponse(results []domain.SummaryResults) (*domain.SummaryResultsResponse, error) {
	// Obtener información del proyecto del primer resultado
	projectInfo, err := uc.repository.GetProjectInfo(results[0].ProjectID)
	if err != nil {
		return nil, fmt.Errorf("error getting project information: %w", err)
	}

	// Calcular totales del proyecto
	totales := uc.calculateProjectTotals(results)

	// Usar el mapper para construir la respuesta
	return uc.summaryMapper.BuildResponse(projectInfo, results, totales), nil
}

// calculateProjectTotals calcula los totales del proyecto.
func (uc *ReportUseCase) calculateProjectTotals(results []domain.SummaryResults) *domain.ProjectTotals {
	// Usar el mapper para convertir a punteros
	resultsPtr := uc.summaryMapper.ConvertToPointers(results)
	return models.CalculateProjectTotals(resultsPtr)
}

func (uc *ReportUseCase) evaluateOperatingResultInsight(ctx context.Context, report *domain.SummaryResultsResponse) {
	if uc == nil || uc.notifier == nil || report == nil || report.ProjectID <= 0 {
		return
	}
	orgID, ok := sharedmodels.OrgIDFromContext(ctx)
	if !ok {
		return
	}
	projectID := strconv.FormatInt(report.ProjectID, 10)
	if !report.Totals.TotalOperatingResultUsd.LessThan(decimal.Zero) {
		_ = uc.notifier.MaybeResolveOperatingResultNegative(ctx, orgID, projectID)
		return
	}
	actor, _ := sharedmodels.ActorFromContext(ctx)
	_ = uc.notifier.NotifyOperatingResultNegative(ctx, orgID, actor, buildOperatingResultNegativeInput(report))
}

func buildOperatingResultNegativeInput(report *domain.SummaryResultsResponse) OperatingResultNegativeInput {
	if report == nil {
		return OperatingResultNegativeInput{}
	}
	negativeCrops := make([]OperatingResultNegativeCrop, 0)
	for _, crop := range report.Crops {
		if !crop.OperatingResultUsd.LessThan(decimal.Zero) {
			continue
		}
		if len(negativeCrops) >= 5 {
			continue
		}
		negativeCrops = append(negativeCrops, OperatingResultNegativeCrop{
			CropID:             strconv.FormatInt(crop.CropID, 10),
			CropName:           crop.CropName,
			OperatingResultUSD: crop.OperatingResultUsd.String(),
			SurfaceHa:          crop.SurfaceHa.String(),
			ReturnPct:          crop.CropReturnPct.String(),
		})
	}
	return OperatingResultNegativeInput{
		ProjectID:               strconv.FormatInt(report.ProjectID, 10),
		CustomerID:              strconv.FormatInt(report.CustomerID, 10),
		CampaignID:              strconv.FormatInt(report.CampaignID, 10),
		TotalOperatingResultUSD: report.Totals.TotalOperatingResultUsd.String(),
		ProjectReturnPct:        report.Totals.ProjectReturnPct.String(),
		TotalInvestedProjectUSD: report.Totals.TotalInvestedProjectUsd.String(),
		NegativeCrops:           negativeCrops,
	}
}
