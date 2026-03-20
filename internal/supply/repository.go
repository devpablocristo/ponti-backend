package supply

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	investormodels "github.com/devpablocristo/ponti-backend/internal/investor/repository/models"
	investordomain "github.com/devpablocristo/ponti-backend/internal/investor/usecases/domain"
	providermodels "github.com/devpablocristo/ponti-backend/internal/provider/repository/models"
	providerdomain "github.com/devpablocristo/ponti-backend/internal/provider/usecases/domain"
	shareddb "github.com/devpablocristo/ponti-backend/internal/shared/db"
	sharedfilters "github.com/devpablocristo/ponti-backend/internal/shared/filters"
	sharedmodels "github.com/devpablocristo/ponti-backend/internal/shared/models"
	sharedrepo "github.com/devpablocristo/ponti-backend/internal/shared/repository"
	models "github.com/devpablocristo/ponti-backend/internal/supply/repository/models"
	domain "github.com/devpablocristo/ponti-backend/internal/supply/usecases/domain"
	workOrderModels "github.com/devpablocristo/ponti-backend/internal/work-order/repository/models"
	types "github.com/devpablocristo/ponti-backend/pkg/types"
	"gorm.io/gorm"
)

type GormEnginePort interface {
	Client() *gorm.DB
}

type Repository struct {
	db GormEnginePort
}

func NewRepository(db GormEnginePort) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ExecuteInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(shareddb.WithTx(ctx, tx))
	})
}

// --- CREATE ---
func (r *Repository) CreateSupply(ctx context.Context, s *domain.Supply) (int64, error) {
	var id int64
	err := r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		model := models.FromDomain(s)
		if err := tx.Create(model).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to create supply", err)
		}
		id = model.ID
		return nil
	})
	return id, err
}

func (r *Repository) CreateSuppliesBulk(ctx context.Context, supplies []domain.Supply) error {
	userID, err := sharedmodels.ActorFromContext(ctx)
	if err != nil {
		return err
	}
	return r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		modelsSlice := make([]*models.Supply, len(supplies))
		for i := range supplies {
			modelsSlice[i] = models.FromDomain(&supplies[i])
			modelsSlice[i].CreatedBy = &userID
		}
		if err := tx.Create(modelsSlice).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to bulk create supplies", err)
		}
		return nil
	})
}

// --- GET ---
func (r *Repository) GetSupply(ctx context.Context, id int64) (*domain.Supply, error) {
	var m models.Supply
	if err := r.getDB(ctx).
		Preload("Category").
		Preload("Type").
		First(&m, id).Error; err != nil {
		return nil, sharedrepo.HandleGormError(err, "supply", id)
	}
	return m.ToDomain(), nil
}

func (r *Repository) GetSuppliesByIDs(ctx context.Context, ids []int64) ([]domain.Supply, error) {
	if len(ids) == 0 {
		return []domain.Supply{}, nil
	}

	var rows []models.Supply
	if err := r.getDB(ctx).
		Preload("Category").
		Preload("Type").
		Where("id IN ?", ids).
		Find(&rows).Error; err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to get supplies by ids", err)
	}

	out := make([]domain.Supply, len(rows))
	for i := range rows {
		out[i] = *rows[i].ToDomain()
	}
	return out, nil
}

func (r *Repository) GetSupplyByProjectAndName(ctx context.Context, projectID int64, name string) (*domain.Supply, error) {
	normalizedName := strings.TrimSpace(name)
	if normalizedName == "" {
		return nil, types.NewError(types.ErrValidation, "supply name is empty", nil)
	}

	var m models.Supply
	err := r.getDB(ctx).
		Preload("Category").
		Preload("Type").
		Where("project_id = ?", projectID).
		Where("LOWER(TRIM(name)) = LOWER(TRIM(?))", normalizedName).
		First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.NewError(types.ErrNotFound, "supply not found", err)
		}
		return nil, types.NewError(types.ErrInternal, "failed to get supply by project and name", err)
	}

	return m.ToDomain(), nil
}

func (r *Repository) GetInvestor(ctx context.Context, id int64) (*investordomain.Investor, error) {
	var model investormodels.Investor
	err := r.getDB(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.NewError(types.ErrNotFound, fmt.Sprintf("investor with id %d not found", id), err)
		}
		return nil, types.NewError(types.ErrInternal, "failed to get investor", err)
	}
	return model.ToDomain(), nil
}

func (r *Repository) GetProvider(ctx context.Context, id int64) (*providerdomain.Provider, error) {
	var model providermodels.Provider
	err := r.getDB(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.NewError(types.ErrNotFound, fmt.Sprintf("provider with id %d not found", id), err)
		}
		return nil, types.NewError(types.ErrInternal, "failed to get provider", err)
	}
	return model.ToDomain(), nil
}

func (r *Repository) ProjectExists(ctx context.Context, projectID int64) (bool, error) {
	var count int64
	if err := r.getDB(ctx).
		Table("projects").
		Where("id = ? AND deleted_at IS NULL", projectID).
		Count(&count).Error; err != nil {
		return false, types.NewError(types.ErrInternal, "failed to check destination project", err)
	}
	return count > 0, nil
}

func (r *Repository) ExistsSupplyMovementByProjectReferenceAndSupply(
	ctx context.Context,
	projectID int64,
	reference string,
	supplyID int64,
) (bool, error) {
	var count int64
	if err := r.getDB(ctx).
		Model(&models.SupplyMovement{}).
		Where("project_id = ? AND reference_number = ? AND supply_id = ?", projectID, reference, supplyID).
		Count(&count).Error; err != nil {
		return false, types.NewError(types.ErrInternal, "failed to check duplicate supply movement", err)
	}
	return count > 0, nil
}

// --- UPDATE ---
func (r *Repository) UpdateSupply(ctx context.Context, s *domain.Supply) error {
	if err := sharedrepo.ValidateID(s.ID, "supply"); err != nil {
		return err
	}
	return r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.Supply{}).Where("id = ?", s.ID).Count(&count).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to check supply existence", err)
		}
		if count == 0 {
			return types.NewError(types.ErrNotFound, fmt.Sprintf("supply %d not found", s.ID), nil)
		}
		updates := map[string]any{
			"name":             s.Name,
			"unit_id":          int64(s.UnitID),
			"price":            s.Price,
			"is_partial_price": s.IsPartialPrice,
			"category_id":      int64(s.CategoryID),
			"type_id":          s.Type.ID,
			"project_id":       s.ProjectID,
			"updated_by":       s.UpdatedBy,
		}
		updateTx := tx.Model(&models.Supply{}).
			Where("id = ?", s.ID)
		if !s.UpdatedAt.IsZero() {
			updateTx = updateTx.Where("updated_at = ?", s.UpdatedAt)
		}
		result := updateTx.Updates(updates)
		if result.Error != nil {
			return types.NewError(types.ErrInternal, "failed to update supply", result.Error)
		}
		if result.RowsAffected == 0 {
			if !s.UpdatedAt.IsZero() {
				return types.NewError(types.ErrConflict, "supply not found or outdated", nil)
			}
			return types.NewError(types.ErrNotFound, fmt.Sprintf("supply %d not found", s.ID), nil)
		}
		return nil
	})
}

// --- DELETE ---
func (r *Repository) GetWorkOrdersBySupplyID(ctx context.Context, supplyID int64) (int64, error) {
	var count int64
	if err := r.getDB(ctx).
		Model(&workOrderModels.WorkOrder{}).
		Joins("JOIN workorder_items ON workorder_items.workorder_id = workorders.id").
		Where("workorder_items.supply_id = ? AND workorders.deleted_at IS NULL", supplyID).
		Count(&count).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		return 0, types.NewError(types.ErrInternal, "failed to get work order", err)
	}
	return count, nil
}

func (r *Repository) DeleteSupply(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "supply"); err != nil {
		return err
	}
	return r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Unscoped().Model(&models.Supply{}).Where("id = ?", id).Count(&count).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to check supply existence", err)
		}
		if count == 0 {
			return types.NewError(types.ErrNotFound, fmt.Sprintf("supply %d not found", id), nil)
		}
		result := tx.Unscoped().Delete(&models.Supply{}, id)
		if result.Error != nil {
			return types.NewError(types.ErrInternal, "failed to delete supply", result.Error)
		}
		if result.RowsAffected == 0 {
			return types.NewError(types.ErrNotFound, fmt.Sprintf("supply %d not found", id), nil)
		}
		return nil
	})
}

func (r *Repository) ArchiveSupply(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "supply"); err != nil {
		return err
	}
	return r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		var supply models.Supply
		if err := tx.Unscoped().Where("id = ?", id).First(&supply).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return types.NewError(types.ErrNotFound, fmt.Sprintf("supply %d not found", id), err)
			}
			return types.NewError(types.ErrInternal, "failed to get supply", err)
		}
		if supply.DeletedAt.Valid {
			return types.NewError(types.ErrConflict, "supply already archived", nil)
		}

		if err := tx.Model(&models.Supply{}).
			Where("id = ?", id).
			Updates(map[string]any{
				"deleted_at": time.Now(),
			}).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to archive supply", err)
		}
		return nil
	})
}

func (r *Repository) RestoreSupply(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "supply"); err != nil {
		return err
	}
	return r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		var supply models.Supply
		if err := tx.Unscoped().Where("id = ?", id).First(&supply).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return types.NewError(types.ErrNotFound, fmt.Sprintf("supply %d not found", id), err)
			}
			return types.NewError(types.ErrInternal, "failed to get supply", err)
		}
		if !supply.DeletedAt.Valid {
			return types.NewError(types.ErrConflict, "supply is not archived", nil)
		}

		if err := tx.Unscoped().Model(&models.Supply{}).
			Where("id = ?", id).
			Updates(map[string]any{
				"deleted_at": nil,
				"updated_at": time.Now(),
			}).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to restore supply", err)
		}
		return nil
	})
}

// --- LIST CENTRALIZADO, con filtros y paginación ---
func (r *Repository) ListSuppliesPaginated(
	ctx context.Context,
	filter domain.SupplyFilter,
	mode string,
	page, perPage int,
) ([]domain.Supply, int64, error) {
	var supplies []models.Supply
	var total int64

	db := r.getDB(ctx).Model(&models.Supply{}).
		Preload("Category").
		Preload("Type")

	// Filtrado flexible
	projectIDs, err := sharedfilters.ResolveProjectIDs(ctx, r.db.Client(), sharedfilters.WorkspaceFilter{
		CustomerID: filter.CustomerID,
		ProjectID:  filter.ProjectID,
		CampaignID: filter.CampaignID,
		FieldID:    filter.FieldID,
	})
	if err != nil {
		return nil, 0, err
	}
	if len(projectIDs) > 0 {
		db = db.Where("project_id IN ?", projectIDs)
	} else if filter.ProjectID != nil || filter.CustomerID != nil || filter.CampaignID != nil || filter.FieldID != nil {
		return []domain.Supply{}, 0, nil
	}

	// Total para paginación
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, types.NewError(types.ErrInternal, "failed to count supplies", err)
	}

	offset := (page - 1) * perPage
	if err := db.Offset(offset).Limit(perPage).Order("name").Find(&supplies).Error; err != nil {
		return nil, 0, types.NewError(types.ErrInternal, "failed to list supplies with filters", err)
	}

	res := make([]domain.Supply, len(supplies))
	for i := range supplies {
		res[i] = *supplies[i].ToDomain()
	}

	if err := r.attachOriginsToSupplies(ctx, res); err != nil {
		return nil, 0, err
	}

	return res, total, nil
}

func (r *Repository) UpdateSuppliesBulk(ctx context.Context, supplies []domain.Supply) error {
	return r.getDB(ctx).Transaction(func(tx *gorm.DB) error {
		for i := range supplies {
			updates := map[string]any{
				"name":             supplies[i].Name,
				"unit_id":          int64(supplies[i].UnitID),
				"price":            supplies[i].Price,
				"is_partial_price": supplies[i].IsPartialPrice,
				"category_id":      int64(supplies[i].CategoryID),
				"type_id":          supplies[i].Type.ID,
				"project_id":       supplies[i].ProjectID,
				"updated_by":       supplies[i].UpdatedBy,
			}
			updateTx := tx.Model(&models.Supply{}).
				Where("id = ?", supplies[i].ID)
			if !supplies[i].UpdatedAt.IsZero() {
				updateTx = updateTx.Where("updated_at = ?", supplies[i].UpdatedAt)
			}
			res := updateTx.Updates(updates)
			if res.Error != nil {
				return types.NewError(types.ErrInternal, fmt.Sprintf("failed to update supply id %d", supplies[i].ID), res.Error)
			}
			if res.RowsAffected == 0 {
				if !supplies[i].UpdatedAt.IsZero() {
					return types.NewError(types.ErrConflict, fmt.Sprintf("supply %d not found or outdated", supplies[i].ID), nil)
				}
				return types.NewError(types.ErrNotFound, fmt.Sprintf("supply %d not found", supplies[i].ID), nil)
			}
		}
		return nil
	})
}

func (r *Repository) ListAllSupplies(ctx context.Context, filter domain.SupplyFilter) ([]domain.Supply, int64, error) {
	base := r.getDB(ctx).Model(&models.Supply{})

	projectIDs, err := sharedfilters.ResolveProjectIDs(ctx, r.db.Client(), sharedfilters.WorkspaceFilter{
		CustomerID: filter.CustomerID,
		ProjectID:  filter.ProjectID,
		CampaignID: filter.CampaignID,
		FieldID:    filter.FieldID,
	})
	if err != nil {
		return nil, 0, err
	}
	if len(projectIDs) > 0 {
		base = base.Where("project_id IN ?", projectIDs)
	} else if filter.ProjectID != nil || filter.CustomerID != nil || filter.CampaignID != nil || filter.FieldID != nil {
		return []domain.Supply{}, 0, nil
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, types.NewError(types.ErrInternal, "failed to count supplies", err)
	}

	var rows []models.Supply
	db := base.
		Preload("Category").
		Preload("Type").
		Order("name")

	if err := db.Find(&rows).Error; err != nil {
		return nil, 0, types.NewError(types.ErrInternal, "failed to list all supplies", err)
	}

	out := make([]domain.Supply, len(rows))
	for i := range rows {
		out[i] = *rows[i].ToDomain()
	}

	if err := r.attachOriginsToSupplies(ctx, out); err != nil {
		return nil, 0, err
	}

	return out, total, nil
}

type supplyOriginRow struct {
	SupplyID        int64      `gorm:"column:supply_id"`
	Type            string     `gorm:"column:type"`
	SourceProjectID *int64     `gorm:"column:source_project_id"`
	SourceProject   *string    `gorm:"column:source_project"`
	MovementID      *int64     `gorm:"column:movement_id"`
	ReferenceNumber *string    `gorm:"column:reference_number"`
	ProviderName    *string    `gorm:"column:provider_name"`
	MovementDate    *time.Time `gorm:"column:movement_date"`
}

func (r *Repository) attachOriginsToSupplies(ctx context.Context, supplies []domain.Supply) error {
	if len(supplies) == 0 {
		return nil
	}

	supplyIDs := make([]int64, 0, len(supplies))
	for i := range supplies {
		supplyIDs = append(supplyIDs, supplies[i].ID)
	}

	query := `
		WITH latest AS (
			SELECT DISTINCT ON (sm.supply_id)
				sm.supply_id,
				sm.id AS movement_id,
				sm.movement_type,
				sm.reference_number,
				sm.movement_date,
				sm.investor_id,
				sm.provider_id,
				sm.quantity
			FROM supply_movements sm
			WHERE sm.deleted_at IS NULL
			  AND sm.is_entry = TRUE
			  AND sm.supply_id IN ?
			ORDER BY sm.supply_id, sm.movement_date DESC, sm.id DESC
		)
		SELECT
			l.supply_id,
			l.movement_type AS type,
			l.movement_id,
			l.reference_number,
			l.movement_date,
			pv.name AS provider_name,
			src.project_id AS source_project_id,
			pj.name AS source_project
		FROM latest l
		LEFT JOIN providers pv ON pv.id = l.provider_id
		LEFT JOIN LATERAL (
			SELECT sm_out.project_id
			FROM supply_movements sm_out
			WHERE sm_out.deleted_at IS NULL
			  AND sm_out.movement_type = 'Movimiento interno'
			  AND sm_out.reference_number = l.reference_number
			  AND sm_out.movement_date = l.movement_date
			  AND sm_out.investor_id = l.investor_id
			  AND sm_out.provider_id = l.provider_id
			  AND sm_out.quantity = (l.quantity * -1)
			ORDER BY sm_out.id DESC
			LIMIT 1
		) src ON l.movement_type = 'Movimiento interno entrada'
		LEFT JOIN projects pj ON pj.id = src.project_id AND pj.deleted_at IS NULL
	`

	var rows []supplyOriginRow
	if err := r.getDB(ctx).Raw(query, supplyIDs).Scan(&rows).Error; err != nil {
		return types.NewError(types.ErrInternal, "failed to resolve supply origins", err)
	}

	originBySupply := make(map[int64]*domain.SupplyOrigin, len(rows))
	for i := range rows {
		row := rows[i]
		origin := &domain.SupplyOrigin{
			Type:         row.Type,
			MovementID:   row.MovementID,
			MovementDate: row.MovementDate,
		}
		if row.ReferenceNumber != nil {
			origin.ReferenceNumber = *row.ReferenceNumber
		}
		if row.ProviderName != nil {
			origin.ProviderName = *row.ProviderName
		}
		origin.SourceProjectID = row.SourceProjectID
		if row.SourceProject != nil {
			origin.SourceProject = *row.SourceProject
		}
		originBySupply[row.SupplyID] = origin
	}

	for i := range supplies {
		supplies[i].Origin = originBySupply[supplies[i].ID]
	}

	return nil
}
