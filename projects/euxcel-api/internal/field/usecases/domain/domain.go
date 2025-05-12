// File: internal/field/usecases/domain/field.go
package domain

import (
	lotdom "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/lot/usecases/domain"
)

// Field represents the domain entity for a field, including its lots.
type Field struct {
	ID              int64        // Auto-generated primary key
	ProjectID       int64        // Foreign key to project
	Name            string       // Human-readable name
	LeasePercentage float64      // Percentage of lease, expressed as decimal
	LeaseType       string       // Type of lease
	Lots            []lotdom.Lot // Child lots belonging to this field
}
