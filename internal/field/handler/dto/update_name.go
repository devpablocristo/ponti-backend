package dto

// UpdateFieldNameRequest es el payload para editar SOLO el nombre del campo
// (edición desde el catálogo unificado / registry). No toca lease_type ni lotes.
type UpdateFieldNameRequest struct {
	Name string `json:"name" binding:"required"`
}
