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
	ListByWorkorder(context.Context, int64, string) ([]domain.LaborRawItem, error)
	ListGroupLabor(context.Context, types.Input, int64, int64, string) ([]domain.LaborRawItem, types.PageInfo, error)
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

func (u *UseCases) ListLaborByWorkorder(ctx context.Context, workorderID int64, usdMonth string) ([]domain.LaborRawItem, error) {
	return u.repo.ListByWorkorder(ctx, workorderID, usdMonth)
}

func (u *UseCases) ListGroupLaborByWorkorder(ctx context.Context, inp types.Input, projectID int64, fieldID int64, usdMonth string) ([]domain.LaborListItem, types.PageInfo, error) {
	rawItems, pageInfo, err := u.repo.ListGroupLabor(ctx, inp, projectID, fieldID, usdMonth)

	items := make([]domain.LaborListItem, len(rawItems))
	for i, r := range rawItems {
		var usdCostHa, usdTotalNet decimal.Decimal
		netTotal := r.CostHa.Mul(r.SurfaceHa)
		totalIVA := netTotal.Mul(decimal.NewFromFloat(0.21))

		if !r.USDAvgValue.IsZero() {
			usdCostHa = r.CostHa.Div(r.USDAvgValue)
			usdTotalNet = netTotal.Div(r.USDAvgValue)
		}

		items[i] = domain.LaborListItem{
			WorkorderID:     r.WorkorderID,
			WorkorderNumber: r.WorkorderNumber,
			Date:            r.Date,
			ProjectName:     r.ProjectName,
			FieldName:       r.FieldName,
			CropName:        r.CropName,
			LaborName:       r.LaborName,
			Contractor:      r.Contractor,
			SurfaceHa:       r.SurfaceHa,
			CostHa:          r.CostHa,
			CategoryName:    r.CategoryName,
			InvestorName:    r.InvestorName,
			USDAvgValue:     r.USDAvgValue,
			NetTotal:        netTotal.Round(2),
			TotalIVA:        totalIVA.Round(2),
			USDCostHa:       usdCostHa.Round(2),
			USDNetTotal:     usdTotalNet.Round(2),
			InvoiceID:       r.InvoiceID,
			InvoiceNumber:   r.InvoiceNumber,
			InvoiceCompany:  r.InvoiceCompany,
			InvoiceDate:     r.InvoiceDate,
			InvoiceStatus:   r.InvoiceStatus,
		}
	}

	return items, pageInfo, err
}
