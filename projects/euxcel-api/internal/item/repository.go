package item

import (
	"context"
	"errors"
	"fmt"

	gorm0 "gorm.io/gorm"

	gorm "github.com/alphacodinggroup/euxcel-backend/pkg/databases/sql/gorm"
	pkgtypes "github.com/alphacodinggroup/euxcel-backend/pkg/types"
	models "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/item/repository/models"
	domain "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/item/usecases/domain"
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
		return 0, pkgtypes.NewError(pkgtypes.ErrValidation, "item is nil", nil)
	}

	// Convertir la entidad de dominio al modelo GORM.
	model := models.FromDomainItem(item)
	// No se asigna ID manualmente: se deja que la base de datos lo genere.
	if err := r.db.Client().WithContext(ctx).Create(model).Error; err != nil {
		return 0, pkgtypes.NewError(pkgtypes.ErrInternal, "failed to create item", err)
	}

	return model.ID, nil
}

// ListItems obtiene todos los items de la base de datos y los convierte a dominio.
func (r *repository) ListItems(ctx context.Context) ([]domain.Item, error) {
	var items []models.Item
	if err := r.db.Client().WithContext(ctx).Find(&items).Error; err != nil {
		return nil, pkgtypes.NewError(pkgtypes.ErrInternal, "failed to list items", err)
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
	err := r.db.Client().WithContext(ctx).Where("id = ?", id).First(&item).Error
	if err != nil {
		// Suponiendo que gorm.ErrRecordNotFound está definido en el paquete gorm.
		if errors.Is(err, gorm0.ErrRecordNotFound) {
			return nil, pkgtypes.NewError(pkgtypes.ErrNotFound, fmt.Sprintf("item with id %d not found", id), err)
		}
		return nil, pkgtypes.NewError(pkgtypes.ErrInternal, "failed to get item", err)
	}
	return item.ToDomain(), nil
}

// UpdateItem actualiza un item existente.
// Se convierte la entidad de dominio al modelo y se guarda.
// Si no se encuentra un item con el ID indicado, retorna un error.
func (r *repository) UpdateItem(ctx context.Context, item *domain.Item) error {
	if item == nil {
		return pkgtypes.NewError(pkgtypes.ErrValidation, "item is nil", nil)
	}

	// Ejecutar el update filtrando por ID y actualizando con los nuevos valores.
	result := r.db.Client().WithContext(ctx).
		Model(&models.Item{}).
		Where("id = ?", item.ID).
		Updates(models.FromDomainItem(item))

	if result.Error != nil {
		return pkgtypes.NewError(pkgtypes.ErrInternal, "failed to update item", result.Error)
	}

	if result.RowsAffected == 0 {
		return pkgtypes.NewError(pkgtypes.ErrNotFound, fmt.Sprintf("item with id %d does not exist", item.ID), nil)
	}

	return nil
}

// DeleteItem elimina un item a partir de su ID.
// Si no se encuentra un item con el ID indicado, retorna un error.
func (r *repository) DeleteItem(ctx context.Context, id int64) error {
	result := r.db.Client().WithContext(ctx).
		Delete(&models.Item{}, "id = ?", id)

	if result.Error != nil {
		return pkgtypes.NewError(pkgtypes.ErrInternal, "failed to delete item", result.Error)
	}

	if result.RowsAffected == 0 {
		return pkgtypes.NewError(pkgtypes.ErrNotFound, fmt.Sprintf("item with id %d does not exist", id), nil)
	}

	return nil
}
