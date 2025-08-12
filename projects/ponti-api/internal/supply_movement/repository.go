package supply_movement

import (
	"context"
	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply_movement/repository/models"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply_movement/usecases/domain"
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

func (r *Repository) CreateSupplyMovement(ctx context.Context, movement *domain.SupplyMovement) (int64, error) {
	if movement == nil {
		return 0, types.NewError(types.ErrValidation, "supply movement is nil", nil)
	}
	model := models.FromDomain(movement)
	db := r.db.Client().WithContext(ctx)
	if err := db.Create(model).Error; err != nil {
		return 0, err
	}
	return model.ID, nil
}

func (r *Repository) GetEntriesSupplyMovementsByProjectID(ctx context.Context, projectId int64) ([]*domain.SupplyMovement, error){
	db := r.db.Client().WithContext(ctx)

	var modelSupplyMovements []models.SupplyMovement

	if err := db.
		Model(&models.SupplyMovement{}).
		Preload("Supply").
		Preload("Supply.Unit").
		Preload("Investor").
		Preload("Provider").
		Joins("JOIN stocks ON supply_movements.stock_id = stocks.id").
		Joins("JOIN projects ON projects.id = stocks.project_id").
		Where("projects.id = ?", projectId).
		Where("is_entry = TRUE").
		Find(&modelSupplyMovements).
		Error; err != nil{
			return nil, types.NewError(types.ErrInternal, "failed to list supplyEntriesMovement", err)
		}
	
	domainSupplyMovements := make([]*domain.SupplyMovement, len(modelSupplyMovements))
	for i, moddomainSupplyMovement := range modelSupplyMovements {
		domainSupplyMovements[i] = moddomainSupplyMovement.ToDomain()
	}

	return domainSupplyMovements, nil
}

func (r *Repository) GetSupplyMovementByID(ctx context.Context, id int64) (*domain.SupplyMovement, error) {
	db := r.db.Client().WithContext(ctx)

	var modelSupplyMovement models.SupplyMovement

	if err := db.
		Preload("Supply").
		Preload("Supply.Unit").
		Preload("Investor").
		Preload("Provider").
		First(&modelSupplyMovement, "id = ?", id).
		Error; err != nil {

		if err == gorm.ErrRecordNotFound {
			return nil, types.NewError(types.ErrNotFound, "supply movement not found", err)
		}
		return nil, types.NewError(types.ErrInternal, "failed to get supply movement", err)
	}

	return modelSupplyMovement.ToDomain(), nil
}

func (r *Repository) UpdateSupplyMovement(ctx context.Context, movement *domain.SupplyMovement) error {
	if movement == nil {
		return types.NewError(types.ErrValidation, "supply movement is nil", nil)
	}

	model := models.FromDomain(movement)
	db := r.db.Client().WithContext(ctx)

	if err := db.Model(&models.SupplyMovement{}).
		Where("id = ?", movement.ID).
		Updates(model). 
		Error; err != nil {

		return types.NewError(types.ErrInternal, "failed to update supply movement", err)
	}

	return nil
}

