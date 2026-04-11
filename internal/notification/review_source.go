package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/devpablocristo/core/governance/go/reviewclient"
)

type ApprovalSource interface {
	ListPendingApprovals(ctx context.Context) ([]PendingApproval, error)
}

type pendingApprovalSourceClient interface {
	ListPendingApprovals(ctx context.Context) (int, []byte, error)
}

type pendingApprovalListPayload struct {
	Data      []PendingApproval `json:"data"`
	Approvals []PendingApproval `json:"approvals"`
}

type ReviewPendingApprovalSource struct {
	client pendingApprovalSourceClient
}

func NewReviewPendingApprovalSource(client pendingApprovalSourceClient) *ReviewPendingApprovalSource {
	return &ReviewPendingApprovalSource{client: client}
}

func (s *ReviewPendingApprovalSource) ListPendingApprovals(ctx context.Context) ([]PendingApproval, error) {
	status, data, err := s.client.ListPendingApprovals(ctx)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("review pending approvals: status %d body %s", status, reviewclient.ParseErrorBody(data))
	}
	var payload pendingApprovalListPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("decode review approvals: %w", err)
	}
	approvals := payload.Data
	if len(approvals) == 0 && len(payload.Approvals) > 0 {
		approvals = payload.Approvals
	}
	return approvals, nil
}
