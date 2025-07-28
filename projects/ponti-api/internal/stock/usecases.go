package stock

import (
	"context"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/stock/usecases/domain"
	"time"
)

type RepositoryPort interface {
	GetStocks(context.Context, int64, int64, time.Time) ([]*domain.Stock, error)
	CreateStock(context.Context, *domain.Stock) (int64, error)
}

type UseCases struct {
	repo RepositoryPort
}

func NewUseCases(repo RepositoryPort) *UseCases {
	return &UseCases{repo: repo}
}

func (u *UseCases) GetStocksSummary(ctx context.Context, projectId int64, fieldId int64, closeDate time.Time) ([]*domain.Stock, error) {
	return u.repo.GetStocks(ctx, projectId, fieldId, closeDate)
}

func (u *UseCases) CreateStock(ctx context.Context, s *domain.Stock) (int64, error) {
	return u.repo.CreateStock(ctx, s)
}
