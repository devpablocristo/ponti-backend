// Package stock contiene casos de uso para stock.
package stock

import (
	"context"
	"strconv"
	"time"

	"github.com/devpablocristo/core/errors/go/domainerr"
	"github.com/devpablocristo/core/security/go/contextkeys"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	projectdomain "github.com/devpablocristo/ponti-backend/internal/project/usecases/domain"
	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
	sharedmodels "github.com/devpablocristo/ponti-backend/internal/shared/models"
	"github.com/devpablocristo/ponti-backend/internal/stock/usecases/domain"
)

// StockNegativeInput es el payload minimo para emitir/resolver una notificacion
// de stock negativo.
type StockNegativeInput struct {
	ProductID   string
	ProductName string
	Quantity    float64
}

// BusinessInsightsNotifier es el contrato que evalua y dispara notificaciones
// reactivas cuando el stock real cambia. Opcional: si es nil, las mutaciones
// siguen funcionando pero no emiten/resuelven notificaciones.
type BusinessInsightsNotifier interface {
	NotifyStockNegative(ctx context.Context, tenantID uuid.UUID, actor string, level StockNegativeInput) error
	MaybeResolveStockNegative(ctx context.Context, tenantID uuid.UUID, productID string) error
}

type RepositoryPort interface {
	GetStocks(context.Context, int64, time.Time) ([]*domain.Stock, error)
	GetActiveStocksByProjectID(context.Context, int64) ([]*domain.Stock, error)
	CreateStock(context.Context, *domain.Stock) (int64, error)
	UpdateCloseDateByProject(context.Context, int64, *domain.Stock) error
	UpdateRealStockUnits(context.Context, int64, *domain.Stock) error
	GetStockByID(context.Context, int64) (*domain.Stock, error)
	GetLastStockByProjectID(context.Context, int64, int64) (*domain.Stock, bool, error)
	GetLastStockByProjectInvestorID(context.Context, int64, int64, int64) (*domain.Stock, bool, error)
	GetStockByPeriodAndProjectID(context.Context, int64) (*domain.Stock, error)
	GetStocksPeriods(context.Context, int64) ([]string, error)
	ListAllStocks(context.Context) ([]*domain.Stock, error)
	UpdateUnitsConsumed(context.Context, domain.Stock, decimal.Decimal) error
	ExecuteInTransaction(context.Context, func(context.Context) error) error
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
	notifier  BusinessInsightsNotifier
}

// NewUseCases crea una instancia de casos de uso para stock.
func NewUseCases(repo RepositoryPort, excel ExporterAdapterPort, projectUC ProjectUseCasesPort) *UseCases {
	return &UseCases{repo: repo, excel: excel, projectUC: projectUC}
}

// SetBusinessInsightsNotifier conecta el notifier despues de wire (la
// dependencia se resuelve recien en bootstrap porque businessinsights vive
// fuera del DI graph de stock).
func (u *UseCases) SetBusinessInsightsNotifier(n BusinessInsightsNotifier) {
	u.notifier = n
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

	activeStocks, err := u.repo.GetActiveStocksByProjectID(ctx, projectID)
	if err != nil {
		return err
	}
	if len(activeStocks) == 0 {
		return domainerr.NotFound("no active stocks to close for project")
	}

	return u.repo.ExecuteInTransaction(ctx, func(txCtx context.Context) error {
		if err := u.repo.UpdateCloseDateByProject(txCtx, projectID, stock); err != nil {
			return err
		}
		for _, active := range activeStocks {
			newStock := createNewStockPeriod(*stock.UpdatedBy, monthPeriod, yearPeriod, active)
			if _, err := u.repo.CreateStock(txCtx, &newStock); err != nil {
				return err
			}
		}
		return nil
	})
}

// UpdateRealStockUnits es el punto unico de mutacion del stock real. Persiste
// el cambio y, si hay notifier configurado, evalua el trigger reactivo:
//   - qty < 0  -> NotifyStockNegative (abre/refresca candidato de notificacion)
//   - qty >= 0 -> MaybeResolveStockNegative (cierra candidato si estaba abierto)
//
// El error del notifier es loggeado pero no propaga: las notificaciones nunca
// rompen el flow principal de stock.
func (u *UseCases) UpdateRealStockUnits(ctx context.Context, stockID int64, stock *domain.Stock) error {
	if err := u.repo.UpdateRealStockUnits(ctx, stockID, stock); err != nil {
		return err
	}
	u.evaluateStockNotification(ctx, stock)
	return nil
}

// evaluateStockNotification dispara el notifier apropiado segun el signo del
// stock. Es no-op si el notifier no esta configurado o si falta info esencial
// (tenant en ctx, producto en stock).
func (u *UseCases) evaluateStockNotification(ctx context.Context, s *domain.Stock) {
	if u.notifier == nil || s == nil || s.Supply == nil {
		return
	}
	orgRaw := ctx.Value(ctxkeys.OrgID)
	orgID, ok := orgRaw.(uuid.UUID)
	if !ok || orgID == uuid.Nil {
		return
	}
	productID := strconv.FormatInt(s.Supply.ID, 10)
	qty, _ := s.RealStockUnits.Float64()
	if qty < 0 {
		actor, _ := sharedmodels.ActorFromContext(ctx)
		_ = u.notifier.NotifyStockNegative(ctx, orgID, actor, StockNegativeInput{
			ProductID:   productID,
			ProductName: s.Supply.Name,
			Quantity:    qty,
		})
		return
	}
	_ = u.notifier.MaybeResolveStockNegative(ctx, orgID, productID)
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

func (u *UseCases) GetLastStockByProjectInvestorID(ctx context.Context, projectID int64, supplyID int64, investorID int64) (*domain.Stock, bool, error) {
	return u.repo.GetLastStockByProjectInvestorID(ctx, projectID, supplyID, investorID)
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
