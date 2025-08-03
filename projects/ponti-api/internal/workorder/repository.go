package workorder

import (
	"context"
	"errors"

	"gorm.io/gorm"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/models"
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

func (r *Repository) CreateWorkorder(ctx context.Context, o *domain.Workorder) (int64, error) {
	// 1) convertir a modelo GORM (cabecera + items sin WorkorderID)
	model := models.FromDomain(o)

	// 2) poblar auditoría
	if userID, err := sharedmodels.ConvertStringToID(ctx); err == nil {
		model.CreatedBy = &userID
		model.UpdatedBy = &userID
	}

	// 3) crear todo en una transacción
	err := r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 3.1) insertar la cabecera para obtener model.ID
		if err := tx.Create(&model).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to create workorder header", err)
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return model.ID, nil
}

func (r *Repository) GetWorkorderByID(ctx context.Context, id int64) (*domain.Workorder, error) {
	var m models.Workorder
	if err := r.db.Client().WithContext(ctx).
		Preload("Items").
		Where("id = ?", id).
		First(&m).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.NewError(types.ErrNotFound, "workorder not found", err)
		}
		return nil, types.NewError(types.ErrInternal, "failed to get workorder", err)
	}
	return m.ToDomain(), nil
}

func (r *Repository) UpdateWorkorderByID(ctx context.Context, o *domain.Workorder) error {
	// 1) Convertimos dominio → GORM y fijamos el ID
	model := models.FromDomain(o)
	model.ID = o.ID

	// 2) Poblar UpdatedBy si hay usuario en contexto
	if userID, err := sharedmodels.ConvertStringToID(ctx); err == nil {
		model.UpdatedBy = &userID
	}

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 3.1) Recuperar original para validar existencia y conservar auditoría
		var orig models.Workorder
		if err := tx.Preload("Items").First(&orig, model.ID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return types.NewError(types.ErrNotFound, "workorder not found", err)
			}
			return types.NewError(types.ErrInternal, "failed to find workorder before update", err)
		}

		// 3.2) Eliminar todos los items antiguos
		if err := tx.
			Where("workorder_id = ?", model.ID).
			Delete(&models.WorkorderItem{}).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to delete old items", err)
		}

		// 3.3) Actualizar sólo la cabecera, omitiendo campos de auditoría y la asociación Items
		if err := tx.Model(&orig).
			Omit("CreatedAt", "CreatedBy", "DeletedAt", "DeletedBy", "Items").
			Updates(model).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to update workorder header", err)
		}

		// 3.4) Insertar los items nuevos, asignando WorkorderID
		for i := range model.Items {
			model.Items[i].WorkorderID = model.ID
		}
		if len(model.Items) > 0 {
			if err := tx.Create(&model.Items).Error; err != nil {
				return types.NewError(types.ErrInternal, "failed to insert new items", err)
			}
		}

		return nil
	})
}

func (r *Repository) DeleteWorkorderByID(ctx context.Context, id int64) error {
	var m models.Workorder

	// 1) Buscamos la workorder por su ID
	if err := r.db.Client().
		WithContext(ctx).
		First(&m, id).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return types.NewError(types.ErrNotFound, "workorder not found", err)
		}
		return types.NewError(types.ErrInternal, "failed to find workorder", err)
	}

	// 2) Soft-delete dentro de una transacción
	if err := r.db.Client().
		WithContext(ctx).
		Transaction(func(tx *gorm.DB) error {
			if err := tx.Delete(&m).Error; err != nil {
				return types.NewError(types.ErrInternal, "failed to soft delete workorder", err)
			}
			return nil
		}); err != nil {
		return err
	}

	return nil
}

// func (r *Repository) DuplicateWorkorder(ctx context.Context, number string) (string, error) {
// 	orig, err := r.GetWorkorderByNumber(ctx, number)
// 	if err != nil {
// 		return "", err
// 	}

// 	return r.CreateWorkorder(ctx, orig)
// }

func (r *Repository) ListWorkorders(
	ctx context.Context,
	filt domain.WorkorderFilter,
	inp types.Input,
) ([]domain.WorkorderListElement, types.PageInfo, error) {
	// 1) Base del query: vinculada a la vista
	base := r.db.Client().
		WithContext(ctx).
		Model(&models.WorkorderListElement{})

	// 2) Aplicar filtros
	if filt.ProjectID != nil {
		base = base.Where("project_id = ?", *filt.ProjectID)
	}
	if filt.FieldID != nil {
		base = base.Where("field_id = ?", *filt.FieldID)
	}

	// 3) Contar total
	var total int64
	if err := base.
		Count(&total).Error; err != nil {
		return nil, types.PageInfo{}, types.NewError(types.ErrInternal,
			"failed to count workorders", err)
	}

	// 4) Paginación
	offset := (int(inp.Page) - 1) * int(inp.PageSize)

	// 5) Recuperar filas paginadas (reutiliza 'base' con filtros)
	var rows []models.WorkorderListElement
	if err := base.
		Limit(int(inp.PageSize)).
		Offset(offset).
		Order("number desc").
		Find(&rows).Error; err != nil {
		return nil, types.PageInfo{}, types.NewError(types.ErrInternal,
			"failed to list workorders", err)
	}

	// 6) Mapear a dominio
	list := make([]domain.WorkorderListElement, len(rows))
	for i, m := range rows {
		list[i] = domain.WorkorderListElement{
			ID:           m.ID,
			Number:       m.Number,
			ProjectName:  m.ProjectName,
			FieldName:    m.FieldName,
			LotName:      m.LotName,
			Date:         m.Date,
			CropName:     m.CropName,
			LaborName:    m.LaborName,
			TypeName:     m.TypeName,
			Contractor:   m.Contractor,
			SurfaceHa:    m.SurfaceHa,
			SupplyName:   m.SupplyName,
			Consumption:  m.Consumption,
			CategoryName: m.CategoryName,
			Dose:         m.Dose,
			CostPerHa:    m.CostPerHa,
			UnitPrice:    m.UnitPrice,
			TotalCost:    m.TotalCost,
		}
	}

	// 7) Construir PageInfo y devolver
	pageInfo := types.NewPageInfo(int(inp.Page), int(inp.PageSize), total)
	return list, pageInfo, nil
}
