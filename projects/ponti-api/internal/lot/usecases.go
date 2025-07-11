package lot

import (
	"context"

	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/usecases/domain"
)

type RepositoryPort interface {
	CreateLot(context.Context, *domain.Lot) (int64, error)
	ListLotsByField(context.Context, int64) ([]domain.Lot, error)
	ListLotsByProject(context.Context, int64) ([]domain.Lot, error)
	ListLotsByProjectAndField(context.Context, int64, int64) ([]domain.Lot, error)
	ListLotsByProjectFieldAndCrop(context.Context, int64, int64, int64, string) ([]domain.Lot, error)
	GetLot(context.Context, int64) (*domain.Lot, error)
	UpdateLot(context.Context, *domain.Lot) error
	DeleteLot(context.Context, int64) error
	ListLotsForKPI(context.Context, int64, int64, int64, string) ([]domain.Lot, error)
	ListLotsTable(context.Context, int64, int64, int64, string, int, int) ([]domain.LotTable, int, float64, float64, error)
}

type UseCases struct {
	repo RepositoryPort
}

func NewUseCases(repo RepositoryPort) *UseCases {
	return &UseCases{repo: repo}
}

func (u *UseCases) CreateLot(ctx context.Context, l *domain.Lot) (int64, error) {
	return u.repo.CreateLot(ctx, l)
}

func (u *UseCases) ListLotsByField(ctx context.Context, fieldID int64) ([]domain.Lot, error) {
	return u.repo.ListLotsByField(ctx, fieldID)
}

func (u *UseCases) GetLot(ctx context.Context, id int64) (*domain.Lot, error) {
	return u.repo.GetLot(ctx, id)
}

func (u *UseCases) UpdateLot(ctx context.Context, l *domain.Lot) error {
	return u.repo.UpdateLot(ctx, l)
}

func (u *UseCases) DeleteLot(ctx context.Context, id int64) error {
	return u.repo.DeleteLot(ctx, id)
}

func (u *UseCases) ListLotsByProject(ctx context.Context, projectID int64) ([]domain.Lot, error) {
	return u.repo.ListLotsByProject(ctx, projectID)
}

func (u *UseCases) ListLotsByProjectAndField(ctx context.Context, projectID, fieldID int64) ([]domain.Lot, error) {
	return u.repo.ListLotsByProjectAndField(ctx, projectID, fieldID)
}

func (u *UseCases) ListLotsByProjectFieldAndCrop(ctx context.Context, projectID, fieldID, cropID int64, cropType string) ([]domain.Lot, error) {
	return u.repo.ListLotsByProjectFieldAndCrop(ctx, projectID, fieldID, cropID, cropType)
}

func (u *UseCases) GetLotKPIs(
	ctx context.Context,
	projectID, fieldID, cropID int64,
	cropType string,
) (*domain.LotKPIs, error) {
	lots, err := u.repo.ListLotsForKPI(ctx, projectID, fieldID, cropID, cropType)
	if err != nil {
		return nil, err
	}

	var (
		seededArea    float64
		harvestedArea float64
		totalHarvest  float64
		totalCost     float64
		lotCount      float64
	)

	for _, lot := range lots {
		seededArea += lot.Hectares
		totalCost += lot.Cost
		lotCount++
		if lot.Status == "cosechado" || lot.Status == "harvested" {
			harvestedArea += lot.Hectares
			totalHarvest += lot.HarvestedTons
		}
	}

	var (
		yieldTnPerHa   float64
		costPerHectare float64
	)

	if harvestedArea > 0 {
		yieldTnPerHa = totalHarvest / harvestedArea
	}
	if lotCount > 0 {
		costPerHectare = totalCost / lotCount
	}

	kpis := &domain.LotKPIs{
		SeededArea:     seededArea,
		HarvestedArea:  harvestedArea,
		YieldTnPerHa:   yieldTnPerHa,
		CostPerHectare: costPerHectare,
	}
	return kpis, nil
}

func (u *UseCases) ListLotsTable(ctx context.Context,
	projectID, fieldID, cropID int64, cropType string,
	page, pageSize int,
) ([]domain.LotTable, int, float64, float64, error) {
	return u.repo.ListLotsTable(ctx, projectID, fieldID, cropID, cropType, page, pageSize)
}
