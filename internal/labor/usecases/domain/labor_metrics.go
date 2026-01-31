// Package domain define entidades de labor.
package domain

import "github.com/shopspring/decimal"

// LaborMetrics agrega superficie y costo promedio por hectárea para labors
type LaborMetrics struct {
	SurfaceHa    decimal.Decimal
	NetTotalCost decimal.Decimal
	AvgCostPerHa decimal.Decimal
}

// LaborFilter permite filtrado opcional por workspace
type LaborFilter struct {
	CustomerID *int64
	ProjectID  *int64
	CampaignID *int64
	FieldID    *int64
}
