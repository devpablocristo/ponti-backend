// Package supply contiene casos de uso para insumos y movimientos.
package supply

import (
	"context"
	"fmt"

	providerdomain "github.com/alphacodinggroup/ponti-backend/internal/provider/usecases/domain"
	stockdomain "github.com/alphacodinggroup/ponti-backend/internal/stock/usecases/domain"
	domain "github.com/alphacodinggroup/ponti-backend/internal/supply/usecases/domain"
	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
)

type RepositoryPort interface {
	CreateSupply(context.Context, *domain.Supply) (int64, error)
	CreateSuppliesBulk(context.Context, []domain.Supply) error
	GetSupply(context.Context, int64) (*domain.Supply, error)
	GetSupplyByProjectAndName(context.Context, int64, string) (*domain.Supply, error)
	GetWorkOrdersBySupplyID(ctx context.Context, supplyID int64) (int64, error)
	UpdateSupply(context.Context, *domain.Supply) error
	DeleteSupply(context.Context, int64) error
	ListSuppliesPaginated(context.Context, domain.SupplyFilter, string, int, int) ([]domain.Supply, int64, error)
	ListAllSupplies(context.Context, domain.SupplyFilter) ([]domain.Supply, int64, error)
	UpdateSuppliesBulk(context.Context, []domain.Supply) error
	CreateProvider(context.Context, *providerdomain.Provider) (int64, error)
	CreateSupplyMovement(context.Context, *domain.SupplyMovement) (int64, error)
	GetEntriesSupplyMovementsByProjectID(context.Context, int64) ([]*domain.SupplyMovement, error)
	UpdateSupplyMovement(context.Context, *domain.SupplyMovement) error
	GetSupplyMovementByID(context.Context, int64) (*domain.SupplyMovement, error)
	DeleteSupplyMovement(context.Context, int64, int64) error
	GetProviders(context.Context) ([]providerdomain.Provider, error)
}

type ExporterAdapterPort interface {
	ExportSupplies(ctx context.Context, items []*domain.Supply) ([]byte, error)
	ExportSupplyMovements(ctx context.Context, items []*domain.SupplyMovement) ([]byte, error)
	Close() error
}

type StockUseCasesPort interface {
	GetLastStockByProjectID(ctx context.Context, projectID int64, supplyID int64) (*stockdomain.Stock, bool, error)
	CreateStock(ctx context.Context, s *stockdomain.Stock) (int64, error)
	UpdateRealStockUnits(ctx context.Context, stockID int64, stock *stockdomain.Stock) error
}

type UseCases struct {
	repo          RepositoryPort
	excel         ExporterAdapterPort
	stockUseCases StockUseCasesPort
}

func NewUseCases(repo RepositoryPort, excel ExporterAdapterPort, stockUseCases StockUseCasesPort) *UseCases {
	return &UseCases{
		repo:          repo,
		excel:         excel,
		stockUseCases: stockUseCases,
	}
}

func (u *UseCases) CreateSupply(ctx context.Context, s *domain.Supply) (int64, error) {
	if s.ProjectID == 0 || s.Name == "" || s.UnitID == 0 || s.CategoryID == 0 || s.Type.ID == 0 {
		return 0, types.NewError(types.ErrInvalidInput, "missing required fields", nil)
	}
	return u.repo.CreateSupply(ctx, s)
}

func (u *UseCases) CreateSuppliesBulk(ctx context.Context, supplies []domain.Supply) error {
	if len(supplies) == 0 {
		return types.NewError(types.ErrInvalidInput, "no supplies provided", nil)
	}

	seen := map[string]bool{}
	projectID := supplies[0].ProjectID

	for _, s := range supplies {
		key := fmt.Sprintf("%d:%s", s.ProjectID, s.Name)
		if seen[key] {
			return types.NewError(types.ErrInvalidInput, fmt.Sprintf("duplicate supply name in request: %s", s.Name), nil)
		}
		seen[key] = true
		if s.ProjectID != projectID {
			return types.NewError(types.ErrInvalidInput, "all supplies must have the same project_id", nil)
		}
		if s.Name == "" || s.UnitID == 0 || s.CategoryID == 0 || s.Type.ID == 0 {
			return types.NewError(types.ErrInvalidInput, fmt.Sprintf("missing fields in supply: %s", s.Name), nil)
		}
	}

	// Chequeo adicional: podés optimizarlo usando un repo más específico
	existing, _, err := u.repo.ListSuppliesPaginated(ctx, domain.SupplyFilter{
		ProjectID: &projectID,
	}, "", 1, 10000)
	if err != nil {
		return err
	}
	for _, s := range supplies {
		for _, e := range existing {
			if e.Name == s.Name {
				return types.NewError(types.ErrConflict, fmt.Sprintf("supply already exists with name: %s", s.Name), nil)
			}
		}
	}

	return u.repo.CreateSuppliesBulk(ctx, supplies)
}

func (u *UseCases) GetSupply(ctx context.Context, id int64) (*domain.Supply, error) {
	return u.repo.GetSupply(ctx, id)
}

func (u *UseCases) UpdateSupply(ctx context.Context, s *domain.Supply) error {
	if s.ID == 0 || s.Name == "" || s.UnitID == 0 || s.CategoryID == 0 || s.Type.ID == 0 {
		return types.NewError(types.ErrInvalidInput, "missing required fields", nil)
	}
	return u.repo.UpdateSupply(ctx, s)
}

func (u *UseCases) DeleteSupply(ctx context.Context, id int64) error {
	return u.repo.DeleteSupply(ctx, id)
}

func (u *UseCases) CountWorkOrdersBySupplyID(ctx context.Context, supplyID int64) (int64, error) {
	return u.repo.GetWorkOrdersBySupplyID(ctx, supplyID)
}

func (u *UseCases) ListSuppliesPaginated(
	ctx context.Context,
	filter domain.SupplyFilter,
	page, perPage int,
	mode string,
) ([]domain.Supply, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage <= 0 || perPage > 1000 {
		perPage = 1000
	}
	return u.repo.ListSuppliesPaginated(ctx, filter, mode, page, perPage)
}

func (u *UseCases) UpdateSuppliesBulk(ctx context.Context, supplies []domain.Supply) error {
	if len(supplies) == 0 {
		return types.NewError(types.ErrInvalidInput, "no supplies provided", nil)
	}
	seen := map[int64]bool{}
	for _, s := range supplies {
		if s.ID == 0 || s.Name == "" || s.UnitID == 0 || s.CategoryID == 0 || s.Type.ID == 0 {
			return types.NewError(types.ErrInvalidInput, fmt.Sprintf("missing fields in supply id: %d", s.ID), nil)
		}
		if seen[s.ID] {
			return types.NewError(types.ErrInvalidInput, fmt.Sprintf("duplicate supply id in request: %d", s.ID), nil)
		}
		seen[s.ID] = true
	}
	return u.repo.UpdateSuppliesBulk(ctx, supplies)
}

func (u *UseCases) ExportTableSupplies(ctx context.Context, filter domain.SupplyFilter) ([]byte, error) {
	if u.excel == nil {
		return nil, types.NewError(types.ErrInternal, "exporter not configured", nil)
	}

	items, total, err := u.repo.ListAllSupplies(ctx, filter)
	if err != nil {
		return nil, types.NewError(types.ErrInternal, "Internal error", err)
	}

	if total == 0 {
		return nil, types.NewError(types.ErrNotFound, "there is no data to export", nil)
	}

	if len(items) == 0 {
		return nil, types.NewError(types.ErrNotFound, "there is no data to export", nil)
	}

	itemPointers := make([]*domain.Supply, len(items))
	for i := range items {
		itemPointers[i] = &items[i]
	}

	return u.excel.ExportSupplies(ctx, itemPointers)
}
