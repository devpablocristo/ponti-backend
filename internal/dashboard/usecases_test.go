package dashboard

import (
	"context"
	"errors"
	"testing"

	domain "github.com/devpablocristo/ponti-backend/internal/dashboard/usecases/domain"
)

type fakeRepo struct {
	calledFilter domain.DashboardFilter
	result       *domain.DashboardData
	err          error
}

func (f *fakeRepo) GetDashboard(_ context.Context, filter domain.DashboardFilter) (*domain.DashboardData, error) {
	f.calledFilter = filter
	return f.result, f.err
}

func ptrInt64(v int64) *int64 { return &v }

func TestUseCases_GetDashboard_ForwardsFilter(t *testing.T) {
	wantData := &domain.DashboardData{Metrics: &domain.DashboardMetrics{}}
	repo := &fakeRepo{result: wantData}
	uc := NewUseCases(repo)

	filter := domain.DashboardFilter{
		CustomerID: ptrInt64(7),
		ProjectID:  ptrInt64(42),
		CampaignID: ptrInt64(3),
	}
	got, err := uc.GetDashboard(context.Background(), filter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != wantData {
		t.Fatalf("expected same pointer back, got %+v", got)
	}
	if *repo.calledFilter.CustomerID != 7 ||
		*repo.calledFilter.ProjectID != 42 ||
		*repo.calledFilter.CampaignID != 3 {
		t.Fatalf("filter not forwarded correctly: %+v", repo.calledFilter)
	}
	if repo.calledFilter.FieldID != nil {
		t.Fatalf("expected FieldID nil, got %v", repo.calledFilter.FieldID)
	}
}

func TestUseCases_GetDashboard_PropagatesError(t *testing.T) {
	repo := &fakeRepo{err: errors.New("db timeout")}
	uc := NewUseCases(repo)

	_, err := uc.GetDashboard(context.Background(), domain.DashboardFilter{})
	if err == nil || err.Error() != "db timeout" {
		t.Fatalf("expected propagated 'db timeout', got %v", err)
	}
}

func TestUseCases_GetDashboard_EmptyFilter(t *testing.T) {
	repo := &fakeRepo{result: &domain.DashboardData{}}
	uc := NewUseCases(repo)

	_, err := uc.GetDashboard(context.Background(), domain.DashboardFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.calledFilter.CustomerID != nil ||
		repo.calledFilter.ProjectID != nil ||
		repo.calledFilter.CampaignID != nil ||
		repo.calledFilter.FieldID != nil {
		t.Fatalf("expected all-nil filter fields, got %+v", repo.calledFilter)
	}
}
