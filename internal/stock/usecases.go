// Package stock contiene casos de uso para stock continuo por proyecto.
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
	GetStockBySupplyID(context.Context, int64, int64, time.Time) (*domain.Stock, error)
	CreateStockCount(context.Context, *domain.StockCount) (int64, error)
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

func (u *UseCases) GetStocksSummary(ctx context.Context, projectID int64, cutoffDate time.Time) ([]*domain.Stock, error) {
	if err := u.validateProject(ctx, projectID); err != nil {
		return nil, err
	}
	return u.repo.GetStocks(ctx, projectID, cutoffDate)
}

func (u *UseCases) GetStockBySupplyID(
	ctx context.Context,
	projectID int64,
	supplyID int64,
	cutoffDate time.Time,
) (*domain.Stock, error) {
	if err := u.validateProject(ctx, projectID); err != nil {
		return nil, err
	}
	if supplyID <= 0 {
		return nil, domainerr.Validation("supply_id must be greater than 0")
	}
	return u.repo.GetStockBySupplyID(ctx, projectID, supplyID, cutoffDate)
}

func (u *UseCases) GetLastStockByProjectID(ctx context.Context, projectID int64, supplyID int64) (*domain.Stock, bool, error) {
	stock, err := u.GetStockBySupplyID(ctx, projectID, supplyID, time.Time{})
	if err != nil {
		if domainerr.IsNotFound(err) {
			return nil, true, nil
		}
		return nil, false, err
	}
	return stock, false, nil
}

func (u *UseCases) GetLastStockByProjectInvestorID(
	ctx context.Context,
	projectID int64,
	supplyID int64,
	_ int64,
) (*domain.Stock, bool, error) {
	return u.GetLastStockByProjectID(ctx, projectID, supplyID)
}

func (u *UseCases) CreateStock(context.Context, *domain.Stock) (int64, error) {
	return 0, domainerr.Validation("legacy stock rows are disabled; use supply movements and stock counts")
}

func (u *UseCases) UpdateRealStockUnits(context.Context, int64, *domain.Stock) error {
	return domainerr.Validation("legacy stock updates are disabled; use stock counts")
}

func (u *UseCases) CreateStockCount(
	ctx context.Context,
	projectID int64,
	supplyID int64,
	count *domain.StockCount,
) (int64, error) {
	if err := u.validateProject(ctx, projectID); err != nil {
		return 0, err
	}
	if count == nil {
		return 0, domainerr.Validation("stock count is nil")
	}
	if supplyID <= 0 {
		return 0, domainerr.Validation("supply_id must be greater than 0")
	}
	if count.CountedAt.IsZero() {
		return 0, domainerr.Validation("counted_at is required")
	}
	if count.CountedUnits.LessThan(decimal.Zero) {
		return 0, domainerr.Validation("counted_units must be greater than or equal to 0")
	}

	stockSummary, err := u.repo.GetStockBySupplyID(ctx, projectID, supplyID, count.CountedAt)
	if err != nil {
		return 0, err
	}

	now := time.Now().UTC()
	count.SupplyID = supplyID
	if count.CreatedAt.IsZero() {
		count.CreatedAt = now
	}
	if count.UpdatedAt.IsZero() {
		count.UpdatedAt = now
	}

	id, err := u.repo.CreateStockCount(ctx, count)
	if err != nil {
		return 0, err
	}

	stockSummary.RealStockUnits = count.CountedUnits
	stockSummary.HasRealStockCount = true
	stockSummary.LastCountAt = &count.CountedAt
	u.evaluateStockNotification(ctx, stockSummary)

	return id, nil
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

// ExportStocksByProject exporta stocks filtrados por proyecto.
func (u *UseCases) ExportStocksByProject(ctx context.Context, projectID int64) ([]byte, error) {
	if u.excel == nil {
		return nil, domainerr.Internal("exporter not configured")
	}

	if err := u.validateProject(ctx, projectID); err != nil {
		return nil, err
	}
	items, err := u.repo.GetStocks(ctx, projectID, time.Time{})
	if err != nil {
		return nil, domainerr.Internal("list stocks")
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
