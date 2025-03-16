package item

import (
	"context"
	"errors"
	"fmt"

	models "github.com/alphacodinggroup/euxcel-backend/internal/item/repository/models"
	domain "github.com/alphacodinggroup/euxcel-backend/internal/item/usecases/domain"
	gorm "github.com/alphacodinggroup/euxcel-backend/pkg/databases/sql/gorm"
)

type repository struct {
	db gorm.Repository
}

// NewRepository crea una instancia del adaptador GORM para items.
func NewRepository(db gorm.Repository) Repository {
	return &repository{
		db: db,
	}
}

// CreateItem inserta un nuevo item en la base de datos.
// Se asume que la columna ID es autoincremental y se asignará automáticamente.
func (r *repository) CreateItem(ctx context.Context, item *domain.Item) (int64, error) {
	if item == nil {
		return 0, errors.New("item is nil")
	}

	// Convertir la entidad de dominio al modelo GORM.
	model := models.FromDomainItem(item)
	// No se asigna ID manualmente: se deja que la base de datos lo genere.
	if err := r.db.Client().WithContext(ctx).Create(model).Error; err != nil {
		return 0, fmt.Errorf("failed to create item: %w", err)
	}

	return model.ID, nil
}

// ListItems obtiene todos los items de la base de datos y los convierte a dominio.
func (r *repository) ListItems(ctx context.Context) ([]domain.Item, error) {
	var items []models.Item
	if err := r.db.Client().WithContext(ctx).Find(&items).Error; err != nil {
		return nil, err
	}

	domainItems := make([]domain.Item, 0, len(items))
	for _, m := range items {
		domainItems = append(domainItems, *m.ToDomain())
	}
	return domainItems, nil
}

// GetItem obtiene un item por su ID y lo convierte a la entidad de dominio.
func (r *repository) GetItem(ctx context.Context, id int64) (*domain.Item, error) {
	var item models.Item
	if err := r.db.Client().WithContext(ctx).Where("id = ?", id).First(&item).Error; err != nil {
		return nil, err
	}
	return item.ToDomain(), nil
}

// UpdateItem actualiza un item existente.
// Se convierte la entidad de dominio al modelo y se guarda.
func (r *repository) UpdateItem(ctx context.Context, item *domain.Item) error {
	model := models.FromDomainItem(item)
	return r.db.Client().WithContext(ctx).Save(model).Error
}

// DeleteItem elimina un item a partir de su ID.
func (r *repository) DeleteItem(ctx context.Context, id int64) error {
	return r.db.Client().WithContext(ctx).Delete(&models.Item{}, "id = ?", id).Error
}
