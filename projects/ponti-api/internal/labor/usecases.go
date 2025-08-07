package labor

import (
	"context"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/labor/usecases/domain"
	"github.com/shopspring/decimal"
)

type RepositoryPort interface {
	CreateLabor(context.Context, *domain.Labor) (int64, error)
	ListLabor(context.Context, int, int, int64) ([]domain.ListedLabor, int64, error)
	deleteLabor(context.Context, int64) error
	UpdateLabor(context.Context, *domain.Labor) error
	ListLaborCategoriesByTypeId(context.Context, int64) ([]domain.LaborCategory, error)
	ListByWorkorder(context.Context, int64) ([]domain.LaborRawItem, error)
}

type UseCases struct {
	repo RepositoryPort
}

func NewUseCases(repo RepositoryPort) *UseCases {
	return &UseCases{repo: repo}
}

func (u *UseCases) CreateLabor(ctx context.Context, labor *domain.Labor) (int64, error) {
	return u.repo.CreateLabor(ctx, labor)
}

func (u *UseCases) ListLabor(ctx context.Context, page, perPage int, projectId int64) ([]domain.ListedLabor, int64, error) {
	return u.repo.ListLabor(ctx, page, perPage, projectId)
}

func (u *UseCases) DeleteLabor(ctx context.Context, laborId int64) error {
	return u.repo.deleteLabor(ctx, laborId)
}

func (u *UseCases) UpdateLabor(ctx context.Context, labor *domain.Labor) error {
	return u.repo.UpdateLabor(ctx, labor)
}

func (u *UseCases) ListLaborCategoriesByTypeId(ctx context.Context, typeId int64) ([]domain.LaborCategory, error) {
	return u.repo.ListLaborCategoriesByTypeId(ctx, typeId)
}

func (u *UseCases) ListLaborByWorkorder(ctx context.Context, workorderID int64) ([]domain.LaborListItem, error) {
	raws, err := u.repo.ListByWorkorder(ctx, workorderID)
	if err != nil {
		return nil, types.NewError(types.ErrInternal, "internal error", err)
	}

	var out []domain.LaborListItem
	for _, r := range raws {
		totalNet := r.CostHa.Mul(r.SurfaceHa)                // costoHa * superficie = netprice
		totalIVA := totalNet.Mul(decimal.NewFromFloat(0.21)) // netprice * 21%  = totalIva

		out = append(out, domain.LaborListItem{
			WorkorderNumber: r.WorkorderNumber,
			Date:            r.Date,
			ProjectName:     r.ProjectName,
			FieldName:       r.FieldName,
			CropName:        r.CropName,
			CategoryName:    r.CategoryName,
			Contractor:      r.Contractor,
			SurfaceHa:       r.SurfaceHa,
			CostHa:          r.CostHa,
			InvestorName:    r.InvestorName,
			NetTotal:        totalNet,
			TotalIVA:        totalIVA,
		})
	}

	return out, err
}
