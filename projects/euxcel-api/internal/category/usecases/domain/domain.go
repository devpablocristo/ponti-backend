package domain

// Category representa una categoría específica (por ejemplo: "Herbicides", "Seeds") asociada a una MacroCategory.
type Category struct {
	ID              int64  // Primary key (numeric)
	Name            string // Category name
	MacroCategoryID int64  // Foreign key referencing MacroCategory
}
