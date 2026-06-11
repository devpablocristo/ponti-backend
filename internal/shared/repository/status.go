package sharedrepo

import "gorm.io/gorm"

// ScopeByStatus aplica el filtro de archivado a una query de catálogo con soft-delete:
//
//	"archived" → solo archivados (deleted_at IS NOT NULL)
//	"all"      → activos + archivados
//	""/default → solo activos (scope por defecto de gorm)
func ScopeByStatus(tx *gorm.DB, status string) *gorm.DB {
	switch status {
	case "archived":
		return tx.Unscoped().Where("deleted_at IS NOT NULL")
	case "all":
		return tx.Unscoped()
	default:
		return tx
	}
}
