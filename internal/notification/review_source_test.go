package notification

import (
	"context"
	"errors"
	"testing"
)

type fakePendingApprovalSourceClient struct {
	status int
	body   []byte
	err    error
}

func (f fakePendingApprovalSourceClient) ListPendingApprovals(ctx context.Context) (int, []byte, error) {
	return f.status, f.body, f.err
}

func TestReviewPendingApprovalSourceListPendingApprovals(t *testing.T) {
	t.Parallel()

	source := NewReviewPendingApprovalSource(fakePendingApprovalSourceClient{
		status: 200,
		body:   []byte(`{"approvals":[{"id":"appr-1","org_id":"org-1","request_id":"req-1","action_type":"approve_expense","target_resource":"expense:123","reason":"Monto alto","risk_level":"high","status":"pending","created_at":"2026-04-09T10:30:00Z"}]}`),
	})

	approvals, err := source.ListPendingApprovals(context.Background())
	if err != nil {
		t.Fatalf("ListPendingApprovals() error = %v", err)
	}
	if len(approvals) != 1 {
		t.Fatalf("expected 1 approval, got %d", len(approvals))
	}
	if approvals[0].ID != "appr-1" {
		t.Fatalf("unexpected approval id %q", approvals[0].ID)
	}
}

func TestReviewPendingApprovalSourceError(t *testing.T) {
	t.Parallel()

	source := NewReviewPendingApprovalSource(fakePendingApprovalSourceClient{
		status: 500,
		body:   []byte(`{"message":"boom"}`),
		err:    errors.New("upstream"),
	})

	if _, err := source.ListPendingApprovals(context.Background()); err == nil {
		t.Fatal("expected error")
	}
}
