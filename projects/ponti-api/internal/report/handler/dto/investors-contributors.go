// Package dto holds the Data Transfer Objects for the Investor Contribution Report.
package dto

import (
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/report/usecases/domain"
	"github.com/shopspring/decimal"
)

/* =========================
   ENUMS
========================= */

// ContributionCategoryType enumerates contribution categories.
type ContributionCategoryType string

const (
	ContributionAgrochemicals           ContributionCategoryType = "agrochemicals"            // Insumos: Agroquímicos
	ContributionSeeds                   ContributionCategoryType = "seeds"                    // Insumos: Semillas
	ContributionGeneralLabors           ContributionCategoryType = "general_labors"           // Labores: Pulverización, Otras (no siembra/riego/cosecha)
	ContributionSowing                  ContributionCategoryType = "sowing"                   // Labor: Siembra
	ContributionIrrigation              ContributionCategoryType = "irrigation"               // Labor: Riego
	ContributionCapitalizableLease      ContributionCategoryType = "capitalizable_lease"      // Arriendo: solo la parte fija
	ContributionAdministrationStructure ContributionCategoryType = "administration_structure" // Admin/Estructura: fijo * ha
)

/* =========================
   RESPONSES — Sección 1: Datos Generales
========================= */

// GeneralProjectDataResponse summarizes base inputs for the report.
type GeneralProjectDataResponse struct {
	// Superficie total del proyecto (suma de lotes/campos)
	SurfaceTotalHa decimal.Decimal `json:"surface_total_ha"`

	// Arriendo considerado (solo la parte fija). Si no hay fijo, será 0.
	LeaseFixedUsd decimal.Decimal `json:"lease_fixed_usd"`
	LeaseIsFixed  bool            `json:"lease_is_fixed"`       // true si hay componente fijo considerado
	LeaseNote     *string         `json:"lease_note,omitempty"` // p.ej. "Se toma solo componente fijo"

	// Administración del proyecto por ha (valor base) y total calculado (valor * ha)
	AdminPerHaUsd decimal.Decimal `json:"admin_per_ha_usd"`
	AdminTotalUsd decimal.Decimal `json:"admin_total_usd"`
}

/* =========================
   RESPONSES — Sección 2: Aportes por Inversor
========================= */

// InvestorShareResponse shows the contribution of a single investor in a category.
type InvestorShareResponse struct {
	InvestorID   *int64          `json:"investor_id,omitempty"`
	InvestorName *string         `json:"investor_name,omitempty"`
	AmountUsd    decimal.Decimal `json:"amount_usd"`
	SharePct     decimal.Decimal `json:"share_pct"` // % respecto al total de la categoría
}

// ContributionCategoryResponse aggregates totals and investor breakdown for a category.
type ContributionCategoryResponse struct {
	Type  ContributionCategoryType `json:"type"`
	Label string                   `json:"label"` // e.g., "Agroquímicos", "Semillas", etc.

	// Totales de la categoría
	TotalUsd   decimal.Decimal `json:"total_usd"`
	TotalUsdHa decimal.Decimal `json:"total_usd_ha"` // Total / SurfaceTotalHa

	// Desglose por inversor
	Investors []InvestorShareResponse `json:"investors"`

	// Para categorías que requieren imputación manual de aportantes (arriendo/admin)
	RequiresManualAttribution bool    `json:"requires_manual_attribution"`
	AttributionNote           *string `json:"attribution_note,omitempty"`
}

/* =========================
   RESPONSES — Sección 3: Comparación Teórico vs Real
========================= */

// InvestorContributionComparisonResponse compares agreed vs actual per investor.
type InvestorContributionComparisonResponse struct {
	InvestorID     *int64          `json:"investor_id,omitempty"`
	InvestorName   *string         `json:"investor_name,omitempty"`
	AgreedSharePct decimal.Decimal `json:"agreed_share_pct"` // % pactado en Clientes y Sociedades
	AgreedUsd      decimal.Decimal `json:"agreed_usd"`       // TOTAL_APORTES * %_ACORDADO
	ActualUsd      decimal.Decimal `json:"actual_usd"`       // Suma de aportes reales (todas las categorías de aporte)
	AdjustmentUsd  decimal.Decimal `json:"adjustment_usd"`   // Actual - Acordado
}

/* =========================
   RESPONSES — Sección 4: Liquidación de Cosecha
========================= */

// HarvestInvestorSettlementResponse shows paid vs agreed and adjustment for a single investor.
type HarvestInvestorSettlementResponse struct {
	InvestorID    *int64          `json:"investor_id,omitempty"`
	InvestorName  *string         `json:"investor_name,omitempty"`
	PaidUsd       decimal.Decimal `json:"paid_usd"`       // lo que efectivamente pagó
	AgreedUsd     decimal.Decimal `json:"agreed_usd"`     // total cosecha * % acordado
	AdjustmentUsd decimal.Decimal `json:"adjustment_usd"` // (Agreed - Paid) o (Cosecha*% - Pagado)
}

// HarvestSettlementResponse aggregates harvest totals and per-investor settlements.
type HarvestSettlementResponse struct {
	TotalHarvestUsd   decimal.Decimal                     `json:"total_harvest_usd"`    // Suma "Total USD Neto" de cosecha
	TotalHarvestUsdHa decimal.Decimal                     `json:"total_harvest_usd_ha"` // Total / SurfaceTotalHa
	Investors         []HarvestInvestorSettlementResponse `json:"investors"`
}

/* =========================
   RESPUESTA GLOBAL DEL REPORTE
========================= */

type InvestorContributionReportResponse struct {
	// Metadatos de proyecto
	ProjectID    int64  `json:"project_id"`
	ProjectName  string `json:"project_name"`
	CustomerID   int64  `json:"customer_id"`
	CustomerName string `json:"customer_name"`
	CampaignID   int64  `json:"campaign_id"`
	CampaignName string `json:"campaign_name"`

	// 1) Datos Generales
	General GeneralProjectDataResponse `json:"general"`

	// 2) Aportes por Inversor (lista de categorías)
	Contributions []ContributionCategoryResponse `json:"contributions"`

	// 3) Comparación Aporte Teórico vs Real
	Comparison []InvestorContributionComparisonResponse `json:"comparison"`

	// 4) Liquidación de Cosecha
	Harvest HarvestSettlementResponse `json:"harvest"`
}

/* =========================
   MAPPERS (dominio → DTO)
   Nota: define en dominio estructuras alineadas a estos mappers.
========================= */

// FromDomainInvestorReport maps the domain model to the DTO response.
// Espera que el dominio ya traiga: SurfaceTotalHa, Lease/Admin, categorías con totales y %,
// comparación acordado vs real, y liquidación de cosecha con pagos y ajustes.
func FromDomainInvestorReport(d *domain.InvestorContributionReport) *InvestorContributionReportResponse {
	if d == nil {
		return nil
	}

	// General
	gen := GeneralProjectDataResponse{
		SurfaceTotalHa: d.General.SurfaceTotalHa,
		LeaseFixedUsd:  d.General.LeaseFixedUsd,
		LeaseIsFixed:   d.General.LeaseIsFixed,
		LeaseNote:      d.General.LeaseNote,
		AdminPerHaUsd:  d.General.AdminPerHaUsd,
		AdminTotalUsd:  d.General.AdminTotalUsd,
	}

	// Contributions
	contribs := make([]ContributionCategoryResponse, 0, len(d.Contributions))
	for _, c := range d.Contributions {
		inv := make([]InvestorShareResponse, 0, len(c.Investors))
		for _, is := range c.Investors {
			inv = append(inv, InvestorShareResponse{
				InvestorID:   is.InvestorID,
				InvestorName: is.InvestorName,
				AmountUsd:    is.AmountUsd,
				SharePct:     is.SharePct,
			})
		}
		contribs = append(contribs, ContributionCategoryResponse{
			Type:                      ContributionCategoryType(c.Type),
			Label:                     c.Label,
			TotalUsd:                  c.TotalUsd,
			TotalUsdHa:                c.TotalUsdHa,
			Investors:                 inv,
			RequiresManualAttribution: c.RequiresManualAttribution,
			AttributionNote:           c.AttributionNote,
		})
	}

	// Comparison
	comp := make([]InvestorContributionComparisonResponse, 0, len(d.Comparison))
	for _, it := range d.Comparison {
		comp = append(comp, InvestorContributionComparisonResponse{
			InvestorID:     it.InvestorID,
			InvestorName:   it.InvestorName,
			AgreedSharePct: it.AgreedSharePct,
			AgreedUsd:      it.AgreedUsd,
			ActualUsd:      it.ActualUsd,
			AdjustmentUsd:  it.AdjustmentUsd,
		})
	}

	// Harvest
	harvestInvestors := make([]HarvestInvestorSettlementResponse, 0, len(d.Harvest.Investors))
	for _, hi := range d.Harvest.Investors {
		harvestInvestors = append(harvestInvestors, HarvestInvestorSettlementResponse{
			InvestorID:    hi.InvestorID,
			InvestorName:  hi.InvestorName,
			PaidUsd:       hi.PaidUsd,
			AgreedUsd:     hi.AgreedUsd,
			AdjustmentUsd: hi.AdjustmentUsd,
		})
	}
	har := HarvestSettlementResponse{
		TotalHarvestUsd:   d.Harvest.TotalHarvestUsd,
		TotalHarvestUsdHa: d.Harvest.TotalHarvestUsdHa,
		Investors:         harvestInvestors,
	}

	return &InvestorContributionReportResponse{
		ProjectID:     d.ProjectID,
		ProjectName:   d.ProjectName,
		CustomerID:    d.CustomerID,
		CustomerName:  d.CustomerName,
		CampaignID:    d.CampaignID,
		CampaignName:  d.CampaignName,
		General:       gen,
		Contributions: contribs,
		Comparison:    comp,
		Harvest:       har,
	}
}
