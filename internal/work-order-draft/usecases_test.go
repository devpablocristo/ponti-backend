package workorderdraft

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	types "github.com/devpablocristo/ponti-backend/internal/shared/types"
	supplydomain "github.com/devpablocristo/ponti-backend/internal/supply/usecases/domain"
	"github.com/devpablocristo/ponti-backend/internal/work-order-draft/usecases/domain"
	workorderdomain "github.com/devpablocristo/ponti-backend/internal/work-order/usecases/domain"
)

type testDraftRepo struct {
	createFn                     func(context.Context, *domain.WorkOrderDraft) (int64, error)
	createBatchFn                func(context.Context, []*domain.WorkOrderDraft) ([]int64, error)
	getByIDFn                    func(context.Context, int64) (*domain.WorkOrderDraft, error)
	listPendingSupplyNamesFn     func(context.Context, []int64) ([]string, error)
	listRelatedFn                func(context.Context, int64, string) ([]*domain.WorkOrderDraft, error)
	listFn                       func(context.Context, string, string, *bool, types.Input) ([]domain.WorkOrderDraftListItem, types.PageInfo, error)
	listOccupiedFn               func(context.Context, int64) ([]string, error)
	listOccupiedExcludingDraftFn func(context.Context, int64, int64) ([]string, error)
	listPublishedFn              func(context.Context, int64) ([]string, error)
	updateFn                     func(context.Context, *domain.WorkOrderDraft) error
	deleteFn                     func(context.Context, int64) error
	markPublishedFn              func(context.Context, int64, int64) error
}

func (r *testDraftRepo) CreateWorkOrderDraft(ctx context.Context, d *domain.WorkOrderDraft) (int64, error) {
	return r.createFn(ctx, d)
}

func (r *testDraftRepo) CreateWorkOrderDraftBatch(ctx context.Context, drafts []*domain.WorkOrderDraft) ([]int64, error) {
	if r.createBatchFn == nil {
		return nil, nil
	}
	return r.createBatchFn(ctx, drafts)
}

func (r *testDraftRepo) GetWorkOrderDraftByID(ctx context.Context, id int64) (*domain.WorkOrderDraft, error) {
	return r.getByIDFn(ctx, id)
}

func (r *testDraftRepo) ListPendingSupplyNamesByIDs(ctx context.Context, ids []int64) ([]string, error) {
	if r.listPendingSupplyNamesFn == nil {
		return nil, nil
	}
	return r.listPendingSupplyNamesFn(ctx, ids)
}

func (r *testDraftRepo) ListRelatedDigitalWorkOrderDraftsByBaseNumber(ctx context.Context, projectID int64, baseNumber string) ([]*domain.WorkOrderDraft, error) {
	if r.listRelatedFn == nil {
		return nil, nil
	}
	return r.listRelatedFn(ctx, projectID, baseNumber)
}

func (r *testDraftRepo) ListWorkOrderDrafts(ctx context.Context, number string, status string, isDigital *bool, inp types.Input) ([]domain.WorkOrderDraftListItem, types.PageInfo, error) {
	if r.listFn == nil {
		return nil, types.PageInfo{}, nil
	}
	return r.listFn(ctx, number, status, isDigital, inp)
}

func (r *testDraftRepo) ListOccupiedWorkOrderNumbersByProject(ctx context.Context, projectID int64) ([]string, error) {
	return r.listOccupiedFn(ctx, projectID)
}
func (r *testDraftRepo) ListOccupiedWorkOrderNumbersByProjectExcludingDraft(ctx context.Context, projectID int64, draftID int64) ([]string, error) {
	return r.listOccupiedExcludingDraftFn(ctx, projectID, draftID)
}
func (r *testDraftRepo) ListPublishedWorkOrderNumbersByProject(ctx context.Context, projectID int64) ([]string, error) {
	return r.listPublishedFn(ctx, projectID)
}
func (r *testDraftRepo) UpdateWorkOrderDraftByID(ctx context.Context, d *domain.WorkOrderDraft) error {
	return r.updateFn(ctx, d)
}
func (r *testDraftRepo) DeleteWorkOrderDraftByID(ctx context.Context, id int64) error {
	if r.deleteFn == nil {
		return nil
	}
	return r.deleteFn(ctx, id)
}
func (r *testDraftRepo) MarkWorkOrderDraftAsPublished(ctx context.Context, draftID int64, workOrderID int64) error {
	return r.markPublishedFn(ctx, draftID, workOrderID)
}

type testPublisher struct {
	createFn func(context.Context, *workorderdomain.WorkOrder) (int64, error)
}

type testPDFExporter struct {
	exportFn      func(context.Context, *domain.WorkOrderDraft) ([]byte, error)
	exportGroupFn func(context.Context, []*domain.WorkOrderDraft) ([]byte, error)
}

type testSupplyReader struct {
	getSupplyFn func(context.Context, int64) (*supplydomain.Supply, error)
}

func (s *testSupplyReader) GetSupply(ctx context.Context, id int64) (*supplydomain.Supply, error) {
	if s.getSupplyFn == nil {
		return &supplydomain.Supply{
			ID:   id,
			Name: "TEST SUPPLY",
		}, nil
	}
	return s.getSupplyFn(ctx, id)
}

func (e *testPDFExporter) ExportDraft(ctx context.Context, draft *domain.WorkOrderDraft) ([]byte, error) {
	if e.exportFn == nil {
		return []byte("pdf"), nil
	}
	return e.exportFn(ctx, draft)
}

func (e *testPDFExporter) ExportDraftGroup(ctx context.Context, drafts []*domain.WorkOrderDraft) ([]byte, error) {
	if e.exportGroupFn == nil {
		return []byte("pdf-group"), nil
	}
	return e.exportGroupFn(ctx, drafts)
}

func (p *testPublisher) CreateWorkOrder(ctx context.Context, wo *workorderdomain.WorkOrder) (int64, error) {
	return p.createFn(ctx, wo)
}

func validDraft() *domain.WorkOrderDraft {
	return &domain.WorkOrderDraft{
		Number:        "",
		Date:          time.Date(2026, 3, 30, 0, 0, 0, 0, time.UTC),
		CustomerID:    1,
		ProjectID:     10,
		FieldID:       20,
		LotID:         30,
		CropID:        40,
		LaborID:       50,
		Contractor:    "Contratista",
		EffectiveArea: decimal.NewFromInt(100),
		InvestorID:    60,
		Items: []domain.WorkOrderDraftItem{
			{
				SupplyID:  70,
				TotalUsed: decimal.NewFromInt(1),
				FinalDose: decimal.NewFromInt(1),
			},
		},
	}
}

func TestCreateDigitalWorkOrderDraft_AssignsNextBaseNumber(t *testing.T) {
	var created *domain.WorkOrderDraft

	repo := &testDraftRepo{
		listOccupiedFn: func(ctx context.Context, projectID int64) ([]string, error) {
			return []string{"40", "D-41", "D-41.1"}, nil
		},
		createFn: func(ctx context.Context, d *domain.WorkOrderDraft) (int64, error) {
			created = d
			return 123, nil
		},
	}

	uc := NewUseCases(repo, &testPublisher{}, &testPDFExporter{}, &testSupplyReader{})

	id, err := uc.CreateDigitalWorkOrderDraft(context.Background(), validDraft())
	require.NoError(t, err)
	require.Equal(t, int64(123), id)
	require.NotNil(t, created)
	require.True(t, created.IsDigital)
	require.Equal(t, domain.StatusDraft, created.Status)
	require.Equal(t, "D-42", created.Number)
}

func TestCreateDigitalWorkOrderDraft_AssignsSplitNumberWhenBaseExists(t *testing.T) {
	var created *domain.WorkOrderDraft

	repo := &testDraftRepo{
		listOccupiedFn: func(ctx context.Context, projectID int64) ([]string, error) {
			return []string{"D-41", "D-41.1"}, nil
		},
		createFn: func(ctx context.Context, d *domain.WorkOrderDraft) (int64, error) {
			created = d
			return 124, nil
		},
	}

	uc := NewUseCases(repo, &testPublisher{}, &testPDFExporter{}, &testSupplyReader{})

	draft := validDraft()
	draft.Number = "D-41"

	_, err := uc.CreateDigitalWorkOrderDraft(context.Background(), draft)
	require.NoError(t, err)
	require.NotNil(t, created)
	require.Equal(t, "D-41.2", created.Number)
}

func TestUpdateWorkOrderDraftByID_RevalidatesDigitalNumberExcludingSelf(t *testing.T) {
	var updated *domain.WorkOrderDraft

	repo := &testDraftRepo{
		getByIDFn: func(ctx context.Context, id int64) (*domain.WorkOrderDraft, error) {
			d := validDraft()
			d.ID = id
			d.Number = "D-41"
			d.IsDigital = true
			d.Status = domain.StatusDraft
			return d, nil
		},
		listOccupiedExcludingDraftFn: func(ctx context.Context, projectID int64, draftID int64) ([]string, error) {
			return []string{"40", "D-42"}, nil
		},
		updateFn: func(ctx context.Context, d *domain.WorkOrderDraft) error {
			updated = d
			return nil
		},
	}

	uc := NewUseCases(repo, &testPublisher{}, &testPDFExporter{}, &testSupplyReader{})

	draft := validDraft()
	draft.ID = 15
	draft.ProjectID = 10
	draft.Number = "D-41"

	err := uc.UpdateWorkOrderDraftByID(context.Background(), draft)
	require.NoError(t, err)
	require.NotNil(t, updated)
	require.True(t, updated.IsDigital)
	require.Equal(t, "D-41", updated.Number)
}

func TestPublishWorkOrderDraft_FailsWhenNumberAlreadyPublished(t *testing.T) {
	repo := &testDraftRepo{
		getByIDFn: func(ctx context.Context, id int64) (*domain.WorkOrderDraft, error) {
			d := validDraft()
			d.ID = id
			d.Number = "D-41"
			d.IsDigital = true
			d.Status = domain.StatusDraft
			return d, nil
		},
		listPublishedFn: func(ctx context.Context, projectID int64) ([]string, error) {
			return []string{"D-41"}, nil
		},
		markPublishedFn: func(ctx context.Context, draftID int64, workOrderID int64) error {
			return nil
		},
	}

	pub := &testPublisher{
		createFn: func(ctx context.Context, wo *workorderdomain.WorkOrder) (int64, error) {
			return 999, nil
		},
	}

	uc := NewUseCases(repo, pub, &testPDFExporter{}, &testSupplyReader{})

	_, err := uc.PublishWorkOrderDraft(context.Background(), 77)
	require.Error(t, err)
	require.Contains(t, err.Error(), "work order already exists for number D-41 and project 10")
}

func TestPublishWorkOrderDraft_FailsWhenSupplyIsPending(t *testing.T) {
	repo := &testDraftRepo{
		getByIDFn: func(ctx context.Context, id int64) (*domain.WorkOrderDraft, error) {
			d := validDraft()
			d.ID = id
			d.Number = "D-55"
			d.IsDigital = true
			d.Status = domain.StatusDraft
			return d, nil
		},
		listPublishedFn: func(ctx context.Context, projectID int64) ([]string, error) {
			return []string{}, nil
		},
		listPendingSupplyNamesFn: func(ctx context.Context, ids []int64) ([]string, error) {
			return []string{"2,4D ENLIST"}, nil
		},
	}

	pub := &testPublisher{
		createFn: func(ctx context.Context, wo *workorderdomain.WorkOrder) (int64, error) {
			return 999, nil
		},
	}

	uc := NewUseCases(repo, pub, &testPDFExporter{}, &testSupplyReader{})

	_, err := uc.PublishWorkOrderDraft(context.Background(), 88)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot publish work order draft with pending supplies")
	require.Contains(t, err.Error(), "2,4D ENLIST")
}
