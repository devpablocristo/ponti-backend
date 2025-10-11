// Package mappers proporciona funciones de mapeo para los casos de uso de reportes
package mappers

import (
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/report/usecases/domain"
)

// SummaryResponseMapper maneja el mapeo de respuestas de resumen
type SummaryResponseMapper struct{}

// NewSummaryResponseMapper crea una nueva instancia del mapper
func NewSummaryResponseMapper() *SummaryResponseMapper {
	return &SummaryResponseMapper{}
}

// BuildEmptyResponse construye una respuesta vacía
func (m *SummaryResponseMapper) BuildEmptyResponse() *domain.SummaryResultsResponse {
	return &domain.SummaryResultsResponse{
		Crops:        []domain.SummaryResults{},
		Totals:       domain.ProjectTotals{},
		GeneralCrops: domain.GeneralCrops{},
	}
}

// BuildResponse construye la respuesta completa con datos
func (m *SummaryResponseMapper) BuildResponse(
	projectInfo *domain.ProjectInfo,
	crops []domain.SummaryResults,
	totals *domain.ProjectTotals,
) *domain.SummaryResultsResponse {
	// GRAL CULTIVOS debe tener los mismos valores que los totales del proyecto
	generalCrops := domain.GeneralCrops{
		TotalSurfaceHa:          totals.TotalSurfaceHa,
		TotalNetIncomeUsd:       totals.TotalNetIncomeUsd,
		TotalDirectCostsUsd:     totals.TotalDirectCostsUsd,
		TotalRentUsd:            totals.TotalRentUsd,
		TotalStructureUsd:       totals.TotalStructureUsd,
		TotalInvestedProjectUsd: totals.TotalInvestedProjectUsd,
		TotalOperatingResultUsd: totals.TotalOperatingResultUsd,
		ProjectReturnPct:        totals.ProjectReturnPct,
	}

	return &domain.SummaryResultsResponse{
		ProjectID:    projectInfo.ProjectID,
		ProjectName:  projectInfo.ProjectName,
		CustomerID:   projectInfo.CustomerID,
		CustomerName: projectInfo.CustomerName,
		CampaignID:   projectInfo.CampaignID,
		CampaignName: projectInfo.CampaignName,
		Crops:        crops,
		Totals:       *totals,
		GeneralCrops: generalCrops,
	}
}

// ConvertToPointers convierte un slice de valores a un slice de punteros
func (m *SummaryResponseMapper) ConvertToPointers(results []domain.SummaryResults) []*domain.SummaryResults {
	pointers := make([]*domain.SummaryResults, len(results))
	for i := range results {
		pointers[i] = &results[i]
	}
	return pointers
}
