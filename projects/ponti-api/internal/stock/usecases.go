package stock

import (
	"context"
	"time"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	shareddomain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/domain"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/stock/usecases/domain"
)

type RepositoryPort interface {
	GetStocks(context.Context, int64, time.Time) ([]*domain.Stock, error)
	CreateStock(context.Context, *domain.Stock) (int64, error)
	UpdateCloseDateByProject(context.Context, int64, *domain.Stock) error
	UpdateRealStockUnits(context.Context, int64, *domain.Stock) error
	GetStockById(context.Context, int64) (*domain.Stock, error)
	GetLastStockByProjectId(context.Context, int64, int64) (*domain.Stock, bool, error)
	GetStockByPeriodAndProjectId(context.Context, int64) (*domain.Stock, error)
	GetStocksPeriods(context.Context, int64) ([]string, error)
}

type UseCases struct {
	repo RepositoryPort
}

func NewUseCases(repo RepositoryPort) *UseCases {
	return &UseCases{repo: repo}
}

func (u *UseCases) GetStocksSummary(ctx context.Context, projectId int64, closeDate time.Time) ([]*domain.Stock, error) {
	return u.repo.GetStocks(ctx, projectId, closeDate)
}

func (u *UseCases) GetStocksPeriods(ctx context.Context, projectId int64) ([]string, error) {
	return u.repo.GetStocksPeriods(ctx, projectId)
}

func (u *UseCases) CreateStock(ctx context.Context, s *domain.Stock) (int64, error) {
	return u.repo.CreateStock(ctx, s)
}

func (u *UseCases) UpdateCloseDateByProject(ctx context.Context, projectId int64, monthPeriod int64, yearPeriod int64, stock *domain.Stock) error {
	stockFromDb, err := u.repo.GetStockByPeriodAndProjectId(ctx, projectId)
	if err != nil {
		return err
	}

	err = u.repo.UpdateCloseDateByProject(ctx, projectId, stock)
	if err != nil {
		return err
	}

	newStock := createNewStockPeriod(*stock.UpdatedBy, monthPeriod, yearPeriod, stockFromDb)
	_, err = u.repo.CreateStock(ctx, &newStock)

	if err != nil {
		return err
	}
	return nil
}

func (u *UseCases) UpdateRealStockUnits(ctx context.Context, stockId int64, stock *domain.Stock) error {
	return u.repo.UpdateRealStockUnits(ctx, stockId, stock)
}

func (u *UseCases) GetStockById(ctx context.Context, stockId int64) (*domain.Stock, error) {
	if stockId <= 0 {
		return nil, types.NewError(types.ErrInvalidInput, "stock id must be greater than 0", nil)
	}
	return u.repo.GetStockById(ctx, stockId)
}

func (u *UseCases) GetLastStockByProjectId(ctx context.Context, projectId int64, supplyId int64) (*domain.Stock, bool, error) {
	return u.repo.GetLastStockByProjectId(ctx, projectId, supplyId)
}

func createNewStockPeriod(userId int64, monthPeriod int64, yearPeriod int64, stock *domain.Stock) domain.Stock {
	newMonthPeriod, newYearPeriod := startNewStockPeriod(monthPeriod, yearPeriod)
	newStock := domain.Stock{
		Project:        stock.Project,
		YearPeriod:     newYearPeriod,
		MonthPeriod:    newMonthPeriod,
		Supply:         stock.Supply,
		Investor:       stock.Investor,
		InitialStock:   stock.RealStockUnits,
		RealStockUnits: stock.RealStockUnits,
		Base: shareddomain.Base{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			CreatedBy: &userId,
			UpdatedBy: &userId,
		},
	}
	return newStock
}

func startNewStockPeriod(monthPeriod int64, yearPeriod int64) (int64, int64) {
	var newMonthPeriod int64
	var newYearPeriod int64

	if monthPeriod == 12 {
		newMonthPeriod = 1
		newYearPeriod = yearPeriod + 1
	} else {
		newMonthPeriod = monthPeriod + 1
		newYearPeriod = yearPeriod
	}
	return newMonthPeriod, newYearPeriod

}
