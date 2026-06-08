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
	listGroupsFn                 func(context.Context, string, string, types.Input) ([]domain.WorkOrderDraftGroupListItem, types.PageInfo, error)
	listOccupiedFn               func(context.Context, int64) ([]string, error)
	listOccupiedExcludingDraftFn func(context.Context, int64, int64) ([]string, error)
	listPublishedFn              func(context.Context, int64) ([]string, error)
	updateFn                     func(context.Context, *domain.WorkOrderDraft) error
	updateGroupFn                func(context.Context, []*domain.WorkOrderDraft) error
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

func (r *testDraftRepo) GetPendingLaborNameByID(ctx context.Context, laborID int64) (string, error) {
	return "", nil
}

func (r *testDraftRepo) ListRelatedDigitalWorkOrderDraftsByBaseNumber(ctx context.Context, projectID int64, baseNumber string) ([]*domain.WorkOrderDraft, error) {
	if r.listRelatedFn == nil {
		return nil, nil
	}
	return r.listRelatedFn(ctx, projectID, baseNumber)
}

func (r *testDraftRepo) GetLaborContractorByID(ctx context.Context, laborID int64) (string, error) {
	return "", nil
}

func (r *testDraftRepo) ListWorkOrderDrafts(ctx context.Context, number string, status string, isDigital *bool, inp types.Input) ([]domain.WorkOrderDraftListItem, types.PageInfo, error) {
	if r.listFn == nil {
		return nil, types.PageInfo{}, nil
	}
	return r.listFn(ctx, number, status, isDigital, inp)
}

func (r *testDraftRepo) ListDigitalWorkOrderDraftGroups(ctx context.Context, number string, status string, inp types.Input) ([]domain.WorkOrderDraftGroupListItem, types.PageInfo, error) {
	if r.listGroupsFn == nil {
		return nil, types.PageInfo{}, nil
	}
	return r.listGroupsFn(ctx, number, status, inp)
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
func (r *testDraftRepo) UpdateWorkOrderDraftGroup(ctx context.Context, drafts []*domain.WorkOrderDraft) error {
	if r.updateGroupFn == nil {
		return nil
	}
	return r.updateGroupFn(ctx, drafts)
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

	uc := NewUseCases(repo, &testPublisher{}, &testSupplyReader{})

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

	uc := NewUseCases(repo, &testPublisher{}, &testSupplyReader{})

	draft := validDraft()
	draft.Number = "D-41"

	_, err := uc.CreateDigitalWorkOrderDraft(context.Background(), draft)
	require.NoError(t, err)
	require.NotNil(t, created)
	require.Equal(t, "D-41.2", created.Number)
}

func TestCreateDigitalWorkOrderDraftBatch_UsesTotalBatchAreaForFinalDose(t *testing.T) {
	var created []*domain.WorkOrderDraft

	repo := &testDraftRepo{
		listOccupiedFn: func(ctx context.Context, projectID int64) ([]string, error) {
			return []string{}, nil
		},
		createBatchFn: func(ctx context.Context, drafts []*domain.WorkOrderDraft) ([]int64, error) {
			created = drafts
			return []int64{101, 102}, nil
		},
	}

	uc := NewUseCases(repo, &testPublisher{}, &testSupplyReader{})

	batch := &domain.WorkOrderDraftBatchCreate{
		Date:       time.Date(2026, 5, 14, 0, 0, 0, 0, time.UTC),
		CustomerID: 1,
		ProjectID:  10,
		FieldID:    20,
		CropID:     30,
		LaborID:    40,
		Contractor: "Contratista",
		InvestorID: 50,
		Lots: []domain.WorkOrderDraftBatchLot{
			{
				LotID:         101,
				EffectiveArea: decimal.NewFromInt(60),
				Items: []domain.WorkOrderDraftBatchLotItem{
					{
						SupplyID:  999,
						TotalUsed: decimal.NewFromInt(120),
					},
				},
			},
			{
				LotID:         102,
				EffectiveArea: decimal.NewFromInt(60),
				Items: []domain.WorkOrderDraftBatchLotItem{
					{
						SupplyID:  999,
						TotalUsed: decimal.NewFromInt(120),
					},
				},
			},
		},
	}

	result, err := uc.CreateDigitalWorkOrderDraftBatch(context.Background(), batch)

	require.NoError(t, err)
	require.Len(t, result, 2)
	require.Len(t, created, 2)

	require.Equal(t, "D-1.1", created[0].Number)
	require.Equal(t, "D-1.2", created[1].Number)

	require.True(t, created[0].Items[0].TotalUsed.Equal(decimal.NewFromInt(60)))
	require.True(t, created[1].Items[0].TotalUsed.Equal(decimal.NewFromInt(60)))

	require.True(t, created[0].Items[0].FinalDose.Equal(decimal.NewFromInt(1)))
	require.True(t, created[1].Items[0].FinalDose.Equal(decimal.NewFromInt(1)))
}

func TestCreateDigitalWorkOrderDraftBatch_DistributesTotalUsedByLotArea(t *testing.T) {
	var created []*domain.WorkOrderDraft

	repo := &testDraftRepo{
		listOccupiedFn: func(ctx context.Context, projectID int64) ([]string, error) {
			return []string{}, nil
		},
		createBatchFn: func(ctx context.Context, drafts []*domain.WorkOrderDraft) ([]int64, error) {
			created = drafts
			return []int64{101, 102}, nil
		},
	}

	uc := NewUseCases(repo, &testPublisher{}, &testSupplyReader{})

	batch := &domain.WorkOrderDraftBatchCreate{
		Date:       time.Date(2026, 5, 14, 0, 0, 0, 0, time.UTC),
		CustomerID: 1,
		ProjectID:  10,
		FieldID:    20,
		CropID:     30,
		LaborID:    40,
		Contractor: "Contratista",
		InvestorID: 50,
		Lots: []domain.WorkOrderDraftBatchLot{
			{
				LotID:         101,
				EffectiveArea: decimal.NewFromInt(25),
				Items: []domain.WorkOrderDraftBatchLotItem{
					{SupplyID: 999, TotalUsed: decimal.NewFromInt(200)},
				},
			},
			{
				LotID:         102,
				EffectiveArea: decimal.NewFromInt(75),
				Items: []domain.WorkOrderDraftBatchLotItem{
					{SupplyID: 999, TotalUsed: decimal.NewFromInt(200)},
				},
			},
		},
	}

	_, err := uc.CreateDigitalWorkOrderDraftBatch(context.Background(), batch)

	require.NoError(t, err)
	require.Len(t, created, 2)
	require.True(t, created[0].Items[0].TotalUsed.Equal(decimal.NewFromInt(50)))
	require.True(t, created[1].Items[0].TotalUsed.Equal(decimal.NewFromInt(150)))
	require.True(t, created[0].Items[0].FinalDose.Equal(decimal.NewFromInt(2)))
	require.True(t, created[1].Items[0].FinalDose.Equal(decimal.NewFromInt(2)))
}

func TestCreateDigitalWorkOrderDraftBatch_AdjustsLastLotForDecimalResidue(t *testing.T) {
	var created []*domain.WorkOrderDraft

	repo := &testDraftRepo{
		listOccupiedFn: func(ctx context.Context, projectID int64) ([]string, error) {
			return []string{}, nil
		},
		createBatchFn: func(ctx context.Context, drafts []*domain.WorkOrderDraft) ([]int64, error) {
			created = drafts
			return []int64{101, 102}, nil
		},
	}

	uc := NewUseCases(repo, &testPublisher{}, &testSupplyReader{})

	batch := &domain.WorkOrderDraftBatchCreate{
		Date:       time.Date(2026, 5, 14, 0, 0, 0, 0, time.UTC),
		CustomerID: 1,
		ProjectID:  10,
		FieldID:    20,
		CropID:     30,
		LaborID:    40,
		Contractor: "Contratista",
		InvestorID: 50,
		Lots: []domain.WorkOrderDraftBatchLot{
			{
				LotID:         101,
				EffectiveArea: decimal.NewFromInt(1),
				Items: []domain.WorkOrderDraftBatchLotItem{
					{SupplyID: 999, TotalUsed: decimal.NewFromInt(100)},
				},
			},
			{
				LotID:         102,
				EffectiveArea: decimal.NewFromInt(2),
				Items: []domain.WorkOrderDraftBatchLotItem{
					{SupplyID: 999, TotalUsed: decimal.NewFromInt(100)},
				},
			},
		},
	}

	_, err := uc.CreateDigitalWorkOrderDraftBatch(context.Background(), batch)

	require.NoError(t, err)
	require.Len(t, created, 2)
	require.True(t, created[0].Items[0].TotalUsed.Equal(decimal.RequireFromString("33.333333")))
	require.True(t, created[1].Items[0].TotalUsed.Equal(decimal.RequireFromString("66.666667")))
	sum := created[0].Items[0].TotalUsed.Add(created[1].Items[0].TotalUsed)
	require.True(t, sum.Equal(decimal.NewFromInt(100)))
}

func TestCreateDigitalWorkOrderDraftBatch_RejectsDifferentSupplySets(t *testing.T) {
	repo := &testDraftRepo{
		listOccupiedFn: func(ctx context.Context, projectID int64) ([]string, error) {
			return []string{}, nil
		},
	}

	uc := NewUseCases(repo, &testPublisher{}, &testSupplyReader{})

	batch := &domain.WorkOrderDraftBatchCreate{
		Date:       time.Date(2026, 5, 14, 0, 0, 0, 0, time.UTC),
		CustomerID: 1,
		ProjectID:  10,
		FieldID:    20,
		CropID:     30,
		LaborID:    40,
		Contractor: "Contratista",
		InvestorID: 50,
		Lots: []domain.WorkOrderDraftBatchLot{
			{
				LotID:         101,
				EffectiveArea: decimal.NewFromInt(50),
				Items: []domain.WorkOrderDraftBatchLotItem{
					{SupplyID: 999, TotalUsed: decimal.NewFromInt(200)},
				},
			},
			{
				LotID:         102,
				EffectiveArea: decimal.NewFromInt(50),
				Items: []domain.WorkOrderDraftBatchLotItem{
					{SupplyID: 888, TotalUsed: decimal.NewFromInt(200)},
				},
			},
		},
	}

	_, err := uc.CreateDigitalWorkOrderDraftBatch(context.Background(), batch)
	require.Error(t, err)
	require.Contains(t, err.Error(), "same supply_id set")
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

	uc := NewUseCases(repo, &testPublisher{}, &testSupplyReader{})

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

	uc := NewUseCases(repo, pub, &testSupplyReader{})

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

	uc := NewUseCases(repo, pub, &testSupplyReader{})

	_, err := uc.PublishWorkOrderDraft(context.Background(), 88)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot publish work order draft with pending supplies")
	require.Contains(t, err.Error(), "2,4D ENLIST")
}

func TestPublishWorkOrderDraft_NormalizesDigitalTotalUsed(t *testing.T) {
	repo := &testDraftRepo{
		getByIDFn: func(ctx context.Context, id int64) (*domain.WorkOrderDraft, error) {
			d := validDraft()
			d.ID = id
			d.Number = "D-60"
			d.IsDigital = true
			d.Status = domain.StatusDraft
			d.EffectiveArea = decimal.NewFromInt(60)
			d.Items = []domain.WorkOrderDraftItem{
				{
					SupplyID:  70,
					TotalUsed: decimal.NewFromInt(120),
					FinalDose: decimal.NewFromInt(1),
				},
			}
			return d, nil
		},
		listPublishedFn: func(ctx context.Context, projectID int64) ([]string, error) {
			return []string{}, nil
		},
		markPublishedFn: func(ctx context.Context, draftID int64, workOrderID int64) error {
			return nil
		},
	}

	var published *workorderdomain.WorkOrder
	pub := &testPublisher{
		createFn: func(ctx context.Context, wo *workorderdomain.WorkOrder) (int64, error) {
			published = wo
			return 999, nil
		},
	}

	uc := NewUseCases(repo, pub, &testSupplyReader{})

	_, err := uc.PublishWorkOrderDraft(context.Background(), 99)
	require.NoError(t, err)
	require.NotNil(t, published)
	require.Len(t, published.Items, 1)
	require.True(t, published.Items[0].TotalUsed.Equal(decimal.NewFromInt(60)))
	require.True(t, published.Items[0].FinalDose.Equal(decimal.NewFromInt(1)))
	require.True(t, published.IsDigital)
}

// --- Helpers para tests de grupo -------------------------------------------------

// groupDraft construye un draft "digital perteneciente a un grupo" listo para
// alimentar mocks de getByIDFn / listRelatedFn.
func groupDraft(id int64, number string, status domain.Status, area decimal.Decimal) *domain.WorkOrderDraft {
	d := validDraft()
	d.ID = id
	d.Number = number
	d.IsDigital = true
	d.ProjectID = 10
	d.Status = status
	d.EffectiveArea = area
	return d
}

// groupUpdateRequest devuelve un payload válido para UpdateWorkOrderDraftGroupByID.
func groupUpdateRequest() *domain.WorkOrderDraftGroup {
	return &domain.WorkOrderDraftGroup{
		Date:       time.Date(2026, 5, 14, 0, 0, 0, 0, time.UTC),
		CustomerID: 1,
		ProjectID:  10,
		FieldID:    20,
		CropID:     40,
		LaborID:    50,
		Contractor: "Contratista",
		InvestorID: 60,
		Items: []domain.WorkOrderDraftItem{
			{
				SupplyID:  70,
				TotalUsed: decimal.NewFromInt(120),
				FinalDose: decimal.NewFromInt(1),
			},
		},
	}
}

func TestUpdateWorkOrderDraftGroupByID_HappyPathSendsAllLotsInSingleCall(t *testing.T) {
	groupRelated := []*domain.WorkOrderDraft{
		groupDraft(101, "D-7.1", domain.StatusDraft, decimal.NewFromInt(60)),
		groupDraft(102, "D-7.2", domain.StatusDraft, decimal.NewFromInt(60)),
	}

	var (
		updateGroupCalls int
		updateOneCalls   int
		captured         []*domain.WorkOrderDraft
	)

	repo := &testDraftRepo{
		getByIDFn: func(ctx context.Context, id int64) (*domain.WorkOrderDraft, error) {
			return groupRelated[0], nil
		},
		listRelatedFn: func(ctx context.Context, projectID int64, baseNumber string) ([]*domain.WorkOrderDraft, error) {
			require.Equal(t, int64(10), projectID)
			require.Equal(t, "D-7", baseNumber)
			return groupRelated, nil
		},
		updateFn: func(ctx context.Context, d *domain.WorkOrderDraft) error {
			updateOneCalls++
			return nil
		},
		updateGroupFn: func(ctx context.Context, drafts []*domain.WorkOrderDraft) error {
			updateGroupCalls++
			captured = drafts
			return nil
		},
	}

	uc := NewUseCases(repo, &testPublisher{}, &testSupplyReader{})

	err := uc.UpdateWorkOrderDraftGroupByID(context.Background(), 101, groupUpdateRequest())
	require.NoError(t, err)
	require.Equal(t, 1, updateGroupCalls, "atomic group method must be called exactly once")
	require.Equal(t, 0, updateOneCalls, "single-draft update must not be used for group updates")
	require.Len(t, captured, 2)
	require.Equal(t, int64(101), captured[0].ID)
	require.Equal(t, int64(102), captured[1].ID)
	require.True(t, captured[0].Items[0].TotalUsed.Equal(decimal.NewFromInt(60)))
	require.True(t, captured[1].Items[0].TotalUsed.Equal(decimal.NewFromInt(60)))
	require.True(t, captured[0].Items[0].FinalDose.Equal(decimal.NewFromInt(1)))
	require.True(t, captured[1].Items[0].FinalDose.Equal(decimal.NewFromInt(1)))
}

func TestUpdateWorkOrderDraftGroupByID_AtomicErrorFromRepoIsPropagated(t *testing.T) {
	groupRelated := []*domain.WorkOrderDraft{
		groupDraft(201, "D-9.1", domain.StatusDraft, decimal.NewFromInt(30)),
		groupDraft(202, "D-9.2", domain.StatusDraft, decimal.NewFromInt(30)),
		groupDraft(203, "D-9.3", domain.StatusDraft, decimal.NewFromInt(30)),
	}

	var (
		updateGroupCalls int
		updateOneCalls   int
	)

	repoErr := types.NewError(types.ErrInternal, "boom from repo", nil)

	repo := &testDraftRepo{
		getByIDFn: func(ctx context.Context, id int64) (*domain.WorkOrderDraft, error) {
			return groupRelated[0], nil
		},
		listRelatedFn: func(ctx context.Context, projectID int64, baseNumber string) ([]*domain.WorkOrderDraft, error) {
			return groupRelated, nil
		},
		updateFn: func(ctx context.Context, d *domain.WorkOrderDraft) error {
			updateOneCalls++
			return nil
		},
		updateGroupFn: func(ctx context.Context, drafts []*domain.WorkOrderDraft) error {
			updateGroupCalls++
			require.Len(t, drafts, 3, "all lots must reach the repo together")
			return repoErr
		},
	}

	uc := NewUseCases(repo, &testPublisher{}, &testSupplyReader{})

	err := uc.UpdateWorkOrderDraftGroupByID(context.Background(), 201, groupUpdateRequest())
	require.ErrorIs(t, err, repoErr)
	require.Equal(t, 1, updateGroupCalls)
	require.Equal(t, 0, updateOneCalls, "single-draft path must not be reachable from group update")
}

func TestUpdateWorkOrderDraftGroupByID_BlocksWhenAnyLotPublished(t *testing.T) {
	groupRelated := []*domain.WorkOrderDraft{
		groupDraft(301, "D-11.1", domain.StatusDraft, decimal.NewFromInt(40)),
		groupDraft(302, "D-11.2", domain.StatusPublished, decimal.NewFromInt(40)),
	}

	var updateGroupCalls int

	repo := &testDraftRepo{
		getByIDFn: func(ctx context.Context, id int64) (*domain.WorkOrderDraft, error) {
			return groupRelated[0], nil
		},
		listRelatedFn: func(ctx context.Context, projectID int64, baseNumber string) ([]*domain.WorkOrderDraft, error) {
			return groupRelated, nil
		},
		updateGroupFn: func(ctx context.Context, drafts []*domain.WorkOrderDraft) error {
			updateGroupCalls++
			return nil
		},
	}

	uc := NewUseCases(repo, &testPublisher{}, &testSupplyReader{})

	err := uc.UpdateWorkOrderDraftGroupByID(context.Background(), 301, groupUpdateRequest())
	require.Error(t, err)
	errType, ok := types.GetErrorType(err)
	require.True(t, ok)
	require.Equal(t, types.ErrConflict, errType)
	require.Equal(t, 0, updateGroupCalls, "repo must NOT be called when any lot is published")
}

func TestUpdateWorkOrderDraftGroupByID_RejectsZeroEffectiveArea(t *testing.T) {
	groupRelated := []*domain.WorkOrderDraft{
		groupDraft(401, "D-13.1", domain.StatusDraft, decimal.Zero),
		groupDraft(402, "D-13.2", domain.StatusDraft, decimal.Zero),
	}

	var updateGroupCalls int

	repo := &testDraftRepo{
		getByIDFn: func(ctx context.Context, id int64) (*domain.WorkOrderDraft, error) {
			return groupRelated[0], nil
		},
		listRelatedFn: func(ctx context.Context, projectID int64, baseNumber string) ([]*domain.WorkOrderDraft, error) {
			return groupRelated, nil
		},
		updateGroupFn: func(ctx context.Context, drafts []*domain.WorkOrderDraft) error {
			updateGroupCalls++
			return nil
		},
	}

	uc := NewUseCases(repo, &testPublisher{}, &testSupplyReader{})

	err := uc.UpdateWorkOrderDraftGroupByID(context.Background(), 401, groupUpdateRequest())
	require.Error(t, err)
	errType, ok := types.GetErrorType(err)
	require.True(t, ok)
	require.Equal(t, types.ErrValidation, errType)
	require.Equal(t, 0, updateGroupCalls)
}

func TestGetWorkOrderDraftGroupByID_RejectsInvalidNumber(t *testing.T) {
	repo := &testDraftRepo{
		getByIDFn: func(ctx context.Context, id int64) (*domain.WorkOrderDraft, error) {
			d := validDraft()
			d.ID = id
			d.Number = "NOT-A-NUMBER"
			d.IsDigital = true
			return d, nil
		},
	}
	uc := NewUseCases(repo, &testPublisher{}, &testSupplyReader{})

	_, err := uc.GetWorkOrderDraftGroupByID(context.Background(), 555)
	require.Error(t, err)
	errType, ok := types.GetErrorType(err)
	require.True(t, ok)
	require.Equal(t, types.ErrValidation, errType)
}

func TestGetWorkOrderDraftGroupByID_RejectsWhenNoRelatedFound(t *testing.T) {
	repo := &testDraftRepo{
		getByIDFn: func(ctx context.Context, id int64) (*domain.WorkOrderDraft, error) {
			d := validDraft()
			d.ID = id
			d.Number = "D-77"
			d.IsDigital = true
			return d, nil
		},
		listRelatedFn: func(ctx context.Context, projectID int64, baseNumber string) ([]*domain.WorkOrderDraft, error) {
			return []*domain.WorkOrderDraft{}, nil
		},
	}
	uc := NewUseCases(repo, &testPublisher{}, &testSupplyReader{})

	_, err := uc.GetWorkOrderDraftGroupByID(context.Background(), 777)
	require.Error(t, err)
	errType, ok := types.GetErrorType(err)
	require.True(t, ok)
	require.Equal(t, types.ErrNotFound, errType)
}

func TestGetWorkOrderDraftGroupByID_AggregatesItemsAcrossLots(t *testing.T) {
	groupRelated := []*domain.WorkOrderDraft{
		groupDraft(101, "D-7.1", domain.StatusDraft, decimal.NewFromInt(50)),
		groupDraft(102, "D-7.2", domain.StatusDraft, decimal.NewFromInt(50)),
	}
	groupRelated[0].Items = []domain.WorkOrderDraftItem{
		{SupplyID: 70, SupplyName: "INSUMO", TotalUsed: decimal.NewFromInt(100), FinalDose: decimal.NewFromInt(2)},
	}
	groupRelated[1].Items = []domain.WorkOrderDraftItem{
		{SupplyID: 70, SupplyName: "INSUMO", TotalUsed: decimal.NewFromInt(100), FinalDose: decimal.NewFromInt(2)},
	}

	repo := &testDraftRepo{
		getByIDFn: func(ctx context.Context, id int64) (*domain.WorkOrderDraft, error) {
			return groupRelated[0], nil
		},
		listRelatedFn: func(ctx context.Context, projectID int64, baseNumber string) ([]*domain.WorkOrderDraft, error) {
			return groupRelated, nil
		},
	}
	uc := NewUseCases(repo, &testPublisher{}, &testSupplyReader{})

	group, err := uc.GetWorkOrderDraftGroupByID(context.Background(), 101)

	require.NoError(t, err)
	require.Equal(t, "D-7", group.Number)
	require.True(t, group.EffectiveArea.Equal(decimal.NewFromInt(100)))
	require.Len(t, group.Items, 1)
	require.True(t, group.Items[0].TotalUsed.Equal(decimal.NewFromInt(200)))
	require.True(t, group.Items[0].FinalDose.Equal(decimal.NewFromInt(2)))
}

func TestListDigitalWorkOrderDraftGroups_DelegatesToRepo(t *testing.T) {
	expected := []domain.WorkOrderDraftGroupListItem{
		{ID: 1, Number: "D-1", ProjectID: 10, Status: domain.StatusDraft},
		{ID: 2, Number: "D-2", ProjectID: 10, Status: domain.StatusPendingReview},
	}
	expectedPage := types.PageInfo{Page: 1, PerPage: 10, MaxPage: 1, Total: 2}

	var captured struct {
		number string
		status string
		inp    types.Input
	}

	repo := &testDraftRepo{
		listGroupsFn: func(ctx context.Context, number string, status string, inp types.Input) ([]domain.WorkOrderDraftGroupListItem, types.PageInfo, error) {
			captured.number = number
			captured.status = status
			captured.inp = inp
			return expected, expectedPage, nil
		},
	}

	uc := NewUseCases(repo, &testPublisher{}, &testSupplyReader{})

	items, page, err := uc.ListDigitalWorkOrderDraftGroups(
		context.Background(),
		"D-",
		"draft",
		types.Input{Page: 1, PageSize: 10},
	)
	require.NoError(t, err)
	require.Equal(t, expected, items)
	require.Equal(t, expectedPage, page)
	require.Equal(t, "D-", captured.number)
	require.Equal(t, "draft", captured.status)
	require.Equal(t, types.Input{Page: 1, PageSize: 10}, captured.inp)
}
