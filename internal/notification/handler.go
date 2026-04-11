package notification

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/devpablocristo/core/errors/go/domainerr"
	sharedhandlers "github.com/devpablocristo/ponti-backend/internal/shared/handlers"
)

type GinEnginePort interface {
	GetRouter() *gin.Engine
	RunServer(ctx context.Context) error
}

type ConfigAPIPort interface {
	APIVersion() string
	APIBaseURL() string
}

type MiddlewaresEnginePort interface {
	GetGlobal() []gin.HandlerFunc
	GetValidation() []gin.HandlerFunc
	GetProtected() []gin.HandlerFunc
}

type Handler struct {
	repo      *Repository
	projector *ProjectionService
	approvals *ApprovalSyncService
	gsv       GinEnginePort
	acf       ConfigAPIPort
	mws       MiddlewaresEnginePort
}

func NewHandler(db *gorm.DB, approvals *ApprovalSyncService, s GinEnginePort, c ConfigAPIPort, m MiddlewaresEnginePort) *Handler {
	repo := NewRepository(db)
	return &Handler{
		repo:      repo,
		projector: NewProjectionService(repo),
		approvals: approvals,
		gsv:       s,
		acf:       c,
		mws:       m,
	}
}

type projectedNotificationRequest struct {
	ProjectID       *int64         `json:"project_id"`
	RecipientActor  string         `json:"recipient_actor"`
	Kind            string         `json:"kind"`
	Source          string         `json:"source"`
	SourceRef       string         `json:"source_ref"`
	NotificationKey string         `json:"notification_key"`
	Title           string         `json:"title"`
	Body            string         `json:"body"`
	Severity        int            `json:"severity"`
	RouteHint       string         `json:"route_hint"`
	CreatedBy       string         `json:"created_by"`
	Payload         map[string]any `json:"payload"`
}

type projectNotificationsRequest struct {
	Items []projectedNotificationRequest `json:"items"`
}

func (h *Handler) Routes() {
	r := h.gsv.GetRouter()
	base := h.acf.APIBaseURL() + "/notifications"
	group := r.Group(base, h.mws.GetValidation()...)
	{
		group.GET("", h.List)
		group.GET("/summary", h.Summary)
		group.GET("/:notification_id", h.Get)
		group.POST("/:notification_id/read", h.MarkRead)
		group.POST("/:notification_id/dismiss", h.Dismiss)
		group.POST("/approvals/sync", h.SyncApprovals)
		group.POST("/projected", h.Project)
	}
}

func (h *Handler) List(c *gin.Context) {
	orgID, actor, projectID, err := parseScope(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err := h.syncApprovals(c.Request.Context(), orgID, actor); err != nil {
		sharedhandlers.RespondError(c, domainerr.Validation(err.Error()))
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	items, err := h.repo.List(c.Request.Context(), ListFilters{
		OrgID:     orgID,
		Actor:     actor,
		ProjectID: projectID,
		Status:    c.Query("status"),
		Kind:      c.Query("kind"),
		Limit:     limit,
		Offset:    offset,
	})
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, gin.H{"items": items})
}

func (h *Handler) Get(c *gin.Context) {
	orgID, actor, _, err := parseScope(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err := h.syncApprovals(c.Request.Context(), orgID, actor); err != nil {
		sharedhandlers.RespondError(c, domainerr.Validation(err.Error()))
		return
	}
	id, err := strconv.ParseInt(c.Param("notification_id"), 10, 64)
	if err != nil || id <= 0 {
		sharedhandlers.RespondError(c, domainerr.Validation("invalid notification_id"))
		return
	}
	item, err := h.repo.Get(c.Request.Context(), orgID, actor, id)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, item)
}

func (h *Handler) Summary(c *gin.Context) {
	orgID, actor, projectID, err := parseScope(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err := h.syncApprovals(c.Request.Context(), orgID, actor); err != nil {
		sharedhandlers.RespondError(c, domainerr.Validation(err.Error()))
		return
	}
	summary, err := h.repo.GetSummary(c.Request.Context(), orgID, actor, projectID)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, summary)
}

func (h *Handler) MarkRead(c *gin.Context) {
	h.updateStatus(c, "read")
}

func (h *Handler) Dismiss(c *gin.Context) {
	h.updateStatus(c, "dismissed")
}

func (h *Handler) SyncApprovals(c *gin.Context) {
	orgID, actor, _, err := parseScope(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	projected, err := h.syncApprovalsCount(c.Request.Context(), orgID, actor)
	if err != nil {
		sharedhandlers.RespondError(c, domainerr.Validation(err.Error()))
		return
	}
	sharedhandlers.RespondOK(c, gin.H{"projected": projected})
}

func (h *Handler) Project(c *gin.Context) {
	orgID, actor, _, err := parseScope(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	var req projectNotificationsRequest
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}
	if len(req.Items) == 0 {
		sharedhandlers.RespondError(c, domainerr.Validation("items is required"))
		return
	}
	items := make([]ProjectedNotificationInput, 0, len(req.Items))
	for _, item := range req.Items {
		projectID := int64(0)
		if item.ProjectID != nil {
			projectID = *item.ProjectID
		}
		items = append(items, ProjectedNotificationInput{
			OrgID:           orgID,
			ProjectID:       projectID,
			Actor:           coalesceString(item.RecipientActor, actor),
			Kind:            item.Kind,
			Source:          item.Source,
			SourceRef:       item.SourceRef,
			NotificationKey: item.NotificationKey,
			Title:           item.Title,
			Body:            item.Body,
			Severity:        item.Severity,
			RouteHint:       item.RouteHint,
			CreatedBy:       coalesceString(item.CreatedBy, actor),
			Payload:         item.Payload,
		})
	}
	if err := h.projector.Project(c.Request.Context(), items); err != nil {
		sharedhandlers.RespondError(c, domainerr.Validation(err.Error()))
		return
	}
	sharedhandlers.RespondOK(c, gin.H{"projected": len(items)})
}

func (h *Handler) updateStatus(c *gin.Context, status string) {
	orgID, actor, _, err := parseScope(c)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	id, err := strconv.ParseInt(c.Param("notification_id"), 10, 64)
	if err != nil || id <= 0 {
		sharedhandlers.RespondError(c, domainerr.Validation("invalid notification_id"))
		return
	}
	if err := h.repo.MarkStatus(c.Request.Context(), orgID, actor, id, status); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func parseScope(c *gin.Context) (uuid.UUID, string, *int64, error) {
	orgID, err := sharedhandlers.ParseOrgID(c)
	if err != nil {
		return uuid.Nil, "", nil, err
	}
	actor, err := sharedhandlers.ParseActor(c)
	if err != nil {
		return uuid.Nil, "", nil, err
	}
	projectID, err := sharedhandlers.ParseOptionalInt64Query(c, "project_id")
	if err != nil {
		return uuid.Nil, "", nil, err
	}
	return orgID, actor, projectID, nil
}

func coalesceString(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func (h *Handler) syncApprovals(ctx context.Context, orgID uuid.UUID, actor string) error {
	_, err := h.syncApprovalsCount(ctx, orgID, actor)
	return err
}

func (h *Handler) syncApprovalsCount(ctx context.Context, orgID uuid.UUID, actor string) (int, error) {
	if h.approvals == nil || !h.approvals.Enabled() {
		return 0, nil
	}
	return h.approvals.SyncForActor(ctx, orgID, actor)
}
