// Package usecases proporciona validadores para los casos de uso de reportes.
package usecases

import (
	"fmt"

	"github.com/devpablocristo/ponti-backend/internal/report/usecases/domain"
)

// ReportFilterValidator valida filtros de reportes.
type ReportFilterValidator struct{}

// NewReportFilterValidator crea una nueva instancia del validador.
func NewReportFilterValidator() *ReportFilterValidator {
	return &ReportFilterValidator{}
}

// ValidateRequiredWorkspaceFilter valida que el workspace esté completo:
// customer_id, project_id y campaign_id son obligatorios; field_id es opcional.
func (v *ReportFilterValidator) ValidateRequiredWorkspaceFilter(filter domain.SummaryResultsFilter) error {
	if filter.ProjectID == nil || filter.CustomerID == nil || filter.CampaignID == nil {
		return fmt.Errorf("customer_id, project_id and campaign_id are required")
	}
	return nil
}
