package investor

import (
	"context"

	domain "github.com/devpablocristo/ponti-backend/internal/investor/usecases/domain"
)

type RepositoryPort interface {
	CreateInvestor(context.Context, *domain.Investor) (int64, error)
	ListInvestors(context.Context, int, int) ([]domain.Investor, int64, error)
	ListArchivedInvestors(context.Context, int, int) ([]domain.Investor, int64, error)
	GetInvestor(context.Context, int64) (*domain.Investor, error)
	UpdateInvestor(context.Context, *domain.Investor) error
	ArchiveInvestor(context.Context, int64) error
	RestoreInvestor(context.Context, int64) error
	HardDeleteInvestor(context.Context, int64) error
}

type UseCases struct {
	repo RepositoryPort
}

func NewUseCases(repo RepositoryPort) *UseCases {
	return &UseCases{repo: repo}
}

func (u *UseCases) CreateInvestor(ctx context.Context, inv *domain.Investor) (int64, error) {
	return u.repo.CreateInvestor(ctx, inv)
}

func (u *UseCases) ListInvestors(ctx context.Context, page, perPage int) ([]domain.Investor, int64, error) {
	return u.repo.ListInvestors(ctx, page, perPage)
}

func (u *UseCases) ListArchivedInvestors(ctx context.Context, page, perPage int) ([]domain.Investor, int64, error) {
	return u.repo.ListArchivedInvestors(ctx, page, perPage)
}

func (u *UseCases) GetInvestor(ctx context.Context, id int64) (*domain.Investor, error) {
	return u.repo.GetInvestor(ctx, id)
}

func (u *UseCases) UpdateInvestor(ctx context.Context, inv *domain.Investor) error {
	return u.repo.UpdateInvestor(ctx, inv)
}

func (u *UseCases) ArchiveInvestor(ctx context.Context, id int64) error {
	return u.repo.ArchiveInvestor(ctx, id)
}

func (u *UseCases) RestoreInvestor(ctx context.Context, id int64) error {
	return u.repo.RestoreInvestor(ctx, id)
}

func (u *UseCases) HardDeleteInvestor(ctx context.Context, id int64) error {
	return u.repo.HardDeleteInvestor(ctx, id)
}
