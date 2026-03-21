// Package stock contiene casos de uso para stock.
package stock

import (
	"context"
	"time"

	"github.com/devpablocristo/core/saas/go/shared/domainerr"
	projectdomain "github.com/devpablocristo/ponti-backend/internal/project/usecases/domain"
	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
	"github.com/devpablocristo/ponti-backend/internal/stock/usecases/domain"
	"github.com/shopspring/decimal"
)

type RepositoryPort interface {
	GetStocks(context.Context, int64, time.Time) ([]*domain.Stock, error)
	CreateStock(context.Context, *domain.Stock) (int64, error)
	UpdateCloseDateByProject(context.Context, int64, *domain.Stock) error
	UpdateRealStockUnits(context.Context, int64, *domain.Stock) error
	GetStockByID(context.Context, int64) (*domain.Stock, error)
	GetLastStockByProjectID(context.Context, int64, int64) (*domain.Stock, bool, error)
	GetStockByPeriodAndProjectID(context.Context, int64) (*domain.Stock, error)
	GetStocksPeriods(context.Context, int64) ([]string, error)
	ListAllStocks(context.Context) ([]*domain.Stock, error)
	UpdateUnitsConsumed(context.Context, domain.Stock, decimal.Decimal) error
}

type ExporterAdapterPort interface {
	Export(ctx context.Context, items []*domain.Stock) ([]byte, error)
	Close() error
}

type ProjectUseCasesPort interface {
	GetProject(ctx context.Context, id int64) (*projectdomain.Project, error)
}

type UseCases struct {
	repo      RepositoryPort
	excel     ExporterAdapterPort
	projectUC ProjectUseCasesPort
}

// NewUseCases crea una instancia de casos de uso para stock.
func NewUseCases(repo RepositoryPort, excel ExporterAdapterPort, projectUC ProjectUseCasesPort) *UseCases {
	return &UseCases{repo: repo, excel: excel, projectUC: projectUC}
}

func (u *UseCases) GetStocksSummary(ctx context.Context, projectID int64, closeDate time.Time) ([]*domain.Stock, error) {
	if err := u.validateProject(ctx, projectID); err != nil {
		return nil, err
	}
	return u.repo.GetStocks(ctx, projectID, closeDate)
}

func (u *UseCases) GetStocksPeriods(ctx context.Context, projectID int64) ([]string, error) {
	if err := u.validateProject(ctx, projectID); err != nil {
		return nil, err
	}
	return u.repo.GetStocksPeriods(ctx, projectID)
}

func (u *UseCases) CreateStock(ctx context.Context, s *domain.Stock) (int64, error) {
	return u.repo.CreateStock(ctx, s)
}

func (u *UseCases) UpdateCloseDateByProject(ctx context.Context, projectID int64, monthPeriod int64, yearPeriod int64, stock *domain.Stock) error {
	if err := u.validateProject(ctx, projectID); err != nil {
		return err
	}
	stockFromDb, err := u.repo.GetStockByPeriodAndProjectID(ctx, projectID)
	if err != nil {
		return err
	}

	err = u.repo.UpdateCloseDateByProject(ctx, projectID, stock)
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

func (u *UseCases) UpdateRealStockUnits(ctx context.Context, stockID int64, stock *domain.Stock) error {
	return u.repo.UpdateRealStockUnits(ctx, stockID, stock)
}

func (u *UseCases) GetStockByID(ctx context.Context, stockID int64) (*domain.Stock, error) {
	if stockID <= 0 {
		return nil, domainerr.Validation("stock_id must be greater than 0")
	}
	return u.repo.GetStockByID(ctx, stockID)
}

func (u *UseCases) GetLastStockByProjectID(ctx context.Context, projectID int64, supplyID int64) (*domain.Stock, bool, error) {
	return u.repo.GetLastStockByProjectID(ctx, projectID, supplyID)
}

func (u *UseCases) UpdateUnitsConsumed(ctx context.Context, stockDomain domain.Stock, quantity decimal.Decimal) error {
	return u.repo.UpdateUnitsConsumed(ctx, stockDomain, quantity)
}

func createNewStockPeriod(userID string, monthPeriod int64, yearPeriod int64, stock *domain.Stock) domain.Stock {
	newMonthPeriod, newYearPeriod := startNewStockPeriod(monthPeriod, yearPeriod)
	newStock := domain.Stock{
		Project:     stock.Project,
		YearPeriod:  newYearPeriod,
		MonthPeriod: newMonthPeriod,
		Supply:      stock.Supply,
		Investor:    stock.Investor,
		// "Stock de campo" es recuento manual, por default arranca en 0 en cada periodo.
		InitialStock:   decimal.Zero,
		RealStockUnits: decimal.Zero,
		Base: shareddomain.Base{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			CreatedBy: &userID,
			UpdatedBy: &userID,
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

// ExportStocksByProject exporta stocks filtrados por proyecto (stocks activos sin close_date)
func (u *UseCases) ExportStocksByProject(ctx context.Context, projectID int64) ([]byte, error) {
	if u.excel == nil {
		return nil, domainerr.Internal("exporter not configured")
	}

	// Usar GetStocks con tiempo vacío para obtener stocks activos del proyecto
	var emptyTime time.Time
	if err := u.validateProject(ctx, projectID); err != nil {
		return nil, err
	}
	items, err := u.repo.GetStocks(ctx, projectID, emptyTime)
	if err != nil {
		return nil, domainerr.Internal("list Stocks")
	}

	if len(items) == 0 {
		return nil, domainerr.NotFound("there is no data to export")
	}

	return u.excel.Export(ctx, items)
}

func (u *UseCases) validateProject(ctx context.Context, projectID int64) error {
	if projectID <= 0 {
		return domainerr.Validation("project_id must be greater than 0")
	}
	if u.projectUC == nil {
		return domainerr.Internal("project usecases not configured")
	}
	_, err := u.projectUC.GetProject(ctx, projectID)
	return err
}
