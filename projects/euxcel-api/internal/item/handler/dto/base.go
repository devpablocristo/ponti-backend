package dto

import (
	"errors"

	"github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/item/usecases/domain"
)

// -----------------------------
// Item DTO y mappers
// -----------------------------

// Item es el DTO para un artículo o ítem.
type Item struct {
	ID         int64   `json:"id"`
	Name       string  `json:"name"`
	PriceUSD   float64 `json:"price_usd"`
	CategoryID int64   `json:"category_id"`
	SupplierID int64   `json:"supplier_id"`
}

// Validate verifica que los campos del DTO Item sean válidos.
func (i Item) Validate() error {
	if i.Name == "" {
		return errors.New("item name cannot be empty")
	}
	if i.PriceUSD < 0 {
		return errors.New("price_usd cannot be negative")
	}
	if i.CategoryID <= 0 {
		return errors.New("category_id must be greater than zero")
	}
	if i.SupplierID <= 0 {
		return errors.New("supplier_id must be greater than zero")
	}
	return nil
}

// ToDomain convierte el DTO Item a la entidad de dominio.
// Retorna (*domain.Item, error) para gestionar errores de validación.
func (i Item) ToDomain() (*domain.Item, error) {
	if err := i.Validate(); err != nil {
		return nil, err
	}
	return &domain.Item{
		ID:         i.ID,
		Name:       i.Name,
		PriceUSD:   i.PriceUSD,
		CategoryID: i.CategoryID,
		SupplierID: i.SupplierID,
	}, nil
}

// FromDomainItem convierte una entidad de dominio a DTO Item.
func FromDomainItem(i domain.Item) *Item {
	return &Item{
		ID:         i.ID,
		Name:       i.Name,
		PriceUSD:   i.PriceUSD,
		CategoryID: i.CategoryID,
		SupplierID: i.SupplierID,
	}
}
