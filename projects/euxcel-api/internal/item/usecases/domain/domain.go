package domain

// MacroCategory representa una macro categoría (por ejemplo: "Agrochemicals", "Fertilizers", "Seeds", "Operations").
type MacroCategory struct {
	ID   int64  // Primary key (numeric)
	Name string // Name of the macro category
}

// Category representa una categoría específica (por ejemplo: "Herbicides", "Seeds") asociada a una MacroCategory.
type Category struct {
	ID              int64  // Primary key (numeric)
	Name            string // Category name
	MacroCategoryID int64  // Foreign key referencing MacroCategory
}

// Supplier representa un proveedor.
type Supplier struct {
	ID   int64  // Primary key (numeric)
	Name string // Supplier name
	// Se pueden agregar campos adicionales (e.g., contact information)
}

// Item representa un artículo o ítem, el cual hace referencia a una Category y un Supplier.
type Item struct {
	ID         int64   // Primary key (numeric)
	Name       string  // Item name
	PriceUSD   float64 // Price in USD
	CategoryID int64   // Foreign key referencing Category
	SupplierID int64   // Foreign key referencing Supplier
}
