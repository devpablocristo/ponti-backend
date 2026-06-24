package dto

// UpdateProjectNameRequest es el payload para editar SOLO el nombre del proyecto
// (edición desde el catálogo unificado / registry). No toca el resto del proyecto.
type UpdateProjectNameRequest struct {
	Name string `json:"name" binding:"required"`
}
