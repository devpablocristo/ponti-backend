package stock

import (
	"context"
	"testing"
	"time"

	"github.com/devpablocristo/core/errors/go/domainerr"
	ctxkeys "github.com/devpablocristo/core/security/go/contextkeys"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	projectdomain "github.com/devpablocristo/ponti-backend/internal/project/usecases/domain"
	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
	"github.com/devpablocristo/ponti-backend/internal/stock/usecases/domain"
	supplydomain "github.com/devpablocristo/ponti-backend/internal/supply/usecases/domain"
)

type stockUseCaseRepoStub struct {
	getStockBySupplyIDFn func(context.Context, int64, int64, time.Time) (*domain.Stock, error)
	createStockCountFn   func(context.Context, *domain.StockCount) (int64, error)

	getStockCalls    []stockRepoCall
	createCountCalls []*domain.StockCount
}

type stockRepoCall struct {
	projectID  int64
	supplyID   int64
	cutoffDate time.Time
}

func (s *stockUseCaseRepoStub) GetStocks(context.Context, int64, time.Time) ([]*domain.Stock, error) {
	return nil, nil
}

func (s *stockUseCaseRepoStub) GetStockBySupplyID(
	ctx context.Context,
	projectID int64,
	supplyID int64,
	cutoffDate time.Time,
) (*domain.Stock, error) {
	s.getStockCalls = append(s.getStockCalls, stockRepoCall{
		projectID:  projectID,
		supplyID:   supplyID,
		cutoffDate: cutoffDate,
	})
	if s.getStockBySupplyIDFn != nil {
		return s.getStockBySupplyIDFn(ctx, projectID, supplyID, cutoffDate)
	}
	return nil, nil
}

func (s *stockUseCaseRepoStub) CreateStockCount(ctx context.Context, count *domain.StockCount) (int64, error) {
	s.createCountCalls = append(s.createCountCalls, count)
	if s.createStockCountFn != nil {
		return s.createStockCountFn(ctx, count)
	}
	return 0, nil
}

type stockProjectUCStub struct {
	getProjectFn func(context.Context, int64) (*projectdomain.Project, error)
}

func (s *stockProjectUCStub) GetProject(ctx context.Context, id int64) (*projectdomain.Project, error) {
	if s.getProjectFn != nil {
		return s.getProjectFn(ctx, id)
	}
	return &projectdomain.Project{ID: id}, nil
}

type stockNotifierStub struct {
	notifyCalls  []notifyCall
	resolveCalls []resolveCall
}

type notifyCall struct {
	tenantID uuid.UUID
	actor    string
	level    StockNegativeInput
}

type resolveCall struct {
	tenantID  uuid.UUID
	productID string
}

func (s *stockNotifierStub) NotifyStockNegative(
	_ context.Context,
	tenantID uuid.UUID,
	actor string,
	level StockNegativeInput,
) error {
	s.notifyCalls = append(s.notifyCalls, notifyCall{
		tenantID: tenantID,
		actor:    actor,
		level:    level,
	})
	return nil
}

func (s *stockNotifierStub) MaybeResolveStockNegative(
	_ context.Context,
	tenantID uuid.UUID,
	productID string,
) error {
	s.resolveCalls = append(s.resolveCalls, resolveCall{
		tenantID:  tenantID,
		productID: productID,
	})
	return nil
}

func TestUseCasesCreateStockCount_PersistsAndResolvesNotifier(t *testing.T) {
	repo := &stockUseCaseRepoStub{
		getStockBySupplyIDFn: func(context.Context, int64, int64, time.Time) (*domain.Stock, error) {
			return &domain.Stock{
				ID:             9,
				ProjectID:      7,
				StockUnits:     decimal.NewFromInt(71),
				RealStockUnits: decimal.NewFromInt(80),
				Supply: &supplydomain.Supply{
					ID:       9,
					Name:     "Urea",
					Price:    decimal.NewFromInt(5),
					UnitName: "kg",
				},
			}, nil
		},
		createStockCountFn: func(context.Context, *domain.StockCount) (int64, error) {
			return 44, nil
		},
	}
	projectUC := &stockProjectUCStub{}
	notifier := &stockNotifierStub{}

	ucs := NewUseCases(repo, nil, projectUC)
	ucs.SetBusinessInsightsNotifier(notifier)

	orgID := uuid.New()
	ctx := context.WithValue(context.Background(), ctxkeys.OrgID, orgID)
	ctx = context.WithValue(ctx, ctxkeys.Actor, "auditor@ponti.test")

	countedAt := time.Date(2026, 4, 22, 15, 4, 5, 0, time.UTC)
	count := &domain.StockCount{
		CountedUnits: decimal.NewFromInt(82),
		CountedAt:    countedAt,
		Note:         "Conteo abril",
		Base:         shareddomain.Base{},
	}

	id, err := ucs.CreateStockCount(ctx, 7, 9, count)
	if err != nil {
		t.Fatalf("CreateStockCount returned error: %v", err)
	}

	if id != 44 {
		t.Fatalf("expected inserted id 44, got %d", id)
	}
	if len(repo.getStockCalls) != 1 {
		t.Fatalf("expected one GetStockBySupplyID call, got %d", len(repo.getStockCalls))
	}
	if repo.getStockCalls[0].projectID != 7 || repo.getStockCalls[0].supplyID != 9 {
		t.Fatalf("unexpected summary lookup call: %+v", repo.getStockCalls[0])
	}
	if !repo.getStockCalls[0].cutoffDate.Equal(countedAt) {
		t.Fatalf("expected cutoff to match countedAt, got %s", repo.getStockCalls[0].cutoffDate)
	}
	if len(repo.createCountCalls) != 1 {
		t.Fatalf("expected one persisted count, got %d", len(repo.createCountCalls))
	}

	persisted := repo.createCountCalls[0]
	if persisted.SupplyID != 9 {
		t.Fatalf("expected persisted supply_id=9, got %d", persisted.SupplyID)
	}
	if !persisted.CreatedAt.IsZero() && persisted.CreatedAt.Location() != time.UTC {
		t.Fatalf("expected CreatedAt in UTC, got %s", persisted.CreatedAt.Location())
	}
	if persisted.CreatedAt.IsZero() || persisted.UpdatedAt.IsZero() {
		t.Fatal("expected created/updated timestamps to be set")
	}
	if len(notifier.notifyCalls) != 0 {
		t.Fatalf("expected no negative notifications, got %d", len(notifier.notifyCalls))
	}
	if len(notifier.resolveCalls) != 1 {
		t.Fatalf("expected one resolve call, got %d", len(notifier.resolveCalls))
	}
	if notifier.resolveCalls[0].tenantID != orgID || notifier.resolveCalls[0].productID != "9" {
		t.Fatalf("unexpected resolve call: %+v", notifier.resolveCalls[0])
	}
}

func TestUseCasesCreateStockCount_RejectsNegativeUnits(t *testing.T) {
	repo := &stockUseCaseRepoStub{}
	projectUC := &stockProjectUCStub{}
	ucs := NewUseCases(repo, nil, projectUC)

	_, err := ucs.CreateStockCount(context.Background(), 7, 9, &domain.StockCount{
		CountedUnits: decimal.NewFromInt(-1),
		CountedAt:    time.Date(2026, 4, 22, 0, 0, 0, 0, time.UTC),
	})
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	if !domainerr.IsValidation(err) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if len(repo.getStockCalls) != 0 || len(repo.createCountCalls) != 0 {
		t.Fatalf("expected repo not to be called on invalid payload, got lookup=%d create=%d", len(repo.getStockCalls), len(repo.createCountCalls))
	}
}
