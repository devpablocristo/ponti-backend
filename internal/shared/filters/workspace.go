// Package filters provee helpers de filtros compartidos.
package filters

import (
	"context"

	"gorm.io/gorm"

	"github.com/devpablocristo/platform/errors/go/domainerr"

	sharedmodels "github.com/devpablocristo/ponti-backend/internal/shared/models"
)

// WorkspaceFilter define filtros comunes de workspace.
type WorkspaceFilter struct {
	CustomerID *int64
	ProjectID  *int64
	CampaignID *int64
	FieldID    *int64
}

// WorkspaceFilterColumns define columnas aplicables en el query.
type WorkspaceFilterColumns struct {
	CustomerID string
	ProjectID  string
	CampaignID string
	FieldID    string
}

// ApplyWorkspaceFilters aplica filtros comunes al query si hay columnas disponibles.
func ApplyWorkspaceFilters(q *gorm.DB, f WorkspaceFilter, cols WorkspaceFilterColumns) *gorm.DB {
	if f.CustomerID != nil && cols.CustomerID != "" {
		q = q.Where(cols.CustomerID+" = ?", *f.CustomerID)
	}
	if f.ProjectID != nil && cols.ProjectID != "" {
		q = q.Where(cols.ProjectID+" = ?", *f.ProjectID)
	}
	if f.CampaignID != nil && cols.CampaignID != "" {
		q = q.Where(cols.CampaignID+" = ?", *f.CampaignID)
	}
	if f.FieldID != nil && cols.FieldID != "" {
		q = q.Where(cols.FieldID+" = ?", *f.FieldID)
	}
	return q
}

// ResolveProjectIDs devuelve project_ids aplicando customer/campaign/field si project_id no viene definido.
func ResolveProjectIDs(ctx context.Context, db *gorm.DB, f WorkspaceFilter) ([]int64, error) {
	if f.ProjectID != nil {
		query := db.WithContext(ctx).
			Table("projects p").
			Where("p.id = ? AND p.deleted_at IS NULL", *f.ProjectID)
		if f.CustomerID != nil {
			query = query.Where("p.customer_id = ?", *f.CustomerID)
		}
		if f.CampaignID != nil {
			query = query.Where("p.campaign_id = ?", *f.CampaignID)
		}
		if f.FieldID != nil {
			query = query.Where(
				"EXISTS (SELECT 1 FROM fields f WHERE f.id = ? AND f.project_id = p.id AND f.deleted_at IS NULL)",
				*f.FieldID,
			)
		}

		// T1.e: acotar al tenant activo (flag-gated). Valida que el project pedido
		// pertenezca al tenant del JWT (cierra la fuga de workspace-ownership).
		if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
			query = query.Where("p.tenant_id = ?", orgID)
		}

		var count int64
		if err := query.Count(&count).Error; err != nil {
			return nil, domainerr.Internal("failed to validate project filters")
		}
		if count == 0 {
			return nil, domainerr.Validation("project_id does not match provided filters")
		}

		return []int64{*f.ProjectID}, nil
	}

	if f.CustomerID == nil && f.CampaignID == nil && f.FieldID == nil {
		return nil, nil
	}

	query := db.WithContext(ctx).
		Table("projects p").
		Select("DISTINCT p.id").
		Where("p.deleted_at IS NULL")

	if f.CustomerID != nil {
		query = query.Where("p.customer_id = ?", *f.CustomerID)
	}
	if f.CampaignID != nil {
		query = query.Where("p.campaign_id = ?", *f.CampaignID)
	}
	if f.FieldID != nil {
		query = query.Where(
			"EXISTS (SELECT 1 FROM fields f WHERE f.id = ? AND f.project_id = p.id AND f.deleted_at IS NULL)",
			*f.FieldID,
		)
	}

	// T1.e: acotar al tenant activo (flag-gated).
	if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
		query = query.Where("p.tenant_id = ?", orgID)
	}

	var projectIDs []int64
	if err := query.Pluck("p.id", &projectIDs).Error; err != nil {
		return nil, domainerr.Internal("failed to resolve project IDs")
	}

	return projectIDs, nil
}

// === T-child guards (Modelo 2): pertenencia al tenant de entidades HIJAS que NO
// tienen tenant_id propio (llegan al tenant vía projects.tenant_id). Todo flag-gated:
// con TENANT_ENFORCEMENT off devuelven nil / predicado vacío (comportamiento idéntico). ===

// GuardProjectForTenant valida que projectID pertenezca al tenant activo. Para
// Create/operaciones que reciben un project_id del cliente. nil si el flag está off o
// el project es del tenant; NotFound si no lo es (404, sin revelar existencia).
func GuardProjectForTenant(ctx context.Context, db *gorm.DB, projectID int64) error {
	orgID, ok := sharedmodels.OrgIDFromContext(ctx)
	if !ok || !sharedmodels.TenantEnforcementEnabled() || projectID <= 0 {
		return nil
	}
	var count int64
	if err := db.WithContext(ctx).
		Table("projects").
		Where("id = ? AND tenant_id = ?", projectID, orgID).
		Count(&count).Error; err != nil {
		return domainerr.Internal("failed to validate project tenant")
	}
	if count == 0 {
		return domainerr.NotFound("project not found")
	}
	return nil
}

// GuardFieldForTenant valida que fieldID pertenezca a un project del tenant activo
// (fields.project_id -> projects.tenant_id). Para Create/operaciones que reciben
// field_id pero no project_id (ej. lots). nil si el flag está off o el field es del tenant.
func GuardFieldForTenant(ctx context.Context, db *gorm.DB, fieldID int64) error {
	orgID, ok := sharedmodels.OrgIDFromContext(ctx)
	if !ok || !sharedmodels.TenantEnforcementEnabled() || fieldID <= 0 {
		return nil
	}
	var count int64
	if err := db.WithContext(ctx).
		Table("fields f").
		Joins("JOIN projects p ON p.id = f.project_id").
		Where("f.id = ? AND p.tenant_id = ?", fieldID, orgID).
		Count(&count).Error; err != nil {
		return domainerr.Internal("failed to validate field tenant")
	}
	if count == 0 {
		return domainerr.NotFound("field not found")
	}
	return nil
}

// TenantProjectScope devuelve un predicado SQL (+args) para acotar mutaciones por-id
// propio de entidades hijas con columna project_id (workorders/supplies/supply_movements/
// fields) al tenant activo. ("", nil) si el flag está off → no se agrega filtro.
// Uso:
//
//	q := tx.Model(&X{}).Where("id = ?", id)
//	if cond, args := filters.TenantProjectScope(ctx); cond != "" { q = q.Where(cond, args...) }
//
// Con el filtro, RowsAffected==0 si la fila no es del tenant → aplica el NotFound existente.
func TenantProjectScope(ctx context.Context) (string, []any) {
	orgID, ok := sharedmodels.OrgIDFromContext(ctx)
	if !ok || !sharedmodels.TenantEnforcementEnabled() {
		return "", nil
	}
	return "project_id IN (SELECT id FROM projects WHERE tenant_id = ?)", []any{orgID}
}

// TenantFieldScope: idem para entidades hija-de-hija con columna field_id (lots),
// vía fields.project_id -> projects.tenant_id. ("", nil) si el flag está off.
func TenantFieldScope(ctx context.Context) (string, []any) {
	orgID, ok := sharedmodels.OrgIDFromContext(ctx)
	if !ok || !sharedmodels.TenantEnforcementEnabled() {
		return "", nil
	}
	return "field_id IN (SELECT id FROM fields WHERE project_id IN (SELECT id FROM projects WHERE tenant_id = ?))", []any{orgID}
}

// ValidateFieldBelongsToProject valida que field_id pertenezca al project_id.
func ValidateFieldBelongsToProject(ctx context.Context, db *gorm.DB, projectID int64, fieldID int64) error {
	if projectID <= 0 || fieldID <= 0 {
		return nil
	}

	var count int64
	if err := db.WithContext(ctx).
		Table("fields f").
		Where("f.id = ? AND f.project_id = ? AND f.deleted_at IS NULL", fieldID, projectID).
		Count(&count).Error; err != nil {
		return domainerr.Internal("failed to validate field against project")
	}
	if count == 0 {
		return domainerr.Validation("field_id does not belong to project_id")
	}
	return nil
}
