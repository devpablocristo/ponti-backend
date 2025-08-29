package domain

import "github.com/shopspring/decimal"

type DashboardPayload struct {
	Metrics struct {
		Sowing struct {
			ProgressPct   decimal.Decimal `json:"progress_pct"`
			Hectares      decimal.Decimal `json:"hectares"`
			TotalHectares decimal.Decimal `json:"total_hectares"`
		} `json:"sowing"`
		Harvest struct {
			ProgressPct   decimal.Decimal `json:"progress_pct"`
			Hectares      decimal.Decimal `json:"hectares"`
			TotalHectares decimal.Decimal `json:"total_hectares"`
		} `json:"harvest"`
		Costs struct {
			ProgressPct decimal.Decimal `json:"progress_pct"`
			ExecutedUSD decimal.Decimal `json:"executed_usd"`
			BudgetUSD   decimal.Decimal `json:"budget_usd"`
		} `json:"costs"`
		InvestorContributions struct {
			ProgressPct decimal.Decimal `json:"progress_pct"`
			Breakdown   interface{}     `json:"breakdown"` // si definís un tipo, reemplazalo
		} `json:"investor_contributions"`
		OperatingResult struct {
			ProgressPct   decimal.Decimal `json:"progress_pct"`
			IncomeUSD     decimal.Decimal `json:"income_usd"`
			TotalCostsUSD decimal.Decimal `json:"total_costs_usd"`
		} `json:"operating_result"`
	} `json:"metrics"`

	ManagementBalance struct {
		Summary struct {
			IncomeUSD              decimal.Decimal `json:"income_usd"`
			DirectCostsExecutedUSD decimal.Decimal `json:"direct_costs_executed_usd"`
			DirectCostsInvestedUSD decimal.Decimal `json:"direct_costs_invested_usd"`
			StockUSD               decimal.Decimal `json:"stock_usd"`
			RentUSD                decimal.Decimal `json:"rent_usd"`
			StructureUSD           decimal.Decimal `json:"structure_usd"`
			OperatingResultUSD     decimal.Decimal `json:"operating_result_usd"`
			OperatingResultPct     decimal.Decimal `json:"operating_result_pct"`
		} `json:"summary"`
		Breakdown []struct {
			Label       string           `json:"label"`
			ExecutedUSD decimal.Decimal  `json:"executed_usd"`
			InvestedUSD decimal.Decimal  `json:"invested_usd"`
			StockUSD    *decimal.Decimal `json:"stock_usd"`
		} `json:"breakdown"`
		TotalsRow struct {
			ExecutedUSD decimal.Decimal `json:"executed_usd"`
			InvestedUSD decimal.Decimal `json:"invested_usd"`
			StockUSD    decimal.Decimal `json:"stock_usd"`
		} `json:"totals_row"`
	} `json:"management_balance"`

	CropIncidence struct {
		Crops []struct {
			Name         string          `json:"name"`
			Hectares     decimal.Decimal `json:"hectares"`
			RotationPct  decimal.Decimal `json:"rotation_pct"`
			CostUSDPerHa decimal.Decimal `json:"cost_usd_per_ha"`
			IncidencePct decimal.Decimal `json:"incidence_pct"`
		} `json:"crops"`
		Total struct {
			Hectares          decimal.Decimal `json:"hectares"`
			RotationPct       decimal.Decimal `json:"rotation_pct"`
			CostUSDPerHectare decimal.Decimal `json:"cost_usd_per_hectare"`
		} `json:"total"`
	} `json:"crop_incidence"`

	OperationalIndicators struct {
		Cards []struct {
			Key           string      `json:"key"`
			Title         string      `json:"title"`
			Date          *string     `json:"date"` // ISO-8601; si querés time.Time cambialo
			WorkorderID   interface{} `json:"workorder_id"`
			WorkorderCode interface{} `json:"workorder_code"`
			AuditID       interface{} `json:"audit_id"`
			AuditCode     interface{} `json:"audit_code"`
			Status        interface{} `json:"status"`
		} `json:"cards"`
	} `json:"operational_indicators"`
}
