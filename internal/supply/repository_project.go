package supply

import (
	"context"

	sharedrepo "github.com/alphacodinggroup/ponti-backend/internal/shared/repository"
)

func (r *Repository) GetProjectNameByID(ctx context.Context, projectID int64) (string, error) {
	var row struct {
		ID   int64  `gorm:"column:id"`
		Name string `gorm:"column:name"`
	}

	err := r.db.Client().WithContext(ctx).
		Table("projects").
		Select("id, name").
		Where("id = ? AND deleted_at IS NULL", projectID).
		Take(&row).Error
	if err != nil {
		return "", sharedrepo.HandleGormError(err, "project", projectID)
	}

	return row.Name, nil
}
