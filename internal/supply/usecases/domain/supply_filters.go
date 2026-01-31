// Package domain define entidades de insumos.
package domain

// SupplyFilter permite filtros de workspace para listados.
type SupplyFilter struct {
	CustomerID *int64
	ProjectID  *int64
	CampaignID *int64
	FieldID    *int64
}
