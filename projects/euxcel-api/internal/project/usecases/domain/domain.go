package domain

type Project struct {
	ID               int64   // Primary key (INT)
	Name             string  // Project name (VARCHAR)
	CustomerID       int64   // Foreign key referencing the responsible company (INT)
	ProjectAdminCost float64 // Administrative cost of the project (DECIMAL)
	AdminResponsible string  // Administrative responsible (VARCHAR)
}
