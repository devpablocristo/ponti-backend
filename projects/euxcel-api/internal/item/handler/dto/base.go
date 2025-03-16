package dto

import (
	"errors"

	"github.com/alphacodinggroup/euxcel-backend/internal/item/usecases/domain"
)

// -----------------------------
// MacroCategory DTO y mappers
// -----------------------------

// MacroCategory es el DTO para una macro categoría.
type MacroCategory struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// ToDomain convierte el DTO MacroCategory a la entidad de dominio.
func (m MacroCategory) ToDomain() *domain.MacroCategory {
	return &domain.MacroCategory{
		ID:   m.ID,
		Name: m.Name,
	}
}

// FromDomainMacroCategory convierte una entidad de dominio a DTO MacroCategory.
func FromDomainMacroCategory(m domain.MacroCategory) *MacroCategory {
	return &MacroCategory{
		ID:   m.ID,
		Name: m.Name,
	}
}

// -----------------------------
// Category DTO y mappers
// -----------------------------

// Category es el DTO para una categoría específica.
type Category struct {
	ID              int64  `json:"id"`
	Name            string `json:"name"`
	MacroCategoryID int64  `json:"macro_category_id"`
}

// ToDomain convierte el DTO Category a la entidad de dominio.
func (c Category) ToDomain() *domain.Category {
	return &domain.Category{
		ID:              c.ID,
		Name:            c.Name,
		MacroCategoryID: c.MacroCategoryID,
	}
}

// FromDomainCategory convierte una entidad de dominio a DTO Category.
func FromDomainCategory(c domain.Category) *Category {
	return &Category{
		ID:              c.ID,
		Name:            c.Name,
		MacroCategoryID: c.MacroCategoryID,
	}
}

// -----------------------------
// Supplier DTO y mappers
// -----------------------------

// Supplier es el DTO para un proveedor.
type Supplier struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// ToDomain convierte el DTO Supplier a la entidad de dominio.
func (s Supplier) ToDomain() *domain.Supplier {
	return &domain.Supplier{
		ID:   s.ID,
		Name: s.Name,
	}
}

// FromDomainSupplier convierte una entidad de dominio a DTO Supplier.
func FromDomainSupplier(s domain.Supplier) *Supplier {
	return &Supplier{
		ID:   s.ID,
		Name: s.Name,
	}
}

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
