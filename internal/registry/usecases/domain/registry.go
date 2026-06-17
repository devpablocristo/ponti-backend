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
	Subtitle   string
}

// RegistryResult es una página de resultados + el total para paginar.
type RegistryResult struct {
	Rows  []RegistryRow
	Total int64
}

// UsageItem es un proyecto que referencia una entidad del catálogo.
type UsageItem struct {
	ID       int64
	Name     string
	Customer string
	Campaign string
}

// UsageResult es la lista de proyectos que usan una entidad dada.
type UsageResult struct {
	Items []UsageItem
	Total int
}
