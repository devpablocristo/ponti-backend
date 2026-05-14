package supply

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/devpablocristo/core/errors/go/domainerr"
	actorsync "github.com/devpablocristo/ponti-backend/internal/actor"
	providermodel "github.com/devpablocristo/ponti-backend/internal/provider/repository/models"
	providerdomain "github.com/devpablocristo/ponti-backend/internal/provider/usecases/domain"
	"github.com/devpablocristo/ponti-backend/internal/shared/authz"
	sharedfilters "github.com/devpablocristo/ponti-backend/internal/shared/filters"
	"github.com/devpablocristo/ponti-backend/internal/shared/lifecycle"
	sharedmodels "github.com/devpablocristo/ponti-backend/internal/shared/models"
	sharedrepo "github.com/devpablocristo/ponti-backend/internal/shared/repository"
	stockmodel "github.com/devpablocristo/ponti-backend/internal/stock/repository/models"
	"github.com/devpablocristo/ponti-backend/internal/supply/repository/models"
	"github.com/devpablocristo/ponti-backend/internal/supply/usecases/domain"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func withSupplyMovementLookups(db *gorm.DB) *gorm.DB {
	return db.
		Preload("Supply", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("Supply.Type").
		Preload("Supply.Category").
		Preload("Investor").
		Preload("Provider")
}

func (r *Repository) CreateSupplyMovement(ctx context.Context, movement *domain.SupplyMovement) (int64, error) {
	if err := sharedrepo.ValidateEntity(movement, "supply movement"); err != nil {
		return 0, err
	}
	model := models.SupplyMovementFromDomain(movement)
	if tenantID, ok, err := authz.OptionalTenantOrStrict(ctx); err != nil {
		return 0, err
	} else if ok {
		model.TenantID = tenantID
	}
	db := r.getDB(ctx)
	if err := db.Transaction(func(tx *gorm.DB) error {
		if movement.Investor != nil && movement.Investor.ID > 0 && strings.TrimSpace(movement.Investor.Name) != "" {
			if _, err := actorsync.SyncLegacyActor(tx, actorsync.LegacyActorSync{
				SourceTable: actorsync.LegacyInvestors,
				SourceID:    movement.Investor.ID,
				Name:        movement.Investor.Name,
				ActorKind:   actorsync.KindUnknown,
				Role:        actorsync.RoleInversor,
				UpdatedAt:   time.Now(),
				UpdatedBy:   movement.UpdatedBy,
			}); err != nil {
				return err
			}
		}
		if movement.Provider != nil && movement.Provider.ID > 0 && strings.TrimSpace(movement.Provider.Name) != "" {
			if _, err := actorsync.SyncLegacyActor(tx, actorsync.LegacyActorSync{
				SourceTable: actorsync.LegacyProviders,
				SourceID:    movement.Provider.ID,
				Name:        movement.Provider.Name,
				ActorKind:   actorsync.KindOrganization,
				Role:        actorsync.RoleProveedor,
				UpdatedAt:   time.Now(),
				UpdatedBy:   movement.UpdatedBy,
			}); err != nil {
				return err
			}
		}
		if err := tx.Create(model).Error; err != nil {
			return err
		}
		if err := actorsync.RefreshSupplyMovementActorColumns(tx, model.ID); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return 0, err
	}
	return model.ID, nil
}

func (r *Repository) ResetFieldStockCounts(ctx context.Context, projectID int64, updatedBy *string) error {
	if projectID <= 0 {
		return domainerr.Validation("project_id must be greater than 0")
	}

	now := time.Now().UTC()

	yearPeriod := int64(now.Year())
	monthPeriod := int64(now.Month())

	tenantID, hasTenant := authz.TenantFromContext(ctx)
	if !hasTenant && authz.TenantStrictModeEnabled() {
		return domainerr.Forbidden("tenant context required")
	}

	tenantFilter := ""
	args := []any{projectID}
	if hasTenant {
		tenantFilter = " AND st.tenant_id = ?"
		args = append(args, tenantID)
	}
	args = append(args, yearPeriod, monthPeriod, now, now, updatedBy, updatedBy)

	if err := r.getDB(ctx).Exec(fmt.Sprintf(`
		WITH ranked_closed AS (
			SELECT
				st.tenant_id,
				st.project_id,
				st.supply_id,
				st.investor_id,
				st.real_stock_units,
				st.has_real_stock_count,
				ROW_NUMBER() OVER (
					PARTITION BY st.supply_id, st.investor_id
					ORDER BY st.close_date DESC, st.id DESC
				) AS row_num
			FROM stocks st
			WHERE st.project_id = ?
			  AND st.close_date IS NOT NULL
			  AND st.deleted_at IS NULL
			  %s
			  AND NOT EXISTS (
				  SELECT 1
				  FROM stocks active
				  WHERE active.project_id = st.project_id
				    AND active.supply_id = st.supply_id
				    AND active.investor_id = st.investor_id
				    AND active.close_date IS NULL
				    AND active.deleted_at IS NULL
				    AND active.tenant_id = st.tenant_id
			  )
		)
		INSERT INTO stocks (
			tenant_id,
			project_id,
			supply_id,
			investor_id,
			close_date,
			real_stock_units,
			initial_units,
			year_period,
			month_period,
			units_entered,
			units_consumed,
			has_real_stock_count,
			created_at,
			updated_at,
			created_by,
			updated_by
		)
		SELECT
			tenant_id,
			project_id,
			supply_id,
			investor_id,
			NULL,
			real_stock_units,
			real_stock_units,
			?,
			?,
			0,
			0,
			has_real_stock_count,
			?,
			?,
			?,
			?
		FROM ranked_closed
		WHERE row_num = 1
	`, tenantFilter), args...).Error; err != nil {
		return domainerr.Internal("failed to prepare field stock counts")
	}

	updates := map[string]any{
		"real_stock_units":     decimal.Zero,
		"has_real_stock_count": true,
		"updated_at":           now,
		"updated_by":           updatedBy,
	}

	if err := r.getDB(ctx).
		Model(&stockmodel.Stock{}).
		Scopes(func(db *gorm.DB) *gorm.DB { return authz.MaybeTenantScope(ctx, db, "stocks") }).
		Where("project_id = ?", projectID).
		Where("close_date IS NULL").
		Where("deleted_at IS NULL").
		Updates(updates).Error; err != nil {
		return domainerr.Internal("failed to reset field stock counts")
	}

	return nil
}

func (r *Repository) CreateProvider(ctx context.Context, provider *providerdomain.Provider) (int64, error) {
	if err := sharedrepo.ValidateEntity(provider, "provider"); err != nil {
		return 0, err
	}

	client := r.getDB(ctx)

	provider.Name = strings.TrimSpace(provider.Name)

	if provider.Name == "" {
		return 0, domainerr.Validation("provider name is empty")
	}

	model := providermodel.FromDomain(provider)
	if tenantID, ok, err := authz.OptionalTenantOrStrict(ctx); err != nil {
		return 0, err
	} else if ok {
		model.TenantID = tenantID
	}
	providerId, err := ensureProvider(client, model)
	if err != nil {
		return 0, err
	}

	provider.ID = providerId
	if actorID, err := actorsync.SyncLegacyActor(client, actorsync.LegacyActorSync{
		SourceTable: actorsync.LegacyProviders,
		SourceID:    providerId,
		Name:        provider.Name,
		ActorKind:   actorsync.KindOrganization,
		Role:        actorsync.RoleProveedor,
		CreatedBy:   provider.CreatedBy,
		UpdatedBy:   provider.UpdatedBy,
	}); err != nil {
		return 0, err
	} else if actorID > 0 {
		provider.ActorID = &actorID
	}

	return providerId, nil
}

func ensureProvider(tx *gorm.DB, i *providermodel.Provider) (int64, error) {
	scope := tx
	if i.TenantID != uuid.Nil {
		scope = scope.Where("tenant_id = ?", i.TenantID)
	}
	if i.ID != 0 {
		var existing providermodel.Provider
		if err := scope.First(&existing, i.ID).Error; err == nil {
			return existing.ID, nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, fmt.Errorf("failed to check provider: %w", err)
		}
	}
	var existing providermodel.Provider
	if err := scope.Where("name = ?", i.Name).First(&existing).Error; err == nil {
		return existing.ID, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, fmt.Errorf("failed to check provider: %w", err)
	}

	if err := tx.Create(i).Error; err != nil {
		return 0, fmt.Errorf("failed to create provider: %w", err)
	}
	return i.ID, nil
}

func (r *Repository) GetEntriesSupplyMovementsByProjectID(ctx context.Context, projectId int64) ([]*domain.SupplyMovement, error) {
	db := r.getDB(ctx)

	var modelSupplyMovements []models.SupplyMovement

	if err := withSupplyMovementLookups(authz.MaybeTenantScope(ctx, db.Model(&models.SupplyMovement{}), "supply_movements")).
		Joins("JOIN stocks ON supply_movements.stock_id = stocks.id AND stocks.tenant_id = supply_movements.tenant_id").
		Joins("JOIN projects ON projects.id = stocks.project_id AND projects.tenant_id = supply_movements.tenant_id").
		Where("projects.id = ?", projectId).
		Where("is_entry = TRUE").
		Find(&modelSupplyMovements).
		Error; err != nil {
		return nil, domainerr.Internal("failed to list supplyEntriesMovement")
	}

	domainSupplyMovements := make([]*domain.SupplyMovement, len(modelSupplyMovements))
	for i, moddomainSupplyMovement := range modelSupplyMovements {
		domainSupplyMovements[i] = moddomainSupplyMovement.ToDomain()
	}

	if err := r.attachOriginsToMovements(ctx, domainSupplyMovements); err != nil {
		return nil, err
	}

	return domainSupplyMovements, nil
}

func (r *Repository) ListEntrySupplyMovements(ctx context.Context, filter domain.SupplyFilter) ([]*domain.SupplyMovement, error) {
	db := r.getDB(ctx)

	projectIDs, err := sharedfilters.ResolveProjectIDs(ctx, r.db.Client(), sharedfilters.WorkspaceFilter{
		CustomerID: filter.CustomerID,
		ProjectID:  filter.ProjectID,
		CampaignID: filter.CampaignID,
		FieldID:    filter.FieldID,
	})
	if err != nil {
		return nil, err
	}

	hasScope := filter.CustomerID != nil || filter.ProjectID != nil || filter.CampaignID != nil || filter.FieldID != nil
	if len(projectIDs) == 0 && hasScope {
		return []*domain.SupplyMovement{}, nil
	}

	query := withSupplyMovementLookups(authz.MaybeTenantScope(ctx, db.Model(&models.SupplyMovement{}), "supply_movements")).
		Joins("JOIN stocks ON supply_movements.stock_id = stocks.id AND stocks.tenant_id = supply_movements.tenant_id").
		Joins("JOIN projects ON projects.id = stocks.project_id AND projects.tenant_id = supply_movements.tenant_id").
		Where("is_entry = TRUE")

	if len(projectIDs) > 0 {
		query = query.Where("projects.id IN ?", projectIDs)
	}

	var modelSupplyMovements []models.SupplyMovement
	if err := query.Find(&modelSupplyMovements).Error; err != nil {
		return nil, domainerr.Internal("failed to list supplyEntriesMovement")
	}

	domainSupplyMovements := make([]*domain.SupplyMovement, len(modelSupplyMovements))
	for i, moddomainSupplyMovement := range modelSupplyMovements {
		domainSupplyMovements[i] = moddomainSupplyMovement.ToDomain()
	}

	if err := r.attachOriginsToMovements(ctx, domainSupplyMovements); err != nil {
		return nil, err
	}

	return domainSupplyMovements, nil
}

type movementOriginRow struct {
	MovementID           int64   `gorm:"column:movement_id"`
	OriginProjectID      *int64  `gorm:"column:origin_project_id"`
	OriginProject        *string `gorm:"column:origin_project_name"`
	DestinationProjectID *int64  `gorm:"column:destination_project_id"`
}

func (r *Repository) attachOriginsToMovements(ctx context.Context, movements []*domain.SupplyMovement) error {
	if len(movements) == 0 {
		return nil
	}

	tenantID, hasTenant := authz.TenantFromContext(ctx)
	if !hasTenant && authz.TenantStrictModeEnabled() {
		return domainerr.Forbidden("tenant context required")
	}

	ids := make([]int64, 0, len(movements))
	for i := range movements {
		ids = append(ids, movements[i].ID)
	}

	query := `
		SELECT
			sm_in.id AS movement_id,
			src.project_id AS origin_project_id,
			pj.name AS origin_project_name,
			CASE
				WHEN sm_in.movement_type = 'Movimiento interno entrada' THEN sm_in.project_id
				ELSE COALESCE(NULLIF(sm_in.project_destination_id, 0), dst.project_id)
			END AS destination_project_id
		FROM supply_movements sm_in
		LEFT JOIN LATERAL (
			SELECT sm_out.project_id
			FROM supply_movements sm_out
			WHERE sm_out.deleted_at IS NULL
			  AND sm_out.movement_type = 'Movimiento interno'
			  AND sm_out.reference_number = sm_in.reference_number
			  AND sm_out.movement_date = sm_in.movement_date
			  AND sm_out.investor_id = sm_in.investor_id
			  AND sm_out.provider_id = sm_in.provider_id
			  AND sm_out.quantity = (sm_in.quantity * -1)
			  AND sm_out.tenant_id = sm_in.tenant_id
			ORDER BY sm_out.id DESC
			LIMIT 1
		) src ON sm_in.movement_type = 'Movimiento interno entrada'
		LEFT JOIN LATERAL (
			SELECT sm_dst.project_id
			FROM supply_movements sm_dst
			WHERE sm_dst.deleted_at IS NULL
			  AND sm_dst.movement_type = 'Movimiento interno entrada'
			  AND sm_dst.reference_number = sm_in.reference_number
			  AND sm_dst.movement_date = sm_in.movement_date
			  AND sm_dst.investor_id = sm_in.investor_id
			  AND sm_dst.provider_id = sm_in.provider_id
			  AND sm_dst.quantity = (sm_in.quantity * -1)
			  AND sm_dst.tenant_id = sm_in.tenant_id
			ORDER BY sm_dst.id DESC
			LIMIT 1
		) dst ON sm_in.movement_type = 'Movimiento interno'
		LEFT JOIN projects pj ON pj.id = src.project_id AND pj.deleted_at IS NULL AND pj.tenant_id = sm_in.tenant_id
		WHERE sm_in.id IN ?
	`
	args := []any{ids}
	if hasTenant {
		query += " AND sm_in.tenant_id = ?"
		args = append(args, tenantID)
	}

	var rows []movementOriginRow
	if err := r.getDB(ctx).Raw(query, args...).Scan(&rows).Error; err != nil {
		return domainerr.Internal("failed to resolve origin project for supply movements")
	}

	originsByMovementID := make(map[int64]movementOriginRow, len(rows))
	for i := range rows {
		originsByMovementID[rows[i].MovementID] = rows[i]
	}

	for i := range movements {
		if row, ok := originsByMovementID[movements[i].ID]; ok {
			movements[i].OriginProjectID = row.OriginProjectID
			movements[i].OriginProjectName = row.OriginProject
			movements[i].DestinationProjectID = row.DestinationProjectID
		}
	}

	destinationIDs := make([]int64, 0, len(movements))
	seenDestination := make(map[int64]struct{}, len(movements))
	for i := range movements {
		if movements[i].DestinationProjectID == nil {
			continue
		}
		id := *movements[i].DestinationProjectID
		if _, seen := seenDestination[id]; seen {
			continue
		}
		seenDestination[id] = struct{}{}
		destinationIDs = append(destinationIDs, id)
	}

	destinationByID, err := r.getDestinationProjectMetadata(ctx, destinationIDs)
	if err != nil {
		return err
	}
	for i := range movements {
		if movements[i].DestinationProjectID == nil {
			continue
		}
		meta, ok := destinationByID[*movements[i].DestinationProjectID]
		if !ok {
			continue
		}
		movements[i].DestinationProject = meta.ProjectName
		movements[i].DestinationCustomer = meta.CustomerName
		movements[i].DestinationCampaign = meta.CampaignName
	}

	return nil
}

type destinationProjectMetadata struct {
	ProjectName  *string
	CustomerName *string
	CampaignName *string
}

type destinationProjectMetadataRow struct {
	ProjectID    int64   `gorm:"column:project_id"`
	ProjectName  *string `gorm:"column:project_name"`
	CustomerName *string `gorm:"column:customer_name"`
	CampaignName *string `gorm:"column:campaign_name"`
}

func (r *Repository) getDestinationProjectMetadata(
	ctx context.Context,
	projectIDs []int64,
) (map[int64]destinationProjectMetadata, error) {
	if len(projectIDs) == 0 {
		return map[int64]destinationProjectMetadata{}, nil
	}

	tenantID, hasTenant := authz.TenantFromContext(ctx)
	if !hasTenant && authz.TenantStrictModeEnabled() {
		return nil, domainerr.Forbidden("tenant context required")
	}

	query := `
		SELECT
			p.id AS project_id,
			p.name AS project_name,
			cu.name AS customer_name,
			ca.name AS campaign_name
		FROM projects p
		LEFT JOIN customers cu ON cu.id = p.customer_id AND cu.deleted_at IS NULL AND cu.tenant_id = p.tenant_id
		LEFT JOIN campaigns ca ON ca.id = p.campaign_id AND ca.deleted_at IS NULL AND ca.tenant_id = p.tenant_id
		WHERE p.id IN ? AND p.deleted_at IS NULL
	`
	args := []any{projectIDs}
	if hasTenant {
		query += " AND p.tenant_id = ?"
		args = append(args, tenantID)
	}

	var rows []destinationProjectMetadataRow
	if err := r.getDB(ctx).Raw(query, args...).Scan(&rows).Error; err != nil {
		return nil, domainerr.Internal("failed to resolve destination project metadata")
	}

	out := make(map[int64]destinationProjectMetadata, len(rows))
	for i := range rows {
		out[rows[i].ProjectID] = destinationProjectMetadata{
			ProjectName:  rows[i].ProjectName,
			CustomerName: rows[i].CustomerName,
			CampaignName: rows[i].CampaignName,
		}
	}

	return out, nil
}

func (r *Repository) GetSupplyMovementByID(ctx context.Context, id int64) (*domain.SupplyMovement, error) {
	db := r.getDB(ctx)

	var modelSupplyMovement models.SupplyMovement

	if err := withSupplyMovementLookups(authz.MaybeTenantScope(ctx, db, "supply_movements")).
		First(&modelSupplyMovement, "id = ?", id).
		Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerr.NotFound("supply movement not found")
		}
		return nil, domainerr.Internal("failed to get supply movement")
	}

	return modelSupplyMovement.ToDomain(), nil
}

func (r *Repository) UpdateSupplyMovement(ctx context.Context, movement *domain.SupplyMovement) error {
	if err := sharedrepo.ValidateEntity(movement, "supply movement"); err != nil {
		return err
	}

	model := models.SupplyMovementFromDomain(movement)
	if tenantID, ok := authz.TenantFromContext(ctx); ok {
		model.TenantID = tenantID
	}
	db := r.getDB(ctx)

	return db.Transaction(func(tx *gorm.DB) error {
		var previous models.SupplyMovement
		if err := authz.MaybeTenantScope(ctx, tx, "supply_movements").
			Where("id = ?", movement.ID).
			First(&previous).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.NotFound("supply movement not found")
			}
			return domainerr.Internal("failed to get supply movement")
		}

		if err := authz.MaybeTenantScope(ctx, tx.Model(&models.SupplyMovement{}), "supply_movements").
			Where("id = ?", movement.ID).
			Updates(model).
			Error; err != nil {
			return domainerr.Internal("failed to update supply movement")
		}
		if movement.Investor != nil && movement.Investor.ID > 0 && strings.TrimSpace(movement.Investor.Name) != "" {
			if _, err := actorsync.SyncLegacyActor(tx, actorsync.LegacyActorSync{
				SourceTable: actorsync.LegacyInvestors,
				SourceID:    movement.Investor.ID,
				Name:        movement.Investor.Name,
				ActorKind:   actorsync.KindUnknown,
				Role:        actorsync.RoleInversor,
				UpdatedAt:   time.Now(),
				UpdatedBy:   movement.UpdatedBy,
			}); err != nil {
				return err
			}
		}
		if movement.Provider != nil && movement.Provider.ID > 0 && strings.TrimSpace(movement.Provider.Name) != "" {
			if _, err := actorsync.SyncLegacyActor(tx, actorsync.LegacyActorSync{
				SourceTable: actorsync.LegacyProviders,
				SourceID:    movement.Provider.ID,
				Name:        movement.Provider.Name,
				ActorKind:   actorsync.KindOrganization,
				Role:        actorsync.RoleProveedor,
				UpdatedAt:   time.Now(),
				UpdatedBy:   movement.UpdatedBy,
			}); err != nil {
				return err
			}
		}
		if err := actorsync.RefreshSupplyMovementActorColumns(tx, movement.ID); err != nil {
			return err
		}

		if previous.StockId != 0 && previous.StockId != movement.StockId {
			var remainingCount int64
			if err := authz.MaybeTenantScope(ctx, tx.Model(&models.SupplyMovement{}), "supply_movements").
				Where("stock_id = ?", previous.StockId).
				Where("deleted_at IS NULL").
				Count(&remainingCount).Error; err != nil {
				return domainerr.Internal("failed to get remaining movements")
			}

			if remainingCount == 0 {
				if err := authz.MaybeTenantScope(ctx, tx, "stocks").Delete(&stockmodel.Stock{}, "id = ?", previous.StockId).Error; err != nil {
					return domainerr.Internal("failed to delete stock")
				}
			}
		}

		return nil
	})
}

func (r *Repository) DeleteSupplyMovement(ctx context.Context, projectId, supplyId int64) error {
	return r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		var supplyModel models.SupplyMovement

		// Obtener el movimiento a eliminar
		err := authz.MaybeTenantScope(ctx, tx, "supply_movements").
			Where("project_id = ?", projectId).
			Where("id = ?", supplyId).
			First(&supplyModel).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.NotFound("supply movement not found")
			}
			return err
		}

		// Verificar si es un movimiento interno
		isInternalMovement := supplyModel.MovementType == "Movimiento interno" ||
			supplyModel.MovementType == "Movimiento interno entrada"

		if isInternalMovement {
			// Buscar todos los registros relacionados del movimiento interno
			// Los registros relacionados comparten: movement_date, reference_number, supply_id, investor_id, provider_id
			var relatedMovements []models.SupplyMovement
			err := authz.MaybeTenantScope(ctx, tx, "supply_movements").Where("movement_date = ? AND reference_number = ? AND supply_id = ? AND investor_id = ? AND provider_id = ?",
				supplyModel.MovementDate,
				supplyModel.ReferenceNumber,
				supplyModel.SupplyID,
				supplyModel.InvestorID,
				supplyModel.ProviderID).
				Find(&relatedMovements).Error
			if err != nil {
				return domainerr.Internal("failed to find related movements")
			}

			// Recolectar todos los stock_ids afectados.
			affectedStocks := make(map[int64]bool)

			for _, mov := range relatedMovements {
				affectedStocks[mov.StockId] = true
			}

			// Eliminar todos los registros relacionados
			if err := authz.MaybeTenantScope(ctx, tx, "supply_movements").Where("movement_date = ? AND reference_number = ? AND supply_id = ? AND investor_id = ? AND provider_id = ?",
				supplyModel.MovementDate,
				supplyModel.ReferenceNumber,
				supplyModel.SupplyID,
				supplyModel.InvestorID,
				supplyModel.ProviderID).
				Delete(&models.SupplyMovement{}).Error; err != nil {
				return domainerr.Internal("failed to delete related supply movements")
			}

			// Eliminar stocks afectados solo si no tienen más movimientos.
			// Nota: `real_stock_units` representa "stock de campo" (recuento manual) y NO se recalcula
			// automáticamente a partir de movimientos.
			for stockID := range affectedStocks {
				var remainingMovements []models.SupplyMovement
				if err := authz.MaybeTenantScope(ctx, tx, "supply_movements").Where("stock_id = ?", stockID).
					Find(&remainingMovements).Error; err != nil {
					return domainerr.Internal("failed to get remaining movements")
				}

				if len(remainingMovements) == 0 {
					if err := authz.MaybeTenantScope(ctx, tx, "stocks").Delete(&stockmodel.Stock{}, "id = ?", stockID).Error; err != nil {
						return domainerr.Internal("failed to delete stock")
					}
				}
			}
		} else {
			// Movimiento normal (no interno)
			// Primero eliminar el movimiento
			if err := authz.MaybeTenantScope(ctx, tx, "supply_movements").Delete(&models.SupplyMovement{}, "project_id = ? AND id = ?", projectId, supplyId).Error; err != nil {
				return domainerr.Internal("failed to delete supply movement")
			}

			var remainingMovements []models.SupplyMovement
			if err := authz.MaybeTenantScope(ctx, tx, "supply_movements").Where("stock_id = ?", supplyModel.StockId).
				Find(&remainingMovements).Error; err != nil {
				return domainerr.Internal("failed to get remaining movements")
			}

			if len(remainingMovements) == 0 {
				if err := authz.MaybeTenantScope(ctx, tx, "stocks").Delete(&stockmodel.Stock{}, "id = ?", supplyModel.StockId).Error; err != nil {
					return domainerr.Internal("failed to delete stock")
				}
			}
		}

		return nil
	})
}

func (r *Repository) ListArchivedSupplyMovements(ctx context.Context, projectID int64) ([]*domain.SupplyMovement, error) {
	if projectID <= 0 {
		return nil, domainerr.Validation("project_id must be greater than 0")
	}

	var modelSupplyMovements []models.SupplyMovement
	if err := withSupplyMovementLookups(authz.MaybeTenantScope(ctx, r.getDB(ctx), "supply_movements")).
		Unscoped().
		Model(&models.SupplyMovement{}).
		Where("project_id = ?", projectID).
		Where("is_entry = TRUE").
		Where("deleted_at IS NOT NULL").
		Find(&modelSupplyMovements).
		Error; err != nil {
		return nil, domainerr.Internal("failed to list archived supply movements")
	}

	movements := make([]*domain.SupplyMovement, len(modelSupplyMovements))
	for i := range modelSupplyMovements {
		movements[i] = modelSupplyMovements[i].ToDomain()
	}

	if err := r.attachOriginsToMovements(ctx, movements); err != nil {
		return nil, err
	}
	return movements, nil
}

func (r *Repository) ArchiveSupplyMovement(ctx context.Context, projectID, movementID int64) error {
	if projectID <= 0 || movementID <= 0 {
		return domainerr.Validation("project_id and movement_id are required")
	}

	actor, err := sharedmodels.ActorFromContext(ctx)
	if err != nil {
		return err
	}
	deletedBy := &actor

	return r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		var movement models.SupplyMovement
		if err := authz.MaybeTenantScope(ctx, tx.Unscoped(), "supply_movements").
			Where("project_id = ? AND id = ?", projectID, movementID).
			First(&movement).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.NotFound("supply movement not found")
			}
			return domainerr.Internal("failed to get supply movement")
		}
		if movement.DeletedAt.Valid {
			return domainerr.Conflict("supply movement already archived")
		}
		archivedAt := time.Now()
		cause, err := lifecycle.RootCause(tx, movement.TenantID, "supply_movements", movementID, nil, deletedBy)
		if err != nil {
			return err
		}
		if err := authz.MaybeTenantScope(ctx, tx.Model(&models.SupplyMovement{}), "supply_movements").
			Where("project_id = ? AND id = ?", projectID, movementID).
			Updates(lifecycle.ArchiveUpdates(tx, "supply_movements", archivedAt, deletedBy, cause)).Error; err != nil {
			return domainerr.Internal("failed to archive supply movement")
		}
		return nil
	})
}

func (r *Repository) RestoreSupplyMovement(ctx context.Context, projectID, movementID int64) error {
	if projectID <= 0 || movementID <= 0 {
		return domainerr.Validation("project_id and movement_id are required")
	}

	return r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		var movement models.SupplyMovement
		if err := authz.MaybeTenantScope(ctx, tx.Unscoped(), "supply_movements").
			Where("project_id = ? AND id = ?", projectID, movementID).
			First(&movement).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.NotFound("supply movement not found")
			}
			return domainerr.Internal("failed to get supply movement")
		}
		if !movement.DeletedAt.Valid {
			return domainerr.Conflict("supply movement is not archived")
		}
		var projectActive int64
		if err := authz.MaybeTenantScope(ctx, tx.Table("projects"), "projects").
			Where("id = ? AND deleted_at IS NULL", projectID).
			Count(&projectActive).Error; err != nil {
			return domainerr.Internal("failed to check project")
		}
		if projectActive == 0 {
			return domainerr.Conflict("cannot restore supply movement while project is archived")
		}
		if err := authz.MaybeTenantScope(ctx, tx.Unscoped().Model(&models.SupplyMovement{}), "supply_movements").
			Where("project_id = ? AND id = ?", projectID, movementID).
			Updates(lifecycle.RestoreUpdates(tx, "supply_movements", time.Now())).Error; err != nil {
			return domainerr.Internal("failed to restore supply movement")
		}
		return nil
	})
}

func (r *Repository) HardDeleteSupplyMovement(ctx context.Context, projectID, movementID int64) error {
	if projectID <= 0 || movementID <= 0 {
		return domainerr.Validation("project_id and movement_id are required")
	}

	return r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		scoped := authz.MaybeTenantScope(ctx, tx.Unscoped().Table("supply_movements"), "supply_movements")
		var count int64
		if err := scoped.Where("project_id = ? AND id = ?", projectID, movementID).Count(&count).Error; err != nil {
			return domainerr.Internal("failed to check supply movement existence")
		}
		if count == 0 {
			return domainerr.NotFound("supply movement not found")
		}
		if err := lifecycle.RequireArchived(scoped.Where("project_id = ?", projectID), "supply_movements", "supply movement", movementID); err != nil {
			return err
		}
		if err := authz.MaybeTenantScope(ctx, tx.Unscoped(), "supply_movements").
			Where("project_id = ? AND id = ?", projectID, movementID).
			Delete(&models.SupplyMovement{}).Error; err != nil {
			return domainerr.Internal("failed to hard delete supply movement")
		}
		return nil
	})
}

func (r *Repository) GetProviders(ctx context.Context) ([]providerdomain.Provider, error) {
	db := r.getDB(ctx)
	if db.Name() == "sqlite" {
		var providers []providermodel.Provider
		if err := authz.MaybeTenantScope(ctx, db, "providers").Find(&providers).Error; err != nil {
			return nil, domainerr.Internal("failed to list providers")
		}
		res := make([]providerdomain.Provider, len(providers))
		for i := range providers {
			res[i] = *providers[i].ToDomain()
		}
		return res, nil
	}

	type providerRow struct {
		ID      int64
		Name    string
		ActorID *int64
	}

	var providers []providerRow
	if err := authz.MaybeTenantScope(ctx, db, "p").
		Table("providers p").
		Select("p.id, p.name, lm.actor_id").
		Joins("LEFT JOIN legacy_actor_map lm ON lm.source_table = 'providers' AND lm.source_id = p.id AND lm.tenant_id = p.tenant_id").
		Where("p.deleted_at IS NULL").
		Order("p.name ASC").
		Scan(&providers).Error; err != nil {
		return nil, domainerr.Internal("failed to list providers")
	}
	res := make([]providerdomain.Provider, len(providers))
	for i := range providers {
		res[i] = providerdomain.Provider{
			ID:      providers[i].ID,
			Name:    providers[i].Name,
			ActorID: providers[i].ActorID,
		}
	}
	return res, nil
}
