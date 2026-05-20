// Package workorder implementa repositorios para work orders.
package workorder

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/devpablocristo/core/errors/go/domainerr"
	actorsync "github.com/devpablocristo/ponti-backend/internal/actor"
	"github.com/devpablocristo/ponti-backend/internal/shared/authz"
	shareddb "github.com/devpablocristo/ponti-backend/internal/shared/db"
	sharedfilters "github.com/devpablocristo/ponti-backend/internal/shared/filters"
	"github.com/devpablocristo/ponti-backend/internal/shared/lifecycle"
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
	tenantID, hasTenant, err := authz.OptionalTenantOrStrict(ctx)
	if err != nil {
		return 0, err
	}
	if hasTenant {
		model.TenantID = tenantID
	}
	if userID, err := sharedmodels.ActorFromContext(ctx); err == nil {
		model.CreatedBy = &userID
		model.UpdatedBy = &userID
	}

	// 3) crear todo en una transacción
	err = r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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
		if model.Contractor != "" {
			if _, err := actorsync.SyncLegacyTextActor(tx, actorsync.LegacyTextActorSync{
				SourceTable: actorsync.LegacyWorkOrderContractor,
				Name:        model.Contractor,
				ActorKind:   actorsync.KindUnknown,
				Role:        actorsync.RoleContratista,
				CreatedAt:   model.CreatedAt,
				UpdatedAt:   model.UpdatedAt,
				CreatedBy:   model.CreatedBy,
				UpdatedBy:   model.UpdatedBy,
			}); err != nil {
				return err
			}
		}

		// 3.2) insertar los items explícitamente asignando WorkOrderID
		if len(model.Items) > 0 {
			for i := range model.Items {
				model.Items[i].WorkOrderID = model.ID
				if hasTenant {
					model.Items[i].TenantID = tenantID
				}
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
				if hasTenant {
					model.InvestorSplits[i].TenantID = tenantID
				}
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
		if err := actorsync.RefreshWorkOrderActorColumns(tx, model.ID); err != nil {
			return err
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
	db := authz.MaybeTenantScope(ctx, r.db.Client().WithContext(ctx), "workorders")
	if err := db.
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Order("id ASC")
		}).
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
	db := authz.MaybeTenantScope(ctx, r.db.Client().WithContext(ctx), "workorders")
	if err := db.
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
	tenantID, hasTenant := authz.TenantFromContext(ctx)
	if hasTenant {
		model.TenantID = tenantID
	}

	// 2) Poblar UpdatedBy si hay usuario en contexto
	if userID, err := sharedmodels.ActorFromContext(ctx); err == nil {
		model.UpdatedBy = &userID
	}

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 3.1) Recuperar original para validar existencia y conservar auditoría
		var orig models.WorkOrder
		query := authz.MaybeTenantScope(ctx, tx.Preload("Items").Preload("InvestorSplits"), "workorders").Where("id = ?", model.ID)
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
			Scopes(func(db *gorm.DB) *gorm.DB { return authz.MaybeTenantScope(ctx, db, "workorder_items") }).
			Where("workorder_id = ?", model.ID).
			Delete(&models.WorkOrderItem{}).Error; err != nil {
			return domainerr.Internal("failed to delete old items")
		}

		// 3.2b) Eliminar splits antiguos
		if err := tx.
			Scopes(func(db *gorm.DB) *gorm.DB { return authz.MaybeTenantScope(ctx, db, "workorder_investor_splits") }).
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
		if model.Contractor != "" {
			if _, err := actorsync.SyncLegacyTextActor(tx, actorsync.LegacyTextActorSync{
				SourceTable: actorsync.LegacyWorkOrderContractor,
				Name:        model.Contractor,
				ActorKind:   actorsync.KindUnknown,
				Role:        actorsync.RoleContratista,
				UpdatedAt:   time.Now(),
				UpdatedBy:   model.UpdatedBy,
			}); err != nil {
				return err
			}
		}

		// 3.4) Insertar los items nuevos, asignando WorkOrderID
		for i := range model.Items {
			model.Items[i].WorkOrderID = model.ID
			if hasTenant {
				model.Items[i].TenantID = tenantID
			}
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
				if hasTenant {
					model.InvestorSplits[i].TenantID = tenantID
				}
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
		if err := actorsync.RefreshWorkOrderActorColumns(tx, model.ID); err != nil {
			return err
		}

		return nil
	})
}

// HardDeleteWorkOrder elimina definitivamente una orden de trabajo.
// Bloquea con 409 si tiene invoices referenciándola.
// Sus children "propios" (items, investor_splits) se eliminan en cascada.
// Nota: la relación con labors es inversa (workorders.labor_id → labors), por lo que
// borrar una work-order no impacta a labors. No se chequea ese sentido.
func (r *Repository) HardDeleteWorkOrder(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "work order"); err != nil {
		return err
	}
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		workOrderDB := authz.MaybeTenantScope(ctx, tx.Unscoped().Table("workorders"), "workorders")
		if err := workOrderDB.Where("id = ?", id).Count(&count).Error; err != nil {
			return domainerr.Internal("failed to check work order existence")
		}
		if count == 0 {
			return domainerr.NotFound("work order not found")
		}
		if err := lifecycle.RequireArchived(workOrderDB, "workorders", "work order", id); err != nil {
			return err
		}

		var invCount int64
		invoiceDB := authz.MaybeTenantScope(ctx, tx.Unscoped().Table("invoices"), "invoices")
		if err := invoiceDB.Where("work_order_id = ?", id).Count(&invCount).Error; err != nil {
			return domainerr.Internal("failed to check invoices")
		}
		if invCount > 0 {
			return domainerr.Conflict(fmt.Sprintf("work order has %d invoice(s); archive or hard-delete them first", invCount))
		}

		// Cascada de "owned children" (no son entidades de negocio independientes).
		if err := authz.MaybeTenantScope(ctx, tx.Unscoped(), "workorder_items").Where("workorder_id = ?", id).Delete(&models.WorkOrderItem{}).Error; err != nil {
			return domainerr.Internal("failed to delete work order items")
		}
		if err := authz.MaybeTenantScope(ctx, tx.Unscoped(), "workorder_investor_splits").Where("workorder_id = ?", id).Delete(&models.WorkOrderInvestorSplit{}).Error; err != nil {
			return domainerr.Internal("failed to delete work order investor splits")
		}
		if err := authz.MaybeTenantScope(ctx, tx.Unscoped(), "workorders").Delete(&models.WorkOrder{}, "id = ?", id).Error; err != nil {
			return domainerr.Internal("failed to hard delete work order")
		}
		return nil
	})
}


// ListArchivedWorkOrders lista ordenes archivadas con nombres joineados (project, field, lot, labor).
// Si filter.LotID > 0, filtra solo las WOs de ese lote.
func (r *Repository) ListArchivedWorkOrders(ctx context.Context, page, perPage int, filter domain.ArchivedWorkOrderFilter) ([]domain.WorkOrderListElement, int64, error) {
	where := []string{"w.deleted_at IS NOT NULL"}
	args := []any{}
	if tenantID, ok := authz.TenantFromContext(ctx); ok {
		where = append(where, "w.tenant_id = ?")
		args = append(args, tenantID)
	}
	if filter.LotID > 0 {
		where = append(where, "w.lot_id = ?")
		args = append(args, filter.LotID)
	}
	whereSQL := strings.Join(where, " AND ")

	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM workorders w WHERE %s", whereSQL)
	if err := r.db.Client().WithContext(ctx).Raw(countQuery, args...).Scan(&total).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to count archived work orders")
	}

	type archivedRow struct {
		ID          int64     `gorm:"column:id"`
		Number      string    `gorm:"column:number"`
		ProjectName string    `gorm:"column:project_name"`
		FieldName   string    `gorm:"column:field_name"`
		LotName     string    `gorm:"column:lot_name"`
		Date        time.Time `gorm:"column:date"`
		SequenceDay int64     `gorm:"column:sequence_day"`
		CropName    string    `gorm:"column:crop_name"`
		LaborName   string    `gorm:"column:labor_name"`
		Contractor  string    `gorm:"column:contractor"`
	}

	offset := (page - 1) * perPage
	listQuery := fmt.Sprintf(`
		SELECT
			w.id, w.number, w.date, w.sequence_day, w.contractor,
			COALESCE(p.name, '') AS project_name,
			COALESCE(f.name, '') AS field_name,
			COALESCE(l.name, '') AS lot_name,
			COALESCE(c.name, '') AS crop_name,
			COALESCE(lb.name, '') AS labor_name
		FROM workorders w
		LEFT JOIN projects p ON p.id = w.project_id
		LEFT JOIN fields f ON f.id = w.field_id
		LEFT JOIN lots l ON l.id = w.lot_id
		LEFT JOIN crops c ON c.id = w.crop_id
		LEFT JOIN labors lb ON lb.id = w.labor_id
		WHERE %s
		ORDER BY w.deleted_at DESC
		LIMIT ? OFFSET ?
	`, whereSQL)
	listArgs := append(append([]any{}, args...), perPage, offset)

	var rows []archivedRow
	if err := r.db.Client().WithContext(ctx).Raw(listQuery, listArgs...).Scan(&rows).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to list archived work orders")
	}

	result := make([]domain.WorkOrderListElement, len(rows))
	for i, row := range rows {
		result[i] = domain.WorkOrderListElement{
			ID:          row.ID,
			Number:      row.Number,
			ProjectName: row.ProjectName,
			FieldName:   row.FieldName,
			LotName:     row.LotName,
			Date:        row.Date,
			SequenceDay: row.SequenceDay,
			CropName:    row.CropName,
			LaborName:   row.LaborName,
			Contractor:  row.Contractor,
		}
	}
	return result, total, nil
}

func (r *Repository) ArchiveWorkOrder(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "work order"); err != nil {
		return err
	}
	actor, err := sharedmodels.ActorFromContext(ctx)
	if err != nil {
		return err
	}
	deletedBy := &actor

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		archivedAt := time.Now()
		var wo models.WorkOrder
		workOrderDB := authz.MaybeTenantScope(ctx, tx.Unscoped(), "workorders")
		if err := workOrderDB.Where("id = ?", id).First(&wo).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.NotFound("work order not found")
			}
			return domainerr.Internal("failed to get work order")
		}
		if wo.DeletedAt.Valid {
			return domainerr.Conflict("work order already archived")
		}

		cause, err := lifecycle.RootCause(tx, wo.TenantID, "workorders", id, nil, deletedBy)
		if err != nil {
			return err
		}
		if err := authz.MaybeTenantScope(ctx, tx.Table("workorder_items"), "workorder_items").
			Where("workorder_id = ? AND deleted_at IS NULL", id).
			Updates(lifecycle.ArchiveUpdates(tx, "workorder_items", archivedAt, deletedBy, cause)).Error; err != nil {
			return domainerr.Internal("failed to archive work order items")
		}
		if err := authz.MaybeTenantScope(ctx, tx.Table("workorder_investor_splits"), "workorder_investor_splits").
			Where("workorder_id = ? AND deleted_at IS NULL", id).
			Updates(lifecycle.ArchiveUpdates(tx, "workorder_investor_splits", archivedAt, deletedBy, cause)).Error; err != nil {
			return domainerr.Internal("failed to archive work order investor splits")
		}
		if err := authz.MaybeTenantScope(ctx, tx.Model(&models.WorkOrder{}), "workorders").
			Where("id = ?", id).
			Updates(lifecycle.ArchiveUpdates(tx, "workorders", archivedAt, deletedBy, cause)).Error; err != nil {
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
		restoredAt := time.Now()
		var wo models.WorkOrder
		workOrderDB := authz.MaybeTenantScope(ctx, tx.Unscoped(), "workorders")
		if err := workOrderDB.Where("id = ?", id).First(&wo).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.NotFound("work order not found")
			}
			return domainerr.Internal("failed to get work order")
		}
		if !wo.DeletedAt.Valid {
			return domainerr.Conflict("work order is not archived")
		}
		// Cascade-up: si field/lot padres están archivados, restaurar solo sus rows
		// (sin cascade-down a otros hijos). Si project está archivado, exigir que
		// el usuario lo restaure manualmente.
		var projectActive int64
		if err := authz.MaybeTenantScope(ctx, tx.Table("projects"), "projects").
			Where("id = ? AND deleted_at IS NULL", wo.ProjectID).
			Count(&projectActive).Error; err != nil {
			return domainerr.Internal("failed to check project")
		}
		if projectActive == 0 {
			return domainerr.Conflict("cannot restore work order while project is archived; restore the project first")
		}
		if err := authz.MaybeTenantScope(ctx, tx.Unscoped().Table("fields"), "fields").
			Where("id = ? AND deleted_at IS NOT NULL", wo.FieldID).
			Updates(lifecycle.RestoreUpdates(tx, "fields", restoredAt)).Error; err != nil {
			return domainerr.Internal("failed to restore parent field")
		}
		if err := authz.MaybeTenantScope(ctx, tx.Unscoped().Table("lots"), "lots").
			Where("id = ? AND deleted_at IS NOT NULL", wo.LotID).
			Updates(lifecycle.RestoreUpdates(tx, "lots", restoredAt)).Error; err != nil {
			return domainerr.Internal("failed to restore parent lot")
		}
		rowState, err := lifecycle.ReadRowState(tx, "workorders", id)
		if err != nil {
			return err
		}
		cause := lifecycle.CauseFromRow(rowState, "workorders", id)

		if err := authz.MaybeTenantScope(ctx, tx.Unscoped().Model(&models.WorkOrder{}), "workorders").
			Where("id = ?", id).
			Updates(lifecycle.RestoreUpdates(tx, "workorders", restoredAt)).Error; err != nil {
			return domainerr.Internal("failed to restore work order")
		}
		itemsRestore := authz.MaybeTenantScope(ctx, tx.Table("workorder_items"), "workorder_items").
			Where("workorder_id = ? AND deleted_at IS NOT NULL", id)
		itemsRestore = lifecycle.ApplyCauseScope(itemsRestore, "workorder_items", cause)
		if err := itemsRestore.Updates(lifecycle.RestoreUpdates(tx, "workorder_items", restoredAt)).Error; err != nil {
			return domainerr.Internal("failed to restore work order items")
		}
		splitsRestore := authz.MaybeTenantScope(ctx, tx.Table("workorder_investor_splits"), "workorder_investor_splits").
			Where("workorder_id = ? AND deleted_at IS NOT NULL", id)
		splitsRestore = lifecycle.ApplyCauseScope(splitsRestore, "workorder_investor_splits", cause)
		if err := splitsRestore.Updates(lifecycle.RestoreUpdates(tx, "workorder_investor_splits", restoredAt)).Error; err != nil {
			return domainerr.Internal("failed to restore work order investor splits")
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
		if err := authz.MaybeTenantScope(ctx, tx.Select("id"), "workorders").Where("id = ?", workOrderID).First(&workOrder).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.NotFound("work order not found")
			}
			return domainerr.Internal("failed to find work order")
		}

		updateTx := tx.Model(&models.WorkOrderInvestorSplit{}).
			Scopes(func(db *gorm.DB) *gorm.DB { return authz.MaybeTenantScope(ctx, db, "workorder_investor_splits") }).
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
	hasWorkspaceFilter := filt.ProjectID != nil || filt.CustomerID != nil || filt.CampaignID != nil || filt.FieldID != nil
	if len(projectIDs) == 0 && hasWorkspaceFilter {
		return &domain.WorkOrderMetrics{
			SurfaceHa:   decimal.Zero,
			Liters:      decimal.Zero,
			Kilograms:   decimal.Zero,
			DirectCost:  decimal.Zero,
			OrdersCount: 0,
		}, nil
	}
	if len(projectIDs) == 0 && authz.TenantStrictModeEnabled() {
		return nil, domainerr.Forbidden("tenant context required")
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

	orderCountQuery := authz.MaybeTenantScope(ctx, r.db.Client().
		WithContext(ctx).
		Table("workorders"), "workorders").
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
		JOIN workorder_items wi ON wi.workorder_id = wo.id AND wi.tenant_id = wo.tenant_id AND wi.deleted_at IS NULL
		JOIN supplies s ON s.id = wi.supply_id AND s.tenant_id = wo.tenant_id AND s.deleted_at IS NULL
		WHERE wo.deleted_at IS NULL
		  AND wi.supply_id = ?
	`
	args := []any{*filt.SupplyID}

	if tenantID, ok := authz.TenantFromContext(ctx); ok {
		q += " AND wo.tenant_id = ?"
		args = append(args, tenantID)
	}
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
	tenantID, hasTenant := authz.TenantFromContext(ctx)
	if !hasTenant && authz.TenantStrictModeEnabled() {
		return decimal.Zero, domainerr.Forbidden("tenant context required")
	}

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
	if hasTenant {
		whereProject += " AND wo.tenant_id = ?"
		args = append(args, tenantID)
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
		      JOIN public.supplies s ON s.id = wi.supply_id AND s.tenant_id = wo.tenant_id AND s.deleted_at IS NULL
		      WHERE wi.workorder_id = wo.id 
		        AND wi.tenant_id = wo.tenant_id
		        AND wi.deleted_at IS NULL
		    ), 0) AS supply_cost
		  FROM public.workorders wo
		  JOIN public.labors l ON l.id = wo.labor_id AND l.tenant_id = wo.tenant_id AND l.deleted_at IS NULL
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

func (r *Repository) GetHarvestAreaSnapshot(
	ctx context.Context,
	lotID int64,
	laborID int64,
	excludeWorkOrderID int64,
) (bool, decimal.Decimal, decimal.Decimal, error) {
	type row struct {
		IsHarvest             bool            `gorm:"column:is_harvest"`
		LotHectares           decimal.Decimal `gorm:"column:lot_hectares"`
		ExistingHarvestedArea decimal.Decimal `gorm:"column:existing_harvested_area"`
	}

	var result row

	tenantID, hasTenant := authz.TenantFromContext(ctx)
	if !hasTenant && authz.TenantStrictModeEnabled() {
		return false, decimal.Zero, decimal.Zero, domainerr.Forbidden("tenant context required")
	}

	laborTenantFilter := ""
	lotTenantFilter := ""
	workOrderTenantFilter := ""
	args := []any{laborID}
	if hasTenant {
		laborTenantFilter = "AND lb.tenant_id = ?"
		args = append(args, tenantID)
	}
	args = append(args, lotID)
	if hasTenant {
		lotTenantFilter = "AND l.tenant_id = ?"
		args = append(args, tenantID)
	}
	args = append(args, lotID)
	if hasTenant {
		workOrderTenantFilter = "AND w.tenant_id = ?"
		args = append(args, tenantID)
	}
	args = append(args, excludeWorkOrderID, excludeWorkOrderID)

	query := fmt.Sprintf(`
		SELECT
			EXISTS (
				SELECT 1
				FROM public.labors lb
				JOIN public.categories cat ON cat.id = lb.category_id AND cat.deleted_at IS NULL
				WHERE lb.id = ?
				  %s
				  AND lb.deleted_at IS NULL
				  AND cat.type_id = 4
				  AND LOWER(TRIM(cat.name)) = 'cosecha'
			) AS is_harvest,
			COALESCE((
				SELECT l.hectares
				FROM public.lots l
				WHERE l.id = ?
				  %s
				  AND l.deleted_at IS NULL
			), 0)::numeric AS lot_hectares,
			COALESCE((
				SELECT SUM(w.effective_area)
				FROM public.workorders w
				JOIN public.labors lb ON lb.id = w.labor_id AND lb.tenant_id = w.tenant_id AND lb.deleted_at IS NULL
				JOIN public.categories cat ON cat.id = lb.category_id AND cat.deleted_at IS NULL
				WHERE w.lot_id = ?
				  %s
				  AND w.deleted_at IS NULL
				  AND w.effective_area > 0
				  AND cat.type_id = 4
				  AND LOWER(TRIM(cat.name)) = 'cosecha'
				  AND (? = 0 OR w.id <> ?)
			), 0)::numeric AS existing_harvested_area
	`, laborTenantFilter, lotTenantFilter, workOrderTenantFilter)

	err := r.db.Client().WithContext(ctx).Raw(query, args...).Scan(&result).Error

	if err != nil {
		return false, decimal.Zero, decimal.Zero, domainerr.Internal("failed to validate harvest area")
	}

	return result.IsHarvest, result.LotHectares, result.ExistingHarvestedArea, nil
}
