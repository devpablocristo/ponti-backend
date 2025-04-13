package domain

type Field struct {
	ID              int64   // Primary key
	Name            string  // Field name
	ProjectID       int64   // Associated project's ID
	LeasePercentage float64 // Lease percentage, expressed as a decimal value
	LeaseType       string  // Type of lease
}
