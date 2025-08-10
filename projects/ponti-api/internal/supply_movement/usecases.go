package supply_movement


import (
	"context"
	fielddom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field/usecases/domain"
	projdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project/usecases/domain"
	stockUseCases "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/stock"
	stockdomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/stock/usecases/domain"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply_movement/usecases/domain"
	"time"
)

type RepositoryPort interface {
	GetEntriesSupplyMovementsByProjectID(context.Context, int64) ([]*domain.SupplyMovement, error)
	CreateSupplyMovement(context.Context, *domain.SupplyMovement) (int64, error)
	GetSupplyMovementByID(context.Context, int64) (*domain.SupplyMovement, error)
	UpdateSupplyMovement(context.Context, *domain.SupplyMovement) error

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

func (u *UseCases) CreateSupplyMovement(ctx context.Context, movement *domain.SupplyMovement) (int64, error) {
	stock, isFirst, err := u.stockUseCases.GetLastStockByProjectIdAndFieldId(ctx, movement.ProjectId, movement.FieldId, movement.Supply.ID)
	if err != nil {
		return 0, err
	}
	if isFirst {
		stock = createStockDomainFromSupplyMovement(movement)
		stockId, err := u.stockUseCases.CreateStock(ctx, stock)
		if err != nil {
			return 0, err
		}
		stock.ID = stockId
	}

	movement.StockId = stock.ID

	stock.RealStockUnits =+ movement.Quantity

	err = u.stockUseCases.UpdateRealStockUnits(ctx, stock.ID, stock)
	if err != nil {
		return 0, err
	}

	return u.repo.CreateSupplyMovement(ctx, movement)
}

func (u *UseCases) GetEntriesSupplyMovementsByProjectID(ctx context.Context, projectId int64) ([]*domain.SupplyMovement, error){
	return u.repo.GetEntriesSupplyMovementsByProjectID(ctx, projectId)
}

func (u *UseCases) UpdateSupplyMovement(ctx context.Context, supplyMovement *domain.SupplyMovement) error {
	return u.repo.UpdateSupplyMovement(ctx, supplyMovement)
}

func (u *UseCases) GetSupplyMovementByID(ctx context.Context, id int64) (*domain.SupplyMovement, error){
	return u.repo.GetSupplyMovementByID(ctx, id)
}


func createStockDomainFromSupplyMovement(supplyMovement *domain.SupplyMovement) *stockdomain.Stock {
	return &stockdomain.Stock{
		Project: &projdom.Project{
			ID: supplyMovement.ProjectId,
		},
		Field: &fielddom.Field{
			ID: supplyMovement.FieldId,
		},
		Supply:       supplyMovement.Supply,
		Investor:     supplyMovement.Investor,
		CloseDate:    nil,
		InitialStock: supplyMovement.Quantity,
		YearPeriod:   int64(time.Now().Year()),
		MonthPeriod:  int64(time.Now().Month()),
		Base:         supplyMovement.Base,
	}
}
