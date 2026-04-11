package notification

import (
	"context"
	"strings"
	"sync"
	"time"

	coreinbox "github.com/devpablocristo/core/notifications/go/inbox"
	"github.com/google/uuid"
)

type ApprovalSyncService struct {
	source   ApprovalSource
	inbox    *coreinbox.Usecases
	repo     *Repository
	cooldown time.Duration
	throttle sync.Map
}

func NewApprovalSyncService(repo *Repository, source ApprovalSource, cooldownSec int) *ApprovalSyncService {
	if cooldownSec <= 0 {
		cooldownSec = 60
	}
	return &ApprovalSyncService{
		source:   source,
		inbox:    coreinbox.NewUsecases(repo),
		repo:     repo,
		cooldown: time.Duration(cooldownSec) * time.Second,
	}
}

func (s *ApprovalSyncService) Enabled() bool {
	return s != nil && s.source != nil && s.inbox != nil && s.repo != nil
}

func (s *ApprovalSyncService) SyncForActor(ctx context.Context, orgID uuid.UUID, actor string) (int, error) {
	if !s.Enabled() || orgID == uuid.Nil || strings.TrimSpace(actor) == "" {
		return 0, nil
	}
	key := orgID.String() + ":" + strings.TrimSpace(actor)
	if s.shouldSkip(key) {
		return 0, nil
	}
	approvals, err := s.source.ListPendingApprovals(ctx)
	if err != nil {
		return 0, err
	}
	filtered := filterApprovalsByOrg(orgID, approvals)
	pendingKeys := make(map[string]struct{}, len(filtered))
	created := 0
	for _, approval := range filtered {
		notification, err := buildApprovalNotification(orgID, actor, approval)
		if err != nil {
			return created, err
		}
		if _, err := s.inbox.Create(ctx, notification); err != nil {
			return created, err
		}
		pendingKeys[notification.ID] = struct{}{}
		created++
	}
	if err := s.repo.ResolveStaleApprovals(ctx, orgID, actor, pendingKeys); err != nil {
		return created, err
	}
	return created, nil
}

func (s *ApprovalSyncService) shouldSkip(key string) bool {
	now := time.Now().UTC()
	last, ok := s.throttle.Load(key)
	if ok {
		if now.Sub(last.(time.Time)) < s.cooldown {
			return true
		}
	}
	s.throttle.Store(key, now)
	return false
}

func filterApprovalsByOrg(orgID uuid.UUID, approvals []PendingApproval) []PendingApproval {
	out := make([]PendingApproval, 0, len(approvals))
	for _, approval := range approvals {
		approvalOrg := strings.TrimSpace(approval.OrgID)
		if approvalOrg == "" || approvalOrg != orgID.String() {
			continue
		}
		out = append(out, approval)
	}
	return out
}
