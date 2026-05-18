// Package labor contiene la implementación del repositorio para el módulo de labor
package labor

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/devpablocristo/ponti-backend/internal/labor/repository/models"
	"github.com/devpablocristo/ponti-backend/internal/labor/usecases/domain"
	"github.com/devpablocristo/ponti-backend/internal/shared/authz"
	shareddb "github.com/devpablocristo/ponti-backend/internal/shared/db"
	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
	sharedfilters "github.com/devpablocristo/ponti-backend/internal/shared/filters"
	"github.com/devpablocristo/ponti-backend/internal/shared/lifecycle"
	sharedmodels "github.com/devpablocristo/ponti-backend/internal/shared/models"
	sharedrepo "github.com/devpablocristo/ponti-backend/internal/shared/repository"
	workOrderModels "github.com/devpablocristo/ponti-backend/internal/work-order/repository/models"
	"gorm.io/gorm"

	"github.com/devpablocristo/core/errors/go/domainerr"
	types "github.com/devpablocristo/ponti-backend/internal/shared/types"
	"github.com/shopspring/decimal"
)

type GormEnginePort interface {
	Client() *gorm.DB
}
type Repository struct {
	db GormEnginePort
}

func NewRepository(db GormEnginePort) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) CreateLabor(ctx context.Context, labor *domain.Labor) (int64, error) {
	if err := sharedrepo.ValidateEntity(labor, "labor"); err != nil {
		return 0, err
	}
	model := models.FromDomain(labor)
	if tenantID, ok, err := authz.OptionalTenantOrStrict(ctx); err != nil {
		return 0, err
	} else if ok {
		model.TenantID = tenantID
	}
	if err := r.db.Client().WithContext(ctx).Create(model).Error; err != nil {
		return 0, domainerr.Internal("failed to create labor")
	}
	return model.ID, nil
}

func (r *Repository) ExistsLaborByProjectAndName(ctx context.Context, projectID int64, name string) (bool, error) {
	var count int64
	err := r.db.Client().WithContext(ctx).
		Model(&models.Labor{}).
		Scopes(func(db *gorm.DB) *gorm.DB { return authz.MaybeTenantScope(ctx, db, "labors") }).
		Where("project_id = ? AND deleted_at IS NULL AND LOWER(TRIM(name)) = LOWER(TRIM(?))", projectID, name).
		Count(&count).Error
	if err != nil {
		return false, domainerr.Internal("failed to check labor duplicate")
	}
	return count > 0, nil
}

func (r *Repository) ExistsOtherLaborByProjectAndName(ctx context.Context, projectID int64, name string, laborID int64) (bool, error) {
	var count int64
	err := r.db.Client().WithContext(ctx).
		Model(&models.Labor{}).
		Scopes(func(db *gorm.DB) *gorm.DB { return authz.MaybeTenantScope(ctx, db, "labors") }).
		Where("project_id = ? AND deleted_at IS NULL AND id <> ? AND LOWER(TRIM(name)) = LOWER(TRIM(?))", projectID, laborID, name).
		Count(&count).Error
	if err != nil {
		return false, domainerr.Internal("failed to check labor duplicate")
	}
	return count > 0, nil
}

func (r *Repository) GetLabor(ctx context.Context, laborID int64) (*domain.Labor, error) {
	var m models.Labor
	if err := authz.MaybeTenantScope(ctx, r.db.Client().WithContext(ctx), "labors").
		Preload("Category").
		First(&m, laborID).Error; err != nil {
		return nil, sharedrepo.HandleGormError(err, "labor", laborID)
	}
	return m.ToDomain(), nil
}

func (r *Repository) GetWorkOrdersByLaborID(ctx context.Context, laborID int64) (int64, error) {
	var count int64
	if err := r.db.Client().WithContext(ctx).
		Model(&workOrderModels.WorkOrder{}).
		Scopes(func(db *gorm.DB) *gorm.DB { return authz.MaybeTenantScope(ctx, db, "workorders") }).
		Joins("JOIN labors ON labors.id = workorders.labor_id").
		Where("labors.id = ? AND labors.tenant_id = workorders.tenant_id AND workorders.deleted_at IS NULL", laborID).
		Count(&count).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, domainerr.Internal("failed to get work order")
	}
	return count, nil
}

func (r *Repository) DeleteLabor(ctx context.Context, id int64) error {
	return r.ArchiveLabor(ctx, id)
}

// ArchiveLabor archiva (soft delete) un labor con validación.
func (r *Repository) ArchiveLabor(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "labor"); err != nil {
		return err
	}
	actor, err := sharedmodels.ActorFromContext(ctx)
	if err != nil {
		return err
	}
	deletedBy := &actor

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		archivedAt := time.Now()
		var l models.Labor
		if err := authz.MaybeTenantScope(ctx, tx.Unscoped(), "labors").Where("id = ?", id).First(&l).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("labor %d not found", id))
			}
			return domainerr.Internal("failed to get labor")
		}
		if l.DeletedAt.Valid {
			return domainerr.Conflict("labor already archived")
		}

		cause, err := lifecycle.RootCause(tx, l.TenantID, "labors", id, nil, deletedBy)
		if err != nil {
			return err
		}
		if err := authz.MaybeTenantScope(ctx, tx.Model(&models.Labor{}), "labors").
			Where("id = ?", id).
			Updates(lifecycle.ArchiveUpdates(tx, "labors", archivedAt, deletedBy, cause)).Error; err != nil {
			return domainerr.Internal("failed to archive labor")
		}
		return nil
	})
}

// RestoreLabor restaura un labor archivado.
func (r *Repository) RestoreLabor(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "labor"); err != nil {
		return err
	}

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		restoredAt := time.Now()
		var l models.Labor
		if err := authz.MaybeTenantScope(ctx, tx.Unscoped(), "labors").Where("id = ?", id).First(&l).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("labor %d not found", id))
			}
			return domainerr.Internal("failed to get labor")
		}
		if !l.DeletedAt.Valid {
			return domainerr.Conflict("labor is not archived")
		}
		var projectActive int64
		if err := authz.MaybeTenantScope(ctx, tx.Table("projects"), "projects").
			Where("id = ? AND deleted_at IS NULL", l.ProjectId).
			Count(&projectActive).Error; err != nil {
			return domainerr.Internal("failed to check project")
		}
		if projectActive == 0 {
			return domainerr.Conflict("cannot restore labor while project is archived")
		}

		if err := authz.MaybeTenantScope(ctx, tx.Unscoped().Model(&models.Labor{}), "labors").
			Where("id = ?", id).
			Updates(lifecycle.RestoreUpdates(tx, "labors", restoredAt)).Error; err != nil {
			return domainerr.Internal("failed to restore labor")
		}
		return nil
	})
}

// HardDeleteLabor elimina definitivamente un labor.
// Bloquea con 409 si tiene workorders (activas o archivadas) referenciándolo.
func (r *Repository) HardDeleteLabor(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "labor"); err != nil {
		return err
	}

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := authz.MaybeTenantScope(ctx, tx.Unscoped().Table("labors"), "labors").Where("id = ?", id).Count(&count).Error; err != nil {
			return domainerr.Internal("failed to check labor existence")
		}
		if count == 0 {
			return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("labor %d not found", id))
		}
		if err := lifecycle.RequireArchived(authz.MaybeTenantScope(ctx, tx.Unscoped().Table("labors"), "labors"), "labors", "labor", id); err != nil {
			return err
		}

		var woCount int64
		if err := authz.MaybeTenantScope(ctx, tx.Unscoped().Model(&workOrderModels.WorkOrder{}), "workorders").Where("labor_id = ?", id).Count(&woCount).Error; err != nil {
			return domainerr.Internal("failed to check work orders")
		}
		if woCount > 0 {
			return domainerr.Conflict(fmt.Sprintf("labor has %d work order(s); archive or hard-delete them first", woCount))
		}

		if err := authz.MaybeTenantScope(ctx, tx.Unscoped(), "labors").Delete(&models.Labor{}, "id = ?", id).Error; err != nil {
			return domainerr.Internal("failed to hard delete labor")
		}
		return nil
	})
}

// ListArchivedLabors lista labors archivados. Si projectID es 0, devuelve labors de todos los proyectos del tenant.
func (r *Repository) ListArchivedLabors(ctx context.Context, page, perPage int, projectID int64) ([]domain.ListedLabor, int64, error) {
	var total int64
	base := r.db.Client().WithContext(ctx).
		Unscoped().
		Model(&models.Labor{}).
		Where("deleted_at IS NOT NULL")
	if projectID > 0 {
		base = base.Where("project_id = ?", projectID)
	}
	base = authz.MaybeTenantScope(ctx, base, "labors")

	if err := base.Count(&total).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to count archived labors")
	}

	var list []models.Labor
	offset := (page - 1) * perPage
	if err := base.
		Preload("Category").
		Select("id, name, contractor_name, price, is_partial_price, category_id, project_id, updated_at").
		Offset(offset).
		Limit(perPage).
		Order("deleted_at DESC").
		Find(&list).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to list archived labors")
	}

	labors := make([]domain.ListedLabor, len(list))
	for i, labor := range list {
		labors[i] = domain.ListedLabor{
			ID:             labor.ID,
			Name:           labor.Name,
			Price:          labor.Price,
			IsPartialPrice: labor.IsPartialPrice,
			ContractorName: labor.ContractorName,
			ProjectId:      labor.ProjectId,
			CategoryId:     labor.LaborCategoryID,
			CategoryName:   labor.Category.Name,
			Base:           shareddomain.Base{UpdatedAt: labor.UpdatedAt},
		}
	}
	return labors, total, nil
}

func (r *Repository) UpdateLabor(ctx context.Context, labor *domain.Labor) error {
	if err := sharedrepo.ValidateEntity(labor, "labor"); err != nil {
		return err
	}

	updates := map[string]any{
		"name":             labor.Name,
		"contractor_name":  labor.ContractorName,
		"price":            labor.Price,
		"is_partial_price": labor.IsPartialPrice,
		"project_id":       labor.ProjectId,
		"category_id":      labor.CategoryId,
		"updated_by":       labor.UpdatedBy,
	}

	result := r.db.Client().WithContext(ctx).
		Model(&models.Labor{}).
		Scopes(func(db *gorm.DB) *gorm.DB { return authz.MaybeTenantScope(ctx, db, "labors") }).
		Where("id = ?", labor.ID).
		Updates(updates)

	if result.Error != nil {
		return domainerr.Internal("failed to update labor")
	}
	if result.RowsAffected == 0 {
		return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("labor with id %d does not exist", labor.ID))
	}
	return nil
}

func (r *Repository) ListLabor(ctx context.Context, page, perPage int, projectID int64) ([]domain.ListedLabor, int64, error) {
	var list []models.Labor
	var total int64

	base := r.db.Client().WithContext(ctx).
		Model(&models.Labor{}).
		Where("project_id = ?", projectID).
		Where("deleted_at IS NULL")
	base = authz.MaybeTenantScope(ctx, base, "labors")

	// Conteo total filtrado por proyecto
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to count labors")
	}

	if err := base.
		Preload("Category").
		Select("id, name, contractor_name, price, is_partial_price, category_id, updated_at").
		Limit(perPage).
		Offset((page - 1) * perPage).
		Find(&list).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to list labor")
	}

	// Mapear a dominio ligero
	labors := make([]domain.ListedLabor, len(list))
	for i, labor := range list {
		labors[i] = domain.ListedLabor{
			ID:             labor.ID,
			Name:           labor.Name,
			Price:          labor.Price,
			IsPartialPrice: labor.IsPartialPrice,
			ContractorName: labor.ContractorName,
			CategoryId:     labor.LaborCategoryID,
			CategoryName:   labor.Category.Name,
			Base:           shareddomain.Base{UpdatedAt: labor.UpdatedAt},
		}
	}

	return labors, total, nil
}

func (r *Repository) ListLaborCategoriesByTypeID(ctx context.Context, typeID int64) ([]domain.LaborCategory, error) {
	var laborCategoriesModels []models.LaborCategory
	db0 := r.db.
		Client().
		WithContext(ctx).
		Model(&models.LaborCategory{}).
		Where("type_id = ?", typeID)

	if err := db0.Find(&laborCategoriesModels).Error; err != nil {
		return nil, domainerr.Internal("failed to list labor categories")
	}

	laborCategories := make([]domain.LaborCategory, len(laborCategoriesModels))
	for i, laborCategory := range laborCategoriesModels {
		laborCategories[i] = domain.LaborCategory{
			ID:   laborCategory.ID,
			Name: laborCategory.Name,
			LaborType: domain.LaborType{
				ID:   laborCategory.LaborType.ID,
				Name: laborCategory.LaborType.Name,
			},
		}
	}
	return laborCategories, nil
}

func (r *Repository) ListByWorkOrder(ctx context.Context, workOrderID int64) ([]domain.LaborRawItem, error) {
	var workOrderCount int64
	if err := authz.MaybeTenantScope(ctx, r.db.Client().WithContext(ctx).Table("workorders"), "workorders").
		Where("id = ? AND deleted_at IS NULL", workOrderID).
		Count(&workOrderCount).Error; err != nil {
		return nil, domainerr.Internal("failed to validate work order")
	}
	if workOrderCount == 0 {
		return nil, domainerr.NotFound("work order not found")
	}

	var v4Models []models.LaborListItem

	query := fmt.Sprintf(`
		SELECT
            v4.workorder_id,
            v4.workorder_number,
            v4.date,
            v4.project_name,
            v4.field_name,
            COALESCE(v4.crop_name, '') AS crop_name,
            v4.labor_name,
            v4.contractor,
            v4.surface_ha,
            v4.cost_per_ha,
            COALESCE(v4.labor_category_name, '') AS category_name,
            COALESCE(v4.investor_name, '') AS investor_name,
			v4.dollar_average_month,
			v4.dollar_average_month AS usd_avg_value,
			i.id AS invoice_id,
			i.number AS invoice_number,
			i.company AS invoice_company,
			i.date AS invoice_date,
			i.status AS invoice_status
		FROM %s AS v4
		LEFT JOIN LATERAL (
    SELECT i.*
    FROM invoices i
    WHERE i.work_order_id = v4.workorder_id
      AND (i.investor_id = v4.investor_id OR i.investor_id IS NULL)
      AND i.deleted_at IS NULL
    ORDER BY
      CASE
        WHEN i.investor_id = v4.investor_id THEN 0
        WHEN i.investor_id IS NULL THEN 1
        ELSE 2
      END,
      i.id DESC
    LIMIT 1
) i ON true
		WHERE v4.workorder_id = ?
	`, shareddb.ReportView("labor_list"))

	err := r.db.Client().WithContext(ctx).Raw(query, workOrderID).Scan(&v4Models).Error
	if err != nil {
		return nil, domainerr.Internal("failed to list labors by work order")
	}

	// Convertir a LaborRawItem para mantener compatibilidad
	raws := make([]domain.LaborRawItem, len(v4Models))
	for i, m := range v4Models {
		// Manejar InvoiceID de forma segura
		var invoiceID int64
		if m.InvoiceID != nil {
			invoiceID = *m.InvoiceID
		}

		raws[i] = domain.LaborRawItem{
			WorkOrderID:     m.WorkOrderID,
			WorkOrderNumber: m.WorkOrderNumber,
			Date:            m.Date,
			ProjectName:     m.ProjectName,
			FieldName:       m.FieldName,
			CropName:        safeStringPtr(m.CropName),
			LaborName:       m.LaborName,
			Contractor:      m.Contractor,
			SurfaceHa:       m.SurfaceHa,
			CostHa:          m.CostPerHa,
			CategoryName:    safeStringPtr(m.LaborCategoryName),
			InvestorName:    safeStringPtr(m.InvestorName),
			USDAvgValue:     m.USDAvgValue,
			InvoiceID:       invoiceID,
			InvoiceNumber:   safeStringPtr(m.InvoiceNumber),
			InvoiceCompany:  safeStringPtr(m.InvoiceCompany),
			InvoiceDate:     m.InvoiceDate,
			InvoiceStatus:   safeStringPtr(m.InvoiceStatus),
		}
	}
	return raws, nil
}

func (r *Repository) ListGroupLabor(
	ctx context.Context,
	inp types.Input,
	filter domain.LaborFilter,
) ([]domain.LaborListItem, types.PageInfo, error) {
	projectIDs, err := sharedfilters.ResolveProjectIDs(ctx, r.db.Client(), sharedfilters.WorkspaceFilter{
		CustomerID: filter.CustomerID,
		ProjectID:  filter.ProjectID,
		CampaignID: filter.CampaignID,
		FieldID:    filter.FieldID,
	})
	if err != nil {
		return nil, types.PageInfo{}, err
	}
	hasWorkspaceFilter := filter.CustomerID != nil || filter.ProjectID != nil || filter.CampaignID != nil || filter.FieldID != nil
	if len(projectIDs) == 0 && hasWorkspaceFilter {
		return []domain.LaborListItem{}, types.NewPageInfo(int(inp.Page), int(inp.PageSize), 0), nil
	}
	if len(projectIDs) == 0 && authz.TenantStrictModeEnabled() {
		return nil, types.PageInfo{}, domainerr.Forbidden("tenant context required")
	}

	where := []string{}
	args := []any{}
	if len(projectIDs) > 0 {
		where = append(where, "v4.project_id IN ?")
		args = append(args, projectIDs)
	}
	if filter.FieldID != nil && *filter.FieldID > 0 {
		where = append(where, "v4.field_id = ?")
		args = append(args, *filter.FieldID)
	}
	whereSQL := "1=1"
	if len(where) > 0 {
		whereSQL = strings.Join(where, " AND ")
	}
	view := shareddb.ReportView("labor_list")

	selectCols := `
			v4.workorder_id,
			v4.workorder_number,
			v4.date,
			v4.project_id,
			v4.field_id,
			v4.project_name,
			v4.field_name,
			v4.lot_id,
			v4.lot_name,
			v4.crop_id,
			COALESCE(v4.crop_name, '') AS crop_name,
			v4.labor_id,
			v4.labor_name,
			v4.labor_category_id,
			COALESCE(v4.labor_category_name, '') AS labor_category_name,
			v4.contractor,
			v4.contractor_name,
			v4.surface_ha,
			v4.cost_per_ha,
			v4.total_labor_cost,
			v4.dollar_average_month,
			v4.dollar_average_month AS usd_avg_value,
			v4.usd_cost_ha,
			v4.usd_net_total,
			v4.investor_id,
			COALESCE(v4.investor_name, '') AS investor_name,
			i.id AS invoice_id,
			i.number AS invoice_number,
			i.company AS invoice_company,
			i.date AS invoice_date,
			i.status AS invoice_status
	`

	// Conteo para paginación
	var total int64
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM %s AS v4
		LEFT JOIN LATERAL (
    SELECT i.*
    FROM invoices i
    WHERE i.work_order_id = v4.workorder_id
      AND (i.investor_id = v4.investor_id OR i.investor_id IS NULL)
      AND i.deleted_at IS NULL
    ORDER BY
      CASE
        WHEN i.investor_id = v4.investor_id THEN 0
        WHEN i.investor_id IS NULL THEN 1
        ELSE 2
      END,
      i.id DESC
    LIMIT 1
) i ON true


		WHERE %s
	`, view, whereSQL)
	if err := r.db.Client().WithContext(ctx).Raw(countQuery, args...).Scan(&total).Error; err != nil {
		return nil, types.PageInfo{}, domainerr.Internal("failed to count labors for work order")
	}

	offset := (int(inp.Page) - 1) * int(inp.PageSize)

	// Datos
	var rows []models.LaborListItem
	dataQuery := fmt.Sprintf(`
		SELECT %s
		FROM %s AS v4
		LEFT JOIN LATERAL (
    SELECT i.*
    FROM invoices i
    WHERE i.work_order_id = v4.workorder_id
      AND (i.investor_id = v4.investor_id OR i.investor_id IS NULL)
      AND i.deleted_at IS NULL
    ORDER BY
      CASE
        WHEN i.investor_id = v4.investor_id THEN 0
        WHEN i.investor_id IS NULL THEN 1
        ELSE 2
      END,
      i.id DESC
    LIMIT 1
) i ON true


		WHERE %s
		ORDER BY v4.workorder_number DESC
		LIMIT ? OFFSET ?
	`, selectCols, view, whereSQL)
	dataArgs := append(append([]any{}, args...), int(inp.PageSize), offset)
	if err := r.db.Client().WithContext(ctx).Raw(dataQuery, dataArgs...).Scan(&rows).Error; err != nil {
		return nil, types.PageInfo{}, domainerr.Internal("failed to list grouped labors")
	}

	// IVA (tasa 0.105; si viene 1.105 se normaliza en getIVAPercentage)
	ivaRate, _ := r.getIVAPercentage(ctx)        // 0.105
	ivaMul := decimal.NewFromInt(1).Add(ivaRate) // 1.105

	// Mapear a dominio
	list := make([]domain.LaborListItem, len(rows))
	for i, m := range rows {
		// Costo ARS/ha SIN IVA = USD/ha × dólar prom
		costHaARS := m.USDCostHa.Mul(m.USDAvgValue)

		// Total neto ARS (SIN IVA)
		netTotal := costHaARS.Mul(m.SurfaceHa)

		// "Total IVA" (histórico): mostramos TOTAL CON IVA para que coincida con la UI actual
		totalConIVA := netTotal.Mul(ivaMul) // net × 1.105

		// USD (vienen de la vista)
		usdCostHa := m.USDCostHa
		usdNetTotal := m.USDNetTotal

		// Invoice seguro
		var invoiceID int64
		if m.InvoiceID != nil {
			invoiceID = *m.InvoiceID
		}

		var investorID int64
		if m.InvestorID != nil {
			investorID = *m.InvestorID
		}

		list[i] = domain.LaborListItem{
			WorkOrderID:     m.WorkOrderID,
			WorkOrderNumber: m.WorkOrderNumber,
			Date:            m.Date,
			ProjectName:     m.ProjectName,
			FieldName:       m.FieldName,
			LotId:           safeInt64Ptr(m.LotID),
			LotName:         safeStringPtr(m.LotName),
			CropName:        safeStringPtr(m.CropName),
			LaborName:       m.LaborName,
			Contractor:      m.Contractor,
			SurfaceHa:       m.SurfaceHa,
			CostHa:          costHaARS, // ARS/ha SIN IVA (10 × 1000 = 10.000)
			CategoryName:    safeStringPtr(m.LaborCategoryName),
			InvestorID:      investorID,
			InvestorName:    safeStringPtr(m.InvestorName),
			USDAvgValue:     m.USDAvgValue,
			NetTotal:        netTotal,    // 10.000 × 100 = 1.000.000
			TotalIVA:        totalConIVA, // MOSTRAMOS TOTAL CON IVA: 1.000.000 × 1.105 = 1.105.000
			USDCostHa:       usdCostHa,   // 10
			USDNetTotal:     usdNetTotal, // 1000
			InvoiceID:       invoiceID,
			InvoiceNumber:   safeStringPtr(m.InvoiceNumber),
			InvoiceCompany:  safeStringPtr(m.InvoiceCompany),
			InvoiceDate:     m.InvoiceDate,
			InvoiceStatus:   safeStringPtr(m.InvoiceStatus),
		}
	}

	pageInfo := types.NewPageInfo(int(inp.Page), int(inp.PageSize), total)
	return list, pageInfo, nil
}

// getIVAPercentage obtiene el porcentaje de IVA desde bparams
func (r *Repository) getIVAPercentage(ctx context.Context) (decimal.Decimal, error) {
	var value string
	err := r.db.Client().WithContext(ctx).
		// business_parameters es la tabla real (bparams era un nombre legacy).
		Table("business_parameters").
		Select("value").
		Where("key = ? AND deleted_at IS NULL", "iva_percentage").
		Scan(&value).Error
	if err != nil || value == "" {
		slog.Warn("IVA percentage not found in business_parameters, using fallback 0.105",
			"error", err, "key", "iva_percentage")
		return decimal.NewFromFloat(0.105), nil
	}
	v, err := decimal.NewFromString(value)
	if err != nil {
		slog.Warn("IVA percentage value is not a valid decimal, using fallback 0.105",
			"error", err, "raw_value", value)
		return decimal.NewFromFloat(0.105), nil
	}
	if v.GreaterThan(decimal.NewFromInt(1)) {
		v = v.Sub(decimal.NewFromInt(1)) // 1.105 -> 0.105
	}
	return v, nil
}

// safeStringPtr convierte un string pointer a string seguro
func safeStringPtr(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

func safeInt64Ptr(ptr *int64) int64 {
	if ptr == nil {
		return 0
	}
	return *ptr
}

func (r *Repository) GetMetrics(ctx context.Context, f domain.LaborFilter) (*domain.LaborMetrics, error) {
	projectIDs, err := sharedfilters.ResolveProjectIDs(ctx, r.db.Client(), sharedfilters.WorkspaceFilter{
		CustomerID: f.CustomerID,
		ProjectID:  f.ProjectID,
		CampaignID: f.CampaignID,
		FieldID:    f.FieldID,
	})
	if err != nil {
		return nil, err
	}
	if len(projectIDs) == 0 && (f.ProjectID != nil || f.CustomerID != nil || f.CampaignID != nil || f.FieldID != nil) {
		return &domain.LaborMetrics{
			SurfaceHa:    decimal.Zero,
			NetTotalCost: decimal.Zero,
			AvgCostPerHa: decimal.Zero,
		}, nil
	}
	if len(projectIDs) == 0 && authz.TenantStrictModeEnabled() {
		return nil, domainerr.Forbidden("tenant context required")
	}

	var row struct {
		SurfaceHa    decimal.Decimal `gorm:"column:surface_ha"`
		NetTotalCost decimal.Decimal `gorm:"column:total_labor_cost"`
		AvgCostPerHa decimal.Decimal `gorm:"column:avg_labor_cost_per_ha"`
	}

	where := []string{"1=1"}
	args := []any{}
	if len(projectIDs) > 0 {
		where = append(where, "project_id IN ?")
		args = append(args, projectIDs)
	}
	if f.FieldID != nil && *f.FieldID > 0 {
		where = append(where, "field_id = ?")
		args = append(args, *f.FieldID)
	}

	q := fmt.Sprintf(`
		SELECT 
			COALESCE(SUM(surface_ha), 0) as surface_ha,
			COALESCE(SUM(total_labor_cost), 0) as total_labor_cost,
			CASE 
				WHEN COALESCE(SUM(surface_ha), 0) > 0 
				THEN COALESCE(SUM(total_labor_cost), 0) / COALESCE(SUM(surface_ha), 0)
				ELSE 0 
			END as avg_labor_cost_per_ha
		FROM %s
		WHERE %s
	`, shareddb.ReportView("labor_metrics"), strings.Join(where, " AND "))
	if err := r.db.Client().WithContext(ctx).Raw(q, args...).Scan(&row).Error; err != nil {
		return nil, domainerr.Internal("failed to get labor metrics")
	}

	return &domain.LaborMetrics{
		SurfaceHa:    row.SurfaceHa,
		NetTotalCost: row.NetTotalCost,
		AvgCostPerHa: row.AvgCostPerHa,
	}, nil
}

func (r *Repository) ListAllGroupLabor(ctx context.Context) ([]domain.LaborRawItem, error) {
	projectIDs, err := sharedfilters.ResolveProjectIDs(ctx, r.db.Client(), sharedfilters.WorkspaceFilter{})
	if err != nil {
		return nil, err
	}

	base := r.db.Client().
		WithContext(ctx).
		Table(shareddb.ReportView("labor_list") + " AS v4").
		Select(`
            v4.workorder_id,
			v4.workorder_number,
			v4.date,
			v4.project_id,
			v4.field_id,
			v4.project_name,
			v4.field_name,
			COALESCE(v4.crop_name, '') AS crop_name,
			v4.labor_name,
			COALESCE(v4.labor_category_name, '') AS category_name,
			v4.contractor,
			v4.surface_ha,
			v4.cost_per_ha,
			v4.contractor_name,
			COALESCE(v4.investor_name, '') AS investor_name,
			v4.dollar_average_month,
			v4.dollar_average_month AS usd_avg_value,
			i.id AS invoice_id,
			i.number AS invoice_number,
			i.company AS invoice_company,
			i.date AS invoice_date,
			i.status AS invoice_status
        `).
		Joins(`LEFT JOIN LATERAL (
    SELECT i.*
    FROM invoices i
    WHERE i.work_order_id = v4.workorder_id
      AND (i.investor_id = v4.investor_id OR i.investor_id IS NULL)
      AND i.deleted_at IS NULL
    ORDER BY
      CASE
        WHEN i.investor_id = v4.investor_id THEN 0
        WHEN i.investor_id IS NULL THEN 1
        ELSE 2
      END,
      i.id DESC
    LIMIT 1
) i ON true`)
	if len(projectIDs) > 0 {
		base = base.Where("v4.project_id IN ?", projectIDs)
	} else if authz.TenantStrictModeEnabled() {
		return nil, domainerr.Forbidden("tenant context required")
	}

	var rows []models.LaborListItem

	if err := base.Order("v4.workorder_number DESC").Scan(&rows).Error; err != nil {
		return nil, domainerr.Internal("failed to list grouped labors")
	}

	list := make([]domain.LaborRawItem, len(rows))
	for i, m := range rows {
		// Calcular valores de USD dinámicamente
		netTotal := m.SurfaceHa.Mul(m.CostPerHa)

		// Obtener porcentaje de IVA dinámicamente desde bparams
		ivaPercentage, err := r.getIVAPercentage(ctx)
		if err != nil {
			slog.Warn("failed to get IVA percentage from bparams, using fallback 0.105",
				"error", err)
			ivaPercentage = decimal.NewFromFloat(0.105)
		}
		totalIVA := netTotal.Mul(ivaPercentage)

		usd := m.USDAvgValue
		var usdCostHa, usdNetTotal decimal.Decimal
		if usd.GreaterThan(decimal.Zero) {
			usdCostHa = m.CostPerHa.Div(usd)
			usdNetTotal = netTotal.Div(usd)
		} else {
			// usd <= 0 dejamos USD en 0 (sin dividir)
			usdCostHa = decimal.Zero
			usdNetTotal = decimal.Zero
		}

		// Manejar campos opcionales
		cropName := ""
		if m.CropName != nil {
			cropName = *m.CropName
		}
		categoryName := ""
		if m.LaborCategoryName != nil {
			categoryName = *m.LaborCategoryName
		}
		investorName := ""
		if m.InvestorName != nil {
			investorName = *m.InvestorName
		}
		invoiceNumber := ""
		if m.InvoiceNumber != nil {
			invoiceNumber = *m.InvoiceNumber
		}
		invoiceCompany := ""
		if m.InvoiceCompany != nil {
			invoiceCompany = *m.InvoiceCompany
		}
		invoiceStatus := ""
		if m.InvoiceStatus != nil {
			invoiceStatus = *m.InvoiceStatus
		}

		list[i] = domain.LaborRawItem{
			WorkOrderID:     m.WorkOrderID,
			WorkOrderNumber: m.WorkOrderNumber,
			Date:            m.Date,
			ProjectName:     m.ProjectName,
			FieldName:       m.FieldName,
			CropName:        cropName,
			LaborName:       m.LaborName,
			Contractor:      m.Contractor,
			SurfaceHa:       m.SurfaceHa,
			CostHa:          m.CostPerHa,
			CategoryName:    categoryName,
			InvestorName:    investorName,
			USDAvgValue:     m.USDAvgValue,
			NetTotal:        netTotal,
			TotalIVA:        totalIVA,
			USDCostHa:       usdCostHa,
			USDNetTotal:     usdNetTotal,
			InvoiceID:       0,
			InvoiceNumber:   invoiceNumber,
			InvoiceCompany:  invoiceCompany,
			InvoiceDate:     m.InvoiceDate,
			InvoiceStatus:   invoiceStatus,
		}
	}
	return list, nil
}
