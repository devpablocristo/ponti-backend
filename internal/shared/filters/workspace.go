// Package filters provee helpers de filtros compartidos.
package filters

import (
	"context"

	"gorm.io/gorm"

	"github.com/devpablocristo/core/saas/go/shared/domainerr"
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

	var projectIDs []int64
	if err := query.Pluck("p.id", &projectIDs).Error; err != nil {
		return nil, domainerr.Internal("failed to resolve project IDs")
	}

	return projectIDs, nil
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
