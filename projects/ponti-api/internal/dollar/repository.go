package dollar

import (
	"context"
	"errors"
	"time"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	"gorm.io/gorm"

	models "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dollar/repository/models"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dollar/usecases/domain"
)

type GormEnginePort interface {
	Client() *gorm.DB
}

type Repository struct {
	db GormEnginePort
}

func NewRepository(db GormEnginePort) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListByProject(ctx context.Context, projecID int64) ([]domain.DollarAverage, error) {
	// preparo la consulta filtrando por project_id
	tx := r.db.Client().
		WithContext(ctx).
		Model(&models.ProjectDollarValue{}).
		Where("project_id = ?", projecID)

	var rows []models.ProjectDollarValue
	// Ejecuto la consulta
	if err := tx.Find(&rows).Error; err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to list project dollar values", err)
	}

	// Convierto los modelos de DB a objetos de domain
	out := make([]domain.DollarAverage, len(rows))
	for i, m := range rows {
		out[i] = *m.ToDomain()
	}

	return out, nil
}

func (r *Repository) Create(ctx context.Context, items *domain.DollarAverage) (int64, error) {
	m := models.FromDomain(items)

	// inserto el registro
	if err := r.db.Client().WithContext(ctx).Create(m).Error; err != nil {
		return 0, types.NewError(types.ErrInternal, "failed to create dolar value", err)
	}

	return m.ID, nil
}

func (r *Repository) Update(ctx context.Context, item *domain.DollarAverage) error {
	// Valido que el ID no sea 0
	if item.ID == 0 {
		return types.NewError(types.ErrInvalidID, "invalid id", nil)
	}

	// Ejecuto una transaccion
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		// compruebo si el registro existe
		if err := tx.Model(&models.ProjectDollarValue{}).Where("id = ?", item.ID).Count(&count).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to check existence", err)
		}
		if count == 0 {
			return types.NewError(types.ErrNotFound, "project dolar value not found", nil)
		}

		// actualizo los campos
		if err := tx.Model(&models.ProjectDollarValue{}).
			Where("id = ?", item.ID).
			Updates(map[string]interface{}{
				"project_id":    item.ProjectID,
				"year":          item.Year,
				"month":         item.Month,
				"start_value":   item.StartValue,
				"end_value":     item.EndValue,
				"average_value": item.AvgValue,
				"updated_at":    time.Now(),
			}).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to update project dollar value", err)
		}

		return nil
	})
}

func (r *Repository) GetByComposite(ctx context.Context, projectID, year int64, month string) (*domain.DollarAverage, error) {
	var m models.ProjectDollarValue

	// consulto por tres campos
	err := r.db.Client().WithContext(ctx).
		Where("project_id = ? AND year = ? AND month = ?", projectID, year, month).
		First(&m).Error

	// si no existe devuelvo un nil, sin error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to query by composite key", err)
	}

	// si existe convierle el modelo a dominio y retorna
	return m.ToDomain(), nil
}
