// Package domain define los tipos del registry (búsqueda unificada de entidades por tipo).
package domain

// RegistryRow es una fila tipada del registry (actor o catálogo).
type RegistryRow struct {
	EntityType string // "actor" | "crops" | "types" | "lease-types" | "campaigns"
	ID         int64
	Name       string
	Tax        string   // CUIT/CUIL/DNI (solo actores)
	Roles      []string // solo actores
	Archived   bool
}

// RegistryResult es una página de resultados + el total para paginar.
type RegistryResult struct {
	Rows  []RegistryRow
	Total int64
}
