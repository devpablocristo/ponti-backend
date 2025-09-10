// Package domain holds the business logic for investor contribution reports.
package domain

import (
	"github.com/shopspring/decimal"
)

// ContributionData represents raw data needed for contribution calculations
type ContributionData struct {
	// Supply data
	SupplyID       int64           `json:"supply_id"`
	SupplyName     string          `json:"supply_name"`
	SupplyPrice    decimal.Decimal `json:"supply_price"`
	SupplyCategory string          `json:"supply_category"`
	SupplyType     string          `json:"supply_type"`
	TotalUsed      decimal.Decimal `json:"total_used"`
	TotalCost      decimal.Decimal `json:"total_cost"`

	// Labor data
	LaborID       int64           `json:"labor_id"`
	LaborName     string          `json:"labor_name"`
	LaborPrice    decimal.Decimal `json:"labor_price"`
	LaborCategory string          `json:"labor_category"`
	EffectiveArea decimal.Decimal `json:"effective_area"`
	LaborCost     decimal.Decimal `json:"labor_cost"`

	// Investor data
	InvestorID   int64           `json:"investor_id"`
	InvestorName string          `json:"investor_name"`
	Percentage   decimal.Decimal `json:"percentage"`
}

// HarvestData represents raw data needed for harvest calculations
type HarvestData struct {
	ProjectID       int64           `json:"project_id"`
	CropID          int64           `json:"crop_id"`
	CropName        string          `json:"crop_name"`
	SurfaceArea     decimal.Decimal `json:"surface_area"`
	NetIncome       decimal.Decimal `json:"net_income"`
	DirectCost      decimal.Decimal `json:"direct_cost"`
	LeaseCost       decimal.Decimal `json:"lease_cost"`
	AdminCost       decimal.Decimal `json:"admin_cost"`
	TotalInvested   decimal.Decimal `json:"total_invested"`
	OperatingResult decimal.Decimal `json:"operating_result"`
}

// InvestorContributionLogic handles the business logic for investor contribution reports
type InvestorContributionLogic struct{}

// NewInvestorContributionLogic creates a new instance of the business logic
func NewInvestorContributionLogic() *InvestorContributionLogic {
	return &InvestorContributionLogic{}
}

// BuildContributionCategories builds contribution categories from raw data
func (l *InvestorContributionLogic) BuildContributionCategories(contributionData []ContributionData, surfaceTotalHa decimal.Decimal) []ContributionCategory {
	categories := make(map[ContributionCategoryType]*ContributionCategory)

	// Agrupar datos por categoría
	for _, data := range contributionData {
		categoryType := l.determineCategoryType(data)

		if category, exists := categories[categoryType]; exists {
			// Actualizar totales de la categoría
			category.TotalUsd = category.TotalUsd.Add(data.TotalCost)
		} else {
			// Crear nueva categoría
			categories[categoryType] = &ContributionCategory{
				Type:                      categoryType,
				Label:                     l.getCategoryLabel(categoryType),
				TotalUsd:                  data.TotalCost,
				TotalUsdHa:                decimal.Zero, // Se calculará después
				Investors:                 []InvestorShare{},
				RequiresManualAttribution: l.requiresManualAttribution(categoryType),
				AttributionNote:           l.getAttributionNote(categoryType),
			}
		}

		// Agregar inversor a la categoría
		category := categories[categoryType]
		investorShare := l.findOrCreateInvestorShare(category.Investors, data.InvestorID, data.InvestorName)
		investorShare.AmountUsd = investorShare.AmountUsd.Add(data.TotalCost)
	}

	// Convertir map a slice y calcular totales por hectárea
	result := make([]ContributionCategory, 0, len(categories))
	for _, category := range categories {
		if surfaceTotalHa.GreaterThan(decimal.Zero) {
			category.TotalUsdHa = category.TotalUsd.Div(surfaceTotalHa)
		}

		// Calcular porcentajes de cada inversor
		l.calculateInvestorPercentages(category.Investors, category.TotalUsd)
		result = append(result, *category)
	}

	return result
}

// BuildInvestorComparison builds theoretical vs actual comparison
func (l *InvestorContributionLogic) BuildInvestorComparison(contributions []ContributionCategory, projectInvestors []ProjectInvestorData) []InvestorContributionComparison {
	comparisons := make(map[int64]*InvestorContributionComparison)

	// Inicializar comparaciones con datos acordados
	for _, pi := range projectInvestors {
		comparisons[pi.InvestorID] = &InvestorContributionComparison{
			InvestorID:     &pi.InvestorID,
			InvestorName:   &pi.InvestorName,
			AgreedSharePct: pi.Percentage,
			AgreedUsd:      decimal.Zero, // Se calculará después
			ActualUsd:      decimal.Zero,
			AdjustmentUsd:  decimal.Zero,
		}
	}

	// Calcular aportes reales por inversor
	totalActualUsd := decimal.Zero
	for _, contribution := range contributions {
		for _, investor := range contribution.Investors {
			if investor.InvestorID != nil {
				if comparison, exists := comparisons[*investor.InvestorID]; exists {
					comparison.ActualUsd = comparison.ActualUsd.Add(investor.AmountUsd)
					totalActualUsd = totalActualUsd.Add(investor.AmountUsd)
				}
			}
		}
	}

	// Calcular aportes acordados basados en el total real
	for _, comparison := range comparisons {
		comparison.AgreedUsd = totalActualUsd.Mul(comparison.AgreedSharePct.Div(decimal.NewFromInt(100)))
		comparison.AdjustmentUsd = comparison.ActualUsd.Sub(comparison.AgreedUsd)
	}

	// Convertir map a slice
	result := make([]InvestorContributionComparison, 0, len(comparisons))
	for _, comparison := range comparisons {
		result = append(result, *comparison)
	}

	return result
}

// BuildHarvestSettlement builds harvest settlement data
func (l *InvestorContributionLogic) BuildHarvestSettlement(harvestData []HarvestData, projectInvestors []ProjectInvestorData) HarvestSettlement {
	settlement := HarvestSettlement{
		TotalHarvestUsd:   decimal.Zero,
		TotalHarvestUsdHa: decimal.Zero,
		Investors:         []HarvestInvestorSettlement{},
	}

	// Calcular totales de cosecha
	totalSurface := decimal.Zero
	for _, data := range harvestData {
		settlement.TotalHarvestUsd = settlement.TotalHarvestUsd.Add(data.NetIncome)
		totalSurface = totalSurface.Add(data.SurfaceArea)
	}

	if totalSurface.GreaterThan(decimal.Zero) {
		settlement.TotalHarvestUsdHa = settlement.TotalHarvestUsd.Div(totalSurface)
	}

	// Calcular liquidación por inversor
	for _, pi := range projectInvestors {
		agreedUsd := settlement.TotalHarvestUsd.Mul(pi.Percentage.Div(decimal.NewFromInt(100)))

		investorSettlement := HarvestInvestorSettlement{
			InvestorID:    &pi.InvestorID,
			InvestorName:  &pi.InvestorName,
			PaidUsd:       decimal.Zero, // TODO: Obtener de pagos reales
			AgreedUsd:     agreedUsd,
			AdjustmentUsd: agreedUsd.Sub(decimal.Zero), // TODO: Calcular con pagos reales
		}

		settlement.Investors = append(settlement.Investors, investorSettlement)
	}

	return settlement
}

// ProjectInvestorData represents project investor relationship data
type ProjectInvestorData struct {
	InvestorID   int64           `json:"investor_id"`
	InvestorName string          `json:"investor_name"`
	Percentage   decimal.Decimal `json:"percentage"`
}

// Helper methods

func (l *InvestorContributionLogic) determineCategoryType(data ContributionData) ContributionCategoryType {
	// Determinar categoría basada en el tipo de insumo o labor
	if data.SupplyType != "" {
		switch data.SupplyType {
		case "Semilla":
			return ContributionSeeds
		case "Agroquímicos":
			return ContributionAgrochemicals
		case "Fertilizantes":
			return ContributionAgrochemicals // Los fertilizantes se consideran agroquímicos
		}
	}

	if data.LaborCategory != "" {
		switch data.LaborCategory {
		case "Siembra":
			return ContributionSowing
		case "Riego":
			return ContributionIrrigation
		case "Cosecha":
			return ContributionGeneralLabors // La cosecha se considera labor general
		case "Pulverización", "Otras Labores":
			return ContributionGeneralLabors
		}
	}

	return ContributionGeneralLabors // Default
}

func (l *InvestorContributionLogic) getCategoryLabel(categoryType ContributionCategoryType) string {
	labels := map[ContributionCategoryType]string{
		ContributionAgrochemicals:           "Agroquímicos",
		ContributionSeeds:                   "Semillas",
		ContributionGeneralLabors:           "Labores Generales",
		ContributionSowing:                  "Siembra",
		ContributionIrrigation:              "Riego",
		ContributionCapitalizableLease:      "Arriendo Capitalizable",
		ContributionAdministrationStructure: "Administración",
	}
	return labels[categoryType]
}

func (l *InvestorContributionLogic) requiresManualAttribution(categoryType ContributionCategoryType) bool {
	// Arriendo y administración requieren imputación manual
	return categoryType == ContributionCapitalizableLease ||
		categoryType == ContributionAdministrationStructure
}

func (l *InvestorContributionLogic) getAttributionNote(categoryType ContributionCategoryType) *string {
	if categoryType == ContributionCapitalizableLease {
		note := "Se requiere imputación manual de aportantes"
		return &note
	}
	if categoryType == ContributionAdministrationStructure {
		note := "Administración del proyecto por hectárea"
		return &note
	}
	return nil
}

func (l *InvestorContributionLogic) findOrCreateInvestorShare(investors []InvestorShare, investorID int64, investorName string) *InvestorShare {
	// Buscar inversor existente
	for i := range investors {
		if investors[i].InvestorID != nil && *investors[i].InvestorID == investorID {
			return &investors[i]
		}
	}

	// Crear nuevo inversor
	newInvestor := InvestorShare{
		InvestorID:   &investorID,
		InvestorName: &investorName,
		AmountUsd:    decimal.Zero,
		SharePct:     decimal.Zero,
	}
	investors = append(investors, newInvestor)
	return &investors[len(investors)-1]
}

func (l *InvestorContributionLogic) calculateInvestorPercentages(investors []InvestorShare, totalUsd decimal.Decimal) {
	if totalUsd.Equal(decimal.Zero) {
		return
	}

	for i := range investors {
		if totalUsd.GreaterThan(decimal.Zero) {
			investors[i].SharePct = investors[i].AmountUsd.Div(totalUsd).Mul(decimal.NewFromInt(100))
		}
	}
}
