package dto

import (
	"time"

	"github.com/devpablocristo/ponti-backend/internal/data-integrity/usecases/domain"
	filters "github.com/devpablocristo/ponti-backend/internal/shared/filters"
)

// maxSummarySampleItems limita la muestra de precios tentativos del resumen.
const maxSummarySampleItems = 5

// DataIntegritySummaryResponse replica el envelope de evidencia de los reads
// operacionales que consume Axis (source, workspace, filters, captured_at,
// summary), agregando costs-check + tentative-prices en una sola respuesta.
type DataIntegritySummaryResponse struct {
	Source     string               `json:"source"`
	Workspace  map[string]int64     `json:"workspace"`
	Filters    map[string]int64     `json:"filters"`
	CapturedAt string               `json:"captured_at"`
	Summary    DataIntegritySummary `json:"summary"`
}

// DataIntegritySummary agrupa los resúmenes por control de integridad.
type DataIntegritySummary struct {
	CostsCheck      CostsCheckSummary      `json:"costs_check"`
	TentativePrices TentativePricesSummary `json:"tentative_prices"`
}

// CostsCheckSummary resume el resultado de los controles de coherencia de costos.
type CostsCheckSummary struct {
	Status         string                 `json:"status"`
	FailedChecks   int                    `json:"failed_checks"`
	TotalChecks    int                    `json:"total_checks"`
	FailedControls []FailedControlSummary `json:"failed_controls"`
}

// FailedControlSummary es la evidencia mínima de un control fuera de tolerancia.
type FailedControlSummary struct {
	ControlNumber int     `json:"control_number"`
	DataToVerify  string  `json:"data_to_verify"`
	Description   string  `json:"description"`
	DifferenceA   string  `json:"difference_a"`
	DifferenceB   *string `json:"difference_b,omitempty"`
	Tolerance     string  `json:"tolerance"`
}

// TentativePricesSummary resume los insumos con precio tentativo del workspace.
type TentativePricesSummary struct {
	Count       int64                        `json:"count"`
	SampleItems []TentativePriceItemResponse `json:"sample_items"`
}

// ToDataIntegritySummaryResponse arma el envelope agregando ambos reportes.
func ToDataIntegritySummaryResponse(
	workspace filters.WorkspaceFilter,
	costsReport *domain.IntegrityReport,
	pricesReport *domain.TentativePricesReport,
) DataIntegritySummaryResponse {
	workspaceFields := workspaceFields(workspace)
	return DataIntegritySummaryResponse{
		Source:     "ponti.data_integrity.summary",
		Workspace:  workspaceFields,
		Filters:    workspaceFields,
		CapturedAt: time.Now().UTC().Format(time.RFC3339),
		Summary: DataIntegritySummary{
			CostsCheck:      toCostsCheckSummary(costsReport),
			TentativePrices: toTentativePricesSummary(pricesReport),
		},
	}
}

func toCostsCheckSummary(report *domain.IntegrityReport) CostsCheckSummary {
	summary := CostsCheckSummary{
		Status:         "OK",
		FailedControls: []FailedControlSummary{},
	}
	if report == nil {
		return summary
	}
	summary.TotalChecks = len(report.Checks)
	for _, check := range report.Checks {
		if check.Status != "ERROR" {
			continue
		}
		summary.FailedChecks++
		control := FailedControlSummary{
			ControlNumber: check.ControlNumber,
			DataToVerify:  check.DataToVerify,
			Description:   check.Description,
			DifferenceA:   formatDecimal(check.DifferenceA),
			Tolerance:     formatDecimal(check.Tolerance),
		}
		if check.DifferenceB != nil {
			v := formatDecimal(*check.DifferenceB)
			control.DifferenceB = &v
		}
		summary.FailedControls = append(summary.FailedControls, control)
	}
	if summary.FailedChecks > 0 {
		summary.Status = "ERROR"
	}
	return summary
}

func toTentativePricesSummary(report *domain.TentativePricesReport) TentativePricesSummary {
	summary := TentativePricesSummary{
		SampleItems: []TentativePriceItemResponse{},
	}
	if report == nil {
		return summary
	}
	summary.Count = report.Count
	for i := range report.Items {
		if len(summary.SampleItems) >= maxSummarySampleItems {
			break
		}
		summary.SampleItems = append(summary.SampleItems, TentativePriceItemResponse{
			SupplyID:     report.Items[i].SupplyID,
			Name:         report.Items[i].Name,
			CategoryName: report.Items[i].CategoryName,
			Price:        report.Items[i].Price.StringFixed(2),
		})
	}
	return summary
}

func workspaceFields(workspace filters.WorkspaceFilter) map[string]int64 {
	out := map[string]int64{}
	set := func(key string, value *int64) {
		if value != nil && *value > 0 {
			out[key] = *value
		}
	}
	set("customer_id", workspace.CustomerID)
	set("project_id", workspace.ProjectID)
	set("campaign_id", workspace.CampaignID)
	set("field_id", workspace.FieldID)
	return out
}
