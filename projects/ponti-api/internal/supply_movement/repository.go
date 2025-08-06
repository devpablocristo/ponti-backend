package supply_movement

import (
	"context"
	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply_movement/repository/models"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply_movement/usecases/domain"
	"gorm.io/gorm"
	"time"
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

// GetSupplyMovements returns supply movements filtered by project, supply, and date range
func (r *Repository) GetSupplyMovements(ctx context.Context, projectId int64, supplyId int64, fromDate, toDate time.Time) ([]*domain.SupplyMovement, error) {
	db := r.db.Client().WithContext(ctx)
	query := db.Model(&models.SupplyMovement{}).
		Preload("Supply").
		Preload("Investor").
		Preload("Provider").
		Where("project_id = ?", projectId)
	if supplyId != 0 {
		query = query.Where("supply_id = ?", supplyId)
	}
	var movementModels []models.SupplyMovement
	if err := query.Find(&movementModels).Error; err != nil {
		return nil, err
	}
	movements := make([]*domain.SupplyMovement, 0, len(movementModels))
	for i := range movementModels {
		movements = append(movements, movementModels[i].ToDomain())
	}
	return movements, nil
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

func (r *Repository) GetSupplyMovementById(ctx context.Context, id int64) (*domain.SupplyMovement, error) {
	db := r.db.Client().WithContext(ctx)
	var model models.SupplyMovement
	if err := db.Preload("Supply").Preload("Investor").Preload("Provider").First(&model, id).Error; err != nil {
		return nil, err
	}
	return model.ToDomain(), nil
}
