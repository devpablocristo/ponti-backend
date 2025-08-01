package workorder

import (
	"context"
	"errors"
	"fmt"

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

func (r *Repository) CreateWorkorder(ctx context.Context, o *domain.Workorder) (string, error) {
	// 1) convertir a modelo GORM
	model := models.FromDomain(o)

	// 2) poblar auditoría
	if userID, err := sharedmodels.ConvertStringToID(ctx); err == nil {
		model.CreatedBy = &userID
		model.UpdatedBy = &userID
	}

	// 3) crear dentro de transacción
	err := r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(model).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to create workorder", err)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return model.Number, nil
}

func (r *Repository) GetWorkorderByNumber(ctx context.Context, number string) (*domain.Workorder, error) {
	var m models.Workorder
	if err := r.db.Client().WithContext(ctx).
		Preload("Items").
		Where("number = ?", number).
		First(&m).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.NewError(types.ErrNotFound, "workorder not found", err)
		}
		return nil, types.NewError(types.ErrInternal, "failed to get workorder", err)
	}
	return m.ToDomain(), nil
}

func (r *Repository) UpdateWorkorder(ctx context.Context, o *domain.Workorder) error {
	// 1) buscar el ID existente por número
	var existing models.Workorder
	if err := r.db.Client().WithContext(ctx).
		Select("id").
		Where("number = ?", o.Number).
		First(&existing).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return types.NewError(types.ErrNotFound, "workorder not found", err)
		}
		return types.NewError(types.ErrInternal, "failed to find workorder ID", err)
	}

	// 2) mapear dominio → GORM e inyectar el ID
	model := models.FromDomain(o)
	model.ID = existing.ID

	// 3) poblar UpdatedBy
	if userID, err := sharedmodels.ConvertStringToID(ctx); err == nil {
		model.UpdatedBy = &userID
	}

	// 4) guardar con asociaciones dentro de transacción
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Session(&gorm.Session{FullSaveAssociations: true}).
			Save(model).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to update workorder", err)
		}
		return nil
	})
}

func (r *Repository) DeleteWorkorder(ctx context.Context, number string) error {
	// buscar el ID por número
	var m models.Workorder
	if err := r.db.Client().WithContext(ctx).
		Select("id").
		Where("number = ?", number).
		First(&m).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return types.NewError(types.ErrNotFound, "workorder not found", err)
		}
		return types.NewError(types.ErrInternal, "failed to find workorder ID", err)
	}

	// borrar por ID (cascade eliminará items si está configurado)
	if err := r.db.Client().WithContext(ctx).
		Delete(&models.Workorder{}, m.ID).Error; err != nil {
		return types.NewError(types.ErrInternal, "failed to delete workorder", err)
	}
	return nil
}

func (r *Repository) DuplicateWorkorder(ctx context.Context, number string) (string, error) {
	orig, err := r.GetWorkorderByNumber(ctx, number)
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
		return "", types.NewError(types.ErrInternal, "failed to get next workorder number", err)
	}
	return fmt.Sprintf("%04d", seq), nil
}

func (r *Repository) ListWorkorders(
	ctx context.Context,
	filt domain.WorkorderFilter,
	inp types.Input,
) ([]domain.WorkorderListElement, types.PageInfo, error) {
	db := r.db.Client().WithContext(ctx)
	if filt.ProjectID != nil {
		db = db.Where("project_id = ?", *filt.ProjectID)
	}
	if filt.FieldID != nil {
		db = db.Where("field_id = ?", *filt.FieldID)
	}

	var total int64
	if err := db.Model(&models.WorkorderListElement{}).
		Count(&total).Error; err != nil {
		return nil, types.PageInfo{}, types.NewError(types.ErrInternal,
			"failed to count workorders", err)
	}

	offset := (int(inp.Page) - 1) * int(inp.PageSize)
	var rows []models.WorkorderListElement
	if err := db.
		Limit(int(inp.PageSize)).
		Offset(offset).
		Order("number desc").
		Find(&rows).Error; err != nil {
		return nil, types.PageInfo{}, types.NewError(types.ErrInternal,
			"failed to list workorders", err)
	}

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

	pageInfo := types.NewPageInfo(int(inp.Page), int(inp.PageSize), total)
	return list, pageInfo, nil
}
