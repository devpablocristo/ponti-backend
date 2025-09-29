package models

import (
	"encoding/json"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/report/usecases/domain"
	"github.com/shopspring/decimal"
)

// ===== MODELOS PARA APORTES DE INVERSORES =====

// InvestorContributionDataModel modelo para los datos de aportes de inversores desde la vista
type InvestorContributionDataModel struct {
	ProjectID         int64           `gorm:"column:project_id"`
	ProjectName       string          `gorm:"column:project_name"`
	CustomerID        int64           `gorm:"column:customer_id"`
	CustomerName      string          `gorm:"column:customer_name"`
	CampaignID        int64           `gorm:"column:campaign_id"`
	CampaignName      string          `gorm:"column:campaign_name"`
	SurfaceTotalHa    decimal.Decimal `gorm:"column:surface_total_ha"`
	LeaseFixedUsd     decimal.Decimal `gorm:"column:lease_fixed_usd"`
	LeaseIsFixed      bool            `gorm:"column:lease_is_fixed"`
	AdminPerHaUsd     decimal.Decimal `gorm:"column:admin_per_ha_usd"`
	AdminTotalUsd     decimal.Decimal `gorm:"column:admin_total_usd"`
	ContributionsData string          `gorm:"column:contributions_data"`
	ComparisonData    string          `gorm:"column:comparison_data"`
	HarvestData       string          `gorm:"column:harvest_data"`
}

// InvestorHeaderModel modelo para headers de inversores
type InvestorHeaderModel struct {
	InvestorID   *int64          `json:"investor_id,omitempty"`
	InvestorName *string         `json:"investor_name,omitempty"`
	SharePct     decimal.Decimal `json:"share_pct"`
}

// InvestorShareModel modelo para shares de inversores
type InvestorShareModel struct {
	InvestorID   *int64          `json:"investor_id,omitempty"`
	InvestorName *string         `json:"investor_name,omitempty"`
	AmountUsd    decimal.Decimal `json:"amount_usd"`
	SharePct     decimal.Decimal `json:"share_pct"`
}

// GeneralProjectDataModel modelo para datos generales del proyecto
type GeneralProjectDataModel struct {
	SurfaceTotalHa decimal.Decimal `json:"surface_total_ha"`
	LeaseFixedUsd  decimal.Decimal `json:"lease_fixed_usd"`
	LeaseIsFixed   bool            `json:"lease_is_fixed"`
	AdminPerHaUsd  decimal.Decimal `json:"admin_per_ha_usd"`
	AdminTotalUsd  decimal.Decimal `json:"admin_total_usd"`
}

// ContributionCategoryModel modelo para categorías de contribución
type ContributionCategoryModel struct {
	Key                       string               `json:"key"`
	SortIndex                 int                  `json:"sort_index"`
	Type                      string               `json:"type"`
	Label                     string               `json:"label"`
	TotalUsd                  decimal.Decimal      `json:"total_usd"`
	TotalUsdHa                decimal.Decimal      `json:"total_usd_ha"`
	Investors                 []InvestorShareModel `json:"investors"`
	RequiresManualAttribution bool                 `json:"requires_manual_attribution"`
	AttributionNote           *string              `json:"attribution_note,omitempty"`
}

// PreHarvestTotalsModel modelo para totales pre-cosecha
type PreHarvestTotalsModel struct {
	TotalUsd   decimal.Decimal      `json:"total_usd"`
	TotalUsdHa decimal.Decimal      `json:"total_us_ha"`
	Investors  []InvestorShareModel `json:"investors"`
}

// InvestorContributionComparisonModel modelo para comparaciones de contribución
type InvestorContributionComparisonModel struct {
	InvestorID     *int64          `json:"investor_id,omitempty"`
	InvestorName   *string         `json:"investor_name,omitempty"`
	AgreedSharePct decimal.Decimal `json:"agreed_share_pct"`
	AgreedUsd      decimal.Decimal `json:"agreed_usd"`
	ActualUsd      decimal.Decimal `json:"actual_usd"`
	AdjustmentUsd  decimal.Decimal `json:"adjustment_usd"`
}

// HarvestRowModel modelo para filas de cosecha
type HarvestRowModel struct {
	Key        string               `json:"key"`
	Type       string               `json:"type"`
	TotalUsd   decimal.Decimal      `json:"total_usd"`
	TotalUsdHa decimal.Decimal      `json:"total_us_ha"`
	Investors  []InvestorShareModel `json:"investors"`
}

// HarvestSettlementModel modelo para liquidación de cosecha
type HarvestSettlementModel struct {
	Rows                    []HarvestRowModel    `json:"rows"`
	FooterPaymentAgreed     []InvestorShareModel `json:"footer_payment_agreed"`
	FooterPaymentAdjustment []InvestorShareModel `json:"footer_payment_adjustment"`
}

// InvestorContributionReportModel modelo completo del reporte
type InvestorContributionReportModel struct {
	ProjectID       int64                                 `json:"project_id"`
	ProjectName     string                                `json:"project_name"`
	CustomerID      int64                                 `json:"customer_id"`
	CustomerName    string                                `json:"customer_name"`
	CampaignID      int64                                 `json:"campaign_id"`
	CampaignName    string                                `json:"campaign_name"`
	InvestorHeaders []InvestorHeaderModel                 `json:"investor_headers"`
	General         GeneralProjectDataModel               `json:"general"`
	Contributions   []ContributionCategoryModel           `json:"contributions"`
	PreHarvest      PreHarvestTotalsModel                 `json:"pre_harvest"`
	Comparison      []InvestorContributionComparisonModel `json:"comparison"`
	Harvest         HarvestSettlementModel                `json:"harvest"`
}

// ===== MAPPERS =====

// ToDomainInvestorContributionReport convierte el modelo al domain
func (m *InvestorContributionDataModel) ToDomainInvestorContributionReport() (*domain.InvestorContributionReport, error) {
	report := &domain.InvestorContributionReport{
		ProjectID:    m.ProjectID,
		ProjectName:  m.ProjectName,
		CustomerID:   m.CustomerID,
		CustomerName: m.CustomerName,
		CampaignID:   m.CampaignID,
		CampaignName: m.CampaignName,
		General: domain.GeneralProjectData{
			SurfaceTotalHa: m.SurfaceTotalHa,
			LeaseFixedUsd:  m.LeaseFixedUsd,
			LeaseIsFixed:   m.LeaseIsFixed,
			AdminPerHaUsd:  m.AdminPerHaUsd,
			AdminTotalUsd:  m.AdminTotalUsd,
		},
	}

	// Parsear contributions si existe
	if m.ContributionsData != "" {
		var contributions []ContributionCategoryModel
		if err := json.Unmarshal([]byte(m.ContributionsData), &contributions); err != nil {
			return nil, err
		}
		report.Contributions = m.mapContributionsToDomain(contributions)
	}

	// Parsear comparison si existe
	if m.ComparisonData != "" {
		var comparisons []InvestorContributionComparisonModel
		if err := json.Unmarshal([]byte(m.ComparisonData), &comparisons); err != nil {
			return nil, err
		}
		report.Comparison = m.mapComparisonsToDomain(comparisons)
	}

	// Parsear harvest si existe
	if m.HarvestData != "" {
		var harvest HarvestSettlementModel
		if err := json.Unmarshal([]byte(m.HarvestData), &harvest); err != nil {
			return nil, err
		}
		report.Harvest = m.mapHarvestToDomain(harvest)
	}

	return report, nil
}

// mapContributionsToDomain mapea contribuciones del modelo al domain
func (m *InvestorContributionDataModel) mapContributionsToDomain(contributions []ContributionCategoryModel) []domain.ContributionCategory {
	domainContributions := make([]domain.ContributionCategory, len(contributions))
	for i, c := range contributions {
		domainContributions[i] = domain.ContributionCategory{
			Key:                       c.Key,
			SortIndex:                 c.SortIndex,
			Type:                      domain.ContributionCategoryType(c.Type),
			Label:                     c.Label,
			TotalUsd:                  c.TotalUsd,
			TotalUsdHa:                c.TotalUsdHa,
			Investors:                 m.mapInvestorSharesToDomain(c.Investors),
			RequiresManualAttribution: c.RequiresManualAttribution,
			AttributionNote:           c.AttributionNote,
		}
	}
	return domainContributions
}

// mapComparisonsToDomain mapea comparaciones del modelo al domain
func (m *InvestorContributionDataModel) mapComparisonsToDomain(comparisons []InvestorContributionComparisonModel) []domain.InvestorContributionComparison {
	domainComparisons := make([]domain.InvestorContributionComparison, len(comparisons))
	for i, c := range comparisons {
		domainComparisons[i] = domain.InvestorContributionComparison{
			InvestorRef: domain.InvestorRef{
				InvestorID:   c.InvestorID,
				InvestorName: c.InvestorName,
			},
			AgreedSharePct: c.AgreedSharePct,
			AgreedUsd:      c.AgreedUsd,
			ActualUsd:      c.ActualUsd,
			AdjustmentUsd:  c.AdjustmentUsd,
		}
	}
	return domainComparisons
}

// mapHarvestToDomain mapea harvest del modelo al domain
func (m *InvestorContributionDataModel) mapHarvestToDomain(harvest HarvestSettlementModel) domain.HarvestSettlement {
	domainRows := make([]domain.HarvestRow, len(harvest.Rows))
	for i, r := range harvest.Rows {
		domainRows[i] = domain.HarvestRow{
			Key:        r.Key,
			Type:       domain.HarvestRowType(r.Type),
			TotalUsd:   r.TotalUsd,
			TotalUsdHa: r.TotalUsdHa,
			Investors:  m.mapInvestorSharesToDomain(r.Investors),
		}
	}

	return domain.HarvestSettlement{
		Rows:                    domainRows,
		FooterPaymentAgreed:     m.mapInvestorSharesToDomain(harvest.FooterPaymentAgreed),
		FooterPaymentAdjustment: m.mapInvestorSharesToDomain(harvest.FooterPaymentAdjustment),
	}
}

// mapInvestorSharesToDomain mapea shares de inversores del modelo al domain
func (m *InvestorContributionDataModel) mapInvestorSharesToDomain(shares []InvestorShareModel) []domain.InvestorShare {
	domainShares := make([]domain.InvestorShare, len(shares))
	for i, s := range shares {
		domainShares[i] = domain.InvestorShare{
			InvestorRef: domain.InvestorRef{
				InvestorID:   s.InvestorID,
				InvestorName: s.InvestorName,
			},
			AmountUsd: s.AmountUsd,
			SharePct:  s.SharePct,
		}
	}
	return domainShares
}
