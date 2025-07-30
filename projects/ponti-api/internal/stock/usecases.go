package stock

import (
	"context"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/stock/usecases/domain"
	"time"
)

type RepositoryPort interface {
	GetStocks(context.Context, int64, int64, int64, int64, time.Time) ([]*domain.Stock, error)
	CreateStock(context.Context, *domain.Stock) (int64, error)
	UpdateCloseDateByProjectAndField(context.Context, int64, int64, int64, int64, *domain.Stock) error
	UpdateRealStockUnits(context.Context, int64, *domain.Stock) error
	GetStockById(context.Context, int64) (*domain.Stock, error)
}

type UseCases struct {
	repo RepositoryPort
}

func NewUseCases(repo RepositoryPort) *UseCases {
	return &UseCases{repo: repo}
}

func (u *UseCases) GetStocksSummary(ctx context.Context, projectId int64, fieldId int64, monthPeriod int64, yearPeriod int64, closeDate time.Time) ([]*domain.Stock, error) {
	return u.repo.GetStocks(ctx, projectId, fieldId, monthPeriod, yearPeriod, closeDate)
}

func (u *UseCases) CreateStock(ctx context.Context, s *domain.Stock) (int64, error) {
	return u.repo.CreateStock(ctx, s)
}

func (u *UseCases) UpdateCloseDateByProjectAndField(ctx context.Context, projectId int64, fieldId int64, monthPeriod int64, yearPeriod int64, stock *domain.Stock) error {
	return u.repo.UpdateCloseDateByProjectAndField(ctx, projectId, fieldId, monthPeriod, yearPeriod, stock)
}

func (u *UseCases) UpdateRealStockUnits(ctx context.Context, stockId int64, stock *domain.Stock) error {
	return u.repo.UpdateRealStockUnits(ctx, stockId, stock)
}

func (u *UseCases) GetStockById(ctx context.Context, stockId int64) (*domain.Stock, error) {
	return u.repo.GetStockById(ctx, stockId)
}
