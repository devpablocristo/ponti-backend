package domain

type Item struct {
	ID         int64   // Primary key (numeric)
	Name       string  // Item name
	PriceUSD   float64 // Price in USD
	CategoryID int64   // Foreign key referencing Category
	SupplierID int64   // Foreign key referencing Supplier
}
