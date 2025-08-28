package dto

import (
	"encoding/json"
	"time"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dashboard/usecases/domain"
	"github.com/shopspring/decimal"
)

// toPtr converts decimal.NullDecimal to *decimal.Decimal
func toPtr(nd decimal.NullDecimal) *decimal.Decimal {
	if !nd.Valid {
		return nil
	}
	d := nd.Decimal
	return &d
}

// toDecimal safely converts interface{} to decimal.Decimal
func toDecimal(value interface{}) decimal.Decimal {
	switch v := value.(type) {
	case float64:
		return decimal.NewFromFloat(v)
	case float32:
		return decimal.NewFromFloat32(v)
	case int:
		return decimal.NewFromInt(int64(v))
	case int64:
		return decimal.NewFromInt(v)
	case string:
		if d, err := decimal.NewFromString(v); err == nil {
			return d
		}
	case decimal.Decimal:
		return v
	}
	return decimal.Zero
}

// CropDetail represents individual crop information for cost incidence
type CropDetail struct {
	Name        string          `json:"name"`
	Hectares    decimal.Decimal `json:"hectares"`
	TotalCost   decimal.Decimal `json:"total_cost"`
	CostPerHa   decimal.Decimal `json:"cost_per_ha"`
	RotationPct decimal.Decimal `json:"rotation_pct"`
}

// DashboardRow represents the dashboard response organized in logical metric groups
type DashboardRow struct {
	// ===== Grupo 1: MÉTRICAS AGRUPADAS POR FUNCIÓN =====
	Metrics struct {
		// AVANCE DE SIEMBRA - Agrupado
		Sowing struct {
			ProgressPct   *decimal.Decimal `json:"progress_pct,omitempty"`
			Hectares      decimal.Decimal  `json:"hectares"`
			TotalHectares decimal.Decimal  `json:"total_hectares"`
		} `json:"sowing"`

		// AVANCE DE COSECHA - Agrupado
		Harvest struct {
			ProgressPct   *decimal.Decimal `json:"progress_pct,omitempty"`
			Hectares      decimal.Decimal  `json:"hectares"`
			TotalHectares decimal.Decimal  `json:"total_hectares"`
		} `json:"harvest"`

		// AVANCE DE COSTOS - Agrupado
		Costs struct {
			ProgressPct *decimal.Decimal `json:"progress_pct,omitempty"`
			Executed    decimal.Decimal  `json:"executed"`
			Budget      *decimal.Decimal `json:"budget,omitempty"`
		} `json:"costs"`

		// APORTES DE INVERSORES - Agrupado (solo detalles, sin porcentaje)
		InvestorContributions struct {
			Details *string `json:"details,omitempty"`
		} `json:"investor_contributions"`

		// RESULTADO OPERATIVO - Agrupado
		OperatingResult struct {
			ProgressPct *decimal.Decimal `json:"progress_pct,omitempty"`
			IncomeNet   decimal.Decimal  `json:"income_net"`
			TotalCosts  decimal.Decimal  `json:"total_costs"`
		} `json:"operating_result"`
	} `json:"metrics"`

	// ===== Grupo 2: INCIDENCIA POR CULTIVO =====
	CropIncidence struct {
		// CULTIVOS INDIVIDUALES
		Crops []CropDetail `json:"crops"`

		// TOTALES
		Total struct {
			Hectares       decimal.Decimal `json:"hectares"`
			RotationPct    decimal.Decimal `json:"rotation_pct"`
			CostPerHectare decimal.Decimal `json:"cost_per_hectare"`
		} `json:"total"`
	} `json:"crop_incidence"`

	// ===== Grupo 3: RENDIMIENTO Y COSTOS =====
	Performance struct {
		YieldPerHectare     *decimal.Decimal `json:"yield_per_hectare,omitempty"`
		TotalCostPerHectare *decimal.Decimal `json:"total_cost_per_hectare,omitempty"`
	} `json:"performance"`

	// ===== Grupo 4: INDICADORES OPERATIVOS =====
	OperationalIndicators struct {
		FirstOrderDate     *time.Time `json:"first_order_date,omitempty"`
		FirstOrderNumber   *string    `json:"first_order_number,omitempty"`
		LastOrderDate      *time.Time `json:"last_order_date,omitempty"`
		LastOrderNumber    *string    `json:"last_order_number,omitempty"`
		LastStockCountDate *time.Time `json:"last_stock_count_date,omitempty"`
	} `json:"operational_indicators"`

	// ===== Grupo 5: BALANCE DE GESTIÓN =====
	ManagementBalance struct {
		IncomeUSD          decimal.Decimal  `json:"income_usd"`
		TotalCostsUSD      decimal.Decimal  `json:"total_costs_usd"`
		OperatingResultUSD decimal.Decimal  `json:"operating_result_usd"`
		OperatingResultPct *decimal.Decimal `json:"operating_result_pct,omitempty"`
		InvestedCostUSD    *decimal.Decimal `json:"invested_cost_usd,omitempty"`
		StockUSD           *decimal.Decimal `json:"stock_usd,omitempty"`
	} `json:"management_balance"`

	// ===== Grupo 6: BALANCE DE GESTIÓN DETALLADO =====
	DetailedManagementBalance struct {
		Balance struct {
			Rows []struct {
				Label    string `json:"label"`
				Executed struct {
					USD decimal.Decimal `json:"usd"`
					Has decimal.Decimal `json:"has"`
				} `json:"executed"`
				Invested struct {
					USD decimal.Decimal `json:"usd"`
					Has decimal.Decimal `json:"has"`
				} `json:"invested"`
				Stock struct {
					USD *decimal.Decimal `json:"usd"`
					Has *decimal.Decimal `json:"has"`
				} `json:"stock"`
			} `json:"rows"`
		} `json:"balance"`
	} `json:"detailed_management_balance"`
}

// FromDomain converts domain.DashboardRow to dto.DashboardRow
func FromDomain(d domain.DashboardRow) DashboardRow {
	// Parse crops breakdown if available
	var crops []CropDetail
	if d.CropsBreakdown != nil {
		var cropsMap map[string]interface{}
		if err := json.Unmarshal([]byte(*d.CropsBreakdown), &cropsMap); err == nil {
			for name, details := range cropsMap {
				if cropDetails, ok := details.(map[string]interface{}); ok {
					crop := CropDetail{
						Name:        name,
						Hectares:    toDecimal(cropDetails["hectares"]),
						TotalCost:   toDecimal(cropDetails["total_cost"]),
						CostPerHa:   toDecimal(cropDetails["cost_per_ha"]),
						RotationPct: toDecimal(cropDetails["rotation_pct"]),
					}
					crops = append(crops, crop)
				}
			}
		}
	}

	return DashboardRow{
		// ===== Grupo 1: MÉTRICAS AGRUPADAS POR FUNCIÓN =====
		Metrics: struct {
			Sowing struct {
				ProgressPct   *decimal.Decimal `json:"progress_pct,omitempty"`
				Hectares      decimal.Decimal  `json:"hectares"`
				TotalHectares decimal.Decimal  `json:"total_hectares"`
			} `json:"sowing"`

			Harvest struct {
				ProgressPct   *decimal.Decimal `json:"progress_pct,omitempty"`
				Hectares      decimal.Decimal  `json:"hectares"`
				TotalHectares decimal.Decimal  `json:"total_hectares"`
			} `json:"harvest"`

			Costs struct {
				ProgressPct *decimal.Decimal `json:"progress_pct,omitempty"`
				Executed    decimal.Decimal  `json:"executed"`
				Budget      *decimal.Decimal `json:"budget,omitempty"`
			} `json:"costs"`

			InvestorContributions struct {
				Details *string `json:"details,omitempty"`
			} `json:"investor_contributions"`

			OperatingResult struct {
				ProgressPct *decimal.Decimal `json:"progress_pct,omitempty"`
				IncomeNet   decimal.Decimal  `json:"income_net"`
				TotalCosts  decimal.Decimal  `json:"total_costs"`
			} `json:"operating_result"`
		}{
			// AVANCE DE SIEMBRA - Agrupado
			Sowing: struct {
				ProgressPct   *decimal.Decimal `json:"progress_pct,omitempty"`
				Hectares      decimal.Decimal  `json:"hectares"`
				TotalHectares decimal.Decimal  `json:"total_hectares"`
			}{
				ProgressPct:   toPtr(d.SowingProgressPct),
				Hectares:      d.SowedArea,
				TotalHectares: d.TotalHectares,
			},

			// AVANCE DE COSECHA - Agrupado
			Harvest: struct {
				ProgressPct   *decimal.Decimal `json:"progress_pct,omitempty"`
				Hectares      decimal.Decimal  `json:"hectares"`
				TotalHectares decimal.Decimal  `json:"total_hectares"`
			}{
				ProgressPct:   toPtr(d.HarvestProgressPct),
				Hectares:      d.HarvestedArea,
				TotalHectares: d.TotalHectares,
			},

			// AVANCE DE COSTOS - Agrupado
			Costs: struct {
				ProgressPct *decimal.Decimal `json:"progress_pct,omitempty"`
				Executed    decimal.Decimal  `json:"executed"`
				Budget      *decimal.Decimal `json:"budget,omitempty"`
			}{
				ProgressPct: toPtr(d.CostsProgressPct),
				Executed:    d.ExecutedCosts,
				Budget:      toPtr(d.BudgetCosts),
			},

			// APORTES DE INVERSORES - Agrupado
			InvestorContributions: struct {
				Details *string `json:"details,omitempty"`
			}{
				Details: d.ContributionDetails,
			},

			// RESULTADO OPERATIVO - Agrupado
			OperatingResult: struct {
				ProgressPct *decimal.Decimal `json:"progress_pct,omitempty"`
				IncomeNet   decimal.Decimal  `json:"income_net"`
				TotalCosts  decimal.Decimal  `json:"total_costs"`
			}{
				ProgressPct: toPtr(d.OperatingResultPct),
				IncomeNet:   d.IncomeNet,
				TotalCosts:  d.TotalCosts,
			},
		},

		// ===== Grupo 2: INCIDENCIA POR CULTIVO =====
		CropIncidence: struct {
			Crops []CropDetail `json:"crops"`

			Total struct {
				Hectares       decimal.Decimal `json:"hectares"`
				RotationPct    decimal.Decimal `json:"rotation_pct"`
				CostPerHectare decimal.Decimal `json:"cost_per_hectare"`
			} `json:"total"`
		}{
			Crops: crops,
			Total: struct {
				Hectares       decimal.Decimal `json:"hectares"`
				RotationPct    decimal.Decimal `json:"rotation_pct"`
				CostPerHectare decimal.Decimal `json:"cost_per_hectare"`
			}{
				Hectares:       d.CropsTotalHectares,
				RotationPct:    d.CropsTotalRotationPct,
				CostPerHectare: d.CropsTotalCostPerHectare,
			},
		},

		// ===== Grupo 3: RENDIMIENTO Y COSTOS =====
		Performance: struct {
			YieldPerHectare     *decimal.Decimal `json:"yield_per_hectare,omitempty"`
			TotalCostPerHectare *decimal.Decimal `json:"total_cost_per_hectare,omitempty"`
		}{
			YieldPerHectare:     toPtr(d.YieldPerHectare),
			TotalCostPerHectare: toPtr(d.TotalCostPerHectare),
		},

		// ===== Grupo 4: INDICADORES OPERATIVOS =====
		OperationalIndicators: struct {
			FirstOrderDate     *time.Time `json:"first_order_date,omitempty"`
			FirstOrderNumber   *string    `json:"first_order_number,omitempty"`
			LastOrderDate      *time.Time `json:"last_order_date,omitempty"`
			LastOrderNumber    *string    `json:"last_order_number,omitempty"`
			LastStockCountDate *time.Time `json:"last_stock_count_date,omitempty"`
		}{
			FirstOrderDate:     d.FirstOrderDate,
			FirstOrderNumber:   d.FirstOrderNumber,
			LastOrderDate:      d.LastOrderDate,
			LastOrderNumber:    d.LastOrderNumber,
			LastStockCountDate: d.LastStockCountDate,
		},

		// ===== Grupo 5: BALANCE DE GESTIÓN =====
		ManagementBalance: struct {
			IncomeUSD          decimal.Decimal  `json:"income_usd"`
			TotalCostsUSD      decimal.Decimal  `json:"total_costs_usd"`
			OperatingResultUSD decimal.Decimal  `json:"operating_result_usd"`
			OperatingResultPct *decimal.Decimal `json:"operating_result_pct,omitempty"`
			InvestedCostUSD    *decimal.Decimal `json:"invested_cost_usd,omitempty"`
			StockUSD           *decimal.Decimal `json:"stock_usd,omitempty"`
		}{
			IncomeUSD:          d.MgmtIncomeUSD,
			TotalCostsUSD:      d.MgmtTotalCostsUSD,
			OperatingResultUSD: d.MgmtOperatingResultUSD,
			OperatingResultPct: toPtr(d.MgmtOperatingResultPct),
			InvestedCostUSD:    toPtr(d.InvestedCostUSD),
			StockUSD:           toPtr(d.StockUSD),
		},

		// ===== Grupo 6: BALANCE DE GESTIÓN DETALLADO =====
		DetailedManagementBalance: struct {
			Balance struct {
				Rows []struct {
					Label    string `json:"label"`
					Executed struct {
						USD decimal.Decimal `json:"usd"`
						Has decimal.Decimal `json:"has"`
					} `json:"executed"`
					Invested struct {
						USD decimal.Decimal `json:"usd"`
						Has decimal.Decimal `json:"has"`
					} `json:"invested"`
					Stock struct {
						USD *decimal.Decimal `json:"usd"`
						Has *decimal.Decimal `json:"has"`
					} `json:"stock"`
				} `json:"rows"`
			} `json:"balance"`
		}{
			Balance: struct {
				Rows []struct {
					Label    string `json:"label"`
					Executed struct {
						USD decimal.Decimal `json:"usd"`
						Has decimal.Decimal `json:"has"`
					} `json:"executed"`
					Invested struct {
						USD decimal.Decimal `json:"usd"`
						Has decimal.Decimal `json:"has"`
					} `json:"invested"`
					Stock struct {
						USD *decimal.Decimal `json:"usd"`
						Has *decimal.Decimal `json:"has"`
					} `json:"stock"`
				} `json:"rows"`
			}{
				Rows: []struct {
					Label    string `json:"label"`
					Executed struct {
						USD decimal.Decimal `json:"usd"`
						Has decimal.Decimal `json:"has"`
					} `json:"executed"`
					Invested struct {
						USD decimal.Decimal `json:"usd"`
						Has decimal.Decimal `json:"has"`
					} `json:"invested"`
					Stock struct {
						USD *decimal.Decimal `json:"usd"`
						Has *decimal.Decimal `json:"has"`
					} `json:"stock"`
				}{
					{
						Label: "Direct costs",
						Executed: struct {
							USD decimal.Decimal `json:"usd"`
							Has decimal.Decimal `json:"has"`
						}{
							USD: d.DirectCostsExecutedUSD,
							Has: d.DirectCostsHectares,
						},
						Invested: struct {
							USD decimal.Decimal `json:"usd"`
							Has decimal.Decimal `json:"has"`
						}{
							USD: d.DirectCostsInvestedUSD,
							Has: d.DirectCostsHectares,
						},
						Stock: struct {
							USD *decimal.Decimal `json:"usd"`
							Has *decimal.Decimal `json:"has"`
						}{
							USD: &d.DirectCostsStockUSD,
							Has: &d.DirectCostsHectares,
						},
					},
					{
						Label: "Seed",
						Executed: struct {
							USD decimal.Decimal `json:"usd"`
							Has decimal.Decimal `json:"has"`
						}{
							USD: d.SeedExecutedUSD,
							Has: d.SeedHectares,
						},
						Invested: struct {
							USD decimal.Decimal `json:"usd"`
							Has decimal.Decimal `json:"has"`
						}{
							USD: d.SeedInvestedUSD,
							Has: d.SeedHectares,
						},
						Stock: struct {
							USD *decimal.Decimal `json:"usd"`
							Has *decimal.Decimal `json:"has"`
						}{
							USD: &d.SeedStockUSD,
							Has: &d.SeedHectares,
						},
					},
					{
						Label: "Supplies",
						Executed: struct {
							USD decimal.Decimal `json:"usd"`
							Has decimal.Decimal `json:"has"`
						}{
							USD: d.SuppliesExecutedUSD,
							Has: d.SuppliesHectares,
						},
						Invested: struct {
							USD decimal.Decimal `json:"usd"`
							Has decimal.Decimal `json:"has"`
						}{
							USD: d.SuppliesInvestedUSD,
							Has: d.SuppliesHectares,
						},
						Stock: struct {
							USD *decimal.Decimal `json:"usd"`
							Has *decimal.Decimal `json:"has"`
						}{
							USD: &d.SuppliesStockUSD,
							Has: &d.SuppliesHectares,
						},
					},
					{
						Label: "Labors",
						Executed: struct {
							USD decimal.Decimal `json:"usd"`
							Has decimal.Decimal `json:"has"`
						}{
							USD: d.LaborsExecutedUSD,
							Has: d.LaborsHectares,
						},
						Invested: struct {
							USD decimal.Decimal `json:"usd"`
							Has decimal.Decimal `json:"has"`
						}{
							USD: d.LaborsInvestedUSD,
							Has: d.LaborsHectares,
						},
						Stock: struct {
							USD *decimal.Decimal `json:"usd"`
							Has *decimal.Decimal `json:"has"`
						}{
							USD: &d.LaborsStockUSD,
							Has: &d.LaborsHectares,
						},
					},
					{
						Label: "Rent",
						Executed: struct {
							USD decimal.Decimal `json:"usd"`
							Has decimal.Decimal `json:"has"`
						}{
							USD: d.RentExecutedUSD,
							Has: d.RentHectares,
						},
						Invested: struct {
							USD decimal.Decimal `json:"usd"`
							Has decimal.Decimal `json:"has"`
						}{
							USD: d.RentInvestedUSD,
							Has: d.RentHectares,
						},
						Stock: struct {
							USD *decimal.Decimal `json:"usd"`
							Has *decimal.Decimal `json:"has"`
						}{
							USD: &d.RentStockUSD,
							Has: &d.RentHectares,
						},
					},
					{
						Label: "Structure",
						Executed: struct {
							USD decimal.Decimal `json:"usd"`
							Has decimal.Decimal `json:"has"`
						}{
							USD: d.StructureExecutedUSD,
							Has: d.StructureHectares,
						},
						Invested: struct {
							USD decimal.Decimal `json:"usd"`
							Has decimal.Decimal `json:"has"`
						}{
							USD: d.StructureInvestedUSD,
							Has: d.StructureHectares,
						},
						Stock: struct {
							USD *decimal.Decimal `json:"usd"`
							Has *decimal.Decimal `json:"has"`
						}{
							USD: &d.StructureStockUSD,
							Has: &d.StructureHectares,
						},
					},
				},
			},
		},
	}
}
