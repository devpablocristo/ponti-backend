// Package workorder implementa repositorios para work orders.
package workorder

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/devpablocristo/core/errors/go/domainerr"
	shareddb "github.com/devpablocristo/ponti-backend/internal/shared/db"
	sharedfilters "github.com/devpablocristo/ponti-backend/internal/shared/filters"
	sharedmodels "github.com/devpablocristo/ponti-backend/internal/shared/models"
	sharedrepo "github.com/devpablocristo/ponti-backend/internal/shared/repository"
	types "github.com/devpablocristo/ponti-backend/internal/shared/types"
	"github.com/devpablocristo/ponti-backend/internal/work-order/repository/models"
	"github.com/devpablocristo/ponti-backend/internal/work-order/usecases/domain"
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
	if userID, err := sharedmodels.ActorFromContext(ctx); err == nil {
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
				return domainerr.Newf(domainerr.KindConflict,
					"work order already exists for number %s and project %d", o.Number, o.ProjectID,
				)
			}
			return domainerr.Internal("failed to create work order header")
		}

		// 3.2) insertar los items explícitamente asignando WorkOrderID
		if len(model.Items) > 0 {
			for i := range model.Items {
				model.Items[i].WorkOrderID = model.ID
				// Asegurar que la PK sea generada por la DB (serial/sequence).
				model.Items[i].ID = 0
			}
			if err := tx.Create(&model.Items).Error; err != nil {
				return domainerr.Internal("failed to create work order items")
			}
		}

		// 3.3) Insertar splits por inversor (si existen)
		if len(model.InvestorSplits) > 0 {
			for i := range model.InvestorSplits {
				model.InvestorSplits[i].WorkOrderID = model.ID
				model.InvestorSplits[i].ID = 0
				model.InvestorSplits[i].PaymentStatus = normalizeSplitPaymentStatus(
					model.InvestorSplits[i].InvestorID,
					model.InvestorSplits[i].PaymentStatus,
					nil,
				)
			}
			if err := tx.Create(&model.InvestorSplits).Error; err != nil {
				return domainerr.Internal("failed to create work order investor splits")
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
			return nil, domainerr.NotFound("work order not found")
		}
		return nil, domainerr.Internal("failed to get work order")
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
		return nil, domainerr.Internal("failed to get work order")
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
	if userID, err := sharedmodels.ActorFromContext(ctx); err == nil {
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
					return domainerr.Conflict("work order not found or outdated")
				}
				return domainerr.NotFound("work order not found")
			}
			return domainerr.Internal("failed to find work order before update")
		}

		// 3.2) Eliminar todos los items antiguos
		if err := tx.
			Where("workorder_id = ?", model.ID).
			Delete(&models.WorkOrderItem{}).Error; err != nil {
			return domainerr.Internal("failed to delete old items")
		}

		// 3.2b) Eliminar splits antiguos
		if err := tx.
			Where("workorder_id = ?", model.ID).
			Delete(&models.WorkOrderInvestorSplit{}).Error; err != nil {
			return domainerr.Internal("failed to delete old investor splits")
		}

		// 3.3) Actualizar sólo la cabecera, omitiendo campos de auditoría y la asociación Items
		updateTx := tx.Model(&orig).
			Omit("CreatedAt", "CreatedBy", "DeletedAt", "DeletedBy", "Items", "InvestorSplits")
		if !o.Base.UpdatedAt.IsZero() {
			updateTx = updateTx.Where("updated_at = ?", o.Base.UpdatedAt)
		}
		updateTx = updateTx.Updates(model)
		if updateTx.Error != nil {
			return domainerr.Internal("failed to update work order header")
		}
		if updateTx.RowsAffected == 0 {
			return domainerr.Conflict("work order not found or outdated")
		}

		// 3.4) Insertar los items nuevos, asignando WorkOrderID
		for i := range model.Items {
			model.Items[i].WorkOrderID = model.ID
		}
		if len(model.Items) > 0 {
			if err := tx.Create(&model.Items).Error; err != nil {
				return domainerr.Internal("failed to insert new items")
			}
		}

		// 3.5) Insertar splits nuevos
		if len(model.InvestorSplits) > 0 {
			existingStatuses := indexSplitPaymentStatuses(orig.InvestorSplits)
			for i := range model.InvestorSplits {
				model.InvestorSplits[i].WorkOrderID = model.ID
				model.InvestorSplits[i].ID = 0
				model.InvestorSplits[i].PaymentStatus = normalizeSplitPaymentStatus(
					model.InvestorSplits[i].InvestorID,
					model.InvestorSplits[i].PaymentStatus,
					existingStatuses,
				)
			}
			if err := tx.Create(&model.InvestorSplits).Error; err != nil {
				return domainerr.Internal("failed to insert new investor splits")
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
		var wo models.WorkOrder
		if err := tx.Unscoped().Preload("Items").Where("id = ?", id).First(&wo).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.NotFound("work order not found")
			}
			return domainerr.Internal("failed to check work order existence")
		}

		if err := tx.Unscoped().Where("workorder_id = ?", id).Delete(&models.WorkOrderItem{}).Error; err != nil {
			return domainerr.Internal("failed to delete work order items")
		}
		if err := tx.Unscoped().Where("workorder_id = ?", id).Delete(&models.WorkOrderInvestorSplit{}).Error; err != nil {
			return domainerr.Internal("failed to delete work order investor splits")
		}
		if err := tx.Unscoped().Delete(&models.WorkOrder{}, "id = ?", id).Error; err != nil {
			return domainerr.Internal("failed to hard delete work order")
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
		if err := tx.Unscoped().Preload("Items").Where("id = ?", id).First(&wo).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.NotFound("work order not found")
			}
			return domainerr.Internal("failed to get work order")
		}
		if wo.DeletedAt.Valid {
			return domainerr.Conflict("work order already archived")
		}

		if err := tx.Model(&models.WorkOrder{}).
			Where("id = ?", id).
			Updates(map[string]any{
				"deleted_at": time.Now(),
			}).Error; err != nil {
			return domainerr.Internal("failed to archive work order")
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
		if err := tx.Unscoped().Preload("Items").Where("id = ?", id).First(&wo).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.NotFound("work order not found")
			}
			return domainerr.Internal("failed to get work order")
		}
		if !wo.DeletedAt.Valid {
			return domainerr.Conflict("work order is not archived")
		}

		if err := tx.Unscoped().Model(&models.WorkOrder{}).
			Where("id = ?", id).
			Updates(map[string]any{
				"deleted_at": nil,
				"updated_at": time.Now(),
			}).Error; err != nil {
			return domainerr.Internal("failed to restore work order")
		}
		return nil
	})
}

func (r *Repository) UpdateInvestorPaymentStatus(
	ctx context.Context,
	workOrderID int64,
	investorID int64,
	paymentStatus string,
) error {
	if err := sharedrepo.ValidateID(workOrderID, "work order"); err != nil {
		return err
	}
	if err := sharedrepo.ValidateID(investorID, "investor"); err != nil {
		return err
	}

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var workOrder models.WorkOrder
		if err := tx.Select("id").Where("id = ?", workOrderID).First(&workOrder).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.NotFound("work order not found")
			}
			return domainerr.Internal("failed to find work order")
		}

		updateTx := tx.Model(&models.WorkOrderInvestorSplit{}).
			Where("workorder_id = ? AND investor_id = ? AND deleted_at IS NULL", workOrderID, investorID).
			Update("payment_status", paymentStatus)
		if updateTx.Error != nil {
			return domainerr.Internal("failed to update investor payment status")
		}
		if updateTx.RowsAffected == 0 {
			return domainerr.NotFound("investor split not found")
		}

		return nil
	})
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func indexSplitPaymentStatuses(
	splits []models.WorkOrderInvestorSplit,
) map[int64]string {
	statuses := make(map[int64]string, len(splits))
	for _, split := range splits {
		if split.InvestorID <= 0 {
			continue
		}
		statuses[split.InvestorID] = split.PaymentStatus
	}
	return statuses
}

func normalizeSplitPaymentStatus(
	investorID int64,
	status string,
	existingStatuses map[int64]string,
) string {
	if status != "" {
		return status
	}
	if existingStatuses != nil {
		if existingStatus, ok := existingStatuses[investorID]; ok && existingStatus != "" {
			return existingStatus
		}
	}
	return domain.InvestorPaymentStatusPending
}

func (r *Repository) ListWorkOrders(
	ctx context.Context,
	filt domain.WorkOrderFilter,
	inp types.Input,
) ([]domain.WorkOrderListElement, types.PageInfo, error) {
	base, empty, err := r.workOrderListBaseQuery(ctx, filt)
	if err != nil {
		return nil, types.PageInfo{}, err
	}
	if empty {
		return []domain.WorkOrderListElement{}, types.NewPageInfo(int(inp.Page), int(inp.PageSize), 0), nil
	}

	var total int64
	if err := base.
		Count(&total).Error; err != nil {
		return nil, types.PageInfo{}, domainerr.Internal(
			"failed to count work orders")
	}

	offset := (int(inp.Page) - 1) * int(inp.PageSize)

	var rows []models.WorkOrderListElement
	if err := base.
		Limit(int(inp.PageSize)).
		Offset(offset).
		Order("date desc, sequence_day desc, id desc").
		Find(&rows).Error; err != nil {
		return nil, types.PageInfo{}, domainerr.Internal(
			"failed to list work orders")
	}

	pageInfo := types.NewPageInfo(int(inp.Page), int(inp.PageSize), total)
	return mapWorkOrderListRows(rows), pageInfo, nil
}

func (r *Repository) ListWorkOrderFilterRows(
	ctx context.Context,
	filt domain.WorkOrderFilter,
) ([]domain.WorkOrderListElement, error) {
	base, empty, err := r.workOrderListBaseQuery(ctx, filt)
	if err != nil {
		return nil, err
	}
	if empty {
		return []domain.WorkOrderListElement{}, nil
	}

	var rows []models.WorkOrderListElement
	if err := base.
		Order("date desc, sequence_day desc, id desc").
		Find(&rows).Error; err != nil {
		return nil, domainerr.Internal("failed to list work order filter rows")
	}

	return mapWorkOrderListRows(rows), nil
}

func (r *Repository) workOrderListBaseQuery(
	ctx context.Context,
	filt domain.WorkOrderFilter,
) (*gorm.DB, bool, error) {
	base := r.db.Client().
		WithContext(ctx).
		Model(&models.WorkOrderListElement{})

	projectIDs, err := sharedfilters.ResolveProjectIDs(ctx, r.db.Client(), sharedfilters.WorkspaceFilter{
		CustomerID: filt.CustomerID,
		ProjectID:  filt.ProjectID,
		CampaignID: filt.CampaignID,
		FieldID:    filt.FieldID,
	})
	if err != nil {
		return nil, false, err
	}
	if len(projectIDs) > 0 {
		base = base.Where("project_id IN ?", projectIDs)
	} else if filt.ProjectID != nil || filt.CustomerID != nil || filt.CampaignID != nil || filt.FieldID != nil {
		return base, true, nil
	}

	if filt.FieldID != nil {
		base = base.Where("field_id = ?", *filt.FieldID)
	}
	if filt.IsDigital != nil {
		base = base.Where("is_digital = ?", *filt.IsDigital)
	}
	if filt.Status != nil {
		base = base.Where("status = ?", *filt.Status)
	}
	if filt.SupplyID != nil {
		base = base.Where(`
			(
				(
					v4_report.workorder_list.is_digital = false
					AND EXISTS (
						SELECT 1
						FROM workorder_items wi
						WHERE wi.workorder_id = v4_report.workorder_list.id
						  AND wi.supply_id = ?
						  AND wi.deleted_at IS NULL
					)
				)
				OR (
					v4_report.workorder_list.is_digital = true
					AND EXISTS (
						SELECT 1
						FROM work_order_draft_items wodi
						WHERE wodi.draft_id = -v4_report.workorder_list.id
						  AND wodi.supply_id = ?
						  AND wodi.deleted_at IS NULL
					)
				)
			)
			AND v4_report.workorder_list.supply_name = (
				SELECT s.name
				FROM supplies s
				WHERE s.id = ?
				  AND s.deleted_at IS NULL
			)
		`, *filt.SupplyID, *filt.SupplyID, *filt.SupplyID)
	}

	return base, false, nil
}

func mapWorkOrderListRows(rows []models.WorkOrderListElement) []domain.WorkOrderListElement {
	list := make([]domain.WorkOrderListElement, len(rows))
	for i, m := range rows {
		list[i] = domain.WorkOrderListElement{
			ID:                m.ID,
			Number:            m.Number,
			ProjectName:       m.ProjectName,
			FieldName:         m.FieldName,
			LotName:           m.LotName,
			Date:              m.Date,
			SequenceDay:       m.SequenceDay,
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
			IsDigital:         m.IsDigital,
			Status:            m.Status,
		}
	}
	return list
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
			SurfaceHa:   decimal.Zero,
			Liters:      decimal.Zero,
			Kilograms:   decimal.Zero,
			DirectCost:  decimal.Zero,
			OrdersCount: 0,
		}, nil
	}

	if filt.SupplyID != nil {
		return r.getSupplyFilteredMetrics(ctx, filt, projectIDs)
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
		return nil, domainerr.Internal("failed to get metrics")
	}

	orderCountQuery := r.db.Client().
		WithContext(ctx).
		Table("workorders").
		Where("deleted_at IS NULL")
	if len(projectIDs) > 0 {
		orderCountQuery = orderCountQuery.Where("project_id IN ?", projectIDs)
	}
	if filt.FieldID != nil {
		orderCountQuery = orderCountQuery.Where("field_id = ?", *filt.FieldID)
	}

	var ordersCount int64
	if err := orderCountQuery.
		Select("COUNT(DISTINCT split_part(number::text, '.', 1))").
		Scan(&ordersCount).Error; err != nil {
		return nil, domainerr.Internal("failed to count work orders")
	}

	return &domain.WorkOrderMetrics{
		SurfaceHa:   row.SurfaceHa,
		Liters:      row.Liters,
		Kilograms:   row.Kilograms,
		DirectCost:  row.DirectCost,
		OrdersCount: ordersCount,
	}, nil
}

func (r *Repository) getSupplyFilteredMetrics(
	ctx context.Context,
	filt domain.WorkOrderFilter,
	projectIDs []int64,
) (*domain.WorkOrderMetrics, error) {
	q := `
		SELECT
			COALESCE(SUM(COALESCE(wo.effective_area, 0)), 0) AS surface_ha,
			COALESCE(SUM(CASE WHEN s.unit_id = 1 THEN COALESCE(wi.total_used, 0) ELSE 0 END), 0) AS liters,
			COALESCE(SUM(CASE WHEN s.unit_id = 2 THEN COALESCE(wi.total_used, 0) ELSE 0 END), 0) AS kilograms,
			COALESCE(SUM(COALESCE(wi.total_used, 0) * COALESCE(s.price, 0)), 0) AS direct_cost,
			COUNT(DISTINCT split_part(wo.number::text, '.', 1)) AS orders_count
		FROM workorders wo
		JOIN workorder_items wi ON wi.workorder_id = wo.id AND wi.deleted_at IS NULL
		JOIN supplies s ON s.id = wi.supply_id AND s.deleted_at IS NULL
		WHERE wo.deleted_at IS NULL
		  AND wi.supply_id = ?
	`
	args := []any{*filt.SupplyID}

	if len(projectIDs) > 0 {
		q += " AND wo.project_id IN ?"
		args = append(args, projectIDs)
	}
	if filt.FieldID != nil {
		q += " AND wo.field_id = ?"
		args = append(args, *filt.FieldID)
	}

	var row struct {
		SurfaceHa   decimal.Decimal `gorm:"column:surface_ha"`
		Liters      decimal.Decimal `gorm:"column:liters"`
		Kilograms   decimal.Decimal `gorm:"column:kilograms"`
		DirectCost  decimal.Decimal `gorm:"column:direct_cost"`
		OrdersCount int64           `gorm:"column:orders_count"`
	}

	if err := r.db.Client().WithContext(ctx).Raw(q, args...).Scan(&row).Error; err != nil {
		return nil, domainerr.Internal("failed to get supply filtered metrics")
	}

	return &domain.WorkOrderMetrics{
		SurfaceHa:   row.SurfaceHa,
		Liters:      row.Liters,
		Kilograms:   row.Kilograms,
		DirectCost:  row.DirectCost,
		OrdersCount: row.OrdersCount,
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
		return decimal.Zero, domainerr.Internal("failed to get raw direct cost")
	}

	return totalCost, nil
}
