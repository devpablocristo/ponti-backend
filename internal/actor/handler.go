package actor

import (
	"context"

	ginmw "github.com/devpablocristo/platform/http/gin/go"
	"github.com/gin-gonic/gin"

	dto "github.com/devpablocristo/ponti-backend/internal/actor/handler/dto"
	domain "github.com/devpablocristo/ponti-backend/internal/actor/usecases/domain"
	sharedhandlers "github.com/devpablocristo/ponti-backend/internal/shared/handlers"
)

type UseCasesPort interface {
	CreateActor(context.Context, *domain.Actor) (int64, error)
	ListActors(context.Context, domain.ListFilters, int, int) ([]domain.Actor, int64, error)
	GetActor(context.Context, int64) (*domain.Actor, error)
	UpdateActor(context.Context, *domain.Actor) error
	ArchiveActor(context.Context, int64) error
	RestoreActor(context.Context, int64) error
	HardDeleteActor(context.Context, int64) error
	AddRole(context.Context, int64, string) error
	AddAlias(context.Context, int64, domain.ActorAlias) (int64, error)
	ListDuplicateCandidates(context.Context) ([]domain.DuplicateCandidate, error)
	MergeActors(context.Context, domain.MergeRequest) (*domain.MergeImpact, error)
}

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
}

type Handler struct {
	ucs UseCasesPort
	gsv GinEnginePort
	acf ConfigAPIPort
	mws MiddlewaresEnginePort
}

type actorIDAction func(context.Context, int64) error

func NewHandler(u UseCasesPort, s GinEnginePort, c ConfigAPIPort, m MiddlewaresEnginePort) *Handler {
	return &Handler{
		ucs: u,
		gsv: s,
		acf: c,
		mws: m,
	}
}

func (h *Handler) Routes() {
	r := h.gsv.GetRouter()
	baseURL := h.acf.APIBaseURL() + "/actors"

	public := r.Group(baseURL, h.mws.GetValidation()...)
	{
		public.POST("", h.CreateActor)
		public.GET("", h.ListActors)
		public.GET("/archived", h.ListArchivedActors)
		public.GET("/duplicate-candidates", h.ListDuplicateCandidates)
		public.POST("/merge", h.MergeActors)
		public.GET("/:actor_id", h.GetActor)
		public.PUT("/:actor_id", h.UpdateActor)
		public.POST("/:actor_id/archive", h.ArchiveActor)
		public.POST("/:actor_id/restore", h.RestoreActor)
		public.DELETE("/:actor_id/hard", h.HardDeleteActor)
		public.POST("/:actor_id/roles", h.AddRole)
		public.POST("/:actor_id/aliases", h.AddAlias)
	}
}

func (h *Handler) CreateActor(c *gin.Context) {
	var req dto.ActorRequest
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}
	id, err := h.ucs.CreateActor(c.Request.Context(), req.ToDomain(0))
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondCreated(c, id)
}

func (h *Handler) ListActors(c *gin.Context) {
	page, perPage := sharedhandlers.ParsePaginationParams(c, 1, 100)
	filters := domain.ListFilters{
		Status: c.DefaultQuery("status", "active"),
		Role:   c.Query("role"),
		Query:  c.Query("q"),
	}
	actors, total, err := h.ucs.ListActors(c.Request.Context(), filters, page, perPage)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, dto.NewListActorsResponse(actors, page, perPage, total))
}

func (h *Handler) ListArchivedActors(c *gin.Context) {
	page, perPage := sharedhandlers.ParsePaginationParams(c, 1, 100)
	filters := domain.ListFilters{
		Status: "archived",
		Role:   c.Query("role"),
		Query:  c.Query("q"),
	}
	actors, total, err := h.ucs.ListActors(c.Request.Context(), filters, page, perPage)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, dto.NewListActorsResponse(actors, page, perPage, total))
}

func (h *Handler) ListDuplicateCandidates(c *gin.Context) {
	candidates, err := h.ucs.ListDuplicateCandidates(c.Request.Context())
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, dto.DuplicateCandidatesFromDomain(candidates))
}

func (h *Handler) GetActor(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "actor_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	actor, err := h.ucs.GetActor(c.Request.Context(), id)
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, dto.FromDomain(actor))
}

func (h *Handler) UpdateActor(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "actor_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	var req dto.ActorRequest
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}
	if err := h.ucs.UpdateActor(c.Request.Context(), req.ToDomain(id)); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}

func (h *Handler) ArchiveActor(c *gin.Context) {
	h.runActorIDAction(c, h.ucs.ArchiveActor)
}

func (h *Handler) RestoreActor(c *gin.Context) {
	h.runActorIDAction(c, h.ucs.RestoreActor)
}

func (h *Handler) HardDeleteActor(c *gin.Context) {
	h.runActorIDAction(c, h.ucs.HardDeleteActor)
}

func (h *Handler) runActorIDAction(c *gin.Context, action actorIDAction) {
	id, err := ginmw.ParseParamID(c, "actor_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	if err := action(c.Request.Context(), id); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}

func (h *Handler) AddRole(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "actor_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	var req dto.RoleRequest
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}
	if err := h.ucs.AddRole(c.Request.Context(), id, req.Role); err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondNoContent(c)
}

func (h *Handler) AddAlias(c *gin.Context) {
	id, err := ginmw.ParseParamID(c, "actor_id")
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	var req dto.AliasRequest
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}
	aliasID, err := h.ucs.AddAlias(c.Request.Context(), id, domain.ActorAlias{
		Alias:  req.Alias,
		Source: req.Source,
	})
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondCreated(c, aliasID)
}

func (h *Handler) MergeActors(c *gin.Context) {
	var req dto.MergeRequest
	if err := sharedhandlers.BindJSON(c, &req); err != nil {
		return
	}
	actorName, _ := sharedhandlers.ParseActor(c)
	impact, err := h.ucs.MergeActors(c.Request.Context(), domain.MergeRequest{
		TargetActorID:  req.TargetActorID,
		SourceActorIDs: req.SourceActorIDs,
		Reason:         req.Reason,
		Confirm:        req.Confirm,
		MergedBy:       actorName,
	})
	if err != nil {
		sharedhandlers.RespondError(c, err)
		return
	}
	sharedhandlers.RespondOK(c, dto.MergeImpactFromDomain(impact))
}
