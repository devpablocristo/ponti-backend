package actor

import (
	"context"
	"errors"
	"testing"

	domain "github.com/devpablocristo/ponti-backend/internal/actor/usecases/domain"
)

// fakeRepo implementa RepositoryPort sin tocar DB. Captura los argumentos
// que UseCases delega y devuelve los stubs que el caller espera.
type fakeRepo struct {
	createCalled bool
	createArg    *domain.Actor
	createID     int64
	createErr    error

	listFilters domain.ListFilters
	listPage    int
	listPerPage int
	listActors  []domain.Actor
	listTotal   int64

	getID    int64
	getActor *domain.Actor
	getErr   error

	updateCalled bool
	updateArg    *domain.Actor

	archiveID  int64
	archiveErr error

	restoreID  int64
	restoreErr error

	hardDeleteID int64

	addRoleID   int64
	addRoleRole string

	addAliasID    int64
	addAliasAlias domain.ActorAlias
	addAliasNew   int64

	listDuplicates []domain.DuplicateCandidate

	mergeReq    domain.MergeRequest
	mergeImpact *domain.MergeImpact
}

func (f *fakeRepo) CreateActor(_ context.Context, a *domain.Actor) (int64, error) {
	f.createCalled = true
	f.createArg = a
	return f.createID, f.createErr
}

func (f *fakeRepo) ListActors(_ context.Context, filters domain.ListFilters, page, perPage int) ([]domain.Actor, int64, error) {
	f.listFilters = filters
	f.listPage = page
	f.listPerPage = perPage
	return f.listActors, f.listTotal, nil
}

func (f *fakeRepo) GetActor(_ context.Context, id int64) (*domain.Actor, error) {
	f.getID = id
	return f.getActor, f.getErr
}

func (f *fakeRepo) UpdateActor(_ context.Context, a *domain.Actor) error {
	f.updateCalled = true
	f.updateArg = a
	return nil
}

func (f *fakeRepo) ArchiveActor(_ context.Context, id int64) error {
	f.archiveID = id
	return f.archiveErr
}

func (f *fakeRepo) RestoreActor(_ context.Context, id int64) error {
	f.restoreID = id
	return f.restoreErr
}

func (f *fakeRepo) HardDeleteActor(_ context.Context, id int64) error {
	f.hardDeleteID = id
	return nil
}

func (f *fakeRepo) AddRole(_ context.Context, id int64, role string) error {
	f.addRoleID = id
	f.addRoleRole = role
	return nil
}

func (f *fakeRepo) AddAlias(_ context.Context, id int64, alias domain.ActorAlias) (int64, error) {
	f.addAliasID = id
	f.addAliasAlias = alias
	return f.addAliasNew, nil
}

func (f *fakeRepo) ListDuplicateCandidates(_ context.Context) ([]domain.DuplicateCandidate, error) {
	return f.listDuplicates, nil
}

func (f *fakeRepo) MergeActors(_ context.Context, req domain.MergeRequest) (*domain.MergeImpact, error) {
	f.mergeReq = req
	return f.mergeImpact, nil
}

// --- tests ---

func TestUseCases_CreateActor_DelegatesAndReturnsID(t *testing.T) {
	repo := &fakeRepo{createID: 42}
	uc := NewUseCases(repo)

	actor := &domain.Actor{ActorKind: "organization", DisplayName: "Acme S.A."}
	id, err := uc.CreateActor(context.Background(), actor)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 42 {
		t.Fatalf("expected id 42, got %d", id)
	}
	if !repo.createCalled {
		t.Fatal("CreateActor should have called repo.CreateActor")
	}
	if repo.createArg.DisplayName != "Acme S.A." {
		t.Fatalf("expected actor to be passed as-is, got %+v", repo.createArg)
	}
}

func TestUseCases_CreateActor_PropagatesError(t *testing.T) {
	repo := &fakeRepo{createErr: errors.New("db down")}
	uc := NewUseCases(repo)

	_, err := uc.CreateActor(context.Background(), &domain.Actor{ActorKind: "organization"})
	if err == nil || err.Error() != "db down" {
		t.Fatalf("expected propagated 'db down' error, got %v", err)
	}
}

func TestUseCases_ListActors_PassesPagination(t *testing.T) {
	repo := &fakeRepo{listTotal: 7, listActors: []domain.Actor{{ID: 1}, {ID: 2}}}
	uc := NewUseCases(repo)

	filters := domain.ListFilters{Role: "cliente"}
	got, total, err := uc.ListActors(context.Background(), filters, 2, 25)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 7 {
		t.Fatalf("expected total 7, got %d", total)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 actors, got %d", len(got))
	}
	if repo.listPage != 2 || repo.listPerPage != 25 {
		t.Fatalf("expected page=2 perPage=25, got page=%d perPage=%d", repo.listPage, repo.listPerPage)
	}
	if repo.listFilters.Role != "cliente" {
		t.Fatalf("expected filters.Role=cliente, got %+v", repo.listFilters)
	}
}

func TestUseCases_GetActor_ReturnsRepoResult(t *testing.T) {
	want := &domain.Actor{ID: 5, ActorKind: "natural_person", DisplayName: "Jane"}
	repo := &fakeRepo{getActor: want}
	uc := NewUseCases(repo)

	got, err := uc.GetActor(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Fatalf("expected same pointer back, got %+v", got)
	}
	if repo.getID != 5 {
		t.Fatalf("expected id 5 forwarded, got %d", repo.getID)
	}
}

func TestUseCases_UpdateActor_Delegates(t *testing.T) {
	repo := &fakeRepo{}
	uc := NewUseCases(repo)

	in := &domain.Actor{ID: 9, ActorKind: "organization"}
	if err := uc.UpdateActor(context.Background(), in); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !repo.updateCalled || repo.updateArg != in {
		t.Fatalf("UpdateActor should have forwarded same pointer")
	}
}

func TestUseCases_ArchiveRestoreHardDelete(t *testing.T) {
	repo := &fakeRepo{}
	uc := NewUseCases(repo)
	ctx := context.Background()

	if err := uc.ArchiveActor(ctx, 11); err != nil {
		t.Fatalf("Archive: %v", err)
	}
	if repo.archiveID != 11 {
		t.Fatalf("expected archiveID=11, got %d", repo.archiveID)
	}

	if err := uc.RestoreActor(ctx, 12); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if repo.restoreID != 12 {
		t.Fatalf("expected restoreID=12, got %d", repo.restoreID)
	}

	if err := uc.HardDeleteActor(ctx, 13); err != nil {
		t.Fatalf("HardDelete: %v", err)
	}
	if repo.hardDeleteID != 13 {
		t.Fatalf("expected hardDeleteID=13, got %d", repo.hardDeleteID)
	}
}

func TestUseCases_AddRoleAndAlias(t *testing.T) {
	repo := &fakeRepo{addAliasNew: 99}
	uc := NewUseCases(repo)
	ctx := context.Background()

	if err := uc.AddRole(ctx, 7, "inversor"); err != nil {
		t.Fatalf("AddRole: %v", err)
	}
	if repo.addRoleID != 7 || repo.addRoleRole != "inversor" {
		t.Fatalf("expected id=7 role=inversor, got id=%d role=%q", repo.addRoleID, repo.addRoleRole)
	}

	alias := domain.ActorAlias{Alias: "Acme SA"}
	newID, err := uc.AddAlias(ctx, 7, alias)
	if err != nil {
		t.Fatalf("AddAlias: %v", err)
	}
	if newID != 99 {
		t.Fatalf("expected new alias ID 99, got %d", newID)
	}
	if repo.addAliasID != 7 || repo.addAliasAlias.Alias != "Acme SA" {
		t.Fatalf("expected id=7 alias=Acme SA, got id=%d alias=%q", repo.addAliasID, repo.addAliasAlias.Alias)
	}
}

func TestUseCases_DuplicatesAndMerge(t *testing.T) {
	repo := &fakeRepo{
		listDuplicates: []domain.DuplicateCandidate{
			{GroupType: "name", GroupKey: "Acme"},
		},
		mergeImpact: &domain.MergeImpact{TargetActorID: 1, SourceActorIDs: []int64{2}, Confirmed: true},
	}
	uc := NewUseCases(repo)
	ctx := context.Background()

	got, err := uc.ListDuplicateCandidates(ctx)
	if err != nil {
		t.Fatalf("ListDuplicateCandidates: %v", err)
	}
	if len(got) != 1 || got[0].GroupType != "name" {
		t.Fatalf("expected one candidate with GroupType=name, got %+v", got)
	}

	req := domain.MergeRequest{TargetActorID: 1, SourceActorIDs: []int64{2}, Confirm: true}
	impact, err := uc.MergeActors(ctx, req)
	if err != nil {
		t.Fatalf("MergeActors: %v", err)
	}
	if impact.TargetActorID != 1 || !impact.Confirmed {
		t.Fatalf("expected impact.TargetActorID=1 and Confirmed=true, got %+v", impact)
	}
	if repo.mergeReq.TargetActorID != 1 || !repo.mergeReq.Confirm {
		t.Fatalf("expected TargetActorID=1 Confirm=true forwarded, got %+v", repo.mergeReq)
	}
}
