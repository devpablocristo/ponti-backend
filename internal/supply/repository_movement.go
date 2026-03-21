package supply

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/devpablocristo/core/saas/go/shared/domainerr"
	providermodel "github.com/devpablocristo/ponti-backend/internal/provider/repository/models"
	providerdomain "github.com/devpablocristo/ponti-backend/internal/provider/usecases/domain"
	sharedrepo "github.com/devpablocristo/ponti-backend/internal/shared/repository"
	stockmodel "github.com/devpablocristo/ponti-backend/internal/stock/repository/models"
	"github.com/devpablocristo/ponti-backend/internal/supply/repository/models"
	"github.com/devpablocristo/ponti-backend/internal/supply/usecases/domain"
	"gorm.io/gorm"
)

func (r *Repository) CreateSupplyMovement(ctx context.Context, movement *domain.SupplyMovement) (int64, error) {
	if err := sharedrepo.ValidateEntity(movement, "supply movement"); err != nil {
		return 0, err
	}
	model := models.SupplyMovementFromDomain(movement)
	db := r.getDB(ctx)
	if err := db.Create(model).Error; err != nil {
		return 0, err
	}
	return model.ID, nil
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
	providerId, err := ensureProvider(client, model)
	if err != nil {
		return 0, err
	}

	provider.ID = providerId

	return providerId, nil
}

func ensureProvider(tx *gorm.DB, i *providermodel.Provider) (int64, error) {
	if i.ID != 0 {
		var existing providermodel.Provider
		if err := tx.First(&existing, i.ID).Error; err == nil {
			return existing.ID, nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, fmt.Errorf("failed to check investor: %w", err)
		}
	}
	var existing providermodel.Provider
	if err := tx.Where("name = ?", i.Name).First(&existing).Error; err == nil {
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

	if err := db.
		Model(&models.SupplyMovement{}).
		Preload("Supply", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("Supply.Type").
		Preload("Supply.Category").
		Preload("Investor").
		Preload("Provider").
		Joins("JOIN stocks ON supply_movements.stock_id = stocks.id").
		Joins("JOIN projects ON projects.id = stocks.project_id").
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
			ORDER BY sm_dst.id DESC
			LIMIT 1
		) dst ON sm_in.movement_type = 'Movimiento interno'
		LEFT JOIN projects pj ON pj.id = src.project_id AND pj.deleted_at IS NULL
		WHERE sm_in.id IN ?
	`

	var rows []movementOriginRow
	if err := r.getDB(ctx).Raw(query, ids).Scan(&rows).Error; err != nil {
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

	const query = `
		SELECT
			p.id AS project_id,
			p.name AS project_name,
			cu.name AS customer_name,
			ca.name AS campaign_name
		FROM projects p
		LEFT JOIN customers cu ON cu.id = p.customer_id AND cu.deleted_at IS NULL
		LEFT JOIN campaigns ca ON ca.id = p.campaign_id AND ca.deleted_at IS NULL
		WHERE p.id IN ? AND p.deleted_at IS NULL
	`

	var rows []destinationProjectMetadataRow
	if err := r.getDB(ctx).Raw(query, projectIDs).Scan(&rows).Error; err != nil {
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

	if err := db.
		Preload("Supply", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Preload("Supply.Type").
		Preload("Supply.Category").
		Preload("Investor").
		Preload("Provider").
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
	db := r.getDB(ctx)

	if err := db.Model(&models.SupplyMovement{}).
		Where("id = ?", movement.ID).
		Updates(model).
		Error; err != nil {

		return domainerr.Internal("failed to update supply movement")
	}

	return nil
}

func (r *Repository) DeleteSupplyMovement(ctx context.Context, projectId, supplyId int64) error {
	return r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		var stockModel stockmodel.Stock
		var supplyModel models.SupplyMovement

		// Obtener el movimiento a eliminar
		err := tx.
			Where("project_id = ?", projectId).
			Where("id = ?", supplyId).
			First(&supplyModel).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.NotFound("supply movement not found")
			}
			return err
		}

		// Verificar si hay stock cerrado
		err = tx.
			Where("project_id = ?", projectId).
			Where("supply_id = ?", supplyModel.SupplyID).
			Where("close_date IS NOT NULL").
			First(&stockModel).Error
		if err == nil {
			return domainerr.Conflict("closed stock movement already exists for this supply in the project")
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		// Verificar si es un movimiento interno
		isInternalMovement := supplyModel.MovementType == "Movimiento interno" ||
			supplyModel.MovementType == "Movimiento interno entrada"

		if isInternalMovement {
			// Buscar todos los registros relacionados del movimiento interno
			// Los registros relacionados comparten: movement_date, reference_number, supply_id, investor_id, provider_id
			var relatedMovements []models.SupplyMovement
			err := tx.Where("movement_date = ? AND reference_number = ? AND supply_id = ? AND investor_id = ? AND provider_id = ?",
				supplyModel.MovementDate,
				supplyModel.ReferenceNumber,
				supplyModel.SupplyID,
				supplyModel.InvestorID,
				supplyModel.ProviderID).
				Find(&relatedMovements).Error
			if err != nil {
				return domainerr.Internal("failed to find related movements")
			}

			// Recolectar todos los project_ids y stock_ids afectados
			affectedProjects := make(map[int64]bool)
			affectedStocks := make(map[int64]bool)

			for _, mov := range relatedMovements {
				affectedProjects[mov.ProjectId] = true
				affectedStocks[mov.StockId] = true
			}

			// Eliminar todos los registros relacionados
			if err := tx.Where("movement_date = ? AND reference_number = ? AND supply_id = ? AND investor_id = ? AND provider_id = ?",
				supplyModel.MovementDate,
				supplyModel.ReferenceNumber,
				supplyModel.SupplyID,
				supplyModel.InvestorID,
				supplyModel.ProviderID).
				Delete(&models.SupplyMovement{}).Error; err != nil {
				return domainerr.Internal("failed to delete related supply movements")
			}

			// Eliminar stocks de todos los proyectos afectados solo si no tienen más movimientos.
			// Nota: `real_stock_units` representa "stock de campo" (recuento manual) y NO se recalcula
			// automáticamente a partir de movimientos.
			for projectIDAffected := range affectedProjects {
				// Obtener movimientos restantes para recalcular el stock
				var remainingMovements []models.SupplyMovement
				if err := tx.Where("project_id = ? AND supply_id = ?", projectIDAffected, supplyModel.SupplyID).
					Find(&remainingMovements).Error; err != nil {
					return domainerr.Internal("failed to get remaining movements")
				}

				if len(remainingMovements) == 0 {
					// No quedan movimientos, eliminar el stock
					if err := tx.Delete(&stockmodel.Stock{}, "project_id = ? AND supply_id = ?", projectIDAffected, supplyModel.SupplyID).Error; err != nil {
						return domainerr.Internal("failed to delete stock")
					}
				}
			}
		} else {
			// Movimiento normal (no interno)
			// Primero eliminar el movimiento
			if err := tx.Delete(&models.SupplyMovement{}, "project_id = ? AND id = ?", projectId, supplyId).Error; err != nil {
				return domainerr.Internal("failed to delete supply movement")
			}

			// Obtener movimientos restantes para recalcular el stock
			var remainingMovements []models.SupplyMovement
			if err := tx.Where("project_id = ? AND supply_id = ?", projectId, supplyModel.SupplyID).
				Find(&remainingMovements).Error; err != nil {
				return domainerr.Internal("failed to get remaining movements")
			}

			if len(remainingMovements) == 0 {
				// No quedan movimientos, eliminar el stock
				if err := tx.Delete(&stockmodel.Stock{}, "project_id = ? AND supply_id = ?", projectId, supplyModel.SupplyID).Error; err != nil {
					return domainerr.Internal("failed to delete stock")
				}
			}
		}

		return nil
	})
}

func (r *Repository) GetProviders(ctx context.Context) ([]providerdomain.Provider, error) {
	var providers []providermodel.Provider
	if err := r.getDB(ctx).Find(&providers).Error; err != nil {
		return nil, domainerr.Internal("failed to list providers")
	}
	res := make([]providerdomain.Provider, len(providers))
	for i := range providers {
		res[i] = *providers[i].ToDomain()
	}
	return res, nil
}
