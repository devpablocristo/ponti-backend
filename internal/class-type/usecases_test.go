package classtype

import (
	"context"
	"errors"
	"testing"

	domain "github.com/devpablocristo/ponti-backend/internal/class-type/usecases/domain"
)

type fakeRepo struct {
	createArg    *domain.ClassType
	createID     int64
	createErr    error
	listPage     int
	listPerPage  int
	listResult   []domain.ClassType
	listTotal    int64
	getID        int64
	getResult    *domain.ClassType
	getErr       error
	updateArg    *domain.ClassType
	archiveID    int64
	archiveErr   error
	restoreID    int64
	hardDeleteID int64
}

func (f *fakeRepo) CreateClassType(_ context.Context, c *domain.ClassType) (int64, error) {
	f.createArg = c
	return f.createID, f.createErr
}

func (f *fakeRepo) ListClassTypes(_ context.Context, page, perPage int) ([]domain.ClassType, int64, error) {
	f.listPage = page
	f.listPerPage = perPage
	return f.listResult, f.listTotal, nil
}

func (f *fakeRepo) ListArchivedClassTypes(_ context.Context, page, perPage int) ([]domain.ClassType, int64, error) {
	f.listPage = page
	f.listPerPage = perPage
	return f.listResult, f.listTotal, nil
}

func (f *fakeRepo) GetClassType(_ context.Context, id int64) (*domain.ClassType, error) {
	f.getID = id
	return f.getResult, f.getErr
}

func (f *fakeRepo) UpdateClassType(_ context.Context, c *domain.ClassType) error {
	f.updateArg = c
	return nil
}

func (f *fakeRepo) ArchiveClassType(_ context.Context, id int64) error {
	f.archiveID = id
	return f.archiveErr
}

func (f *fakeRepo) RestoreClassType(_ context.Context, id int64) error {
	f.restoreID = id
	return nil
}

func (f *fakeRepo) HardDeleteClassType(_ context.Context, id int64) error {
	f.hardDeleteID = id
	return nil
}

func TestUseCases_CreateClassType_Delegates(t *testing.T) {
	repo := &fakeRepo{createID: 7}
	uc := NewUseCases(repo)

	in := &domain.ClassType{Name: "Agroquímicos"}
	id, err := uc.CreateClassType(context.Background(), in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 7 {
		t.Fatalf("expected id 7, got %d", id)
	}
	if repo.createArg.Name != "Agroquímicos" {
		t.Fatalf("expected forwarded class type Name, got %+v", repo.createArg)
	}
}

func TestUseCases_CreateClassType_PropagatesError(t *testing.T) {
	repo := &fakeRepo{createErr: errors.New("duplicate name")}
	uc := NewUseCases(repo)

	_, err := uc.CreateClassType(context.Background(), &domain.ClassType{Name: "X"})
	if err == nil || err.Error() != "duplicate name" {
		t.Fatalf("expected propagated 'duplicate name', got %v", err)
	}
}

func TestUseCases_ListClassTypes_ForwardsPagination(t *testing.T) {
	repo := &fakeRepo{listResult: []domain.ClassType{{ID: 1}, {ID: 2}}, listTotal: 30}
	uc := NewUseCases(repo)

	got, total, err := uc.ListClassTypes(context.Background(), 3, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 30 {
		t.Fatalf("expected total 30, got %d", total)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 results, got %d", len(got))
	}
	if repo.listPage != 3 || repo.listPerPage != 10 {
		t.Fatalf("expected page=3 perPage=10, got page=%d perPage=%d", repo.listPage, repo.listPerPage)
	}
}

func TestUseCases_ListArchived_ForwardsPagination(t *testing.T) {
	repo := &fakeRepo{listResult: nil, listTotal: 0}
	uc := NewUseCases(repo)

	_, _, err := uc.ListArchivedClassTypes(context.Background(), 1, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.listPage != 1 || repo.listPerPage != 5 {
		t.Fatalf("expected page=1 perPage=5, got page=%d perPage=%d", repo.listPage, repo.listPerPage)
	}
}

func TestUseCases_GetClassType_ReturnsRepoResult(t *testing.T) {
	want := &domain.ClassType{ID: 42, Name: "Semillas"}
	repo := &fakeRepo{getResult: want}
	uc := NewUseCases(repo)

	got, err := uc.GetClassType(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Fatalf("expected same pointer back, got %+v", got)
	}
	if repo.getID != 42 {
		t.Fatalf("expected id 42 forwarded, got %d", repo.getID)
	}
}

func TestUseCases_Update_Archive_Restore_HardDelete(t *testing.T) {
	repo := &fakeRepo{}
	uc := NewUseCases(repo)
	ctx := context.Background()

	in := &domain.ClassType{ID: 5}
	if err := uc.UpdateClassType(ctx, in); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if repo.updateArg != in {
		t.Fatalf("UpdateClassType should forward same pointer")
	}

	if err := uc.ArchiveClassType(ctx, 5); err != nil {
		t.Fatalf("Archive: %v", err)
	}
	if repo.archiveID != 5 {
		t.Fatalf("expected archiveID=5, got %d", repo.archiveID)
	}

	if err := uc.RestoreClassType(ctx, 5); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if repo.restoreID != 5 {
		t.Fatalf("expected restoreID=5, got %d", repo.restoreID)
	}

	if err := uc.HardDeleteClassType(ctx, 5); err != nil {
		t.Fatalf("HardDelete: %v", err)
	}
	if repo.hardDeleteID != 5 {
		t.Fatalf("expected hardDeleteID=5, got %d", repo.hardDeleteID)
	}
}
