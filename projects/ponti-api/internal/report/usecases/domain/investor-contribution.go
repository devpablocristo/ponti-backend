// Package domain holds the domain models for investor contribution reports.
package domain

import (
	"github.com/shopspring/decimal"
)

// ContributionCategoryType enumerates contribution categories.
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

// GeneralProjectData represents base project data for the report.
type GeneralProjectData struct {
	SurfaceTotalHa decimal.Decimal `json:"surface_total_ha"`
	LeaseFixedUsd  decimal.Decimal `json:"lease_fixed_usd"`
	LeaseIsFixed   bool            `json:"lease_is_fixed"`
	LeaseNote      *string         `json:"lease_note,omitempty"`
	AdminPerHaUsd  decimal.Decimal `json:"admin_per_ha_usd"`
	AdminTotalUsd  decimal.Decimal `json:"admin_total_usd"`
}

// InvestorShare represents the contribution of a single investor in a category.
type InvestorShare struct {
	InvestorID   *int64          `json:"investor_id,omitempty"`
	InvestorName *string         `json:"investor_name,omitempty"`
	AmountUsd    decimal.Decimal `json:"amount_usd"`
	SharePct     decimal.Decimal `json:"share_pct"`
}

// ContributionCategory represents a category of contributions.
type ContributionCategory struct {
	Type                      ContributionCategoryType `json:"type"`
	Label                     string                   `json:"label"`
	TotalUsd                  decimal.Decimal          `json:"total_usd"`
	TotalUsdHa                decimal.Decimal          `json:"total_usd_ha"`
	Investors                 []InvestorShare          `json:"investors"`
	RequiresManualAttribution bool                     `json:"requires_manual_attribution"`
	AttributionNote           *string                  `json:"attribution_note,omitempty"`
}

// InvestorContributionComparison compares agreed vs actual per investor.
type InvestorContributionComparison struct {
	InvestorID     *int64          `json:"investor_id,omitempty"`
	InvestorName   *string         `json:"investor_name,omitempty"`
	AgreedSharePct decimal.Decimal `json:"agreed_share_pct"`
	AgreedUsd      decimal.Decimal `json:"agreed_usd"`
	ActualUsd      decimal.Decimal `json:"actual_usd"`
	AdjustmentUsd  decimal.Decimal `json:"adjustment_usd"`
}

// HarvestInvestorSettlement represents harvest settlement for a single investor.
type HarvestInvestorSettlement struct {
	InvestorID    *int64          `json:"investor_id,omitempty"`
	InvestorName  *string         `json:"investor_name,omitempty"`
	PaidUsd       decimal.Decimal `json:"paid_usd"`
	AgreedUsd     decimal.Decimal `json:"agreed_usd"`
	AdjustmentUsd decimal.Decimal `json:"adjustment_usd"`
}

// HarvestSettlement represents harvest totals and per-investor settlements.
type HarvestSettlement struct {
	TotalHarvestUsd   decimal.Decimal             `json:"total_harvest_usd"`
	TotalHarvestUsdHa decimal.Decimal             `json:"total_harvest_usd_ha"`
	Investors         []HarvestInvestorSettlement `json:"investors"`
}

// InvestorContributionReport represents the complete investor contribution report.
type InvestorContributionReport struct {
	ProjectID     int64                            `json:"project_id"`
	ProjectName   string                           `json:"project_name"`
	CustomerID    int64                            `json:"customer_id"`
	CustomerName  string                           `json:"customer_name"`
	CampaignID    int64                            `json:"campaign_id"`
	CampaignName  string                           `json:"campaign_name"`
	General       GeneralProjectData               `json:"general"`
	Contributions []ContributionCategory           `json:"contributions"`
	Comparison    []InvestorContributionComparison `json:"comparison"`
	Harvest       HarvestSettlement                `json:"harvest"`
}
