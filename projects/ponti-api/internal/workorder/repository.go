package workorder

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/workorder/repository/models"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/workorder/usecases/domain"
)

type GormEngine interface {
	Client() *gorm.DB
}

type Repository struct {
	db GormEngine
}

func NewRepository(db GormEngine) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateWorkorder(ctx context.Context, o *domain.Workorder) (string, error) {
	// 1) mapeamos dominio -> modelo
	ordModel := models.FromDomain(o)

	err := r.db.Client().WithContext(ctx).
		Transaction(func(tx *gorm.DB) error {
			// 2) FullSaveAssociations para que cree también los Items
			if err := tx.Session(&gorm.Session{FullSaveAssociations: true}).
				Create(ordModel).Error; err != nil {
				return types.NewError(types.ErrInternal, "failed to create order", err)
			}
			return nil
		})

	return ordModel.Number, err
}

func (r *Repository) GetWorkorder(ctx context.Context, number string) (*domain.Workorder, error) {
	var ordModel models.Workorder
	if err := r.db.Client().WithContext(ctx).
		Preload("Items").
		First(&ordModel, "number = ?", number).Error; err != nil {

		if err == gorm.ErrRecordNotFound {
			return nil, types.NewError(types.ErrNotFound, "order not found", err)
		}
		return nil, types.NewError(types.ErrInternal, "failed to get order", err)
	}

	// 3) modelo -> dominio
	return ordModel.ToDomain(), nil
}

func (r *Repository) UpdateWorkorder(ctx context.Context, o *domain.Workorder) error {
	ordModel := models.FromDomain(o)

	return r.db.Client().WithContext(ctx).
		Transaction(func(tx *gorm.DB) error {
			// verificamos existencia
			var count int64
			if err := tx.Model(&models.Workorder{}).
				Where("number = ?", ordModel.Number).
				Count(&count).Error; err != nil {
				return types.NewError(types.ErrInternal, "failed to check order existence", err)
			}
			if count == 0 {
				return types.NewError(types.ErrNotFound, "order not found", nil)
			}

			// actualizamos campos y asociaciones de una vez
			if err := tx.Session(&gorm.Session{FullSaveAssociations: true}).
				Updates(ordModel).Error; err != nil {
				return types.NewError(types.ErrInternal, "failed to update order", err)
			}
			return nil
		})
}

func (r *Repository) DeleteWorkorder(ctx context.Context, number string) error {
	return r.db.Client().WithContext(ctx).
		Transaction(func(tx *gorm.DB) error {
			// verificar si existe
			var count int64
			if err := tx.Model(&models.Workorder{}).
				Where("number = ?", number).
				Count(&count).Error; err != nil {
				return types.NewError(types.ErrInternal, "failed to check order existence", err)
			}
			if count == 0 {
				return types.NewError(types.ErrNotFound, "order not found", nil)
			}

			// eliminamos items y luego la orden
			if err := tx.Where("order_number = ?", number).
				Delete(&models.WorkorderItem{}).Error; err != nil {
				return types.NewError(types.ErrInternal, "failed to delete order items", err)
			}
			if err := tx.Delete(&models.Workorder{}, "number = ?", number).Error; err != nil {
				return types.NewError(types.ErrInternal, "failed to delete order", err)
			}
			return nil
		})
}

func (r *Repository) DuplicateWorkorder(ctx context.Context, number string) (string, error) {
	// reuso Get + Create
	orig, err := r.GetWorkorder(ctx, number)
	if err != nil {
		return "", err
	}
	newNum, err := getNextNumber(ctx, r.db.Client())
	if err != nil {
		return "", err
	}
	orig.Number = newNum
	return r.CreateWorkorder(ctx, orig)
}

func getNextNumber(ctx context.Context, db *gorm.DB) (string, error) {
	var seq int64
	if err := db.WithContext(ctx).
		Raw("SELECT nextval('workorder_number_seq')").
		Scan(&seq).Error; err != nil {
		return "", types.NewError(types.ErrInternal, "failed to get next work order number from sequence", err)
	}
	return fmt.Sprintf("%04d", seq), nil
}
