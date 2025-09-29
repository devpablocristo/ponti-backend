package dto

import (
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/report/usecases/domain"
	"github.com/shopspring/decimal"
)

// -------------------------------
// Entidades básicas de inversores
// -------------------------------

// InvestorRef: referencia mínima de un inversor (id + nombre).
type InvestorRef struct {
	InvestorID   *int64  `json:"investor_id,omitempty"`
	InvestorName *string `json:"investor_name,omitempty"`
}

// InvestorHeader: chapita de cabecera (ej: "Agrolaits 50%").
type InvestorHeader struct {
	InvestorRef
	SharePct decimal.Decimal `json:"share_pct"` // % global acordado
}

// InvestorShare: celda por inversor en una fila.
type InvestorShare struct {
	InvestorRef
	AmountUsd decimal.Decimal `json:"amount_usd"` // Monto USD en la celda
	SharePct  decimal.Decimal `json:"share_pct"`  // % de esa fila
}

// -------------------------------
// Datos generales del proyecto
// -------------------------------
type GeneralProjectData struct {
	SurfaceTotalHa decimal.Decimal `json:"surface_total_ha"` // Hectáreas totales
	LeaseFixedUsd  decimal.Decimal `json:"lease_fixed_usd"`  // Arriendo fijo por ha
	LeaseIsFixed   bool            `json:"lease_is_fixed"`   // true = arriendo fijo
	AdminPerHaUsd  decimal.Decimal `json:"admin_per_ha_usd"` // Administración por ha
	AdminTotalUsd  decimal.Decimal `json:"admin_total_usd"`  // Administración total
}

// -------------------------------
// Aportes pre-cosecha (tabla sup.)
// -------------------------------

// Enum con tipos de categoría
type ContributionCategoryType string

const (
	ContributionAgrochemicals           ContributionCategoryType = "agrochemicals"
	ContributionSeeds                   ContributionCategoryType = "seeds"
	ContributionGeneralLabors           ContributionCategoryType = "general_labors"
	ContributionSowing                  ContributionCategoryType = "sowing"
	ContributionIrrigation              ContributionCategoryType = "irrigation"
	ContributionCapitalizableLease      ContributionCategoryType = "capitalizable_lease"
	ContributionAdministrationStructure ContributionCategoryType = "administration_structure"
)

// ContributionCategory: fila de la tabla de aportes pre-cosecha
type ContributionCategory struct {
	Key                       string                   `json:"key"` // clave estable en inglés (ej: "agrochemicals")
	SortIndex                 int                      `json:"sort_index"`
	Type                      ContributionCategoryType `json:"type"`
	Label                     string                   `json:"label"` // etiqueta visible (ej: "Agroquímicos")
	TotalUsd                  decimal.Decimal          `json:"total_usd"`
	TotalUsdHa                decimal.Decimal          `json:"total_usd_ha"`
	Investors                 []InvestorShare          `json:"investors"`
	RequiresManualAttribution bool                     `json:"requires_manual_attribution"`
	AttributionNote           *string                  `json:"attribution_note,omitempty"`
}

// PreHarvestTotals: fila "Totales" en la tabla de aportes pre-cosecha
type PreHarvestTotals struct {
	TotalUsd   decimal.Decimal `json:"total_usd"`
	TotalUsdHa decimal.Decimal `json:"total_us_ha"`
	Investors  []InvestorShare `json:"investors"`
}

// -------------------------------------------------
// Aporte acordado / Ajuste de aporte (bloque medio)
// -------------------------------------------------
type InvestorContributionComparison struct {
	InvestorRef
	AgreedSharePct decimal.Decimal `json:"agreed_share_pct"`
	AgreedUsd      decimal.Decimal `json:"agreed_usd"`
	ActualUsd      decimal.Decimal `json:"actual_usd"`
	AdjustmentUsd  decimal.Decimal `json:"adjustment_usd"`
}

// -------------------------------
// Pagos de cosecha (tabla inferior)
// -------------------------------

type HarvestRowType string

const (
	HarvestRowHarvest HarvestRowType = "harvest" // fila detalle "Cosecha"
	HarvestRowTotals  HarvestRowType = "totals"  // fila "Totales"
)

// HarvestRow: representa una fila en pagos de cosecha
type HarvestRow struct {
	Key        string          `json:"key"`  // "harvest" o "totals"
	Type       HarvestRowType  `json:"type"` // enum backend
	TotalUsd   decimal.Decimal `json:"total_usd"`
	TotalUsdHa decimal.Decimal `json:"total_us_ha"`
	Investors  []InvestorShare `json:"investors"`
}

// HarvestSettlement: sección completa de pagos de cosecha
type HarvestSettlement struct {
	Rows                    []HarvestRow    `json:"rows"`                      // 2 filas: harvest y totals
	FooterPaymentAgreed     []InvestorShare `json:"footer_payment_agreed"`     // fila "Pago acordado"
	FooterPaymentAdjustment []InvestorShare `json:"footer_payment_adjustment"` // fila "Ajuste de pago"
}

// -------------------------------
// Informe raíz completo
// -------------------------------
type InvestorContributionReport struct {
	ProjectID       int64                            `json:"project_id"`
	ProjectName     string                           `json:"project_name"`
	CustomerID      int64                            `json:"customer_id"`
	CustomerName    string                           `json:"customer_name"`
	CampaignID      int64                            `json:"campaign_id"`
	CampaignName    string                           `json:"campaign_name"`
	InvestorHeaders []InvestorHeader                 `json:"investor_headers"`
	General         GeneralProjectData               `json:"general"`
	Contributions   []ContributionCategory           `json:"contributions"`
	PreHarvest      PreHarvestTotals                 `json:"pre_harvest"`
	Comparison      []InvestorContributionComparison `json:"comparison"`
	Harvest         HarvestSettlement                `json:"harvest"`
}

// ===== MAPPERS =====

// FromDomainInvestorReport mapea del domain al DTO
func FromDomainInvestorReport(domainReport *domain.InvestorContributionReport) *InvestorContributionReport {
	if domainReport == nil {
		return nil
	}

	return &InvestorContributionReport{
		ProjectID:       domainReport.ProjectID,
		ProjectName:     domainReport.ProjectName,
		CustomerID:      domainReport.CustomerID,
		CustomerName:    domainReport.CustomerName,
		CampaignID:      domainReport.CampaignID,
		CampaignName:    domainReport.CampaignName,
		InvestorHeaders: fromDomainInvestorHeaders(domainReport.InvestorHeaders),
		General:         fromDomainGeneralProjectData(domainReport.General),
		Contributions:   fromDomainContributionCategories(domainReport.Contributions),
		PreHarvest:      fromDomainPreHarvestTotals(domainReport.PreHarvest),
		Comparison:      fromDomainInvestorContributionComparisons(domainReport.Comparison),
		Harvest:         fromDomainHarvestSettlement(domainReport.Harvest),
	}
}

// fromDomainInvestorHeaders mapea headers de inversores
func fromDomainInvestorHeaders(domainHeaders []domain.InvestorHeader) []InvestorHeader {
	headers := make([]InvestorHeader, len(domainHeaders))
	for i, h := range domainHeaders {
		headers[i] = InvestorHeader{
			InvestorRef: InvestorRef{
				InvestorID:   h.InvestorID,
				InvestorName: h.InvestorName,
			},
			SharePct: h.SharePct,
		}
	}
	return headers
}

// fromDomainGeneralProjectData mapea datos generales del proyecto
func fromDomainGeneralProjectData(domainGeneral domain.GeneralProjectData) GeneralProjectData {
	return GeneralProjectData{
		SurfaceTotalHa: domainGeneral.SurfaceTotalHa,
		LeaseFixedUsd:  domainGeneral.LeaseFixedUsd,
		LeaseIsFixed:   domainGeneral.LeaseIsFixed,
		AdminPerHaUsd:  domainGeneral.AdminPerHaUsd,
		AdminTotalUsd:  domainGeneral.AdminTotalUsd,
	}
}

// fromDomainContributionCategories mapea categorías de contribución
func fromDomainContributionCategories(domainCategories []domain.ContributionCategory) []ContributionCategory {
	categories := make([]ContributionCategory, len(domainCategories))
	for i, c := range domainCategories {
		categories[i] = ContributionCategory{
			Key:                       c.Key,
			SortIndex:                 c.SortIndex,
			Type:                      ContributionCategoryType(c.Type),
			Label:                     c.Label,
			TotalUsd:                  c.TotalUsd,
			TotalUsdHa:                c.TotalUsdHa,
			Investors:                 fromDomainInvestorShares(c.Investors),
			RequiresManualAttribution: c.RequiresManualAttribution,
			AttributionNote:           c.AttributionNote,
		}
	}
	return categories
}

// fromDomainPreHarvestTotals mapea totales pre-cosecha
func fromDomainPreHarvestTotals(domainTotals domain.PreHarvestTotals) PreHarvestTotals {
	return PreHarvestTotals{
		TotalUsd:   domainTotals.TotalUsd,
		TotalUsdHa: domainTotals.TotalUsdHa,
		Investors:  fromDomainInvestorShares(domainTotals.Investors),
	}
}

// fromDomainInvestorContributionComparisons mapea comparaciones de contribución
func fromDomainInvestorContributionComparisons(domainComparisons []domain.InvestorContributionComparison) []InvestorContributionComparison {
	comparisons := make([]InvestorContributionComparison, len(domainComparisons))
	for i, c := range domainComparisons {
		comparisons[i] = InvestorContributionComparison{
			InvestorRef: InvestorRef{
				InvestorID:   c.InvestorID,
				InvestorName: c.InvestorName,
			},
			AgreedSharePct: c.AgreedSharePct,
			AgreedUsd:      c.AgreedUsd,
			ActualUsd:      c.ActualUsd,
			AdjustmentUsd:  c.AdjustmentUsd,
		}
	}
	return comparisons
}

// fromDomainHarvestSettlement mapea liquidación de cosecha
func fromDomainHarvestSettlement(domainHarvest domain.HarvestSettlement) HarvestSettlement {
	return HarvestSettlement{
		Rows:                    fromDomainHarvestRows(domainHarvest.Rows),
		FooterPaymentAgreed:     fromDomainInvestorShares(domainHarvest.FooterPaymentAgreed),
		FooterPaymentAdjustment: fromDomainInvestorShares(domainHarvest.FooterPaymentAdjustment),
	}
}

// fromDomainHarvestRows mapea filas de cosecha
func fromDomainHarvestRows(domainRows []domain.HarvestRow) []HarvestRow {
	rows := make([]HarvestRow, len(domainRows))
	for i, r := range domainRows {
		rows[i] = HarvestRow{
			Key:        r.Key,
			Type:       HarvestRowType(r.Type),
			TotalUsd:   r.TotalUsd,
			TotalUsdHa: r.TotalUsdHa,
			Investors:  fromDomainInvestorShares(r.Investors),
		}
	}
	return rows
}

// fromDomainInvestorShares mapea shares de inversores
func fromDomainInvestorShares(domainShares []domain.InvestorShare) []InvestorShare {
	shares := make([]InvestorShare, len(domainShares))
	for i, s := range domainShares {
		shares[i] = InvestorShare{
			InvestorRef: InvestorRef{
				InvestorID:   s.InvestorID,
				InvestorName: s.InvestorName,
			},
			AmountUsd: s.AmountUsd,
			SharePct:  s.SharePct,
		}
	}
	return shares
}
