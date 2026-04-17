// Package workorder implementa repositorios para work orders.
package workorder

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgconn"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	shareddb "github.com/alphacodinggroup/ponti-backend/internal/shared/db"
	sharedfilters "github.com/alphacodinggroup/ponti-backend/internal/shared/filters"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/internal/shared/models"
	sharedrepo "github.com/alphacodinggroup/ponti-backend/internal/shared/repository"
	"github.com/alphacodinggroup/ponti-backend/internal/work-order/repository/models"
	"github.com/alphacodinggroup/ponti-backend/internal/work-order/usecases/domain"
	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
)

type GormEngine interface {
	Client() *gorm.DB
}

type Repository struct {
	db GormEngine
}

// NewRepository crea una instancia de repositorio de work orders.
func NewRepository(db GormEngine) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateWorkOrder(ctx context.Context, o *domain.WorkOrder) (int64, error) {
	// 1) convertir a modelo GORM (cabecera + items sin WorkOrderID)
	model := models.FromDomain(o)

	// 2) poblar auditoría
	if userID, err := sharedmodels.ConvertStringToID(ctx); err == nil {
		model.CreatedBy = &userID
		model.UpdatedBy = &userID
	}

	// 3) crear todo en una transacción
	err := r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 3.1) insertar la cabecera para obtener model.ID
		// Importante: evitamos que GORM intente crear también las asociaciones (Items) acá,
		// porque abajo insertamos los items explícitamente. Si se insertan dos veces puede
		// terminar en violación de PK (duplicate key) como "pk_workorder_items".
		if err := tx.Omit("Items", "InvestorSplits").Create(&model).Error; err != nil {
			if isUniqueViolation(err) {
				return types.NewError(
					types.ErrConflict,
					fmt.Sprintf("work order already exists for number %s and project %d", o.Number, o.ProjectID),
					err,
				)
			}
			return types.NewError(types.ErrInternal, "failed to create work order header", err)
		}

		// 3.2) insertar los items explícitamente asignando WorkOrderID
		if len(model.Items) > 0 {
			for i := range model.Items {
				model.Items[i].WorkOrderID = model.ID
				// Asegurar que la PK sea generada por la DB (serial/sequence).
				model.Items[i].ID = 0
			}
			if err := tx.Create(&model.Items).Error; err != nil {
				return types.NewError(types.ErrInternal, "failed to create work order items", err)
			}
		}

		// 3.3) Insertar splits por inversor (si existen)
		if len(model.InvestorSplits) > 0 {
			for i := range model.InvestorSplits {
				model.InvestorSplits[i].WorkOrderID = model.ID
				model.InvestorSplits[i].ID = 0
			}
			if err := tx.Create(&model.InvestorSplits).Error; err != nil {
				return types.NewError(types.ErrInternal, "failed to create work order investor splits", err)
			}
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return model.ID, nil
}

func (r *Repository) GetWorkOrderByID(ctx context.Context, id int64) (*domain.WorkOrder, error) {
	var m models.WorkOrder
	if err := r.db.Client().WithContext(ctx).
		Preload("Items").
		Preload("InvestorSplits").
		Where("id = ?", id).
		First(&m).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.NewError(types.ErrNotFound, "work order not found", err)
		}
		return nil, types.NewError(types.ErrInternal, "failed to get work order", err)
	}
	return m.ToDomain(), nil
}

func (r *Repository) GetWorkOrderByNumberAndProjectID(ctx context.Context, number string, projectID int64) (*domain.WorkOrder, error) {
	var m models.WorkOrder
	if err := r.db.Client().WithContext(ctx).
		Where("number = ?", number).
		Where("project_id = ?", projectID).
		First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, types.NewError(types.ErrInternal, "failed to get work order", err)
	}
	return m.ToDomain(), nil
}

func (r *Repository) UpdateWorkOrderByID(ctx context.Context, o *domain.WorkOrder) error {
	if err := sharedrepo.ValidateEntity(o, "work order"); err != nil {
		return err
	}
	if err := sharedrepo.ValidateID(o.ID, "work order"); err != nil {
		return err
	}
	// 1) Convertimos dominio → GORM y fijamos el ID
	model := models.FromDomain(o)
	model.ID = o.ID

	// 2) Poblar UpdatedBy si hay usuario en contexto
	if userID, err := sharedmodels.ConvertStringToID(ctx); err == nil {
		model.UpdatedBy = &userID
	}

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 3.1) Recuperar original para validar existencia y conservar auditoría
		var orig models.WorkOrder
		query := tx.Preload("Items").Preload("InvestorSplits").Where("id = ?", model.ID)
		if !o.Base.UpdatedAt.IsZero() {
			query = query.Where("updated_at = ?", o.Base.UpdatedAt)
		}
		if err := query.First(&orig).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if !o.Base.UpdatedAt.IsZero() {
					return types.NewError(types.ErrConflict, "work order not found or outdated", err)
				}
				return types.NewError(types.ErrNotFound, "work order not found", err)
			}
			return types.NewError(types.ErrInternal, "failed to find work order before update", err)
		}

		// 3.2) Eliminar todos los items antiguos
		if err := tx.
			Where("workorder_id = ?", model.ID).
			Delete(&models.WorkOrderItem{}).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to delete old items", err)
		}

		// 3.2b) Eliminar splits antiguos
		if err := tx.
			Where("workorder_id = ?", model.ID).
			Delete(&models.WorkOrderInvestorSplit{}).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to delete old investor splits", err)
		}

		// 3.3) Actualizar sólo la cabecera, omitiendo campos de auditoría y la asociación Items
		updateTx := tx.Model(&orig).
			Omit("CreatedAt", "CreatedBy", "DeletedAt", "DeletedBy", "Items", "InvestorSplits")
		if !o.Base.UpdatedAt.IsZero() {
			updateTx = updateTx.Where("updated_at = ?", o.Base.UpdatedAt)
		}
		updateTx = updateTx.Updates(model)
		if updateTx.Error != nil {
			return types.NewError(types.ErrInternal, "failed to update work order header", updateTx.Error)
		}
		if updateTx.RowsAffected == 0 {
			return types.NewError(types.ErrConflict, "work order not found or outdated", nil)
		}

		// 3.4) Insertar los items nuevos, asignando WorkOrderID
		for i := range model.Items {
			model.Items[i].WorkOrderID = model.ID
		}
		if len(model.Items) > 0 {
			if err := tx.Create(&model.Items).Error; err != nil {
				return types.NewError(types.ErrInternal, "failed to insert new items", err)
			}
		}

		// 3.5) Insertar splits nuevos
		if len(model.InvestorSplits) > 0 {
			for i := range model.InvestorSplits {
				model.InvestorSplits[i].WorkOrderID = model.ID
				model.InvestorSplits[i].ID = 0
			}
			if err := tx.Create(&model.InvestorSplits).Error; err != nil {
				return types.NewError(types.ErrInternal, "failed to insert new investor splits", err)
			}
		}

		return nil
	})
}

func (r *Repository) DeleteWorkOrderByID(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "work order"); err != nil {
		return err
	}
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Unscoped().Model(&models.WorkOrder{}).Where("id = ?", id).Count(&count).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to check work order existence", err)
		}
		if count == 0 {
			return types.NewError(types.ErrNotFound, "work order not found", nil)
		}

		if err := tx.Unscoped().Where("workorder_id = ?", id).Delete(&models.WorkOrderItem{}).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to delete work order items", err)
		}
		if err := tx.Unscoped().Where("workorder_id = ?", id).Delete(&models.WorkOrderInvestorSplit{}).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to delete work order investor splits", err)
		}
		if err := tx.Unscoped().Delete(&models.WorkOrder{}, "id = ?", id).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to hard delete work order", err)
		}
		return nil
	})
}

func (r *Repository) ArchiveWorkOrder(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "work order"); err != nil {
		return err
	}
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var wo models.WorkOrder
		if err := tx.Unscoped().Where("id = ?", id).First(&wo).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return types.NewError(types.ErrNotFound, "work order not found", err)
			}
			return types.NewError(types.ErrInternal, "failed to get work order", err)
		}
		if wo.DeletedAt.Valid {
			return types.NewError(types.ErrConflict, "work order already archived", nil)
		}

		if err := tx.Model(&models.WorkOrder{}).
			Where("id = ?", id).
			Updates(map[string]any{
				"deleted_at": time.Now(),
			}).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to archive work order", err)
		}
		return nil
	})
}

func (r *Repository) RestoreWorkOrder(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "work order"); err != nil {
		return err
	}
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var wo models.WorkOrder
		if err := tx.Unscoped().Where("id = ?", id).First(&wo).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return types.NewError(types.ErrNotFound, "work order not found", err)
			}
			return types.NewError(types.ErrInternal, "failed to get work order", err)
		}
		if !wo.DeletedAt.Valid {
			return types.NewError(types.ErrConflict, "work order is not archived", nil)
		}

		if err := tx.Unscoped().Model(&models.WorkOrder{}).
			Where("id = ?", id).
			Updates(map[string]any{
				"deleted_at": nil,
				"updated_at": time.Now(),
			}).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to restore work order", err)
		}
		return nil
	})
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func (r *Repository) ListWorkOrders(
	ctx context.Context,
	filt domain.WorkOrderFilter,
	inp types.Input,
) ([]domain.WorkOrderListElement, types.PageInfo, error) {
	// 1) Base del query: vinculada a la vista
	base := r.db.Client().
		WithContext(ctx).
		Model(&models.WorkOrderListElement{})

	// 2) Resolver filtros de proyecto (customer/campaign/field)
	projectIDs, err := sharedfilters.ResolveProjectIDs(ctx, r.db.Client(), sharedfilters.WorkspaceFilter{
		CustomerID: filt.CustomerID,
		ProjectID:  filt.ProjectID,
		CampaignID: filt.CampaignID,
		FieldID:    filt.FieldID,
	})
	if err != nil {
		return nil, types.PageInfo{}, err
	}
	if len(projectIDs) > 0 {
		base = base.Where("project_id IN ?", projectIDs)
	} else if filt.ProjectID != nil || filt.CustomerID != nil || filt.CampaignID != nil || filt.FieldID != nil {
		// filtros presentes pero sin resultados: devolver vacío
		return []domain.WorkOrderListElement{}, types.NewPageInfo(int(inp.Page), int(inp.PageSize), 0), nil
	}

	// 3) Aplicar filtros directos
	if filt.FieldID != nil {
		base = base.Where("field_id = ?", *filt.FieldID)
	}

	// 4) Contar total
	var total int64
	if err := base.
		Count(&total).Error; err != nil {
		return nil, types.PageInfo{}, types.NewError(types.ErrInternal,
			"failed to count work orders", err)
	}

	// 5) Paginación
	offset := (int(inp.Page) - 1) * int(inp.PageSize)

	// 6) Recuperar filas paginadas (reutiliza 'base' con filtros)
	var rows []models.WorkOrderListElement
	if err := base.
		Limit(int(inp.PageSize)).
		Offset(offset).
		Order("number desc").
		Find(&rows).Error; err != nil {
		return nil, types.PageInfo{}, types.NewError(types.ErrInternal,
			"failed to list work orders", err)
	}

	// 7) Mapear a dominio
	list := make([]domain.WorkOrderListElement, len(rows))
	for i, m := range rows {
		list[i] = domain.WorkOrderListElement{
			ID:                m.ID,
			Number:            m.Number,
			ProjectName:       m.ProjectName,
			FieldName:         m.FieldName,
			LotName:           m.LotName,
			Date:              m.Date,
			CropName:          m.CropName,
			LaborName:         m.LaborName,
			LaborCategoryName: m.LaborCategoryName,
			TypeName:          m.TypeName,
			Contractor:        m.Contractor,
			SurfaceHa:         m.SurfaceHa,
			SupplyName:        m.SupplyName,
			Consumption:       m.Consumption,
			CategoryName:      m.CategoryName,
			Dose:              m.Dose,
			CostPerHa:         m.CostPerHa,
			UnitPrice:         m.UnitPrice,
			TotalCost:         m.TotalCost,
		}
	}

	// 8) Construir PageInfo y devolver
	pageInfo := types.NewPageInfo(int(inp.Page), int(inp.PageSize), total)
	return list, pageInfo, nil
}

func (r *Repository) GetMetrics(ctx context.Context, filt domain.WorkOrderFilter) (*domain.WorkOrderMetrics, error) {
	projectIDs, err := sharedfilters.ResolveProjectIDs(ctx, r.db.Client(), sharedfilters.WorkspaceFilter{
		CustomerID: filt.CustomerID,
		ProjectID:  filt.ProjectID,
		CampaignID: filt.CampaignID,
		FieldID:    filt.FieldID,
	})
	if err != nil {
		return nil, err
	}
	if len(projectIDs) == 0 && (filt.ProjectID != nil || filt.CustomerID != nil || filt.CampaignID != nil || filt.FieldID != nil) {
		return &domain.WorkOrderMetrics{
			SurfaceHa:  decimal.Zero,
			Liters:     decimal.Zero,
			Kilograms:  decimal.Zero,
			DirectCost: decimal.Zero,
		}, nil
	}

	// Construimos el WHERE dinámico según los filtros presentes
	q := fmt.Sprintf(`
		SELECT
		  COALESCE(SUM(surface_ha), 0) AS surface_ha,
		  COALESCE(SUM(liters), 0) AS liters,
		  COALESCE(SUM(kilograms), 0) AS kilograms,
		  COALESCE(SUM(direct_cost_usd), 0) AS direct_cost
		FROM %s
		WHERE 1=1
	`, shareddb.ReportView("workorder_metrics"))
	var args []any

	if len(projectIDs) > 0 {
		q += " AND project_id IN ?"
		args = append(args, projectIDs)
	}
	if filt.FieldID != nil {
		q += " AND field_id = ?"
		args = append(args, *filt.FieldID)
	}

	var row struct {
		SurfaceHa  decimal.Decimal `gorm:"column:surface_ha"`
		Liters     decimal.Decimal `gorm:"column:liters"`
		Kilograms  decimal.Decimal `gorm:"column:kilograms"`
		DirectCost decimal.Decimal `gorm:"column:direct_cost"`
	}

	if err := r.db.Client().WithContext(ctx).Raw(q, args...).Scan(&row).Error; err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to get metrics", err)
	}

	return &domain.WorkOrderMetrics{
		SurfaceHa:  row.SurfaceHa,
		Liters:     row.Liters,
		Kilograms:  row.Kilograms,
		DirectCost: row.DirectCost,
	}, nil
}

// GetRawDirectCost calcula el costo directo RAW desde las tablas workorders y workorder_items
// Calcula ∑(Órdenes_de_trabajo.costo_total) como indica el CSV de controles
// Este cálculo es INDEPENDIENTE de las vistas SSOT para validar coherencia
func (r *Repository) GetRawDirectCost(ctx context.Context, projectID int64) (decimal.Decimal, error) {
	// Query RAW: suma directa desde workorders + workorder_items
	// Labor cost: effective_area × labor.price
	// Supply cost: total_used × price (consistente con v4_calc.workorder_metrics).
	//
	// Importante:
	// - Respetar soft-delete (deleted_at) como en vistas/reportes.
	// - Usar COALESCE para no "perder" items con price NULL.
	whereProject := ""
	args := []any{}
	if projectID > 0 {
		whereProject = "AND wo.project_id = ?"
		args = append(args, projectID)
	}

	q := fmt.Sprintf(`
		WITH workorder_costs AS (
		  SELECT 
		    wo.id,
		    -- Costo de la labor (área efectiva × precio de la labor)
		    (COALESCE(wo.effective_area, 0) * COALESCE(l.price, 0)) AS labor_cost,
		    -- Costo de insumos (suma de items: total_used × price)
		    COALESCE((
		      SELECT SUM(COALESCE(wi.total_used, 0) * COALESCE(s.price, 0))
		      FROM public.workorder_items wi
		      JOIN public.supplies s ON s.id = wi.supply_id AND s.deleted_at IS NULL
		      WHERE wi.workorder_id = wo.id 
		        AND wi.deleted_at IS NULL
		    ), 0) AS supply_cost
		  FROM public.workorders wo
		  JOIN public.labors l ON l.id = wo.labor_id AND l.deleted_at IS NULL
		  WHERE wo.deleted_at IS NULL
		    AND wo.effective_area IS NOT NULL
		    AND wo.effective_area > 0
		    %s
		)
		SELECT COALESCE(SUM(labor_cost + supply_cost), 0) AS total_cost
		FROM workorder_costs
	`, whereProject)

	var totalCost decimal.Decimal
	if err := r.db.Client().WithContext(ctx).Raw(q, args...).Scan(&totalCost).Error; err != nil {
		return decimal.Zero, types.NewError(types.ErrInternal, "failed to get raw direct cost", err)
	}

	return totalCost, nil
}
