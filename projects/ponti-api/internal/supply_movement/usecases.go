package supply_movement

// Usecases for supply_movement operations

import (
	"context"
	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	stockUseCases "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/stock"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply_movement/usecases/domain"
	stockdomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/stock/usecases/domain"
	"time"
)

type RepositoryPort interface {
	GetSupplyMovements(context.Context, int64, int64, time.Time, time.Time) ([]*domain.SupplyMovement, error)
	CreateSupplyMovement(context.Context, *domain.SupplyMovement) (int64, error)
	GetSupplyMovementById(context.Context, int64) (*domain.SupplyMovement, error)
}

type UseCases struct {
	repo          RepositoryPort
	stockUseCases stockUseCases.UseCasesPort
}

func NewUseCases(
	repo RepositoryPort,
	stockUseCases stockUseCases.UseCasesPort) *UseCases {
	return &UseCases{repo: repo, stockUseCases: stockUseCases}
}

func (u *UseCases) GetSupplyMovements(ctx context.Context, projectId int64, supplyId int64, fromDate, toDate time.Time) ([]*domain.SupplyMovement, error) {
	return u.repo.GetSupplyMovements(ctx, projectId, supplyId, fromDate, toDate)
}

func (u *UseCases) CreateSupplyMovement(ctx context.Context, m *domain.SupplyMovement) (int64, error) {
	stock, isFirst, error := u.stockUseCases.GetLastStockByProjectIdAndFieldId(ctx, m.ProjectId, m.FieldId)
	if error != nil {
		return 0, error
	}
	if isFirst{
		stockId, err := u.stockUseCases.CreateStock(ctx, )
	}

	m.StockId = stock.ID
	return u.repo.CreateSupplyMovement(ctx, m)
}

func (u *UseCases) GetSupplyMovementById(ctx context.Context, id int64) (*domain.SupplyMovement, error) {
	if id <= 0 {
		return nil, types.NewError(types.ErrInvalidInput, "supply movement id must be greater than 0", nil)
	}
	return u.repo.GetSupplyMovementById(ctx, id)
}

func createStockFromSupplyMovement(supplyMovement *domain.SupplyMovement) *stockdomain.Stock {
	return 